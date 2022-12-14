/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package config

import (
	"testing"
)

func Test_Genesis_Config(t *testing.T) {
	configFile := []string{
		"../../wienerchain-java-sdk/conf/flic-config.yaml.template",
		"../../wienerchain-java-sdk/conf/raft-config.yaml.template",
		"../../wienerchain-java-sdk/conf/solo-config.yaml.template",
		"../../wienerchain-java-sdk/conf/hotstuff-config.yaml.template",
		"../../wienerchain-java-sdk/src/test/resources/net_config.yaml",
		"../../wienerchain-java-sdk/conf/default-config.template.yaml",
	}
	for _, file := range configFile {
		cfg, err := NewGenesisConfig(file)
		if err != nil {
			t.Errorf("genesis config error: %v\n", err)
			return
		}
		t.Logf("[FilePath] %s\n", configFile)
		t.Logf("genesis config: %+v\n", cfg.GenesisBlock)
	}
}

func Test_Genesis_Config_WithNetworkConfig(t *testing.T) {
	configFile := []string{
		"../../wienerchain-java-sdk/src/test/resources/net_config.yaml",
		"../../wienerchain-java-sdk/conf/default-config.template.yaml",
	}
	for _, file := range configFile {
		cfg, err := NewGenesisConfig(file)
		if err != nil {
			t.Errorf("genesis config error: %v\n", err)
			return
		}
		t.Logf("[FilePath] %s\n", configFile)
		t.Logf("genesis config: %+v\n", cfg.GenesisBlock)
		t.Logf("net: %v\n", cfg.GenesisBlock.Net)
		t.Logf("cons_zone: %v\n", cfg.GenesisBlock.Net.ConsZone)
		for i, zone := range cfg.GenesisBlock.Net.Zones {
			t.Logf("zone %v:  %v\n", i, zone)
		}
	}
}
