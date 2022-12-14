/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package bccryptoutil

import (
	"bytes"
	"encoding/base64"
	"testing"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

var testAesKeyBase64Str = "PuLWPGL2TbfJP6NY2ItCBg=="

func getEncMagicNumber() []byte {
	return []byte{'0', '0', '0', '0', '0', '0', '0', '0'}
}

func TestShortValue(t *testing.T) {
	value := []byte("80")
	magicValue := getEncMagicNumber()
	magicValue = append(magicValue, value...)
	keyBase64, err := NewAes128KeyWithKeyBase64(testAesKeyBase64Str)
	if err != nil {
		t.Fatalf("new aes 128 key with base64 string")
	}
	encryptedValue, err := keyBase64.Encrypt(magicValue)
	if err != nil {
		t.Fatalf("encrypt failed")
	}
	magicWithEncValue := getEncMagicNumber()
	magicWithEncValue = append(magicWithEncValue, encryptedValue...)
	log.Infof("the magicWithEncValue len is %d", len(magicWithEncValue))
}

func TestNewAes128KeyWithKeyBase64(t *testing.T) {
	keyBase64, err := NewAes128KeyWithKeyBase64(testAesKeyBase64Str)
	if err != nil {
		t.Fatalf("new aes 128 key with base64 string")
	}
	var plainTxt = "qwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopggggg"
	encAndDec(t, keyBase64, plainTxt)

	log.Infof("new aes 128 key with base64 string success!")
}

func TestAesDecryptWrongLength(t *testing.T) {
	keyBase64, err := NewAes128KeyWithKeyBase64(testAesKeyBase64Str)
	if err != nil {
		t.Fatalf("new aes 128 key with base64 string")
	}
	var plainTxt = "qwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiop"
	encrypt, err := keyBase64.Encrypt([]byte(plainTxt))
	if err != nil {
		t.Fatalf("aes 128 encrypt failed:%s", err.Error())
	}
	_, err = keyBase64.Decrypt(encrypt[3:])
	if err == nil {
		t.Fatalf("decrypt should be failed")
	}
	log.Infof("test wrong decrypt length success")
}

func TestNewAes128Key(t *testing.T) {
	key, err := NewAes128Key()
	if err != nil {
		t.Fatalf("new aes 128 key failed")
	}
	var plainTxt = "qwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiop"
	encAndDec(t, key, plainTxt)

	plainTxt = "hello"
	encAndDec(t, key, plainTxt)

	str := key.GetKeyBase64Str()
	log.Infof("the aes key str is \n\n%s\n\n", str)

	log.Infof("aes 128 key encrypt and decrypt success!")
}

func encAndDec(t *testing.T, key cryptomgr.SymmetricKey, plainTxt string) {
	encrypt, err := key.Encrypt([]byte(plainTxt))
	if err != nil {
		t.Fatalf("aes 128 key encrypt failed")
	}
	decrypt, err := key.Decrypt(encrypt)
	if err != nil {
		t.Fatalf("aes 128 key decrypt failed")
	}
	if !bytes.Equal([]byte(plainTxt), decrypt) {
		t.Fatalf("aes 128 encrypt and decrypt failed")
	}
}

func TestGenerateAES128Key(t *testing.T) {
	keyMap := make(map[string][]byte)

	for i := 0; i < 1000; i++ {
		key, err := generateAES128Key()
		if err != nil {
			t.Fatalf("generate aes 128 key failed")
		}
		if len(key) != symmetricKeyLen {
			t.Fatalf("the aes key length is not correct:%d", len(key))
		}
		keyStr := base64.StdEncoding.EncodeToString(key)
		_, ok := keyMap[keyStr]
		if ok {
			t.Fatalf("the ase 128 key has been exist")
		}
		keyMap[keyStr] = key
	}
	log.Infof("generate aes 128 key success\n")
}

func TestGenerateAES128IV(t *testing.T) {
	keyMap := make(map[string][]byte)

	for i := 0; i < 1000; i++ {
		key, err := generateAES128IV()
		if err != nil {
			t.Fatalf("generate aes 128 iv failed")
		}
		if len(key) != symmetricKeyLen {
			t.Fatalf("the aes iv length is not correct:%d", len(key))
		}
		keyStr := base64.StdEncoding.EncodeToString(key)
		_, ok := keyMap[keyStr]
		if ok {
			t.Fatalf("the ase 128 iv has been exist")
		}
		keyMap[keyStr] = key
	}
	log.Infof("generate aes 128 iv success\n")
}

func TestDecWithAES128(t *testing.T) {
	text := "qwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuiopqwertyuio"
	key, err := generateSm4Key()
	if err != nil {
		t.Fatalf("generate key failed")
	}
	if len(key) != symmetricKeyLen {
		t.Fatalf("the aes key len is not correct")
	}

	iv, err := generateSm4IV()
	if err != nil {
		t.Fatalf("generate rand iv failed")
	}
	if len(iv) != ivLen {
		t.Fatalf("the aes iv len is not correct")
	}

	encInfo, err := encWithAES128IV(key, iv, []byte(text))
	if err != nil {
		t.Fatalf("enc with aes 128 failed")
	}

	if len(encInfo) < ivLen+symmetricKeyLen {
		t.Fatalf("the enc info len is not correct:%d", len(encInfo))
	}

	plainTxt, err := decWithAES128IV(key, iv, encInfo)
	if err != nil {
		t.Fatalf("dec with aes 128 failed")
	}

	if !bytes.Equal([]byte(text), plainTxt) {
		t.Fatalf("the aes encrypt and decrypt is not match")
	}
	log.Infof("aes 128 encrypt and decrypt success!")
}
