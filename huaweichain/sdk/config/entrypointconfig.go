/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"regexp"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto/common"
)

// NewEntrypointConfig is used to new an instance of EntrypointConfig.
func NewEntrypointConfig(entrypointConfigPath string) (*EntrypointConfig, error) {
	v, errMsg := parseCfg(entrypointConfigPath)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "parse config file error")
	}

	entrypointConfig := &EntrypointConfig{}

	errMsg = unmarshal(v, &entrypointConfig)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "error unmarshaling config into struct")
	}
	return entrypointConfig, nil
}

// ReadEntryPointFromFile is used to read entrypoint from file.
func ReadEntryPointFromFile(entrypointConfigPath string) (*common.Entrypoint, error) {
	entrypointConfig, err := NewEntrypointConfig(entrypointConfigPath)
	if err != nil {
		return nil, errors.WithMessage(err, "new entry point config error")
	}
	entrypoint := &common.Entrypoint{ZoneId: entrypointConfig.Zone, Coordinator: entrypointConfig.Coordinator,
		InitialMaster: entrypointConfig.InitialMaster}
	if entrypointConfig.InitialMaster {
		entrypoint.Coordinator = true
		entrypoint.Linkers = entrypointConfig.Linkers
	} else {
		entrypoint.Seeds = entrypointConfig.Seeds
	}
	return entrypoint, nil
}

// EntrypointConfig is the definition of entry point config.
type EntrypointConfig struct {
	Zone          string
	Coordinator   bool
	InitialMaster bool `mapstructure:"initial_master"`
	Seeds         []string
	Linkers       []string
}

// CheckEntryPoint returns error when entrypoint information has invalid field.
func CheckEntryPoint(entrypoint *common.Entrypoint) error {
	if err := verifyHostPort(entrypoint.Linkers); err != nil {
		return err
	}
	if err := verifyHostPort(entrypoint.Seeds); err != nil {
		return err
	}
	if err := CheckZoneID(entrypoint.ZoneId); err != nil {
		return err
	}
	if err := verifyEntry(entrypoint); err != nil {
		return err
	}
	return nil
}

const hostPortReg = `^(.*)(:)([0-9]+)$`

func verifyHostPort(s []string) error {
	r, err := regexp.Compile(hostPortReg)
	if err != nil {
		return errors.WithMessagef(err, "regexp compile error: %v", hostPortReg)
	}
	for i := 0; i < len(s); i++ {
		if !r.Match([]byte(s[i])) {
			return errors.Errorf("port error: %v", s[i])
		}
	}
	return nil
}

func verifyEntry(e *common.Entrypoint) error {
	// Master of sub zone should have linkers
	if e.ZoneId != rootZone && e.InitialMaster && len(e.Linkers) == 0 {
		return errors.Errorf("node is Master of subZone[%v], should have at least one linker", e.ZoneId)
	}
	// peer should have seeds
	if !e.Coordinator && len(e.Seeds) == 0 {
		return errors.New("peer should have at least one seed")
	}
	return nil
}
