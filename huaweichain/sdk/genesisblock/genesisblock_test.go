/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package genesisblock

import (
	"testing"

	"git.huawei.com/huaweichain/sdk/config"
)

func Test_GetChainConfig(t *testing.T) {
	configFile := []string{
		"../../wienerchain-java-sdk/src/test/resources/net_config.yaml",
		"../../wienerchain-java-sdk/conf/default-config.template.yaml",
	}
	for _, file := range configFile {
		cfg, err := config.NewGenesisConfig(file)
		if err != nil {
			t.Errorf("new genesis config error: %v", err)
			return
		}
		nwc, err := getNetWorkConfig(cfg)
		if err != nil {
			t.Errorf("get chain config error: %v", err)
			return
		}
		t.Logf("test file path: %v\n", file)
		t.Logf("zones: %v\n", nwc.Zones)
		t.Logf("Domains: %v\n", nwc.Domains)
		t.Logf("zone_1::zone_4 Domains: %v\n", nwc.Zones[3].Domains)
		t.Logf("ZoneTemplate: %v\n", nwc.ZoneTemplate)
		t.Logf("ZoneTemplate alloc type: %v\n", nwc.ZoneTemplate.Alloc.Type)
		t.Logf("ConsZone: %v\n", nwc.ConsZone)
		t.Logf("ConsZone alloc type: %v\n", nwc.ConsZone.Alloc.Type)
		t.Logf("Sync: %v\n", nwc.Sync)
		t.Logf("AutoGenZone: %v\n", nwc.AutoGenZone)
		t.Logf("\n")
	}
}
