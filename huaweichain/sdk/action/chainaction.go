/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package action provide the implementation of action module.
package action

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/connectivity"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"git.huawei.com/huaweichain/sdk/utils"
)

// ChainAction is the action for chain operations.
type ChainAction struct {
	action
	client nodeservice.ChainManagerClient
}

// NewChainAction is used to create chain action instance with a action config.
func NewChainAction(config *Config) *ChainAction {
	action := &ChainAction{action: action{config: config}}
	return action
}

// JoinChain is used to send join chain request raw message by grpc.
func (action *ChainAction) JoinChain(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.CreateChain(ctx, rawMsg)
}

// QuitChain is used to send quit chain request raw message by grpc.
func (action *ChainAction) QuitChain(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.DeleteChain(ctx, rawMsg)
}

// QueryChain is used to send query chain request raw message by grpc.
func (action *ChainAction) QueryChain(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.QueryChainInfo(ctx, rawMsg)
}

// QueryAllChains is used to send query all chains request raw message by grpc.
func (action *ChainAction) QueryAllChains(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.QueryAllChainInfos(ctx, rawMsg)
}

func (action *ChainAction) getClient() (nodeservice.ChainManagerClient, error) {
	if action.conn == nil || action.conn.GetState() == connectivity.Shutdown {
		if err := action.newClient(); err != nil {
			return nil, errors.WithMessage(err, "new client error")
		}
	}
	return action.client, nil
}

func (action *ChainAction) newClient() error {
	cc, err := action.newClientConn()
	if err != nil {
		return errors.WithMessage(err, "get client connection error")
	}
	action.conn = cc
	action.client = nodeservice.NewChainManagerClient(action.conn)
	return nil
}
