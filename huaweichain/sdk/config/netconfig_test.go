/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"os"
	"testing"
)

func Test_Networking(t *testing.T) {
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
	t.Logf("fan: %v\n", fan)
}

func TestVerifyDomainCfgRegx(t *testing.T) {
	type args struct {
		domainPath string
	}

	tests := []struct {
		name   string
		args   args
		except bool
	}{
		{name: "empty", args: args{domainPath: ""}, except: false},
		{name: "root domain", args: args{domainPath: "/"}, except: true},
		{name: "root domain-all", args: args{domainPath: "/*"}, except: false},
		{name: "root domain-all recursive", args: args{domainPath: "/**"}, except: false},
		{name: "domainPath", args: args{domainPath: "/domain_a"}, except: true},
		{name: "upper case", args: args{domainPath: "/domain_A"}, except: true},
		{name: "sub domain", args: args{domainPath: "/domain_a/domain_b"}, except: true},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c"}, except: true},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c/*"}, except: false},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c/**"}, except: false},
		{name: "sub domain end with/", args: args{domainPath: "/domain_a/domain_b/"}, except: false},
		{name: "split :", args: args{domainPath: "/domain_a:domain_b"}, except: false},
		{name: "split .", args: args{domainPath: "/domain_a.domain_b"}, except: false},
		{name: "split //", args: args{domainPath: "/domain_a//domain_b"}, except: false},
		{name: "split -", args: args{domainPath: "/domain_a-domain_b"}, except: false},
		{name: "split *", args: args{domainPath: "/domain_a*domain_b"}, except: false},
		{name: "no sub domain", args: args{domainPath: "/domain_a/"}, except: false},
	}

	for _, tt := range tests {
		testFunc := func(t *testing.T) {
			res := CheckDomainPath([]string{tt.args.domainPath})
			if (res == nil) != tt.except {
				t.Errorf("[test] %v, args: %v, exp: %v, actual: %v", tt.name, tt.args, tt.except, res)
			}
		}
		t.Run(tt.name, testFunc)
	}
}

func TestVerifyTxDomainRegx(t *testing.T) {
	type args struct {
		domainPath string
	}

	tests := []struct {
		name   string
		args   args
		except bool
	}{
		{name: "empty", args: args{domainPath: ""}, except: false},
		{name: "root domain", args: args{domainPath: "/"}, except: true},
		{name: "root domain-all", args: args{domainPath: "/*"}, except: true},
		{name: "root domain-all recursive", args: args{domainPath: "/**"}, except: true},
		{name: "domainPath", args: args{domainPath: "/domain_a"}, except: true},
		{name: "upper case", args: args{domainPath: "/domain_A"}, except: true},
		{name: "sub domain", args: args{domainPath: "/domain_a/domain_b"}, except: true},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c"}, except: true},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c/*"}, except: true},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c/**"}, except: true},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c/*/*"}, except: false},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c/*/**"}, except: false},
		{name: "sub domain 3", args: args{domainPath: "/domain_a/domain_b/domain_c/**/**"}, except: false},
		{name: "sub domain end with/", args: args{domainPath: "/domain_a/domain_b/"}, except: false},
		{name: "split :", args: args{domainPath: "/domain_a:domain_b"}, except: false},
		{name: "split .", args: args{domainPath: "/domain_a.domain_b"}, except: false},
		{name: "split //", args: args{domainPath: "/domain_a//domain_b"}, except: false},
		{name: "split -", args: args{domainPath: "/domain_a-domain_b"}, except: false},
		{name: "split *", args: args{domainPath: "/domain_a*domain_b"}, except: false},
		{name: "no sub domain", args: args{domainPath: "/domain_a/"}, except: false},
	}

	for _, tt := range tests {
		testFunc := func(t *testing.T) {
			res := CheckTxDomainPath([]string{tt.args.domainPath})
			if (res == nil) != tt.except {
				t.Errorf("[test] %v, args: %v, exp: %v, actual: %v", tt.name, tt.args, tt.except, res)
			}
		}
		t.Run(tt.name, testFunc)
	}
}
