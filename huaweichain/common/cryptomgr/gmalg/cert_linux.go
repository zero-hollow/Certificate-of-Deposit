/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

// Package gmalg the gm algorithm for generate of certificate and key
package gmalg

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/common/logger"
	"git.huawei.com/huaweichain/gmssl"
)

type gmcert struct {
	pemCert           []byte
	gmsslCert         *gmssl.Certificate
	pubKey            *gmssl.PublicKey
	derPubKey         []byte
	subjectCN         string
	subjectOU         string
	subjectO          string
	subjectPostalCode string
	subjectStreet     string
	subjectLocality   string
	subjectCountry    string
}

const sm2SignFlag = "sm2sign"
const defaultSm2ID = "1234567812345678"

var log = logger.GetDefaultLogger()

// GenerateMiddleCACert generate middle CA cert
func GenerateMiddleCACert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return geneCert(info, pubKey, caPriKey, caCert, true)
}

// GenerateCert generate cert
func GenerateCert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	return geneCert(info, pubKey, caPriKey, caCert, false)
}

func geneCert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert, isMiddleCert bool) (cryptomgr.Cert, error) {
	if info == nil || pubKey == nil || caPriKey == nil || caCert == nil {
		return nil, fmt.Errorf("input parameter is not correct")
	}
	gmPubKey, ok := pubKey.(*gmKey)
	if !ok {
		return nil, fmt.Errorf("the public key is not gm key")
	}
	gmCaPriKey, ok := caPriKey.(*gmKey)
	if !ok {
		return nil, fmt.Errorf("the ca private key is not gm key")
	}
	gmCaCert, ok := caCert.(*gmcert)
	if !ok {
		return nil, fmt.Errorf("the ca cert is not gm cert")
	}

	var gmCert *gmssl.Certificate
	var err error
	if isMiddleCert {
		gmCert, err = gmssl.GenerateMiddleCACert(info.Country, info.Province, info.Locality, info.Organization,
			info.OrganizationUnit, info.CommonName, info.ValidationYears,
			gmPubKey.pubKey, gmCaPriKey.priKey, gmCaCert.gmsslCert)
		if err != nil {
			fmt.Printf("the cert is : %s\n", err.Error())
			return nil, errors.WithMessage(err, "generate middle cert failed")
		}
	} else {
		gmCert, err = gmssl.GenerateCert(info.Country, info.Province, info.Locality, info.Organization,
			info.OrganizationUnit, info.CommonName, info.ValidationYears,
			gmPubKey.pubKey, gmCaPriKey.priKey, gmCaCert.gmsslCert)
		if err != nil {
			fmt.Printf("the cert is : %s\n", err.Error())
			return nil, errors.WithMessage(err, "generate middle cert failed")
		}
	}
	pemCert := gmssl.GetPemFromCert(gmCert)
	return GetCert(pemCert)
}

// GenerateSelfSignCert create self signed certificate
func GenerateSelfSignCert(info *cryptomgr.CertBasicInfo,
	key cryptomgr.Key) (cryptomgr.Cert, error) {
	if info == nil || key == nil {
		return nil, fmt.Errorf("input parameter is not correct")
	}
	gk, ok := key.(*gmKey)
	if !ok {
		log.Errorf("the key is not gm key")
		return nil, fmt.Errorf("the key is not gm key")
	}
	certificate, err := gmssl.GenerateSelfSignCert(info.Country, info.Province, info.Locality, info.Organization,
		info.OrganizationUnit, info.CommonName, info.ValidationYears, gk.priKey)
	if err != nil {
		return nil, err
	}
	pemCert := gmssl.GetPemFromCert(certificate)
	return GetCert(pemCert)
}

// GetCert create certificate from pem
func GetCert(pem []byte) (cryptomgr.Cert, error) {
	if pem == nil {
		log.Infof("the pem cert is nil")
		return nil, errors.New("the pem cert is nil")
	}
	certificate, err := gmssl.NewCertificateFromPEM(string(pem), "")
	if err != nil {
		log.Errorf("change gm pem to certificate failed:%s", err.Error())

		return nil, errors.WithMessage(err, "change gm pem to certificate failed")
	}

	cert := gmcert{}
	cert.pemCert = pem
	cert.gmsslCert = certificate
	cert.setSubjects()
	publicKey, err := cert.gmsslCert.GetPublicKey()
	if err != nil {
		log.Errorf("Get public key failed from gm certificate.")
		return nil, fmt.Errorf("get public key failed from gm certificate")
	}
	cert.pubKey = publicKey
	derPubKey, err := publicKey.GetDer()
	if err != nil {
		return nil, errors.WithMessage(err, "get der public key failed")
	}
	cert.derPubKey = derPubKey
	return &cert, nil
}

func (g *gmcert) setSubjects() {
	subject, err := g.gmsslCert.GetSubject()
	if err != nil {
		return
	}
	subjects := strings.Split(subject, "/")
	for _, v := range subjects {
		index := strings.Index(v, "=")
		if index == -1 {
			continue
		}
		subjectLabel := v[:index]
		resStr := v[index+1:]
		switch subjectLabel {
		case "CN":
			g.subjectCN = resStr
		case "OU":
			g.subjectOU = resStr
		case "O":
			g.subjectO = resStr
		case "postalCode":
			g.subjectPostalCode = resStr
		case "street":
			g.subjectStreet = resStr
		case "L":
			g.subjectLocality = resStr
		case "ST":
			g.subjectStreet = resStr
		case "C":
			g.subjectCountry = resStr
		default:
			log.Infof("Unknown subject")
		}
	}
}

// GetCommonName get common name of cert
func (g *gmcert) GetCommonName() string {
	if g.gmsslCert == nil {
		return ""
	}
	return g.subjectCN
}

// GetExpireTime get expire time of cert
func (g *gmcert) GetExpireTime() time.Time {
	return g.gmsslCert.GetExpireTime()
}

// GetOrganizationalUnit get organization unit of cert
func (g *gmcert) GetOrganizationalUnit() []string {
	if g.gmsslCert == nil {
		return nil
	}
	return []string{g.subjectOU}
}

// GetOrganization get organizations of cert
func (g *gmcert) GetOrganization() []string {
	if g.gmsslCert == nil {
		return nil
	}
	return []string{g.subjectO}
}

// VerifyBatch verify batch
func (g *gmcert) VerifyBatch(msg [][]byte, signature [][]byte, hashOpt string) error {
	return errors.New("not support batch verify for gm")
}

// Verify verify signature
func (g *gmcert) Verify(msg []byte, signature []byte, hashOpt string) error {
	digest := calcSm2SignDigest(g.pubKey, msg)
	if digest == nil {
		return fmt.Errorf("calc digest failed in verify")
	}
	return g.pubKey.Verify(sm2SignFlag, digest, signature, nil)
}

// GetDerPublicKey get public key from cert
func (g *gmcert) GetDerPublicKey() ([]byte, error) {
	return g.derPubKey, nil
}

// CheckValidation check the validation of this cert by root certs
func (g *gmcert) CheckValidation(rootcerts []cryptomgr.Cert) (string, error) {
	if len(rootcerts) == 0 {
		return "", errors.New("there is no root certs")
	}
	for _, rootCert := range rootcerts {
		gmRootCert, ok := rootCert.(*gmcert)
		if !ok {
			continue
		}
		verifyRes := g.gmsslCert.VerifyCert(gmRootCert.gmsslCert)
		if verifyRes {
			return gmRootCert.subjectO, nil
		}
	}

	return "", errors.Errorf("verify certificate %s failed by root certs %s", g.GetCommonName(),
		rootcerts[0].GetCommonName())
}

// GetPemCertBytes get pem cert bytes
func (g *gmcert) GetPemCertBytes() []byte {
	return g.pemCert
}

// GetFingerPrint get finger print of cert
func (g *gmcert) GetFingerPrint() string {
	return g.gmsslCert.GetFingerPrint()
}

// GetSerialNumber get serial number of cert
func (g *gmcert) GetSerialNumber() string {
	serialNumber, err := g.gmsslCert.GetSerialNumber()
	if err != nil {
		return ""
	}
	return serialNumber
}
