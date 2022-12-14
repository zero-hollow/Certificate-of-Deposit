/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

package ecdsaalg

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"math/big"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

type ecdsaKey struct {
	pemKey       []byte
	pemPubKey    []byte
	priKey       *ecdsa.PrivateKey
	isPrivateKey bool
	pubKey       *ecdsa.PublicKey
	curve        elliptic.Curve
	X            *big.Int
	D            *big.Int
	Y            *big.Int
}

// CreateKeyWithCurve create key with elliptic curve.
func CreateKeyWithCurve(cruve elliptic.Curve) (cryptomgr.Key, error) {
	ecKey := &ecdsaKey{}
	ecKey.curve = cruve
	privateKey, err := ecdsa.GenerateKey(ecKey.curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	ecKey.priKey = privateKey
	ecKey.isPrivateKey = true
	ecKey.pubKey = &ecKey.priKey.PublicKey
	ecKey.X = privateKey.X
	ecKey.D = privateKey.D
	ecKey.Y = privateKey.Y

	bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	memory := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bytes})
	ecKey.pemKey = memory

	bytes, err = x509.MarshalPKIXPublicKey(ecKey.pubKey)
	if err != nil {
		return nil, err
	}
	memory = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: bytes})
	ecKey.pemPubKey = memory

	return ecKey, nil
}

// CreateKeyWithCurveP256 create key with curve p256.
func CreateKeyWithCurveP256() (cryptomgr.Key, error) {
	curve := elliptic.P256()
	return CreateKeyWithCurve(curve)
}

// CreatePublicKeyP256WithParam Create ecdsa PublicKey With P256 X & Y.
func CreatePublicKeyP256WithParam(x *big.Int, y *big.Int) (cryptomgr.Key, error) {
	ecKey := &ecdsaKey{}
	ecKey.curve = elliptic.P256()

	ecKey.isPrivateKey = false
	var pubKey ecdsa.PublicKey
	pubKey.X = x
	pubKey.Y = y
	pubKey.Curve = elliptic.P256()
	ecKey.pubKey = &pubKey

	bytes, err := x509.MarshalPKIXPublicKey(ecKey.pubKey)
	if err != nil {
		return nil, err
	}
	memory := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: bytes})
	ecKey.pemPubKey = memory

	return ecKey, nil
}

// CreateKeyFromBuffer Create ecdsa public Key From tee Buffer
func CreateKeyFromBuffer(buffer []byte) (cryptomgr.Key, error) {
	// lv decode
	const lenOccupation = 2
	const bitsNumInOneByte = 8
	bufLen := len(buffer)
	if bufLen <= lenOccupation {
		return nil, errors.Errorf("the buffer len is <= lenOccupation %d", lenOccupation)
	}
	xLen := int(buffer[0]) + int(buffer[1])<<bitsNumInOneByte

	if bufLen < xLen+lenOccupation {
		return nil, errors.Errorf("the buffer len is <= xLen+lenOccupation %d", xLen+lenOccupation)
	}

	bigX := new(big.Int).SetBytes(buffer[lenOccupation : xLen+lenOccupation])

	if bufLen <= xLen+lenOccupation+lenOccupation {
		return nil, errors.Errorf("bufLen %d <= xLen+lenOccupation+"+
			"lenOccupation", xLen+lenOccupation+lenOccupation)
	}
	yLen := int(buffer[xLen+lenOccupation]) + int(buffer[xLen+lenOccupation+1])<<bitsNumInOneByte
	if bufLen < lenOccupation+xLen+lenOccupation+yLen {
		return nil, errors.New("the buffer len is smaller than xlen+ylen+3")
	}
	bigY := new(big.Int).SetBytes(buffer[lenOccupation+xLen+lenOccupation : lenOccupation+xLen+lenOccupation+yLen])

	return CreatePublicKeyP256WithParam(bigX, bigY)
}

// GetKeyFromPem get key from pem bytes
func GetKeyFromPem(pemKey []byte) (cryptomgr.Key, error) {
	if pemKey == nil {
		return nil, errors.New("the pem key is nil")
	}
	block, _ := pem.Decode(pemKey)
	if block == nil {
		return nil, errors.New("decode pem key failed")
	}

	switch block.Type {
	case "PRIVATE KEY":
		return initPriKey(pemKey, block.Bytes)
	case "PUBLIC KEY":
		return initPubKey(pemKey, block.Bytes)
	default:
		return nil, errors.Errorf("unknown block type:%s", block.Type)
	}
}

func initPriKey(pemPriKey []byte, derPriKey []byte) (cryptomgr.Key, error) {
	ecPriKey, err := x509.ParsePKCS8PrivateKey(derPriKey)
	if err != nil {
		return nil, errors.WithMessage(err, "parse ec private key failed")
	}
	ecdsaPriKey, ok := ecPriKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("this is not ecdsa private key")
	}

	pubKeyDer, err := x509.MarshalPKIXPublicKey(&ecdsaPriKey.PublicKey)
	if err != nil {
		return nil, errors.New("get public key from private key failed")
	}

	publicKey, err := x509.ParsePKIXPublicKey(pubKeyDer)
	if err != nil {
		return nil, errors.New("change der public key to ecdsa.PublicKey ")
	}
	ecdsaPubKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("the public key is not ecdsa.PublicKey")
	}

	pubKeyBlock := &pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   pubKeyDer,
	}
	pemPubKey := pem.EncodeToMemory(pubKeyBlock)

	ecKey := &ecdsaKey{
		pemKey:       pemPriKey,
		pemPubKey:    pemPubKey,
		priKey:       ecdsaPriKey,
		isPrivateKey: true,
		pubKey:       ecdsaPubKey,
		curve:        ecdsaPriKey.Curve,
		X:            ecdsaPriKey.X,
		D:            ecdsaPriKey.D,
		Y:            ecdsaPriKey.Y,
	}
	return ecKey, nil
}

func initPubKey(pemPubKey []byte, derPubKey []byte) (cryptomgr.Key, error) {
	ecPubKey, err := x509.ParsePKIXPublicKey(derPubKey)
	if err != nil {
		return nil, errors.WithMessage(err, "parse ec public key failed")
	}
	pubKey, ok := ecPubKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("this is not a ecdsa public key")
	}
	ecdsaPubKey := &ecdsaKey{
		pemKey:       nil,
		pemPubKey:    pemPubKey,
		priKey:       nil,
		isPrivateKey: false,
		pubKey:       pubKey,
		curve:        pubKey.Curve,
		X:            pubKey.X,
		D:            nil,
		Y:            pubKey.Y,
	}

	return ecdsaPubKey, nil
}

// Sign sign the msg with this private key.
func (e *ecdsaKey) Sign(msg []byte, hashAlg string) ([]byte, error) {
	if e.priKey == nil {
		return nil, errors.New("private key is nil")
	}
	digest := Hash(msg)
	bytes, err := e.priKey.Sign(rand.Reader, digest, nil)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Sign sign the msg with this private key.
func (e *ecdsaKey) SignForBatchVerify(msg []byte, hashAlg string) ([]byte, error) {
	if e.priKey == nil {
		return nil, errors.New("private key is nil")
	}
	digest := Hash(msg)
	bytes, err := signBatchMarshal(e.priKey, rand.Reader, digest)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func verifyWithPubKey(pubKey *ecdsa.PublicKey, msg []byte, signature []byte) error {
	var R, S *big.Int
	ecdsaSigRich := &ecdsaSignatureRich{}
	ecdsaSig := &cryptomgr.ECDSASignature{}
	if _, err := asn1.Unmarshal(signature, ecdsaSig); err != nil {
		if _, err := asn1.Unmarshal(signature, ecdsaSigRich); err != nil {
			log.Debugf("asn1.Unmarshal signature failed")
			return errors.WithMessage(err, "unmarshal the signature with ecdsaSignatureRich failed")
		}
		R = ecdsaSigRich.R
		S = ecdsaSigRich.S
	} else {
		R = ecdsaSig.R
		S = ecdsaSig.S
	}
	// Validate ecdsaSig
	if R == nil {
		return errors.New("invalid signature, R is nil")
	}
	if S == nil {
		return errors.New("invalid signature, S is nil")
	}

	if R.Sign() != 1 {
		return errors.New("invalid signature, R not equal to 1")
	}
	if S.Sign() != 1 {
		return errors.New("invalid signature, S not equal to 1")
	}

	hashInfo := Hash(msg)

	if ecdsa.Verify(pubKey, hashInfo, ecdsaSig.R, ecdsaSig.S) {
		return nil
	}

	return errors.New("ecdsa verify failed")
}

// Verify verify the signature of this msg with this public key.
func (e *ecdsaKey) Verify(msg []byte, signature []byte, hashAlg string) error {
	pubKey := e.pubKey
	return verifyWithPubKey(pubKey, msg, signature)
}

// GetPemBytes get pem bytes of this key.
func (e *ecdsaKey) GetPemBytes() []byte {
	if e.isPrivateKey {
		return e.pemKey
	}
	return e.pemPubKey
}

// GetDerBytes get der bytes of this key
func (e *ecdsaKey) GetDerBytes() ([]byte, error) {
	if e.isPrivateKey {
		priKeyDer, err := x509.MarshalPKCS8PrivateKey(e.priKey)
		if err != nil {
			return nil, err
		}
		return priKeyDer, nil
	}
	pubKeyDer, err := x509.MarshalPKIXPublicKey(e.pubKey)
	if err != nil {
		return nil, err
	}
	return pubKeyDer, nil
}

// IsSymmetric judge whether this is a symmetric key or not.
func (e *ecdsaKey) IsSymmetric() bool {
	return false
}

// IsPrivate is this a private key or not.
func (e *ecdsaKey) IsPrivate() bool {
	return e.isPrivateKey
}

// GetPublicKey get public key of this key.
// If this is a private key, then export public key from private key.
// If this is a public key, then return this public key directly.
func (e *ecdsaKey) GetPublicKey() cryptomgr.Key {
	key, err := GetKeyFromPem(e.pemPubKey)
	if err != nil {
		return nil
	}
	return key
}

// GetPrivateKey get private key of this key.
// If this is a private key, then return this private key directly.
// If this is a public key, then return nil.
func (e *ecdsaKey) GetPrivateKey() (priKey interface{}, err error) {
	if e.isPrivateKey {
		return e.priKey, nil
	}
	return nil, errors.New("this is not a private key")
}
