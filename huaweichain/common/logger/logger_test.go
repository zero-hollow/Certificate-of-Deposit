/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package logger

import (
	"testing"
)

func TestGetLogger(t *testing.T) {
	config := &LogCfg{
		Type:         "pretty",
		Level:        "info",
		ModuleLevels: make(map[string]string),
	}
	config.ModuleLevels["ledger"] = "error"

	logger := GetDefaultLogger()
	mlogger := GetModuleLogger("ledger", "test")
	Init(config)

	logger.Info("default test")
	mlogger.Info("module test")
	mlogger.Errorf("module error test: %s", "ledger")
}
