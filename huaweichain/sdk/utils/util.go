/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

// Package utils provide the common utils.
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	mrand "math/rand"
	"time"

	"git.huawei.com/huaweichain/common/cryptomgr/gmalg"
)

const (
	// Seed is the random factor for generating tx id.
	Seed = 10000
)

var timeout time.Duration = 60

// HashSha256 computes hash by sha256 algorithm.
func HashSha256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// HashSM3 computes hash by sm3 algorithm.
func HashSM3(data []byte) []byte {
	return gmalg.Hash(data)
}

// GenerateTimestamp is used to generate timestamp.
func GenerateTimestamp() uint64 {
	return uint64(time.Now().UTC().UnixNano())
}

// GenerateNonce generates a random number.
func GenerateNonce() uint64 {
	n, err := rand.Int(rand.Reader, big.NewInt(Seed))
	if err != nil {
		mrand.Seed(time.Now().UnixNano())
		return mrand.Uint64() // nolint:gosec // this is just a backup operation for secure random
	}
	return n.Uint64()
}

// SetTimeout set client timeout (/s)
func SetTimeout(seconds time.Duration) {
	timeout = seconds
}

// GetTimeout get client timeout (/s)
func GetTimeout() time.Duration {
	return timeout
}
