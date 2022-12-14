/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package ecdsaalg

import (
	"crypto/elliptic"
	"reflect"
	"strings"
	"testing"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

func TestGetPriKeyFromPem(t *testing.T) {
	var key1Str = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgeQ1WPoPDyUmMF+nX
Vhfch3DRVt8pM9oK7di9IZkiiOWhRANCAAQKoizhzJsSbegWZsOmFFdge7gwDP1b
ACLvfmhTIQYJppWIKMYGJFYJ+O73l2MxBRpyBBOnayywMdWArGK0WF7x
-----END PRIVATE KEY-----`
	var message = []byte("hello world")
	var message2 = []byte("helloworld")
	key, err := GetKeyFromPem([]byte(key1Str))
	if err != nil {
		t.Fatalf(err.Error())
	}
	isPri := key.IsPrivate()
	if isPri != true {
		t.Fatalf("expect this key is private key, but not")
	}

	if strings.EqualFold(string(key.GetPemBytes()), key1Str) != true {
		t.Fatalf("the pem bytes of private key is not correct")
	}

	sig, err := key.Sign(message, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("sign failed:%s", err.Error())
	}

	publicKey := key.GetPublicKey()
	if publicKey == nil {
		t.Fatalf("get public key from private key failed")
	}

	err = publicKey.Verify(message, sig, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("expect verify success, but verify failed:%s", err.Error())
	}

	err = key.GetPublicKey().Verify(message2, sig, cryptomgr.Sha256)
	if err == nil {
		t.Fatalf("expect verify failed, but verify success")
	}
}

func TestCreateKeyWithCurve(t *testing.T) {
	var message = []byte("hello")
	var message2 = []byte("helloworld")

	curve := elliptic.P256()

	key, err := CreateKeyWithCurve(curve)
	if err != nil {
		t.Fatalf("Create key with curve failed")
	}
	sig, err := key.Sign(message, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("sign failed with new key:%s", err.Error())
	}
	pubKey := key.GetPublicKey()

	err = pubKey.Verify(message, sig, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("expect verify success, but verify failed")
	}

	err = pubKey.Verify(message2, sig, cryptomgr.Sha256)
	if err == nil {
		t.Fatalf("expect verify failed, but verify success")
	}

	info := &cryptomgr.CertBasicInfo{
		Organization:     "",
		OrganizationUnit: "",
		Country:          "",
		Province:         "",
		Locality:         "",
		CommonName:       "www.test.com",
		ValidationYears:  20,
	}
	cert, err := GenerateSelfSignCert(info, key)
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = cert.Verify(message, sig, cryptomgr.Sha256)
	if err != nil {
		t.Fatalf("cert verify, expect success, but failed")
	}

	newKey, err := GetKeyFromPem(key.GetPemBytes())
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !reflect.DeepEqual(newKey, key) {
		t.Fatalf("the key is not equal to new key")
	}
}

func TestGetPubKeyFromPem(t *testing.T) {
	var key1Str = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgeQ1WPoPDyUmMF+nX
Vhfch3DRVt8pM9oK7di9IZkiiOWhRANCAAQKoizhzJsSbegWZsOmFFdge7gwDP1b
ACLvfmhTIQYJppWIKMYGJFYJ+O73l2MxBRpyBBOnayywMdWArGK0WF7x
-----END PRIVATE KEY-----`
	priKey, err := GetKeyFromPem([]byte(key1Str))
	if err != nil {
		t.Fatalf(err.Error())
	}
	if priKey == nil {
		t.Fatalf("get private key failed")
	}

	_, ok := priKey.(*ecdsaKey)
	if !ok {
		t.Fatalf("the key is not ecdsa private key")
	}

	publicKey := priKey.GetPublicKey()
	if publicKey == nil {
		t.Fatalf("get public key from private key failed")
	}
	isPri := publicKey.IsPrivate()
	if isPri == true {
		t.Fatalf("for public key , expect return false, but true")
	}

	pubPemBytes := publicKey.GetPemBytes()

	publicKey2, err := GetKeyFromPem(pubPemBytes)
	if err != nil {
		t.Fatalf(err.Error())
	}

	isEqual := reflect.DeepEqual(pubPemBytes, publicKey2.GetPemBytes())
	if isEqual != true {
		t.Fatalf("for public key, get pem bytes failed")
	}
}
