/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package cryptomgr crypto management.
package cryptomgr

import (
	"math/big"
	"time"
)

// 对于签名，我们需要用wnode节点本地的私钥进行签名

// 对于验签，如果是链外请求（比如创建链，查询链），需要用wnode节点本地的CA证书验证消息中证书的可靠性，保证请求者是本节点所属组织的管理员

// 对于验签，如果是链内请求，需要用链上的各个组织内的CA验证消息中证书的可靠性

// Symmetric key type
const (
	AES256 string = "aes_256"
	SM4    string = "sm4"
)

const (
	// Sm2WithSm3  gm sign algorithm
	Sm2WithSm3 = "sm2_with_sm3"
	// EcdsaWithSha256 ecdsa sign algorithm
	EcdsaWithSha256 = "ecdsa_with_sha256"
	// Ed25519 ed25519 sign algorithm
	Ed25519 = "ed25519"
	// Sm3 sm3 hash algorithm
	Sm3 = "sm3"
	// Sha256 sha256 hash algorithm
	Sha256 = "sha256"

	// Sm4 symmetric encryption algorithm
	Sm4 = "sm4"

	// Aes128 symmetric encryption algorithm
	Aes128 = "aes128"
)

// ECDSASignature ecdsa signature struct with R and S.
type ECDSASignature struct {
	R, S *big.Int
}

// Key describe the key
type Key interface {
	// GetType() string
	GetPemBytes() []byte

	GetDerBytes() ([]byte, error)

	IsSymmetric() bool

	IsPrivate() bool

	GetPublicKey() Key

	Signature

	SignForBatchVerify

	Verification
}

// Signature sign interface.
type Signature interface {
	Sign(msg []byte, hashAlg string) ([]byte, error)
}

// SignForBatchVerify sign  for batch verify interface.
type SignForBatchVerify interface {
	SignForBatchVerify(msg []byte, hashAlg string) ([]byte, error)
}

// Verification verify interface.
type Verification interface {
	Verify(msg []byte, signature []byte, hashAlg string) error
}

// SymmetricKey symmetric key interface
type SymmetricKey interface {
	Encrypt(plainTxt []byte) ([]byte, error)
	Decrypt(cipherTxtAndIv []byte) ([]byte, error)
	GetKeyBase64Str() string
	GetKeyBytes() []byte
}

// CertBasicInfo the certificate basic information
type CertBasicInfo struct {
	Organization     string
	OrganizationUnit string
	Country          string
	Province         string
	Locality         string
	CommonName       string
	ValidationYears  uint
}

// Cert describe the certificate
type Cert interface {
	GetCommonName() string
	GetDerPublicKey() ([]byte, error)
	GetExpireTime() time.Time
	GetOrganizationalUnit() []string
	GetOrganization() []string
	Verification
	CheckValidation(rootcerts []Cert) (string, error)
	GetPemCertBytes() []byte
	GetFingerPrint() string
	GetSerialNumber() string
	VerifyBatch(msg [][]byte, signature [][]byte, hashOpt string) error
}
