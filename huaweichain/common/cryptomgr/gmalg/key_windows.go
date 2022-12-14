/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package gmalg the gm algorithm for generate of certificate and key
package gmalg

import (
	"git.huawei.com/huaweichain/common/cryptomgr"
)

// GmKey gm key
type GmKey struct {
}

// GeneratePriKey generate private key
func GeneratePriKey() (cryptomgr.Key, error) {
	return nil, nil
}

// GetPubKey generate public key from pem
func GetPubKey(pemKey []byte) cryptomgr.Key {
	return nil
}

// GetPriKey generate private key from pem
func GetPriKey(pemKey []byte) (cryptomgr.Key, error) {
	return nil, nil
}

// GetPemBytes get pem bytes
func (g *GmKey) GetPemBytes() []byte {
	return nil
}

// IsSymmetric is symmetric
func (g *GmKey) IsSymmetric() bool {
	return false
}

// IsPrivate is private key
func (g *GmKey) IsPrivate() bool {
	return false
}

// GetPublicKey get public key
func (g *GmKey) GetPublicKey() (pubKey interface{}, err error) {
	return nil, nil
}

// GetPrivateKey get private key
func (g *GmKey) GetPrivateKey() (priKey interface{}, err error) {
	return nil, nil
}

// Sign sign the message
func (g *GmKey) Sign(msg []byte, hashAlg string) ([]byte, error) {
	return nil, nil
}

// Verify verify the message
func (g *GmKey) Verify(msg []byte, signature []byte, hashAlg string) error {
	return nil
}
