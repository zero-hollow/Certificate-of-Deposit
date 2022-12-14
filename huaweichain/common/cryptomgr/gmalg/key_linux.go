/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

// Package gmalg the gm algorithm for generate of certificate and key
package gmalg

import (
	"fmt"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/gmssl"
)

type gmKey struct {
	isPrivate bool
	priKey    *gmssl.PrivateKey
	pemPriKey []byte
	derPriKey []byte
	pubKey    *gmssl.PublicKey
	pemPubKey []byte
	derPubKey []byte
}

const (
	ecParamgenCurve = "ec_paramgen_curve"
	sm2Curve        = "sm2p256v1"
	ecParamEnc      = "ec_param_enc"
	namedCurve      = "named_curve"
	ecAlg           = "EC"
)

// GeneratePriKey generate private key
func GeneratePriKey() (cryptomgr.Key, error) {
	sm2keygenargs := [][2]string{
		{ecParamgenCurve, sm2Curve},
		{ecParamEnc, namedCurve},
	}
	sm2Key, err := gmssl.GeneratePrivateKey(ecAlg, sm2keygenargs, nil)
	if err != nil {
		log.Errorf("gmssl generate private key failed.")
		return nil, fmt.Errorf("gmssl generate private key failed")
	}
	pemSm2Key, err := sm2Key.GetPemWithoutEnc()
	if err != nil {
		log.Errorf("get pem format gm key failed.")
		return nil, fmt.Errorf("get pem format gm key failed")
	}
	return GetPriKey([]byte(pemSm2Key))
}

// GetPubKey generate public key from pem
func GetPubKey(pemKey []byte) (cryptomgr.Key, error) {
	if pemKey == nil {
		log.Infof("the pem public key is nil")
		return nil, fmt.Errorf("the pem public key is nil")
	}
	key := &gmKey{}
	publicKey, err := gmssl.NewPublicKeyFromPEM(string(pemKey))
	if err != nil {
		log.Infof("This is not a gm public key.")
		return nil, fmt.Errorf("this is not a gm public key")
	}

	key.pubKey = publicKey
	der, err := publicKey.GetDer()
	if err != nil {
		return nil, err
	}
	key.derPubKey = der
	key.pemPubKey = pemKey
	key.isPrivate = false
	return key, nil
}

// GetPriKey generate private key from pem
func GetPriKey(pemKey []byte) (cryptomgr.Key, error) {
	if pemKey == nil {
		log.Infof("the pem private key is nil")
		return nil, fmt.Errorf("the pem private key is nil")
	}
	key := &gmKey{}
	privateKey, err := gmssl.NewPrivateKeyFromPEM(string(pemKey), "test")
	if err != nil {
		return nil, err
	}
	key.priKey = privateKey
	key.pemPriKey = pemKey
	derPriKey, err := privateKey.GetDer()
	if err != nil {
		return nil, errors.WithMessagef(err, "get der private key failed")
	}
	key.derPriKey = derPriKey

	publicKeyPEM, err := privateKey.GetPublicKeyPEM()
	if err != nil {
		log.Infof("Get gm public key from private key failed")
		return nil, err
	}
	key.pemPubKey = []byte(publicKeyPEM)
	publicKey, err := gmssl.NewPublicKeyFromPEM(string(key.pemPubKey))
	if err != nil {
		log.Infof("Generate gm public key from pem publie key failed")
		return nil, err
	}
	derPubKey, err := publicKey.GetDer()
	if err != nil {
		return nil, err
	}
	key.derPubKey = derPubKey
	key.pubKey = publicKey
	key.isPrivate = true
	return key, nil
}

// GetPemBytes get pem format key
func (g *gmKey) GetPemBytes() []byte {
	if g.isPrivate {
		return g.pemPriKey
	}
	return g.pemPubKey
}

// GetDerBytes get der format key
func (g *gmKey) GetDerBytes() ([]byte, error) {
	if g.isPrivate {
		return g.derPriKey, nil
	}
	return g.derPubKey, nil
}

// IsSymmetric is symmetric key
func (g *gmKey) IsSymmetric() bool {
	return false
}

// IsPrivate is private key
func (g *gmKey) IsPrivate() bool {
	return g.isPrivate
}

// GetPublicKey get public key object from private key
func (g *gmKey) GetPublicKey() cryptomgr.Key {
	pubKey, err := GetPubKey(g.pemPubKey)
	if err != nil {
		return nil
	}
	return pubKey
}

// GetPrivateKey get private key
func (g *gmKey) GetPrivateKey() (priKey interface{}, err error) {
	return g.priKey, nil
}

// SignForBatchVerify get the signature of msg for batch verify
func (g *gmKey) SignForBatchVerify(msg []byte, hashAlg string) ([]byte, error) {
	// todo don't support sign for batch now, so call Sign function as a substitute
	return g.Sign(msg, hashAlg)
}

// Sign get signature
func (g *gmKey) Sign(msg []byte, hashAlg string) ([]byte, error) {
	if !g.isPrivate {
		return nil, fmt.Errorf("this is not private key")
	}
	digest := calcSm2SignDigest(g.pubKey, msg)
	if digest == nil {
		return nil, fmt.Errorf("calc digest failed in sign")
	}

	sig, err := g.priKey.Sign(sm2SignFlag, digest, nil)
	if err != nil {
		return nil, fmt.Errorf("sign error")
	}
	return sig, nil
}

// Verify verify signature
func (g *gmKey) Verify(msg []byte, signature []byte, hashAlg string) error {
	digest := calcSm2SignDigest(g.pubKey, msg)
	if digest == nil {
		return fmt.Errorf("calc digest failed when public verify")
	}
	err := g.pubKey.Verify(sm2SignFlag, digest, signature, nil)
	return err
}

func calcSm2SignDigest(sm2pk *gmssl.PublicKey, message []byte) []byte {
	sm3ctx, _ := gmssl.NewDigestContext(sm3HashFlag)
	sm2zid, _ := sm2pk.ComputeSM2IDDigest(defaultSm2ID)
	err := sm3ctx.Reset()
	if err != nil {
		log.Errorf("Reset sm3 ctx failed:%s", err.Error())
		return nil
	}
	err = sm3ctx.Update(sm2zid)
	if err != nil {
		log.Errorf("Update sm3 failed for sm2zid:%s", err.Error())
		return nil
	}
	err = sm3ctx.Update(message)
	if err != nil {
		log.Errorf("Update sm3 failed for message:%s", err.Error())
		return nil
	}

	hashRes, err := sm3ctx.Final()
	if err != nil {
		log.Errorf("Final sm3 failed:%s", err.Error())
		return nil
	}
	return hashRes
}
