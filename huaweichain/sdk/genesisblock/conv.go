/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package genesisblock

import (
	"git.huawei.com/huaweichain/proto/consensus"
	"git.huawei.com/huaweichain/sdk/config"
)

type consenter config.Consenter

func (c *consenter) conv(reeCert []byte, teeCert []byte) *consensus.Consenter {
	return &consensus.Consenter{
		Name:     c.ID,
		Org:      c.Org,
		Host:     c.Host,
		Port:     c.Port,
		ReqPort:  c.ReqPort,
		RestPort: c.RestPort,
		ReeCert:  reeCert,
		TeeCert:  teeCert,
	}
}
