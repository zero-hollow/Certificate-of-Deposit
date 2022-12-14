/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package bccryptoutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

const symmetricKeyLen = 16
const ivLen = 16

type aes128 struct {
	key []byte
}

// NewAes128KeyWithKeyBase64 new aes 128 key with base64 key string
func NewAes128KeyWithKeyBase64(keyStr string) (cryptomgr.SymmetricKey, error) {
	decodeString, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	aesKey := aes128{key: decodeString}
	return &aesKey, nil
}

// NewAes128Key new aes 128 key
func NewAes128Key() (cryptomgr.SymmetricKey, error) {
	key, err := generateAES128Key()
	if err != nil {
		return nil, err
	}
	aeskey := aes128{key: key}
	return &aeskey, nil
}

// GetKeyBytes get bytes format key
func (a *aes128) GetKeyBytes() []byte {
	return a.key
}

// GetKeyBase64Str get base64 format key
func (a *aes128) GetKeyBase64Str() string {
	return base64.StdEncoding.EncodeToString(a.key)
}

// Encrypt encrypt plain txt
func (a *aes128) Encrypt(plainTxt []byte) ([]byte, error) {
	iv, err := generateAES128IV()
	if err != nil {
		return nil, err
	}

	cipherMsg, err := encWithAES128IV(a.key, iv, plainTxt)
	if err != nil {
		return nil, err
	}
	cipherMsg = append(cipherMsg, iv...)
	return cipherMsg, nil
}

// Decrypt decrypt message with cipher txt and iv
func (a *aes128) Decrypt(cipherTxtAndIv []byte) ([]byte, error) {
	length := len(cipherTxtAndIv)
	if length < ivLen+symmetricKeyLen || length%ivLen != 0 {
		return nil, errors.Errorf("the cipher msg length:%d is not correct", length)
	}

	iv := cipherTxtAndIv[length-ivLen:]
	encMgs := cipherTxtAndIv[:length-ivLen]
	return decWithAES128IV(a.key, iv, encMgs)
}

// generateAES128Key generate secure random number(128bit) as AES 128 key
func generateAES128Key() ([]byte, error) {
	aesKey := make([]byte, symmetricKeyLen)
	n, err := rand.Read(aesKey)
	if err != nil {
		return nil, err
	}
	if n != symmetricKeyLen {
		return nil, errors.Errorf("the aes 128 key len:%d is not correct", n)
	}
	return aesKey, nil
}

func generateAES128IV() ([]byte, error) {
	aes128IV := make([]byte, ivLen)
	n, err := rand.Read(aes128IV)
	if err != nil {
		return nil, err
	}
	if n != ivLen {
		return nil, errors.Errorf("the aes 128 iv len:%d is not correct", n)
	}
	return aes128IV, nil
}

// encWithAES128IV enc with aes 128 CBC mode, the enc result length is len(result)%16==0 && len(result)-len(text)<16
func encWithAES128IV(key []byte, iv []byte, text []byte) ([]byte, error) {
	// generate key data block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// pad to be 16x
	blockSize := block.BlockSize()
	originData := pad(text, blockSize)
	// get the cbc block mode
	blockMode := cipher.NewCBCEncrypter(block, iv)
	// crypto and output to crypted
	crypted := make([]byte, len(originData))
	blockMode.CryptBlocks(crypted, originData)
	return crypted, nil
}

// PKCS7 padding mode
func pad(ciphertext []byte, blockSize int) []byte {
	if blockSize == 0 {
		return ciphertext
	}
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func decWithAES128IV(key []byte, iv []byte, text []byte) ([]byte, error) {
	// todo: 去掉中文注释
	// get key block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// get the cbc mode
	blockMode := cipher.NewCBCDecrypter(block, iv)
	// output to the originData
	originData := make([]byte, len(text))
	blockMode.CryptBlocks(originData, text)
	// delete the padding
	return unpad(originData), nil
}

func unpad(ciphertext []byte) []byte {
	length := len(ciphertext)
	// delete the last time padding
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}
