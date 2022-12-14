/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

package ecdsaalg

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr/bccryptoutil"

	"git.huawei.com/huaweichain/common/cryptomgr"

	"testing"
)

var pemStr = `-----BEGIN CERTIFICATE-----
MIICWzCCAgGgAwIBAgIRAISV0qOuBJLraUb7VhnMt0UwCgYIKoZIzj0EAwIwgZMx
DTALBgNVBAYTBHRlc3QxDTALBgNVBAgTBHRlc3QxDTALBgNVBAcTBHRlc3QxDTAL
BgNVBAkTBHRlc3QxDTALBgNVBBETBHRlc3QxGTAXBgNVBAoTEG9yZzEuZXhhbXBs
ZS5jb20xDTALBgNVBAsTBHRlc3QxHDAaBgNVBAMTE2NhLm9yZzEuZXhhbXBsZS5j
b20wHhcNMjAwMjI5MDYwNTAwWhcNMzAwMjI2MDYwNTAwWjB7MQ0wCwYDVQQGEwR0
ZXN0MQ0wCwYDVQQIEwR0ZXN0MQ0wCwYDVQQHEwR0ZXN0MQ0wCwYDVQQJEwR0ZXN0
MQ0wCwYDVQQREwR0ZXN0MQ0wCwYDVQQLEwR0ZXN0MR8wHQYDVQQDExZwZWVyMC5v
cmcxLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAECqIs4cyb
Em3oFmbDphRXYHu4MAz9WwAi735oUyEGCaaViCjGBiRWCfju95djMQUacgQTp2ss
sDHVgKxitFhe8aNNMEswDgYDVR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwKwYD
VR0jBCQwIoAgaTOBXgpb9p10WundQSHdDehm3Afuv6CzcpDSWeqGb9owCgYIKoZI
zj0EAwIDSAAwRQIhAPnWr9qhDbR8XYkxELinkho7ymB2J6BOiRZDQKpwPyDpAiBX
Q4hmCmX/wxXNwElctbQnWJXerNJtJlABaZ5eDOLzVw==
-----END CERTIFICATE-----`

var keyStr = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgeQ1WPoPDyUmMF+nX
Vhfch3DRVt8pM9oK7di9IZkiiOWhRANCAAQKoizhzJsSbegWZsOmFFdge7gwDP1b
ACLvfmhTIQYJppWIKMYGJFYJ+O73l2MxBRpyBBOnayywMdWArGK0WF7x
-----END PRIVATE KEY-----`

var caStr = `-----BEGIN CERTIFICATE-----
MIIClDCCAjqgAwIBAgIRAMRutSXS4HefEZyiuCZ5/owwCgYIKoZIzj0EAwIwgZMx
DTALBgNVBAYTBHRlc3QxDTALBgNVBAgTBHRlc3QxDTALBgNVBAcTBHRlc3QxDTAL
BgNVBAkTBHRlc3QxDTALBgNVBBETBHRlc3QxGTAXBgNVBAoTEG9yZzEuZXhhbXBs
ZS5jb20xDTALBgNVBAsTBHRlc3QxHDAaBgNVBAMTE2NhLm9yZzEuZXhhbXBsZS5j
b20wHhcNMjAwMjI5MDYwNTAwWhcNMzAwMjI2MDYwNTAwWjCBkzENMAsGA1UEBhME
dGVzdDENMAsGA1UECBMEdGVzdDENMAsGA1UEBxMEdGVzdDENMAsGA1UECRMEdGVz
dDENMAsGA1UEERMEdGVzdDEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTENMAsG
A1UECxMEdGVzdDEcMBoGA1UEAxMTY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqG
SM49AgEGCCqGSM49AwEHA0IABOx1Zo5ZWC8HouOqDFczRoj3aSyEyBsOoOZR2Wfc
ASB7JL0PfN1NnO9d67E2+bYk9xrIWfEYWFGI+wrqF8+kUiejbTBrMA4GA1UdDwEB
/wQEAwIBpjAdBgNVHSUEFjAUBggrBgEFBQcDAgYIKwYBBQUHAwEwDwYDVR0TAQH/
BAUwAwEB/zApBgNVHQ4EIgQgaTOBXgpb9p10WundQSHdDehm3Afuv6CzcpDSWeqG
b9owCgYIKoZIzj0EAwIDSAAwRQIhAMkizmIv1bpj3P0phvBwp7m/x95HAIcm6nFa
I8AEi3/cAiBDlS/pZSWkg5SRShU1xIPNOfEbH+wjJPKZbe3DJNEGOQ==
-----END CERTIFICATE-----`

// TestCreateSelfSignCACert test create self sign ca cert
func TestCreateSelfSignCACert(t *testing.T) {
	key, err := CreateKeyWithCurveP256()
	if err != nil {
		t.Fatalf("Create key with curve p256 failed:%s", err.Error())
		return
	}
	info := &cryptomgr.CertBasicInfo{
		Organization:     "hangyan",
		OrganizationUnit: "hua wei",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "example.com",
		ValidationYears:  20,
	}
	cert, err := GenerateSelfSignCert(info, key)
	if err != nil {
		t.Fatalf("Create cert failed")
		return
	}

	organizations := cert.GetOrganization()
	if organizations[0] != "hangyan" {
		t.Fatalf("Create ecdsa certificate failed, the organization is wrong:%s", organizations[0])
	}
	commonName := cert.GetCommonName()
	if commonName != "example.com" {
		t.Fatalf("Create ecdsa certificate failed, the common name is wrond:%s", commonName)
	}

	ou := cert.GetOrganizationalUnit()
	if ou[0] != "hua wei" {
		t.Fatalf("Create ecdsa certificate failed,  the ou is wrong")
	}

	sig, err := key.Sign([]byte("hello"), cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("key sign failed")
	}

	err = cert.Verify([]byte("hello"), sig, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("cert verify failed")
	}

	pemCertBytes := cert.GetPemCertBytes()
	if pemCertBytes == nil {
		t.Fatalf("the pem cert is nil ")
	}
	if len(pemCertBytes) < 100 {
		t.Fatalf("the pem cert bytes len is too small:%d", len(pemCertBytes))
	}
	fingerPrint := cert.GetFingerPrint()
	if fingerPrint == "" {
		t.Fatalf("the fingerPrint is empty")
	}
	if len(fingerPrint) != 64 {
		t.Fatalf("the finger print is wrong:%d", len(fingerPrint))
	}

	timeExpire := cert.GetExpireTime()
	year := time.Now().Year() + 20
	if timeExpire.Year() != year {
		t.Fatalf("the cert expire time is wrong:%d", timeExpire.Year())
	}

	orgName, err := cert.CheckValidation([]cryptomgr.Cert{cert})
	if err != nil {
		t.Fatalf("check validation of cert failed:%s", err.Error())
	}
	if orgName != "hangyan" {
		t.Fatalf("the org name is wrong:%s", orgName)
	}

}

// TestGenerateMiddleCACert test generate middle ca cert
func TestGenerateMiddleCACert(t *testing.T) {
	key, err := CreateKeyWithCurveP256()
	if err != nil {
		t.Fatalf("Create key with curve p256 failed:%s", err.Error())
		return
	}
	info := &cryptomgr.CertBasicInfo{
		Organization:     "hangyan",
		OrganizationUnit: "hua wei",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "example.com",
		ValidationYears:  20,
	}
	cert, err := GenerateSelfSignCert(info, key)
	if err != nil {
		t.Fatalf("Create cert failed")
		return
	}

	caBytes := cert.GetPemCertBytes()
	err = ioutil.WriteFile("/superchain/temp/1117/rootca.crt", caBytes, 0600)
	if err != nil {
		t.Fatalf("Write root ca file failed")
	}

	middleKey, err := CreateKeyWithCurveP256()
	if err != nil {
		t.Fatalf("Create middle key failed")
	}
	middleCertInfo := &cryptomgr.CertBasicInfo{
		Organization:     "hangyan",
		OrganizationUnit: "hua wei",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "middle.example.com",
		ValidationYears:  20,
	}
	middleCaCert, err := GenerateMiddleCACert(middleCertInfo, middleKey.GetPublicKey(), key, cert)
	if err != nil {
		t.Fatalf("Create middle ca cert failed")
	}
	middleCaBytes := middleCaCert.GetPemCertBytes()
	err = ioutil.WriteFile("/superchain/temp/1117/middleca.crt", middleCaBytes, 0600)
	if err != nil {
		t.Fatalf("Write middle ca file failed")
	}

	endKey, err := CreateKeyWithCurveP256()
	if err != nil {
		t.Fatalf("Create middle key failed")
	}
	endCertInfo := &cryptomgr.CertBasicInfo{
		Organization:     "hangyan",
		OrganizationUnit: "hua wei",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "admin.middle.example.com",
		ValidationYears:  20,
	}
	endCert, err := GenerateCert(endCertInfo, endKey.GetPublicKey(), middleKey, middleCaCert)
	if err != nil {
		t.Fatalf("Create end cert failed")
	}

	endCertBytes := endCert.GetPemCertBytes()
	err = ioutil.WriteFile("/superchain/temp/1117/end.crt", endCertBytes, 0600)
	if err != nil {
		t.Fatalf("Write middle ca file failed")
	}

}

// TestGetCert test get cert
func TestGetCert(t *testing.T) {
	cert, err := GetCert([]byte(pemStr))
	if err != nil {
		t.Fatalf(err.Error())
	}

	key, err := GetKeyFromPem([]byte(keyStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	signRes, err := key.Sign([]byte("hello"), cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("sign failed with key")
	}
	err = cert.Verify([]byte("hello"), signRes, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("verify failed")
	}

	certNew, err := GetCert(cert.GetPemCertBytes())
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !reflect.DeepEqual(certNew, cert) {
		t.Fatalf("the pem and cert interface can not change to each other")
	}
}

// TestNewEcdsaCert test new ecdsa cert
func TestNewEcdsaCert(t *testing.T) {
	cert, err := GetCert([]byte(pemStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	pemBytes := cert.GetPemCertBytes()
	if strings.EqualFold(string(pemBytes), pemStr) != true {
		t.Fatalf("the pem bytes not correct")
	}
	commonName := cert.GetCommonName()
	if strings.EqualFold(commonName, "peer0.org1.example.com") {
		t.Fatalf("the common name is not correct")
	}

	expireTime := cert.GetExpireTime()

	fmt.Printf("The expire time is %s\n", expireTime.String())

	orgs := cert.GetOrganization()

	fmt.Printf("The orgs is %v\n", orgs)

	msg := []byte("hello")
	key, err := GetKeyFromPem([]byte(keyStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	sign, err := key.Sign(msg, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("key sign failed")
	}
	err = cert.Verify(msg, sign, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("cert verify failed")
	}
}

// TestEcdsacert_Verify test ecdsa cert verify
func TestEcdsacert_Verify(t *testing.T) {
	cert, err := GetCert([]byte(pemStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	msg := []byte("hello")
	key, err := GetKeyFromPem([]byte(keyStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	sign, err := key.Sign(msg, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("key sign failed")
	}

	e := cert.Verify(msg, sign, cryptomgr.Sha256)
	if e != nil {
		t.Fatalf("verify failed")
	}
}

// TestEcdsacert_CheckValidation test ecdsa cert check validation
func TestEcdsacert_CheckValidation(t *testing.T) {
	cert, err := GetCert([]byte(pemStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	caCert, err := GetCert([]byte(caStr))
	if err != nil {
		t.Fatalf(err.Error())
	}

	_, e := cert.CheckValidation([]cryptomgr.Cert{caCert})
	if e != nil {
		t.Fatalf("check validation failed")
	}
}

// TestEcdsacert_GetFingerPrint test ecdsa cert get finger print
func TestEcdsacert_GetFingerPrint(t *testing.T) {
	cert, err := GetCert([]byte(pemStr))
	if err != nil {
		t.Fatalf(err.Error())
	}
	finger := cert.GetFingerPrint()
	if strings.EqualFold(finger, "3a5dab8c7e067ae9ec18e234d4acbfdc9baeac8135e078d1a0a94a15f87c18ee") != true {
		t.Fatalf("the finger is not correct")
	}
}

// TestCreateTLSCACert test create TLS CA cert
func TestCreateTLSCACert(t *testing.T) {
	key, err := CreateKeyWithCurveP256()
	if err != nil {
		t.Fatalf("create key with curve p256 failed")
	}

	endCertInfo := &cryptomgr.CertBasicInfo{
		Organization:     "hangyan",
		OrganizationUnit: "hua wei",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "peer0.org1.example.com",
		ValidationYears:  30,
	}
	_, err = GenerateSelfSignCert(endCertInfo, key)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

type mytest struct {
	number1 string
	number2 int32
}

// TestCreateCertBasic test create cert basic
func TestCreateCertBasic(t *testing.T) {
	myType := reflect.TypeOf(mytest{})
	fmt.Printf("the myType is %s \n", myType)
}

// PemCertAndKey the pem format certificate and key
type PemCertAndKey struct {
	Name string

	// CertPem is the pem format cert which should not be empty
	CertPem string

	// KeyPem is the pem format cert which should not be empty
	KeyPem string
}

// CertInfo the information for certificate creation
type CertInfo struct {
	Usage           int    // it can only be set to SIGN = 1 or TLS = 2;
	SignAlg         string // it can only be set to "sm2_with_sm3" or "ecdsa_with_sha256"
	NodePrefix      string // the prefix for node cert common name, which should not be empty
	ClientPrefix    string // the prefix for client cert common name, which should not be empty
	ValidationYears uint   // validation years ranges from 1 to 30

	Organization     string
	OrganizationUnit string
	Country          string
	Province         string
	Locality         string
}

const (
	// EcdsaWithSha256 : the alg is ecdsa_with_sha256
	EcdsaWithSha256 = "ecdsa_with_sha256"
	// Sm2WithSm3 : the alg is sm2_with_sm3
	Sm2WithSm3 = "sm2_with_sm3"
)
const (
	// SIGN : the certificate usage is sign
	SIGN = 1
	// TLS : the certificate usage is tls communication
	TLS = 2
)

func createSelfSignCACertAndKey(caInfo *CertInfo, commonName string) (*PemCertAndKey, error) {
	if caInfo == nil || commonName == "" {
		return nil, fmt.Errorf("the parameter is not correct")
	}

	basicInfo := &cryptomgr.CertBasicInfo{
		Organization:     caInfo.Organization,
		OrganizationUnit: caInfo.OrganizationUnit,
		Country:          caInfo.Country,
		Province:         caInfo.Province,
		Locality:         caInfo.Locality,
		CommonName:       commonName,
		ValidationYears:  caInfo.ValidationYears,
	}

	p256Key, err := CreateKeyWithCurveP256()
	if err != nil {
		return nil, errors.New("create p256 key failed")
	}

	caCert, err := GenerateSelfSignCert(basicInfo, p256Key)
	if err != nil {
		return nil, errors.New("generate self sign cert failed")
	}
	return &PemCertAndKey{
		Name:    commonName,
		CertPem: string(caCert.GetPemCertBytes()),
		KeyPem:  string(p256Key.GetPemBytes()),
	}, nil
}

// TestEcdsacert_VerifyBatch test ecdsa cert verify batch
func TestEcdsacert_VerifyBatch(t *testing.T) {
	bccryptoutil.InitRandom()
	caInfo := &CertInfo{
		Usage:            SIGN,
		SignAlg:          Sm2WithSm3,
		NodePrefix:       "node",
		ClientPrefix:     "user",
		ValidationYears:  20,
		Organization:     "",
		OrganizationUnit: "",
		Country:          "",
		Province:         "",
		Locality:         "",
	}
	certAndKey, err := createSelfSignCACertAndKey(caInfo, "www.example.com")
	if err != nil {
		t.Fatalf("create cert and key failed")
	}

	key, err := GetKeyFromPem([]byte(certAndKey.KeyPem))
	if err != nil {
		t.Fatalf("get key failed")
	}
	var msgSlice [][]byte
	var hashSlice [][]byte
	var sigSlice [][]byte
	for i := 0; i < 33; i++ {
		msgTmp := fmt.Sprintf("%d", rand.Int())
		msgSlice = append(msgSlice, []byte(msgTmp))
		sign, err := key.SignForBatchVerify([]byte(msgTmp), "")
		if err != nil {
			t.Fatalf("sign failed")
		}
		hashSlice = append(hashSlice, Hash([]byte(msgTmp)))

		sigSlice = append(sigSlice, sign)
	}

	cert, err := GetCert([]byte(certAndKey.CertPem))
	if err != nil {
		t.Fatalf("get cert failed")
	}

	err = cert.VerifyBatch(msgSlice, sigSlice, "")
	if err != nil {
		t.Fatalf("batch verify failed")
	}
	t.Logf("batch verify success")
}

// TestEcdsacert_Verify2 test ecdsa cert verify2
func TestEcdsacert_Verify2(t *testing.T) {
	caInfo := &CertInfo{
		Usage:            SIGN,
		SignAlg:          Sm2WithSm3,
		NodePrefix:       "node",
		ClientPrefix:     "user",
		ValidationYears:  20,
		Organization:     "",
		OrganizationUnit: "",
		Country:          "",
		Province:         "",
		Locality:         "",
	}
	certAndKey, err := createSelfSignCACertAndKey(caInfo, "www.example.com")
	if err != nil {
		t.Fatalf("create cert and key failed")
	}

	key, err := GetKeyFromPem([]byte(certAndKey.KeyPem))
	if err != nil {
		t.Fatalf("get key failed")
	}

	msgTmp := "hello"
	sign, err := key.SignForBatchVerify([]byte(msgTmp), "")
	if err != nil {
		t.Fatalf("sign for batch failed")
		return
	}
	cert, err := GetCert([]byte(certAndKey.CertPem))
	if err != nil {
		t.Fatalf("get cert failed")
	}

	err = cert.Verify([]byte(msgTmp), sign, "")
	if err != nil {
		t.Fatalf("verify2 failed")
	}
	t.Logf("Verify2 success for %s", msgTmp)

	msg2 := "kkkkkk"
	signInfo, err := key.Sign([]byte(msg2), "")
	if err != nil {
		t.Fatalf("key sign failed")
	}
	err = cert.Verify([]byte(msg2), signInfo, "")
	if err != nil {
		t.Fatalf("verify2 failed for %s", msg2)
	}
	t.Logf("verify 2 success for %s", msg2)
}
