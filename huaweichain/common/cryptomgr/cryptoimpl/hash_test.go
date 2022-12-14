/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

package cryptoimpl

import (
	"testing"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

func TestHash(t *testing.T) {
	alg := cryptomgr.Sha256
	msg := []byte("hello")
	hashRes := Hash(msg, alg)
	if hashRes == nil {
		t.Fatalf("calc sha256 hash failed")
	}

	alg = cryptomgr.Sm3
	hashRes = Hash(msg, alg)
	if hashRes == nil {
		t.Fatalf("calc sm3 hash failed")
	}

	hashRes = Hash(msg, "test")
	if hashRes != nil {
		t.Fatalf("calc hash with unkown algrithm should be failed")
	}
}
