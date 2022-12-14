/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package node is the definition of node.
package node

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"git.huawei.com/huaweichain/sdk/config"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"

	"git.huawei.com/huaweichain/sdk/action"
)

// WNode is the definition of WNode.
type WNode struct {
	*config.Node
	tls            *TLS
	ChainAction    *action.ChainAction
	ContractAction *action.ContractAction
	EventAction    *action.EventAction
	QueryAction    *action.QueryAction
	tlsEnable      bool
}

// NewNode is used to create a wnode proxy.
func NewNode(node *config.Node, tlsEnable bool, tls *TLS) (*WNode, error) {
	config, err := getConfig(node, tlsEnable, tls)
	if err != nil {
		return nil, errors.WithMessage(err, "get config error")
	}
	return &WNode{
		Node:           node,
		tlsEnable:      tlsEnable,
		tls:            tls,
		ChainAction:    action.NewChainAction(config),
		ContractAction: action.NewContractAction(config),
		EventAction:    action.NewEventAction(config),
		QueryAction:    action.NewQueryAction(config),
	}, nil
}

// GetNodeAddr is used to get node address.
func (n *WNode) GetNodeAddr() string {
	return fmt.Sprintf("%s:%d", n.Host, n.Port)
}

// TLS is the definition of TLS.
type TLS struct {
	certPEMBlock []byte
	keyPEMBlock  []byte
	roots        [][]byte
}

// NewTLS is use to new an instance of TLS.
func NewTLS(certPEMBlock []byte, keyPEMBlock []byte, roots [][]byte) *TLS {
	return &TLS{
		certPEMBlock: certPEMBlock,
		keyPEMBlock:  keyPEMBlock,
		roots:        roots,
	}
}

// ConvertTLS is used to convert config TLS to node TLS.
func ConvertTLS(tls config.TLS, decrypt func(bytes []byte) ([]byte, error)) (*TLS, error) {
	certPEMBlock, err := ioutil.ReadFile(filepath.Clean(tls.CertPath))
	if err != nil {
		return nil, errors.WithMessagef(err, "read file failed, file: %v", tls.CertPath)
	}
	certPEMBlock, err = decrypt(certPEMBlock)
	if err != nil {
		return nil, errors.WithMessagef(err, "decrypt cert failed, file: %v", tls.CertPath)
	}
	keyPEMBlock, err := ioutil.ReadFile(filepath.Clean(tls.KeyPath))
	if err != nil {
		return nil, errors.WithMessagef(err, "read file failed, file: %v", tls.KeyPath)
	}
	keyPEMBlock, err = decrypt(keyPEMBlock)
	if err != nil {
		return nil, errors.WithMessagef(err, "decrypt key failed, file: %v", tls.KeyPath)
	}
	roots := make([][]byte, len(tls.RootPath))
	for i, rootCertPath := range tls.RootPath {
		root, err := ioutil.ReadFile(filepath.Clean(rootCertPath))
		if err != nil {
			return nil, errors.WithMessagef(err, "read file failed, file: %v", rootCertPath)
		}
		root, err = decrypt(root)
		if err != nil {
			return nil, errors.WithMessagef(err, "decrypt root cert failed, file: %v", rootCertPath)
		}
		roots[i] = root
	}
	return NewTLS(certPEMBlock, keyPEMBlock, roots), nil
}

func getConfig(node *config.Node, tlsEnable bool, tls *TLS) (*action.Config, error) {
	if node.HostOverride == "" {
		node.HostOverride = node.Host
	}

	if tlsEnable {
		transportCreds, err := getTransportCredentials(node.HostOverride, tls)
		if err != nil {
			return nil, errors.WithMessage(err, "get ssl context error")
		}
		return action.NewConfig(node.HostOverride, node.Host, node.Port, tlsEnable, transportCreds), nil
	}
	return action.NewConfig(node.HostOverride, node.Host, node.Port, tlsEnable, nil), nil
}

func getTransportCredentials(serverName string, t *TLS) (credentials.TransportCredentials, error) {
	cert, err := tls.X509KeyPair(t.certPEMBlock, t.keyPEMBlock)
	if err != nil {
		return nil, errors.WithMessage(err, "x509 key pair error")
	}
	certPool := x509.NewCertPool()
	for _, root := range t.roots {
		if !certPool.AppendCertsFromPEM(root) {
			return nil, errors.New("fail to append root ca")
		}
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:               serverName,
		Certificates:             []tls.Certificate{cert},
		RootCAs:                  certPool,
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	})
	return transportCreds, nil
}
