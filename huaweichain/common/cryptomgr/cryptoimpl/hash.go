/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

package cryptoimpl

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	// Register some standard stuff
	"strings"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/common/cryptomgr/ecdsaalg"
	"git.huawei.com/huaweichain/common/cryptomgr/gmalg"
)

// HashSha256 get sha256 hash for msg.
func HashSha256(msg []byte) []byte {
	// todo unit test
	return Hash(msg, cryptomgr.Sha256)
}

// Hash get hash for msg with certain algorithm.
func Hash(msg []byte, algorithm string) []byte {
	switch algorithm {
	case cryptomgr.Sha256, "":
		return ecdsaalg.Hash(msg)
	case cryptomgr.Sm3:
		return gmalg.Hash(msg)
	default:
		return nil
	}
}

// GetCfgHashAlg get the hash option from config
func GetCfgHashAlg(signAlg string) string {
	alg := strings.ToLower(signAlg)
	switch alg {
	case cryptomgr.Sm2WithSm3:
		return cryptomgr.Sm3
	case cryptomgr.EcdsaWithSha256:
		return cryptomgr.Sha256
	case cryptomgr.Ed25519:
		return cryptomgr.Sha256
	}
	return cryptomgr.Sha256
}

// Sha256ForFile sha256 for file
func Sha256ForFile(filePath string) ([]byte, error) {
	h := sha256.New()
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, errors.WithMessage(err, "open file err")
	}
	defer func() {
		if err = file.Close(); err != nil {
			return
		}
	}()
	_, err = io.Copy(h, file)
	if err != nil {
		return nil, errors.WithMessage(err, "io copy err")
	}
	hashInfo := h.Sum(nil)
	return hashInfo, nil
}
