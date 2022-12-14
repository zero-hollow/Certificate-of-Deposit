/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package gmalg the gm algorithm for generate of certificate and key
package gmalg

import (
	"time"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

// Gmcert gm certificate
type Gmcert struct {
}

// GenerateCert create certificate
func GenerateCert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return nil, nil
}

// GenerateSelfSignCert create self signed certificate
func GenerateSelfSignCert(info *cryptomgr.CertBasicInfo,
	key cryptomgr.Key) (cryptomgr.Cert, error) {
	return nil, nil
}

// GenerateMiddleCACert generate middle CA cert
func GenerateMiddleCACert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return nil, nil
}

// GetCert create certificate from pem
func GetCert(pem []byte) (cryptomgr.Cert, error) {
	return nil, nil
}

// GetCommonName get common name
func (g *Gmcert) GetCommonName() string {
	return ""
}

// GetExpireTime get expire time
func (g *Gmcert) GetExpireTime() time.Time {
	return time.Time{}
}

// GetOrganizationalUnit get organization unit
func (g *Gmcert) GetOrganizationalUnit() []string {
	return nil
}

// GetOrganization get organization
func (g *Gmcert) GetOrganization() []string {
	return nil
}

// Verify verify the message signature
func (g *Gmcert) Verify(msg []byte, signature []byte, hashOpt string) error {
	return nil
}

// CheckValidation check validation
func (g *Gmcert) CheckValidation(rootcerts []cryptomgr.Cert) (string, error) {
	return "", nil
}

// GetPemCertBytes get pem bytes
func (g *Gmcert) GetPemCertBytes() []byte {
	return nil
}

// GetFingerPrint get finger print
func (g *Gmcert) GetFingerPrint() string {
	return ""
}
