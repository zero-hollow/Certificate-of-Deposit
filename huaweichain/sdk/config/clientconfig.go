/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package config provide the implementation of config module, which is used to parse config file.
package config

import (
	"os"

	"github.com/pkg/errors"

	"github.com/spf13/viper"
)

// ClientConfig is the definition of config.
type ClientConfig struct {
	Client *Client
	Nodes  []*Node
}

// Client is the definition of Client.
type Client struct {
	Type     string
	Identity Identity
	TLS      TLS
}

// Identity is the definition fo Identity.
type Identity struct {
	CertPath string
	KeyPath  string
}

// TLS is the definition of TLS.
type TLS struct {
	Enable   bool
	CertPath string
	KeyPath  string
	RootPath []string
}

// Node is the definition of Node.
type Node struct {
	ID           string
	HostOverride string
	Host         string
	Port         int
}

// NewClientConfig is used to load config file and parse config file to ClientConfig struct.
func NewClientConfig(configPath string) (*ClientConfig, error) {
	v, errMsg := parseCfg(configPath)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "parse config file error")
	}

	gatewayConfig := &ClientConfig{}
	errMsg = unmarshalClientConfig(v, &gatewayConfig)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "error unmarshaling config into struct")
	}

	return gatewayConfig, nil
}

func parseCfg(configPath string) (*viper.Viper, error) {
	_, errMsg := os.Stat(configPath)
	if errMsg != nil {
		return nil, errors.WithMessagef(errMsg, "CfgFilePath does not exist: %s", configPath)
	}
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}
