/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package gmalg

import (
	"strings"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/gmssl"
)

const sm3HashFlag = "SM3"

// Hash for sm3
func Hash(msg []byte) []byte {
	sm3ctx, err := gmssl.NewDigestContext(strings.ToUpper(cryptomgr.Sm3))
	if err != nil {
		log.Errorf("New sm3 digest context failed:%s", err.Error())
		return nil
	}
	err = sm3ctx.Reset()
	if err != nil {
		log.Errorf("Sm3 context reset failed:%s", err.Error())
		return nil
	}
	err = sm3ctx.Update(msg)
	if err != nil {
		log.Errorf("Sm3 context update message failed:%s", err.Error())
		return nil
	}
	sm3Hash, err := sm3ctx.Final()
	if err != nil {
		log.Errorf("Sm3 ctx final failed:%s", err.Error())
		return nil
	}
	return sm3Hash
}
