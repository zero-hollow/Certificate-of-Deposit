/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

package ecdsaalg

import "crypto/sha256"

// Hash for sha256
func Hash(msg []byte) []byte {
	h := sha256.New()
	_, err := h.Write(msg)
	if err != nil {
		return nil
	}
	hashInfo := h.Sum(nil)
	return hashInfo
}
