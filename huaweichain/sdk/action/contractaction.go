/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
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

// ContractAction is the action for contract operations.
type ContractAction struct {
	action
	contractClient    nodeservice.ContractClient
	transactionClient nodeservice.TransactionSenderClient
}

// NewContractAction is used to create contract action instance with action config.
func NewContractAction(config *Config) *ContractAction {
	action := &ContractAction{action: action{config: config}}
	return action
}

// Invoke is used to send invoke request raw message to contract by grpc.
func (action *ContractAction) Invoke(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	if rawMsg == nil {
		return nil, errors.New("raw message is nil")
	}
	client, err := action.getContractClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get contract client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.Invoke(ctx, rawMsg)
}

// Query is used to send query request raw message to contract by grpc.
func (action *ContractAction) Query(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	if rawMsg == nil {
		return nil, errors.New("raw message is nil")
	}
	client, err := action.getContractClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get contract client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.Query(ctx, rawMsg)
}

// ContractImport is used to send import contract request raw message by grpc.
func (action *ContractAction) ContractImport(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getContractClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get contract client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.Import(ctx, rawMsg)
}

// ContractUnImport is used to send unimport contract request raw message by grpc.
func (action *ContractAction) ContractUnImport(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getContractClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get contract client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.UnImport(ctx, rawMsg)
}

// Transaction is used to send transaction request raw message by grpc.
func (action *ContractAction) Transaction(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	if rawMsg == nil {
		return nil, errors.New("raw message is nil")
	}
	client, err := action.getTransactionClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get transaction client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.SendTransaction(ctx, rawMsg)
}

func (action *ContractAction) getContractClient() (nodeservice.ContractClient, error) {
	if err := action.resetClients(); err != nil {
		return nil, errors.WithMessage(err, "reset clients error")
	}
	return action.contractClient, nil
}

func (action *ContractAction) getTransactionClient() (nodeservice.TransactionSenderClient, error) {
	if err := action.resetClients(); err != nil {
		return nil, errors.WithMessage(err, "reset clients error")
	}
	return action.transactionClient, nil
}

func (action *ContractAction) resetClients() error {
	if action.conn == nil || action.conn.GetState() == connectivity.Shutdown {
		if err := action.newClients(); err != nil {
			return errors.WithMessage(err, "new client error")
		}
	}
	return nil
}

func (action *ContractAction) newClients() error {
	cc, err := action.newClientConn()
	if err != nil {
		return errors.WithMessage(err, "get client connection error")
	}
	action.conn = cc
	action.contractClient = nodeservice.NewContractClient(action.conn)
	action.transactionClient = nodeservice.NewTransactionSenderClient(action.conn)
	return nil
}

// QueryState is used to send import contract request raw message by grpc.
func (action *ContractAction) QueryState(rawMsg *common.RawMessage) (*common.RawMessage, error) {
	client, err := action.getContractClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get contract client error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	return client.QueryState(ctx, rawMsg)
}
