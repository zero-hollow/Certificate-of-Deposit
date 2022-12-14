/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package client

import (
	"time"

	"git.huawei.com/huaweichain/sdk/utils"
	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/sdk/config"
	"git.huawei.com/huaweichain/sdk/node"
	"git.huawei.com/huaweichain/sdk/rawmessage"
)

// RelayerClient is the definition of relayer client.
type RelayerClient struct {
	CrossChainRawMessage *rawmessage.CrossChainRawMessage
	Relayers             map[string]*node.Relayer
}

// NewRelayerClient is used to create an instance of relayer client by specifying the config file path.
func NewRelayerClient(configPath string, decrypts ...func(bytes []byte) ([]byte, error)) (*RelayerClient, error) {
	config, err := config.NewRelayerConfig(configPath)
	if err != nil {
		return nil, errors.WithMessage(err, "load config file error")
	}

	decrypt := func(bytes []byte) ([]byte, error) {
		return bytes, nil
	}
	if len(decrypts) > 0 {
		decrypt = decrypts[0]
	}

	crypto, err := getCrypto(config.Client, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get crypto error")
	}
	msgBuilder, err := rawmessage.NewMsgBuilderImpl(crypto)
	if err != nil {
		return nil, errors.WithMessage(err, "new MsgBuilderImpl error")
	}
	crossChainRawMessage := rawmessage.NewCrossChainRawMessage(msgBuilder)

	relayers, err := getRelayers(config, decrypt)
	if err != nil {
		return nil, errors.WithMessage(err, "get relayers error")
	}

	return &RelayerClient{CrossChainRawMessage: crossChainRawMessage, Relayers: relayers}, nil
}

func getRelayers(c *config.RelayerConfig,
	decrypt func(bytes []byte) ([]byte, error)) (map[string]*node.Relayer, error) {
	relayers := make(map[string]*node.Relayer)
	for _, n := range c.Relayers {
		var relayer *node.Relayer
		var tls *node.TLS
		var err error
		tls, err = node.ConvertTLS(c.Client.TLS, decrypt)
		if err != nil {
			return nil, errors.WithMessage(err, "convert node tls from config tls error")
		}
		relayer, err = node.NewRelayer(n, c.Client.TLS.Enable, tls)
		if err != nil {
			return nil, errors.WithMessage(err, "new relayer error")
		}
		relayers[n.ID] = relayer
	}
	return relayers, nil
}

// SetTimeout set client timeout (/s)
func (r *RelayerClient) SetTimeout(seconds int64) {
	utils.SetTimeout(time.Duration(seconds))
}
