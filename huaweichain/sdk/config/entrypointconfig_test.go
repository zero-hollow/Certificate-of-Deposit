/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"os"
	"testing"
)

func Test_EntrypointConfig(t *testing.T) {
	configFile := "../../wienerchain-java-sdk/src/test/resources/entrypoint.yaml"
	_, errMsg := os.Stat(configFile)
	if errMsg != nil {
		t.Errorf("CfgFilePath %s does not exist", configFile)
		return
	}

	cfg, err := NewEntrypointConfig(configFile)
	if err != nil {
		t.Errorf("new entrypoint config error: %v\n", err)
		return
	}
	t.Logf("[FilePath] %s\n", configFile)
	t.Logf("entrypoint config: %+v\n", cfg)
}
