/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package bccryptoutil

import (
	"encoding/base64"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/gmssl"
)

type gmSm4 struct {
	key []byte
}

// NewSm4KeyWithKeyBase64 new sm4 key with base64 key string
func NewSm4KeyWithKeyBase64(keyStr string) (cryptomgr.SymmetricKey, error) {
	decodeString, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	sm4 := &gmSm4{key: decodeString}
	return sm4, nil
}

// NewSm4Key new sm4 key
func NewSm4Key() (cryptomgr.SymmetricKey, error) {
	key, err := generateSm4Key()
	if err != nil {
		return nil, err
	}
	sm4 := &gmSm4{key: key}
	return sm4, nil
}

// GetKeyBytes get bytes format key
func (g *gmSm4) GetKeyBytes() []byte {
	return g.key
}

// GetKeyBase64Str get base64 format key
func (g *gmSm4) GetKeyBase64Str() string {
	return base64.StdEncoding.EncodeToString(g.key)
}

// Encrypt encrypt plain txt
func (g *gmSm4) Encrypt(plainTxt []byte) ([]byte, error) {
	iv, err := generateSm4IV()
	if err != nil {
		return nil, err
	}

	cipherMsg, err := encWithSm4IV(g.key, iv, plainTxt)
	if err != nil {
		return nil, err
	}
	cipherMsg = append(cipherMsg, iv...)
	return cipherMsg, nil
}

// Decrypt decrypt cipher txt and iv
func (g *gmSm4) Decrypt(cipherTxtAndIv []byte) ([]byte, error) {
	length := len(cipherTxtAndIv)
	if length < ivLen+symmetricKeyLen {
		return nil, errors.Errorf("the cipher msg length:%d is not correct", length)
	}
	iv := cipherTxtAndIv[length-ivLen:]
	encMgs := cipherTxtAndIv[:length-ivLen]
	return decWithSm4IV(g.key, iv, encMgs)
}

// generateSm4Key generate random key
func generateSm4Key() ([]byte, error) {
	keyLength, err := gmssl.GetCipherKeyLength("SMS4")
	if err != nil {
		return nil, err
	}
	key, err := gmssl.GenerateRandom(keyLength)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// generateSm4IV generate random iv
func generateSm4IV() ([]byte, error) {
	ivLength, err := gmssl.GetCipherIVLength("SMS4")
	if err != nil {
		return nil, err
	}
	key, err := gmssl.GenerateRandom(ivLength)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// encWithSm4IV enc with sm4
func encWithSm4IV(key []byte, iv []byte, msg []byte) ([]byte, error) {
	return procWithSm4IV(key, iv, msg, true)
}

func procWithSm4IV(key []byte, iv []byte, msg []byte, isEncrypt bool) ([]byte, error) {
	encryptor, err := gmssl.NewCipherContext("SMS4", key, iv, isEncrypt)
	if err != nil {
		return nil, err
	}
	text1, err := encryptor.Update(msg)
	if err != nil {
		return nil, err
	}
	text2, err := encryptor.Final()
	if err != nil {
		return nil, err
	}
	text := make([]byte, 0, len(text1)+len(text2))
	text = append(text, text1...)
	text = append(text, text2...)
	return text, nil
}

// decWithSm4IV dec with sm4
func decWithSm4IV(key []byte, iv []byte, encMsg []byte) ([]byte, error) {
	return procWithSm4IV(key, iv, encMsg, false)
}
