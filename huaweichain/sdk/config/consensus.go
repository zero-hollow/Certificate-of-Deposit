/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	goerrors "errors"
	"io/ioutil"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/logger"
	"git.huawei.com/huaweichain/proto/consensus"
)

var log = logger.GetModuleLogger("go sdk", "config")

// DecryptFunc is the type for Decrypt Func
type DecryptFunc func(bytes []byte) ([]byte, error)

var (
	errConfigNodeMismatch = goerrors.New("consensus node mismatch config file")
	errConfigItemEmpty    = goerrors.New("consensus config has empty config item")
)

// NewConsensusConfig get consenter config from file
func NewConsensusConfig(nodeName string, path string, decrypt DecryptFunc) (*consensus.Consenter, error) {
	v, err := parseCfg(path)
	if err != nil {
		return nil, errors.WithMessage(err, "parse consensus config file error")
	}
	var cf ConsenterInfo
	err = unmarshal(v, &cf)
	if err != nil {
		return nil, errors.WithMessage(err, "error unmarshaling config into struct")
	}
	if err = checkConsensusPara(nodeName, &cf); err != nil {
		return nil, err
	}
	cs, err := cf.decrypt(decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "decrypt config cert fail")
	}
	return cs, nil
}

func checkConsensusPara(nodeName string, c *ConsenterInfo) error {
	if c.Name != nodeName {
		return errConfigNodeMismatch
	}
	if len(c.Org) == 0 || len(c.Addr) == 0 || len(c.Org) == 0 ||
		len(c.TeeCert) == 0 || len(c.ReeCert) == 0 {
		return errConfigItemEmpty
	}
	return nil
}

// ConsenterInfo contains all information of consensus node.
type ConsenterInfo struct {
	Name     string
	Org      string
	Addr     string
	Port     uint64
	ReqPort  uint32 `mapstructure:"req_port"`
	RestPort uint32 `mapstructure:"rest_port"`
	ReeCert  string
	TeeCert  string
}

func (c *ConsenterInfo) decrypt(decrypt DecryptFunc) (*consensus.Consenter, error) {
	reeCert, err := ioutil.ReadFile(c.ReeCert)
	if err != nil {
		return nil, errors.WithMessage(err, "load ree certificate error")
	}
	teeCert, err := ioutil.ReadFile(c.TeeCert)
	if err != nil {
		log.Warnf("failed to read tee cert from file because %+v, ignore and replaced by ree cert.",
			errors.WithMessage(err, "load tee certificate error"))
		teeCert = reeCert
	}
	reeCert, err = decrypt(reeCert)
	if err != nil {
		return nil, errors.WithMessage(err, "decrypt message error")
	}
	teeCert, err = decrypt(teeCert)
	if err != nil {
		return nil, errors.WithMessage(err, "decrypt message error")
	}

	return c.conv(reeCert, teeCert), nil
}

func (c *ConsenterInfo) conv(reeCert []byte, teeCert []byte) *consensus.Consenter {
	return &consensus.Consenter{
		Name:     c.Name,
		Org:      c.Org,
		Host:     c.Addr,
		Port:     c.Port,
		ReqPort:  c.ReqPort,
		RestPort: c.RestPort,
		ReeCert:  reeCert,
		TeeCert:  teeCert,
	}
}
