/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"os"
	"testing"
)

func Test_ZoneList(t *testing.T) {
	configFile := "../../wienerchain-java-sdk/src/test/resources/zones.yaml"
	_, errMsg := os.Stat(configFile)
	if errMsg != nil {
		t.Errorf("CfgFilePath %s does not exist", configFile)
		return
	}

	zoneList, err := NewZoneList(configFile)
	if err != nil {
		t.Errorf("new zone list error: %v\n", err)
		return
	}
	t.Logf("[FilePath] %s\n", configFile)
	t.Logf("zone: %+v\n", zoneList)
	m := zoneList.Zones[0]
	value, ok, err := getValue(m, allocConfFanKey)
	if err != nil {
		t.Errorf("get value error: %v\n", err)
	}
	if !ok {
		t.Errorf("not contains feild: %v\n", allocConfFanKey)
		return
	}
	fan, ok := value.(int)
	if !ok {
		t.Errorf("type assert error: %v\n", value)
	}
	if fan != 3 {
		t.Errorf("expected fan: 3, but got: %v\n", fan)
	}
}

func Test_BuildProperties(t *testing.T) {
	configFile := "../../wienerchain-java-sdk/src/test/resources/zones.yaml"
	_, errMsg := os.Stat(configFile)
	if errMsg != nil {
		t.Errorf("CfgFilePath %s does not exist", configFile)
		return
	}

	zoneList, err := NewZoneList(configFile)
	if err != nil {
		t.Errorf("new zone list error: %v\n", err)
		return
	}
	t.Logf("[FilePath] %s\n", configFile)
	for _, zone := range zoneList.Zones {
		ps, err := BuildProperties(zone)
		if err != nil {
			t.Errorf("build properties error: %v", err)
			return
		}
		t.Logf("id: %v, properties: %v\n", zone["id"], ps)
		for _, p := range ps {
			if p.K == domainKey {
				t.Logf("property: %v, properties: %v\n", p.K, p.GetStrArr().Ss)
			}
		}
	}
}
