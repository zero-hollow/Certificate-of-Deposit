/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package crypto provide the implementation of crypto.
package crypto

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/common/cryptomgr/cryptoimpl"

	"git.huawei.com/huaweichain/sdk/utils"
)

// Info is the definition of Info
type Info struct {
	alg    string
	cert   cryptomgr.Cert
	key    cryptomgr.Key
	hashFn func([]byte) []byte
}

// Sign is used to sign for a message.
func (c *Info) Sign(message []byte) ([]byte, error) {
	return c.key.Sign(message, c.alg)
}

// GetCertificate is used to get certificate.
func (c *Info) GetCertificate() ([]byte, error) {
	return c.cert.GetPemCertBytes(), nil
}

// GetCommonName is used to get common name from certificate.
func (c *Info) GetCommonName() string {
	return c.cert.GetCommonName()
}

// GetOrg is used to get organization from certificate.
func (c *Info) GetOrg() string {
	if len(c.cert.GetOrganization()) == 0 {
		return ""
	}
	return c.cert.GetOrganization()[0]
}

// Hash is the function to compute hash. It could be sha256 or sm3 depends on the identity
// algorithm config.
func (c *Info) Hash(data []byte) []byte {
	return c.hashFn(data)
}

// Crypto is the definition of crypto interface.
type Crypto interface {
	Sign(message []byte) ([]byte, error)
	GetCertificate() ([]byte, error)
	GetCommonName() string
	GetOrg() string
	Hash(data []byte) []byte
}

// NewCrypto is used to new an instance of crypto.
func NewCrypto(alg string, certPath string, keyPath string,
	decrypt func(bytes []byte) ([]byte, error)) (Crypto, error) {
	keyPem, err := getPemInfoFromPath(keyPath, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "getPemInfoFromPath error")
	}
	certPem, err := getPemInfoFromPath(certPath, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "getPemInfoFromPath error")
	}
	return NewCryptoWithIdentity(alg, certPem, keyPem)
}

// NewCryptoWithIdentity is used to create an instance by specify certificate and key byte array.
func NewCryptoWithIdentity(alg string, certPem []byte, keyPem []byte) (Crypto, error) {
	var certFactory cryptoimpl.CertFactory
	var keyFactory cryptoimpl.KeyFactory
	var hashFn func([]byte) []byte

	switch alg {
	case cryptomgr.EcdsaWithSha256:
		certFactory = &cryptoimpl.EcdsaCertFactory{}
		keyFactory = &cryptoimpl.EcdsaP256KeyFactory{}
		hashFn = utils.HashSha256
	case cryptomgr.Sm2WithSm3:
		certFactory = &cryptoimpl.GmCertFactory{}
		keyFactory = &cryptoimpl.GmKeyFactory{}
		hashFn = utils.HashSM3
	default:
		return nil, errors.New("not support crypto algorithm")
	}
	cert, err := certFactory.GetCertFromPem(certPem)
	if err != nil {
		return nil, errors.WithMessage(err, "GetCertFromPem error")
	}
	key, err := keyFactory.GetKeyFromPem(keyPem)
	if err != nil {
		return nil, errors.WithMessage(err, "GetKeyFromPem error")
	}
	return &Info{
		alg:    alg,
		cert:   cert,
		key:    key,
		hashFn: hashFn,
	}, nil
}

func getPemInfoFromPath(path string, decrypt func(bytes []byte) ([]byte, error)) ([]byte, error) {
	info, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.WithMessage(err, "read file error")
	}
	info, err = decrypt(info)
	if err != nil {
		return nil, errors.WithMessage(err, "decrypt message error")
	}
	return info, nil
}
