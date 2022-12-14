/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package cryptoimpl is the crypto implementation
package cryptoimpl

import (
	"git.huawei.com/huaweichain/common/cryptomgr"
)

// CryptoFactory the crypto factory for create cert or key
type CryptoFactory struct {
	ecdsaCertFactory CertFactory
	ecdsaKeyFactory  KeyFactory
	gmCertFactory    CertFactory
	gmKeyFactory     KeyFactory
}

// NewFactory new crypto factory
func NewFactory() *CryptoFactory {
	return &CryptoFactory{
		ecdsaCertFactory: &EcdsaCertFactory{},
		ecdsaKeyFactory:  &EcdsaP256KeyFactory{},
		gmCertFactory:    &GmCertFactory{},
		gmKeyFactory:     &GmKeyFactory{},
	}
}

// GetEcdsaCertFactory get ecdsa cert factory
func (f *CryptoFactory) GetEcdsaCertFactory() CertFactory {
	return f.ecdsaCertFactory
}

// GetEcdsaKeyFactory get ecdsa key factory
func (f *CryptoFactory) GetEcdsaKeyFactory() KeyFactory {
	return f.ecdsaKeyFactory
}

// GetGmCertFactory get gm cert factory
func (f *CryptoFactory) GetGmCertFactory() CertFactory {
	return f.gmCertFactory
}

// GetGmKeyFactory get gm key factory
func (f *CryptoFactory) GetGmKeyFactory() KeyFactory {
	return f.gmKeyFactory
}

// KeyFactory the key factory.
type KeyFactory interface {
	GenerateKey() (cryptomgr.Key, error)
	GetKeyFromPem(pemKey []byte) (cryptomgr.Key, error)
	CreateKeyFromBuffer(buffer []byte) (cryptomgr.Key, error)
}

// CertFactory the certificate factory.
type CertFactory interface {
	GenerateSelfSignCert(info *cryptomgr.CertBasicInfo,
		key cryptomgr.Key) (cryptomgr.Cert, error)
	GenerateMiddleCACert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
		caCert cryptomgr.Cert) (cryptomgr.Cert, error)
	GenerateCert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
		caCert cryptomgr.Cert) (cryptomgr.Cert, error)
	GetCertFromPem(pemCert []byte) (cryptomgr.Cert, error)
}

// CsrFactory the csr factory.
type CsrFactory interface {
}
