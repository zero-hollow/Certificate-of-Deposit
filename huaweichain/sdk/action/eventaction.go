/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package action

import (
	"context"
	"encoding/hex"
	"sync"

	"git.huawei.com/huaweichain/sdk/action/event"

	"github.com/pkg/errors"

	"google.golang.org/grpc/connectivity"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"github.com/gogo/protobuf/proto"
)

// EventAction is the action for event operations.
type EventAction struct {
	action
	client nodeservice.EventServiceClient
}

// NewEventAction is used to create event action instance with a grpc client connection.
func NewEventAction(config *Config) *EventAction {
	action := &EventAction{action: action{config: config}}
	return action
}

// GetBlockEventService is used to get block event service.
func (action *EventAction) GetBlockEventService(chainID string) (*event.BlockEventService, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	return event.NewBlockEventService(client, chainID), nil
}

// GetTxEventService is used to get tx event service by default event source type. default tx event.
func (action *EventAction) GetTxEventService(chainID string) (*event.TxEventService, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	return event.NewTxEventService(client, chainID)
}

// GetTxEventServiceWithSourceType is used to get tx event service with specified event source type.
func (action *EventAction) GetTxEventServiceWithSourceType(chainID string,
	source event.SourceType) (*event.TxEventService, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	return event.NewTxEventServiceWithSourceType(client, chainID, source)
}

// Listen is used to register result event to server and get the register result client.
func (action *EventAction) Listen(chainID string) (nodeservice.EventService_RegisterResultEventClient, error) {
	in := &common.RawMessage{}
	startPoint := &nodeservice.EventStartPoint{}
	startPoint.ChainId = chainID
	startPoint.Type = nodeservice.LATEST
	bytes, err := startPoint.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal EventStartPoint error")
	}
	in.Payload = bytes

	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	event, err := client.RegisterResultEvent(context.Background(), in)
	if err != nil {
		return nil, errors.WithMessage(err, "event action RegisterResultEvent failed")
	}
	return event, nil
}

// ListenBlockAndResult is used to register block and result event to server and get the register result client.
func (action *EventAction) ListenBlockAndResult(chainID string,
	number uint64) (nodeservice.EventService_RegisterBlockAndResultEventClient, error) {
	in := &common.RawMessage{}
	startPoint := &nodeservice.EventStartPoint{
		ChainId:  chainID,
		Type:     nodeservice.SPECIFIC,
		BlockNum: number,
	}
	bytes, err := startPoint.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal EventStartPoint error")
	}
	in.Payload = bytes

	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	event, err := client.RegisterBlockAndResultEvent(context.Background(), in)
	if err != nil {
		return nil, errors.WithMessage(err, "event action RegisterBlockAndResultEvent failed")
	}
	return event, nil
}

// RegisterTxEvent is used to register tx event to server and get the register tx event client.
func (action *EventAction) RegisterTxEvent(chainID string) (*TxEvent, error) {
	client, err := action.getClient()
	if err != nil {
		return nil, errors.WithMessage(err, "get client error")
	}
	txEventClient, err := client.RegisterTxEvent(context.Background())
	if err != nil {
		return nil, errors.WithMessage(err, "client action RegisterTxEvent failed")
	}
	txEvent := &nodeservice.TxEvent{ChainId: chainID, Type: nodeservice.REGISTER_CLIENT}
	bytes, err := txEvent.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal tx event error: %v")
	}
	rawMsg := &common.RawMessage{Payload: bytes}
	if err = txEventClient.Send(rawMsg); err != nil {
		return nil, errors.WithMessage(err, "client action send register info failed")
	}

	rawMsg, err = txEventClient.Recv()
	if err != nil {
		return nil, errors.WithMessage(err, "RegisterTxEvent error: event receive response message error")
	}
	res := &nodeservice.TxEventRes{}
	err = proto.Unmarshal(rawMsg.Payload, res)
	if err != nil {
		return nil, errors.WithMessage(err, "RegisterTxEvent error: unmarshal tx event response error")
	}
	if res.Status == nodeservice.SUCCESS {
		return NewTxEvent(txEventClient), nil
	}
	return nil, errors.WithMessage(err, "RegisterTxEvent error")
}

func (action *EventAction) getClient() (nodeservice.EventServiceClient, error) {
	if action.conn == nil || action.conn.GetState() == connectivity.Shutdown {
		if err := action.newClient(); err != nil {
			return nil, errors.WithMessage(err, "new client error: %v")
		}
	}
	return action.client, nil
}

func (action *EventAction) newClient() error {
	cc, err := action.newClientConn()
	if err != nil {
		return errors.WithMessage(err, "get client connection error: %v")
	}
	action.conn = cc
	action.client = nodeservice.NewEventServiceClient(action.conn)
	return nil
}

// TxEvent is the definition of TxEvent.
type TxEvent struct {
	client nodeservice.EventService_RegisterTxEventClient
	txMap  sync.Map
}

// NewTxEvent is used to create instance of tx event.
func NewTxEvent(client nodeservice.EventService_RegisterTxEventClient) *TxEvent {
	txEvent := &TxEvent{client: client}
	go txEvent.listen()
	return txEvent
}

// RegisterTx is used to register tx id to server.
func (e *TxEvent) RegisterTx(txHash []byte) (chan *common.TxResult, error) {
	txEvent := &nodeservice.TxEvent{TxHash: txHash, Type: nodeservice.REGISTER_TX_HASH}
	bytes, err := txEvent.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal tx event error: %v")
	}
	rawMsg := &common.RawMessage{Payload: bytes}
	if err := e.client.Send(rawMsg); err != nil {
		return nil, errors.WithMessage(err, "send register txid info error")
	}
	ch := make(chan *common.TxResult, 1)
	e.txMap.Store(hex.EncodeToString(txHash), ch)
	return ch, nil
}

func (e *TxEvent) listen() {
	for {
		responseMsg, err := e.client.Recv()
		if err != nil {
			log.Errorf("event receive response message error: %v", err)
			return
		}
		resp := &nodeservice.TxEventRes{}
		err = proto.Unmarshal(responseMsg.Payload, resp)
		if err != nil {
			log.Errorf("unmarshal tx event response result error: %v", err)
			return
		}
		txResult := &common.TxResult{}
		err = proto.Unmarshal(resp.Payload, txResult)
		if err != nil {
			log.Errorf("unmarshal tx result error: %v", err)
			return
		}
		txID := hex.EncodeToString(txResult.TxHash)
		value, ok := e.txMap.Load(txID)
		if !ok {
			continue
		}
		ch, ok := value.(chan *common.TxResult)
		if !ok {
			log.Error("type assert error: chan *common.TxResult")
			continue
		}
		if ch == nil {
			log.Error("chan *common.TxResult init incorrectly.")
			continue
		}
		ch <- txResult
		e.txMap.Delete(txID)
	}
}
