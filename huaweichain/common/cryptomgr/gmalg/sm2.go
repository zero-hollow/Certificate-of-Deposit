/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package gmalg

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
)

type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

// pkcs1PrivateKey is a structure which mirrors the PKCS #1 ASN.1 for an RSA private key.
type pkcs1PrivateKey struct {
	Version int
	N       *big.Int
	E       int
	D       *big.Int
	P       *big.Int
	Q       *big.Int
	// We ignore these values, if present, because rsa will calculate them.
	Dp   *big.Int `asn1:"optional"`
	Dq   *big.Int `asn1:"optional"`
	Qinv *big.Int `asn1:"optional"`

	AdditionalPrimes []pkcs1AdditionalRSAPrime `asn1:"optional,omitempty"`
}

type pkcs1AdditionalRSAPrime struct {
	Prime *big.Int

	// We ignore these values because rsa will calculate them.
	Exp   *big.Int
	Coeff *big.Int
}

type pkcs8 struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
	// optional attributes omitted.
}

var (
	oidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	oidsm2p256v1      = asn1.ObjectIdentifier{1, 2, 156, 10197, 1, 301}
)

const ecPrivKeyVersion = 1

// ExtractPrivateKeyFromPKCS8 parses big unsigned integer bytes from PKCS8 private key for SM2.
func ExtractPrivateKeyFromPKCS8(der []byte) ([]byte, error) {
	var privKey pkcs8
	if _, err := asn1.Unmarshal(der, &privKey); err != nil {
		if _, err = asn1.Unmarshal(der, &ecPrivateKey{}); err == nil {
			return nil, errors.New("x509: failed to parse private key (use ParseECPrivateKey instead for " +
				"this key format)")
		}
		if _, err = asn1.Unmarshal(der, &pkcs1PrivateKey{}); err == nil {
			return nil, errors.New("x509: failed to parse private key (use ParsePKCS1PrivateKey instead for " +
				"this key format)")
		}
		return nil, err
	}
	if !privKey.Algo.Algorithm.Equal(oidPublicKeyECDSA) {
		return nil, errors.Errorf("invalid oid %s, not equals %s",
			privKey.Algo.Algorithm.String(), oidPublicKeyECDSA.String())
	}

	namedCurveOID := new(asn1.ObjectIdentifier)
	if _, err := asn1.Unmarshal(privKey.Algo.Parameters.FullBytes, namedCurveOID); err != nil {
		namedCurveOID = nil
	}
	key, err := parseECPrivateKey(namedCurveOID, privKey.PrivateKey)
	if err != nil {
		return nil, errors.New("x509: failed to parse EC private key embedded in PKCS#8: " + err.Error())
	}
	return key, nil
}

func parseECPrivateKey(namedCurveOID *asn1.ObjectIdentifier, der []byte) ([]byte, error) {
	if !namedCurveOID.Equal(oidsm2p256v1) {
		return nil, errors.Errorf("invalid curve oid %s, not equals %s", namedCurveOID, oidsm2p256v1)
	}
	var privKey ecPrivateKey
	if _, err := asn1.Unmarshal(der, &privKey); err != nil {
		if _, err = asn1.Unmarshal(der, &pkcs8{}); err == nil {
			return nil, errors.New("x509: failed to parse private key (use ParsePKCS8PrivateKey instead for " +
				"this key format)")
		}
		if _, err = asn1.Unmarshal(der, &pkcs1PrivateKey{}); err == nil {
			return nil, errors.New("x509: failed to parse private key (use ParsePKCS1PrivateKey instead for " +
				"this key format)")
		}
		return nil, errors.New("x509: failed to parse EC private key: " + err.Error())
	}
	if privKey.Version != ecPrivKeyVersion {
		return nil, fmt.Errorf("x509: unknown EC private key version %d", privKey.Version)
	}
	return privKey.PrivateKey, nil
}
