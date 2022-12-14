/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package config

import (
	"os"
	"testing"
)

func Test_Client_Config(t *testing.T) {
	configFile := "../../wienerchain-java-sdk/conf/client-config.yaml.template"
	_, errMsg := os.Stat(configFile)
	if errMsg != nil {
		t.Errorf("CfgFilePath %s does not exist", configFile)
		return
	}

	cfg, err := NewClientConfig(configFile)
	if err != nil {
		t.Errorf("new genesis config error: %v\n", err)
		return
	}

	t.Logf("client: %v\n", cfg.Client)
	for i, node := range cfg.Nodes {
		t.Logf("node %v:  %v\n", i, node)
	}
}
