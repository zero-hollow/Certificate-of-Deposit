/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

// Package ecdsaalg the ecdsa algorithm for generation of certificate and key
package ecdsaalg

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/common/cryptomgr/bccryptoutil"
	"git.huawei.com/huaweichain/common/logger"
)

var log = logger.GetModuleLogger("crypto", "ecdsaalg")

const daysInOneYear = 365
const hoursInOneDay = 24

type ecdsacert struct {
	pemCert []byte
	cert    *x509.Certificate
	pubKey  *ecdsa.PublicKey
}

func generateCertBasic(info *cryptomgr.CertBasicInfo) *x509.Certificate {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil
	}

	currentTime := time.Now()
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    currentTime,
		NotAfter: currentTime.Add(time.Hour * daysInOneYear * hoursInOneDay *
			time.Duration(info.ValidationYears)),
		BasicConstraintsValid: true,
	}

	// todo judge the country length is 0
	cert.Subject.Country = []string{info.Country}

	cert.Subject.Province = []string{info.Province}

	cert.Subject.Locality = []string{info.Locality}

	cert.Subject.Organization = []string{info.Organization}

	cert.Subject.OrganizationalUnit = []string{info.OrganizationUnit}

	cert.Subject.CommonName = info.CommonName

	cert.DNSNames = []string{info.CommonName}

	return cert
}

func generateEcdsaCert(template *x509.Certificate, pubKey *ecdsa.PublicKey, caCert *x509.Certificate,
	caPriKey *ecdsa.PrivateKey) (*ecdsacert, error) {
	cert, err := x509.CreateCertificate(rand.Reader, template, caCert, pubKey, caPriKey)
	if err != nil {
		log.Errorf("Create X509 self sign cert failed:%s", err.Error())
		return nil, errors.WithMessage(err, "create X509 self sign cert failed")
	}
	certificate, err := x509.ParseCertificate(cert)
	if err != nil {
		log.Errorf("Get x509 certificate failed:%s", err)
		return nil, errors.WithMessage(err, "get x509 certificate failed")
	}

	memory := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})

	ecdsaPubKey, ok := certificate.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("the public key is not ecdsa key")
	}

	ecdsaCert := &ecdsacert{
		pemCert: memory,
		cert:    certificate,
		pubKey:  ecdsaPubKey,
	}

	return ecdsaCert, nil
}

// GenerateMiddleCACert generate middle ca cert
func GenerateMiddleCACert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	if info == nil || pubKey == nil || caPriKey == nil || caCert == nil {
		return nil, errors.New("input parameter is not correct")
	}
	certBasic := generateCertBasic(info)
	if certBasic == nil {
		return nil, errors.New("create cert basic failed")
	}

	certBasic.IsCA = true
	keyUsage := x509.KeyUsageDigitalSignature |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	certBasic.KeyUsage = keyUsage
	certBasic.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}

	ecdsaCaPriKey, ok := caPriKey.(*ecdsaKey)
	if !ok {
		log.Errorf("PriKey is not a ecdsa private key")
		return nil, errors.New("priKey is not a ecdsa private key")
	}

	ecdsaCaCert, ok := caCert.(*ecdsacert)
	if !ok {
		return nil, errors.New("the cert is not ecdsa cert")
	}

	ecdsaPubKey, ok := pubKey.(*ecdsaKey)
	if !ok {
		return nil, errors.New("the pub key is not ecdsa key")
	}
	certBasic.SubjectKeyId = Hash(elliptic.Marshal(ecdsaPubKey.curve, ecdsaPubKey.X, ecdsaPubKey.Y))

	return generateEcdsaCert(certBasic, ecdsaPubKey.pubKey, ecdsaCaCert.cert, ecdsaCaPriKey.priKey)
}

// GenerateCert create self signed ca cert.
func GenerateCert(info *cryptomgr.CertBasicInfo, pubKey cryptomgr.Key, caPriKey cryptomgr.Key,
	caCert cryptomgr.Cert) (cryptomgr.Cert, error) {
	if info == nil || pubKey == nil || caPriKey == nil || caCert == nil {
		return nil, fmt.Errorf("input parameter is not correct")
	}

	certBasic := generateCertBasic(info)
	if certBasic == nil {
		return nil, errors.New("generate cert basic template failed")
	}
	certBasic.IsCA = false
	certBasic.KeyUsage = x509.KeyUsageDigitalSignature |
		x509.KeyUsageKeyEncipherment

	ecdsaCaCert, ok := caCert.(*ecdsacert)
	if !ok {
		return nil, errors.New("this is not a ecdsa certification")
	}
	caPrivateKey, ok := caPriKey.(*ecdsaKey)
	if !ok {
		return nil, errors.New("this is not a ecdsa private key")
	}
	ecdsaPubKey, ok := pubKey.(*ecdsaKey)
	if !ok {
		return nil, errors.New("this is not a ecdsa pub key")
	}

	return generateEcdsaCert(certBasic, ecdsaPubKey.pubKey, ecdsaCaCert.cert, caPrivateKey.priKey)
}

// GenerateSelfSignCert create self signed ca cert.
func GenerateSelfSignCert(info *cryptomgr.CertBasicInfo,
	key cryptomgr.Key) (cryptomgr.Cert, error) {
	if info == nil || key == nil {
		return nil, errors.New("input parameter is not correct")
	}
	certBasic := generateCertBasic(info)
	if certBasic == nil {
		return nil, fmt.Errorf("generate cert basic template failed")
	}
	certBasic.IsCA = true
	keyUsage := x509.KeyUsageDigitalSignature |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
	certBasic.KeyUsage = keyUsage
	certBasic.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}

	privateKey, ok := key.(*ecdsaKey)
	if !ok {
		log.Errorf("PriKey is not a ecdsa private key")
		return nil, errors.New("private key is not ecdsa key")
	}
	certBasic.SubjectKeyId = Hash(elliptic.Marshal(privateKey.curve, privateKey.X, privateKey.Y))

	return generateEcdsaCert(certBasic, privateKey.pubKey, certBasic, privateKey.priKey)
}

// GetCert get certificate from pem bytes.
func GetCert(pemCert []byte) (cryptomgr.Cert, error) {
	if pemCert == nil {
		log.Infof("the pem cert is nil")
		return nil, errors.New("the pem cert is nil")
	}

	block, _ := pem.Decode(pemCert)

	if block == nil {
		return nil, errors.New("decode the pem cert failed")
	}
	x509cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.New("parse certificate failed")
	}
	pubKey := x509cert.PublicKey
	ecdsaPubKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("the key is not ecdsa public key")
	}

	eccert := &ecdsacert{
		pemCert: pemCert,
		cert:    x509cert,
		pubKey:  ecdsaPubKey,
	}

	return eccert, nil
}

// GetFingerPrint get the ginger print of this cert.
func (e *ecdsacert) GetFingerPrint() string {
	hashInfo := Hash(e.cert.Raw)
	fingerStr := fmt.Sprintf("%x", hashInfo)
	return fingerStr
}

// GetSerialNumber get serial number of cert
func (e *ecdsacert) GetSerialNumber() string {
	return e.cert.SerialNumber.String()
}

// GetCommonName get common name of this cert.
func (e *ecdsacert) GetCommonName() string {
	return e.cert.Subject.CommonName
}

// GetDerPublicKey get der public key from cert
func (e *ecdsacert) GetDerPublicKey() ([]byte, error) {
	if e.cert == nil {
		return nil, errors.New("the x509 cert is nil")
	}
	pubKey, ok := e.cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("the public key is not ecdsa public key")
	}
	pubKeyDer, err := x509.MarshalPKIXPublicKey(pubKey)

	if err != nil {
		return nil, errors.New("get der format public key failed")
	}
	return pubKeyDer, nil
}

// GetExpireTime get expire time for this cert.
func (e *ecdsacert) GetExpireTime() time.Time {
	return e.cert.NotAfter
}

// GetOrganizationalUnit get organization unit of this cert.
func (e *ecdsacert) GetOrganizationalUnit() []string {
	return e.cert.Subject.OrganizationalUnit
}

// GetOrganization get organization.
func (e *ecdsacert) GetOrganization() []string {
	return e.cert.Subject.Organization
}

// VerifyBatch verify batch signatures
func (e *ecdsacert) VerifyBatch(msg [][]byte, signature [][]byte, hashOpt string) error {
	msgLen := len(msg)
	var hashSlice = make([][]byte, 0, msgLen)
	for _, v := range msg {
		hashInfo := Hash(v)
		hashSlice = append(hashSlice, hashInfo)
	}

	randArr, ok := bccryptoutil.RandMap.Load(msgLen)
	if !ok {
		return errors.New("get rand array error")
	}
	arrays, ok := randArr.(*bccryptoutil.RandomArrays)
	if !ok {
		return errors.New("this is not rand arrays")
	}
	verifyRes := verifyBatchWithRandArray(e.pubKey, hashSlice, signature, arrays)
	if verifyRes {
		return nil
	}
	return errors.New("verify batch failed")
}

// Verify verify the msg signature.
func (e *ecdsacert) Verify(msg []byte, signature []byte, hashOpt string) error {
	pubKey, ok := e.cert.PublicKey.(*ecdsa.PublicKey)
	if ok {
		return verifyWithPubKey(pubKey, msg, signature)
	}

	return errors.New("verify failed the public key is not ecdsa publickey")
}

// CheckValidation check whether this cert is signed by one of the root certs or not.
func (e ecdsacert) CheckValidation(rootcerts []cryptomgr.Cert) (string, error) {
	var opts x509.VerifyOptions
	certPool := x509.NewCertPool()

	// Get bytes
	for _, v := range rootcerts {
		ecdsaCert, ok := v.(*ecdsacert)
		if !ok {
			continue
		}
		certPool.AddCert(ecdsaCert.cert)
	}
	opts.Roots = certPool

	chains, err := e.cert.Verify(opts)
	if err != nil {
		return "", err
	}
	return chains[0][len(chains[0])-1].Subject.Organization[0], nil
}

// GetPemCertBytes get the pem cert bytes of this cert.
func (e ecdsacert) GetPemCertBytes() []byte {
	return e.pemCert
}
