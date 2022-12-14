/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package gmalg

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

var gmCaCert = `-----BEGIN CERTIFICATE-----
MIICwzCCAmmgAwIBAgIRAJ3DNbiyL/Jj/IRW/d25NAAwCgYIKoEcz1UBg3UwgZMx
DTALBgNVBAYTBHRlc3QxDTALBgNVBAgTBHRlc3QxDTALBgNVBAcTBHRlc3QxDTAL
BgNVBAkTBHRlc3QxDTALBgNVBBETBHRlc3QxGTAXBgNVBAoTEG9yZzEuZXhhbXBs
ZS5jb20xDTALBgNVBAsTBHRlc3QxHDAaBgNVBAMTE2NhLm9yZzEuZXhhbXBsZS5j
b20wHhcNMjAwODE4MTIwNDAwWhcNMzAwODE2MTIwNDAwWjCBkzENMAsGA1UEBhME
dGVzdDENMAsGA1UECBMEdGVzdDENMAsGA1UEBxMEdGVzdDENMAsGA1UECRMEdGVz
dDENMAsGA1UEERMEdGVzdDEZMBcGA1UEChMQb3JnMS5leGFtcGxlLmNvbTENMAsG
A1UECxMEdGVzdDEcMBoGA1UEAxMTY2Eub3JnMS5leGFtcGxlLmNvbTBZMBMGByqG
SM49AgEGCCqBHM9VAYItA0IABEc3yCGBrsaj8Anh+Pway5c6LaZCar10JrsYH7bc
7Qmp1YlDDPj/BV7iWS2vS5wBQ2o0mdW8hAsUVPwTL+hkGiqjgZswgZgwDgYDVR0P
AQH/BAQDAgGmMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcDATAPBgNVHRMB
Af8EBTADAQH/MCkGA1UdDgQiBCC7wlVdjmzSGnpWnKahbu64Qu3CJi4ad4teDRNm
UMjQRjArBgNVHSMEJDAigCC7wlVdjmzSGnpWnKahbu64Qu3CJi4ad4teDRNmUMjQ
RjAKBggqgRzPVQGDdQNIADBFAiEAvuntSPrhjio3IhwtVW/RIgzWgM2GgX0Ozw8S
MxzcnoMCICBJNB5LnH2BpAuQQHkGtp3Zo8BW0kvXHjqwQLvPc5cp
-----END CERTIFICATE-----
`

// TestGetCert test get cert
func TestGetCert(t *testing.T) {
	_, err := GetCert([]byte(gmCaCert))
	if err != nil {
		t.Errorf("get gm cert from pem failed")
		return
	}
	return
}

// TestGenerateSelfSignCert test generate self sign cert
func TestGenerateSelfSignCert(t *testing.T) {
	key, err := GeneratePriKey()
	if err != nil {
		t.Fatalf("create gm key failed\n")
		return
	}
	years := 25
	info := &cryptomgr.CertBasicInfo{
		Organization:     "huawei",
		OrganizationUnit: "hanyansuo",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "www.test.com",
		ValidationYears:  uint(years),
	}
	cert, err := GenerateSelfSignCert(info, key)
	if err != nil {
		t.Fatalf("create cert failed, err:%s ", err.Error())
	}
	if strings.EqualFold(cert.GetCommonName(), "www.test.com") != true {
		t.Fatalf("the cert common name is not correct")
	}

	if (cert.GetExpireTime().Year()) != (time.Now().Year() + years) {
		t.Fatalf("the time for expire is not %d years, the expire year is %d", years, cert.GetExpireTime().Year())
	}
	fingerPrint := cert.GetFingerPrint()
	t.Logf("the finger print is %s", fingerPrint)

	t.Logf("Generate self sign cert success\n")
}

// TestGenerateCert test generate cert
func TestGenerateCert(t *testing.T) {
	caKey, err := GeneratePriKey()
	if err != nil {
		t.Fatalf("create gm key failed\n")
		return
	}
	years := 17
	info := &cryptomgr.CertBasicInfo{
		Organization:     "huawei",
		OrganizationUnit: "hangyansuo",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "www.test.com",
		ValidationYears:  uint(years),
	}
	caCert, err := GenerateSelfSignCert(info, caKey)
	if err != nil {
		t.Fatalf("create cert failed ")
	}

	middlePriKey, err := GeneratePriKey()
	if err != nil {
		t.Fatalf("create myPriKey failed\n")
		return
	}
	middleCertInfo := &cryptomgr.CertBasicInfo{
		Organization:     "huawei",
		OrganizationUnit: "hangyansuo",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "middle.www.test.com",
		ValidationYears:  uint(years),
	}
	middleCert, err := GenerateMiddleCACert(middleCertInfo, middlePriKey.GetPublicKey(), caKey, caCert)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if strings.EqualFold(middleCert.GetCommonName(), "middle.www.test.com") != true {
		t.Fatalf("the cert common name is not correct")
	}

	if (middleCert.GetExpireTime().Year()) != (time.Now().Year() + years) {
		t.Fatalf("the time for expire is not %d years", years)
	}

	adminPriKey, err := GeneratePriKey()
	if err != nil {
		t.Fatalf("create myPriKey failed\n")
		return
	}
	endCertInfo := &cryptomgr.CertBasicInfo{
		Organization:     "huawei",
		OrganizationUnit: "hangyansuo",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "admin.www.test.com",
		ValidationYears:  uint(years),
	}
	cert, err := GenerateCert(endCertInfo, adminPriKey.GetPublicKey(), middlePriKey, middleCert)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if strings.EqualFold(cert.GetCommonName(), "admin.www.test.com") != true {
		t.Fatalf("the cert common name is not correct")
	}

	if (cert.GetExpireTime().Year()) != (time.Now().Year() + years) {
		t.Fatalf("the time for expire is not %d years", years)
	}
}

// TestGmcert_Verify test gm cert verify
func TestGmcert_Verify(t *testing.T) {
	caKey, err := GeneratePriKey()
	if err != nil {
		t.Fatalf("create gm key failed\n")
		return
	}
	years := 18
	info := &cryptomgr.CertBasicInfo{
		Organization:     "huawei",
		OrganizationUnit: "hangyansuo",
		Country:          "CN",
		Province:         "zhejiang",
		Locality:         "hangzhou",
		CommonName:       "www.test.com",
		ValidationYears:  uint(years),
	}
	caCert, err := GenerateSelfSignCert(info, caKey)
	if err != nil {
		t.Fatalf("create cert failed ")
	}

	msg := []byte("hello")
	sig, err := caKey.Sign(msg, cryptomgr.Sm3)
	if err != nil {
		t.Fatalf("sign failed")
	}

	err = caCert.Verify(nil, sig, cryptomgr.Sm3)
	if err == nil {
		t.Fatalf("should verify failed, but not")
	}
}

var myCert = `-----BEGIN CERTIFICATE-----
MIICJTCCAcqgAwIBAgIRAK15TYbgmTRp4GDkYALXz1QwCgYIKoZIzj0EAwIwWTEJ
MAcGA1UEBhMAMQkwBwYDVQQIEwAxCTAHBgNVBAcTADENMAsGA1UEChMEb3JnMTEJ
MAcGA1UECxMAMRwwGgYDVQQDExNjYS5vcmcxLmV4YW1wbGUuY29tMB4XDTIxMTAx
ODA3NDQxMFoXDTIyMTAxODA3NDQxMFowXDEJMAcGA1UEBhMAMQkwBwYDVQQIEwAx
CTAHBgNVBAcTADENMAsGA1UEChMEb3JnMTEJMAcGA1UECxMAMR8wHQYDVQQDExZw
ZWVyMC5vcmcxLmV4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE
YgxMRl4HmUkKEtAwW2TahifUAYJg5trnZe22XsMEG0OnuXyjRrf8Q0Fn9fLXC8aa
25aCZlhr3BvNiKcVYFtw7KNwMG4wDgYDVR0PAQH/BAQDAgWgMAwGA1UdEwEB/wQC
MAAwKwYDVR0jBCQwIoAgfEBgcph26VskTkd1vPTIISHYnK1CXPGxdq45FhH13Zsw
IQYDVR0RBBowGIIWcGVlcjAub3JnMS5leGFtcGxlLmNvbTAKBggqhkjOPQQDAgNJ
ADBGAiEAjEUAcj+fTJQkuln6tU1ajW5CKo5DRpOHUWQUL9QkGVoCIQCiJ9cWZ4+6
kbQVSu3p0LGMZT0loMnbUPURq7KkQfP2hg==
-----END CERTIFICATE-----`

var myKey = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgZwd9zmBdWuqN7Yya
9ez0kHWy/OlREYHVY1zhoJ6k+KqhRANCAARiDExGXgeZSQoS0DBbZNqGJ9QBgmDm
2udl7bZewwQbQ6e5fKNGt/xDQWf18tcLxprbloJmWGvcG82IpxVgW3Ds
-----END PRIVATE KEY-----`

func TestGmcert_GetDerPublicKey(t *testing.T) {
	cert, err := GetCert([]byte(myCert))
	if err != nil {
		t.Fatalf("get cert failed")
	}
	derPublicKey1, err := cert.GetDerPublicKey()
	if err != nil {
		t.Fatalf("get der public key failed:%s", err.Error())
	}
	fmt.Printf("the public key len is %d\n", len(derPublicKey1))

	priKey, err := GetPriKey([]byte(myKey))
	if err != nil {
		t.Fatalf("get der private key failed:%s", err.Error())
	}
	publicKey := priKey.GetPublicKey()
	derPublieKey2, err := publicKey.GetDerBytes()
	if err != nil {
		t.Fatalf("get private key failed:%s", err.Error())
	}
	if len(derPublicKey1) != len(derPublieKey2) {
		t.Fatalf("the der public key length is not equal")
	}
	fmt.Printf("test der success!\n")
}
