/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package cryptoimpl is the crypto implementation
package cryptoimpl

import (
	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/common/cryptomgr/gmalg"
)

// GmKeyFactory gm key factory
type GmKeyFactory struct {
}

// GenerateKey create key.
func (g *GmKeyFactory) GenerateKey() (cryptomgr.Key, error) {
	return gmalg.GeneratePriKey()
}

// GetKeyFromPem get key from pem.
func (g *GmKeyFactory) GetKeyFromPem(pemKey []byte) (cryptomgr.Key, error) {
	return gmalg.GetPriKey(pemKey)
}

// CreateKeyFromBuffer create key from buffer
func (g *GmKeyFactory) CreateKeyFromBuffer(buffer []byte) (cryptomgr.Key, error) {
	return nil, nil
}

// GmCertFactory gm cert factory
type GmCertFactory struct {
}

// GenerateSelfSignCert create self sign CA cert
func (g *GmCertFactory) GenerateSelfSignCert(info *cryptomgr.CertBasicInfo,
	key cryptomgr.Key) (cryptomgr.Cert, error) {
	return gmalg.GenerateSelfSignCert(info, key)
}

// GenerateMiddleCACert generate middle ca cert
func (g *GmCertFactory) GenerateMiddleCACert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key,
	caPriKey cryptomgr.Key, caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return gmalg.GenerateMiddleCACert(info, pubKey, caPriKey, caCert)
}

// GenerateCert create cert by gm cert factory.
func (g *GmCertFactory) GenerateCert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return gmalg.GenerateCert(info, pubKey, caPriKey, caCert)
}

// GetCertFromPem Get gm cert from pem bytes.
func (g *GmCertFactory) GetCertFromPem(pemCert []byte) (cryptomgr.Cert, error) {
	return gmalg.GetCert(pemCert)
}
