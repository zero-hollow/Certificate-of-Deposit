/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

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

// QueryAction is the action for query operations.
type QueryAction struct {
	action
	chainServiceClient nodeservice.ChainServiceClient
	nodeServiceClient  nodeservice.NodeServiceClient
	voteManagerClient  nodeservice.VoteManagerClient
}

// NewQueryAction is used to create query action instance with action config.
func NewQueryAction(config *Config) *QueryAction {
	action := &QueryAction{action: action{config: config}}
	return action
}

// GetLatestChainState is used to send get latest chain request raw message by grpc.
func (action *QueryAction) GetLatestChainState(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getChainServiceClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get chain service client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.GetLatestChainState(ctx, rawMsg)
}

// GetBlockByNum is used to send get block by number request raw message by grpc.
func (action *QueryAction) GetBlockByNum(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getChainServiceClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get chain service client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.GetBlockByNum(ctx, rawMsg)
}

// GetBlockAndResultByNum is used to send get block and block result by number request raw message by grpc.
func (action *QueryAction) GetBlockAndResultByNum(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getChainServiceClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get chain service client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.GetBlockAndResultByNum(ctx, rawMsg)
}

// GetBlockByTxHash is used to send get block by transaction id request raw message by grpc.
func (action *QueryAction) GetBlockByTxHash(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getChainServiceClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get chain service client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.GetBlockByTxHash(ctx, rawMsg)
}

// GetTxByHash is used to send get transaction by transaction id request raw message by grpc.
func (action *QueryAction) GetTxByHash(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	if rawMsg == nil {
		return nil, errors.New("raw message is nil")
	}
	client, err := action.getChainServiceClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get chain service client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.GetTxByHash(ctx, rawMsg)
}

// GetTxResultByTxHash is used to send get tx result by transaction id request raw message by grpc.
func (action *QueryAction) GetTxResultByTxHash(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	if rawMsg == nil {
		return nil, errors.New("raw message is nil")
	}
	client, err := action.getChainServiceClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get chain service client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.GetTxResultByTxHash(ctx, rawMsg)
}

// GetContractInfo is used to send get contract info request raw message by grpc.
func (action *QueryAction) GetContractInfo(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getChainServiceClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get chain service client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.GetContractInfo(ctx, rawMsg)
}

// GetVote is used to send get vote request raw message by grpc.
func (action *QueryAction) GetVote(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getVoteManagerClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get vote manager client error: %v")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.Query(ctx, rawMsg)
}

func (action *QueryAction) getChainServiceClient() (nodeservice.ChainServiceClient, error) {
	if err := action.resetClients(); err != nil {
		return nil, errors.WithMessage(err, "reset clients error: %v")
	}
	return action.chainServiceClient, nil
}

func (action *QueryAction) getVoteManagerClient() (nodeservice.VoteManagerClient, error) {
	if err := action.resetClients(); err != nil {
		return nil, errors.WithMessage(err, "reset clients error: %v")
	}
	return action.voteManagerClient, nil
}

func (action *QueryAction) resetClients() error {
	if action.conn == nil || action.conn.GetState() == connectivity.Shutdown {
		if err := action.newClients(); err != nil {
			return errors.WithMessage(err, "new client error: %v")
		}
	}
	return nil
}

func (action *QueryAction) newClients() error {
	cc, err := action.newClientConn()
	if err != nil {
		return errors.WithMessage(err, "get client connection error: %v")
	}
	action.conn = cc
	action.chainServiceClient = nodeservice.NewChainServiceClient(action.conn)
	action.nodeServiceClient = nodeservice.NewNodeServiceClient(action.conn)
	action.voteManagerClient = nodeservice.NewVoteManagerClient(action.conn)
	return nil
}
