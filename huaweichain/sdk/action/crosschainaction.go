/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package action

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/connectivity"

	"git.huawei.com/huaweichain/proto/relayer"
	"git.huawei.com/huaweichain/sdk/utils"
)

// CrossChainAction is the definition of cross chain action.
type CrossChainAction struct {
	action
	relayerClient relayer.RelayerClient
}

// NewCrossChainAction is used to create an instance of cross chain action.
func NewCrossChainAction(config *Config) *CrossChainAction {
	action := &CrossChainAction{action: action{config: config}}
	return action
}

// RegisterConfig is used to register config.
func (action *CrossChainAction) RegisterConfig(rawMsg *relayer.RawMessage) (*relayer.RawMessage, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.RegisterConfig(ctx, rawMsg)
}

// QueryConfig is used to send query cross chain config.
func (action *CrossChainAction) QueryConfig(rawMsg *relayer.RawMessage) (*relayer.RawMessage, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.QueryConfig(ctx, rawMsg)
}

func (action *CrossChainAction) getClient() (relayer.RelayerClient, error) {
	if action.conn == nil || action.conn.GetState() == connectivity.Shutdown {
		if err := action.newClient(); err != nil {
			return nil, errors.WithMessage(err, "new client error")
		}
	}
	return action.relayerClient, nil
}

func (action *CrossChainAction) newClient() error {
	cc, err := action.newClientConn()
	if err != nil {
		return errors.WithMessage(err, "get client connection error")
	}
	action.conn = cc
	action.relayerClient = relayer.NewRelayerClient(action.conn)
	return nil
}
