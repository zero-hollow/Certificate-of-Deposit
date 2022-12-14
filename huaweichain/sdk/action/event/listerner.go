/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package event

import (
	"context"

	"git.huawei.com/huaweichain/proto"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"github.com/pkg/errors"
)

type eventListener interface {
	registerTx(txHash []byte) (chan *common.TxResult, error)
	listen()
}

type txListener struct {
	ch     chan msg
	client nodeservice.EventService_RegisterTxEventClient
	ctx    context.Context
}

func newTxListener(ctx context.Context, client nodeservice.EventServiceClient, chainID string,
	ch chan msg) (*txListener, error) {
	txEventClient, err := client.RegisterTxEvent(ctx)
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
	if res.Status != nodeservice.SUCCESS {
		return nil, errors.WithMessage(err, "RegisterTxEvent error")
	}
	return &txListener{
		ch:     ch,
		client: txEventClient,
		ctx:    ctx,
	}, nil
}

func (l *txListener) registerTx(txHash []byte) (chan *common.TxResult, error) {
	ch := make(chan *common.TxResult, 1)
	msg := &registerMsg{
		txHash: txHash,
		ch:     ch,
	}
	l.ch <- msg
	txEvent := &nodeservice.TxEvent{TxHash: txHash, Type: nodeservice.REGISTER_TX_HASH}
	bytes, err := txEvent.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal tx event error: %v")
	}
	rawMsg := &common.RawMessage{Payload: bytes}
	if err := l.client.Send(rawMsg); err != nil {
		return nil, errors.WithMessage(err, "send register txid info error")
	}
	return ch, nil
}

func (l *txListener) listen() {
	processor := func(rawMsg *common.RawMessage) {
		resp := &nodeservice.TxEventRes{}
		err := proto.Unmarshal(rawMsg.Payload, resp)
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
		msg := &resultMsg{txResult: txResult}
		l.ch <- msg
	}
	listen(l.client.Recv, processor, l.ctx.Done())
}

type blockListener struct {
	ch     chan msg
	client nodeservice.EventService_RegisterResultEventClient
	ctx    context.Context
}

func newBlockListener(ctx context.Context, client nodeservice.EventServiceClient, chainID string,
	ch chan msg) (*blockListener, error) {
	startPoint := &nodeservice.EventStartPoint{}
	startPoint.ChainId = chainID
	startPoint.Type = nodeservice.LATEST
	bytes, err := startPoint.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal EventStartPoint error")
	}
	rawMsg := &common.RawMessage{
		Payload: bytes,
	}
	blockResultClient, err := client.RegisterResultEvent(ctx, rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "block listener RegisterResultEvent failed")
	}
	return &blockListener{
		ch:     ch,
		client: blockResultClient,
		ctx:    ctx,
	}, nil
}

func (l *blockListener) registerTx(txHash []byte) (chan *common.TxResult, error) {
	ch := make(chan *common.TxResult, 1)
	msg := &registerMsg{
		txHash: txHash,
		ch:     ch,
	}
	l.ch <- msg
	return ch, nil
}

func (l *blockListener) listen() {
	processor := func(rawMsg *common.RawMessage) {
		res := &common.BlockResult{}
		err := proto.Unmarshal(rawMsg.Payload, res)
		if err != nil {
			log.Errorf("unmarshal block result error: %v", err)
			return
		}
		for _, txResult := range res.TxResults {
			msg := &resultMsg{txResult: txResult}
			l.ch <- msg
		}
	}
	listen(l.client.Recv, processor, l.ctx.Done())
}

func listen(recv func() (*common.RawMessage, error), processor func(rawMsg *common.RawMessage),
	done <-chan struct{}) {
	if done == nil {
		log.Error(" context done chan is nil")
		return
	}
	for {
		responseMsg, err := recv()
		if err != nil {
			log.Errorf("event receive response message error: %v", err)
			return
		}
		select {
		case <-done:
			return
		default:
			processor(responseMsg)
		}
	}
}
