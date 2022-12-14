/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package client provide the implementation of client module, which is the entrance of go sdk.
package client

import (
	"time"

	"git.huawei.com/huaweichain/proto"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"git.huawei.com/huaweichain/sdk/utils"
	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/sdk/config"
	"git.huawei.com/huaweichain/sdk/crypto"
	"git.huawei.com/huaweichain/sdk/node"
	"git.huawei.com/huaweichain/sdk/rawmessage"
)

// GatewayClient is the definition of gateway client.
type GatewayClient struct {
	tls                 *node.TLS
	ChainRawMessage     *rawmessage.ChainRawMessage
	ContractRawMessage  *rawmessage.ContractRawMessage
	LifecycleRawMessage *rawmessage.LifecycleRawMessage
	QueryRawMessage     *rawmessage.QueryRawMessage
	ConfigRawMessage    *rawmessage.ConfigRawMessage
	Nodes               map[string]*node.WNode
	Crypto              crypto.Crypto
}

// NewEmptyGatewayClient is used to create an instance of gateway client which is empty.
func NewEmptyGatewayClient() *GatewayClient {
	return &GatewayClient{Nodes: make(map[string]*node.WNode)}
}

// NewGatewayClient is used to create an instance of gateway client by specifying the config file path.
func NewGatewayClient(configPath string, decrypts ...func(bytes []byte) ([]byte, error)) (*GatewayClient, error) {
	cfg, err := config.NewClientConfig(configPath)
	if err != nil {
		return nil, errors.WithMessage(err, "load config file error")
	}

	return NewGatewayClientWithCfg(cfg, decrypts...)
}

// NewGatewayClientWithCfg is used to create an instance of gateway client by specifying the config data.
func NewGatewayClientWithCfg(cfg *config.ClientConfig, decrypts ...func(bytes []byte) ([]byte,
	error)) (*GatewayClient, error) {
	client := NewEmptyGatewayClient()
	decrypt := func(bytes []byte) ([]byte, error) {
		return bytes, nil
	}
	if len(decrypts) > 0 {
		decrypt = decrypts[0]
	}

	c, err := getCrypto(cfg.Client, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get crypto error")
	}
	err = client.setRawMsg(c)
	if err != nil {
		return nil, errors.WithMessage(err, " set raw message error")
	}
	client.setCrypto(c)

	client.Nodes, err = getNodes(cfg, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get nodes error")
	}
	return client, nil
}

// SetIdentity is used to set identity for sdk.
func (client *GatewayClient) SetIdentity(algorithm string, key []byte, cert []byte) error {
	c, err := crypto.NewCryptoWithIdentity(algorithm, cert, key)
	if err != nil {
		return errors.WithMessage(err, "new crypto with identity error")
	}
	return client.setRawMsg(c)
}

// SetTLS is used to set TLS for sdk.
func (client *GatewayClient) SetTLS(certPEMBlock []byte, keyPEMBlock []byte, roots [][]byte) {
	client.tls = node.NewTLS(certPEMBlock, keyPEMBlock, roots)
}

// AddNode is used to add node proxy for sdk.
func (client *GatewayClient) AddNode(name string, hostOverride string, host string, port int) error {
	n := &config.Node{
		ID:           name,
		HostOverride: hostOverride,
		Host:         hostOverride,
		Port:         port,
	}

	tlsEnable := false
	if client.tls != nil {
		tlsEnable = true
	}
	var err error
	client.Nodes[name], err = node.NewNode(n, tlsEnable, client.tls)
	if err != nil {
		return errors.WithMessage(err, "new node error")
	}
	return nil
}

// SetTimeout set client timeout (/s)
func (client *GatewayClient) SetTimeout(seconds int64) {
	utils.SetTimeout(time.Duration(seconds))
}

// GetSingleApplier provides func to get latest block number when build txHeader.
func (client *GatewayClient) GetSingleApplier(addr, chainID string) func() (uint64, error) {
	return func() (uint64, error) {
		var wnode *node.WNode
		for _, n := range client.Nodes {
			if n.GetNodeAddr() == addr {
				wnode = n
				break
			}
		}
		if wnode == nil {
			return 0, errors.Errorf("node addr %s is unknown", addr)
		}
		reqRaw, err := client.QueryRawMessage.BuildLatestChainStateRawMessage(chainID)
		if err != nil {
			return 0, err
		}
		resRaw, err := wnode.QueryAction.GetLatestChainState(reqRaw)
		if err != nil {
			return 0, err
		}
		response := &common.Response{}
		if err := proto.Unmarshal(resRaw.Payload, response); err != nil {
			return 0, errors.WithMessage(err, "unmarshal response error")
		}
		if response.Status != common.SUCCESS {
			return 0, errors.Errorf("response error: status: %v, info: %v",
				response.Status.String(), response.StatusInfo)
		}
		latestChainState := &nodeservice.LatestChainState{}
		if err := proto.Unmarshal(response.Payload, latestChainState); err != nil {
			return 0, errors.WithMessage(err, "unmarshal latest chain state error")
		}
		return latestChainState.Height + 1, nil
	}
}

func (client *GatewayClient) setRawMsg(crypto crypto.Crypto) error {
	msgBuilder, err := rawmessage.NewMsgBuilderImpl(crypto)
	if err != nil {
		return errors.WithMessage(err, "new MsgBuilderImpl error")
	}

	client.ChainRawMessage = rawmessage.NewChainRawMessage(msgBuilder, crypto)
	client.ContractRawMessage = rawmessage.NewContractRawMessage(msgBuilder)
	client.LifecycleRawMessage = rawmessage.NewLifecycleRawMessage(msgBuilder, crypto)
	client.QueryRawMessage = rawmessage.NewQueryRawMessage(msgBuilder)
	client.ConfigRawMessage = rawmessage.NewConfigRawMessage(msgBuilder)
	return nil
}

func (client *GatewayClient) setCrypto(crypto crypto.Crypto) {
	client.Crypto = crypto
}

func getNodes(c *config.ClientConfig, decrypt config.DecryptFunc) (map[string]*node.WNode, error) {
	nodes := make(map[string]*node.WNode)
	for _, n := range c.Nodes {
		var wnode *node.WNode
		var tls *node.TLS
		var err error
		if c.Client.TLS.Enable {
			tls, err = node.ConvertTLS(c.Client.TLS, decrypt)
			if err != nil {
				return nil, errors.WithMessage(err, "convert node tls from config tls error")
			}
		}
		wnode, err = node.NewNode(n, c.Client.TLS.Enable, tls)
		if err != nil {
			return nil, errors.WithMessage(err, "new node error")
		}
		nodes[n.ID] = wnode
	}
	return nodes, nil
}

func getCrypto(c *config.Client, decrypt config.DecryptFunc) (crypto.Crypto, error) {
	if c == nil {
		return nil, errors.New("the client config is nil")
	}
	return crypto.NewCrypto(c.Type, c.Identity.CertPath, c.Identity.KeyPath, decrypt)
}
