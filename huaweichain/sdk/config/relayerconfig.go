/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package config

import (
	"github.com/pkg/errors"
)

// RelayerConfig is the definition of config.
type RelayerConfig struct {
	Client   *Client
	Relayers []*Node
}

// NewRelayerConfig is used to load config file and parse config file to ClientConfig struct.
func NewRelayerConfig(configPath string) (*RelayerConfig, error) {
	v, errMsg := parseCfg(configPath)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "parse config file error")
	}

	relayerConfig := &RelayerConfig{}
	errMsg = unmarshalRelayerConfig(v, &relayerConfig)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "error unmarshaling config into struct")
	}

	return relayerConfig, nil
}
