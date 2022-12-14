/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"os"
	"sort"
	"testing"
)

func Test_DomainConfig(t *testing.T) {
	configFile := "../../wienerchain-java-sdk/src/test/resources/domain_conf.yaml"
	_, errMsg := os.Stat(configFile)
	if errMsg != nil {
		t.Errorf("CfgFilePath %s does not exist", configFile)
		return
	}

	cfg, err := NewDomainConfig(configFile)
	if err != nil {
		t.Errorf("new domain config error: %v\n", err)
		return
	}
	t.Logf("[FilePath] %s\n", configFile)
	t.Logf("domain config: %+v\n", cfg)
}

func Test_RemovePrefix(t *testing.T) {
	match := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}

		if (a == nil) != (b == nil) {
			return false
		}
		sort.Strings(a)
		sort.Strings(b)

		b = b[:len(a)]
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}

		return true
	}

	ds := map[string]struct{}{
		"/a":   {},
		"/a/b": {},
		"/c":   {},
	}
	expect := []string{"/a/b", "/c"}
	res := removePrefix(ds)
	if !match(expect, res) {
		t.Errorf("check fail, expect: %v, actual: %v", expect, res)
	}
}
