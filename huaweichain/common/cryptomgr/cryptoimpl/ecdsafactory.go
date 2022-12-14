/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package cryptoimpl is the crypto implementation
package cryptoimpl

import (
	"crypto/elliptic"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/common/cryptomgr/ecdsaalg"
)

// EcdsaP256KeyFactory ecdsa p256 key factory
type EcdsaP256KeyFactory struct{}

// GenerateKey create key.
func (ek *EcdsaP256KeyFactory) GenerateKey() (cryptomgr.Key, error) {
	curve := elliptic.P256()
	return ecdsaalg.CreateKeyWithCurve(curve)
}

// GetKeyFromPem get key from pem bytes.
func (ek *EcdsaP256KeyFactory) GetKeyFromPem(pemKey []byte) (cryptomgr.Key, error) {
	return ecdsaalg.GetKeyFromPem(pemKey)
}

// CreateKeyFromBuffer create key from tee buffer
func (ek *EcdsaP256KeyFactory) CreateKeyFromBuffer(buffer []byte) (cryptomgr.Key, error) {
	return ecdsaalg.CreateKeyFromBuffer(buffer)
}

// EcdsaCertFactory ecdsa certificate factory
type EcdsaCertFactory struct {
}

// GenerateSelfSignCert create self sign CA certificate
func (ec *EcdsaCertFactory) GenerateSelfSignCert(info *cryptomgr.CertBasicInfo,
	key cryptomgr.Key) (cryptomgr.Cert, error) {
	return ecdsaalg.GenerateSelfSignCert(info, key)
}

// GenerateMiddleCACert generate middle ca cert
func (ec *EcdsaCertFactory) GenerateMiddleCACert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key,
	caPriKey cryptomgr.Key, caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return ecdsaalg.GenerateMiddleCACert(info, pubKey, caPriKey, caCert)
}

// GenerateCert create certificate
func (ec *EcdsaCertFactory) GenerateCert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return ecdsaalg.GenerateCert(info, pubKey, caPriKey, caCert)
}

// GetCertFromPem get certificate from pem.
func (ec *EcdsaCertFactory) GetCertFromPem(pemCert []byte) (cryptomgr.Cert, error) {
	return ecdsaalg.GetCert(pemCert)
}
