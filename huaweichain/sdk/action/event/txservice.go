/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package event

import (
	"context"
	"encoding/hex"
	"reflect"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
)

// SourceType is the data source type of tx event service.
type SourceType int32

const (
	// Block is the source type of block event.
	Block SourceType = 0
	// Tx is the source type of tx event.
	Tx SourceType = 1
)

// TxEventService 支持Block和Tx两种不同的Event来源
type TxEventService struct {
	txMap    map[string]chan *common.TxResult
	chainID  string
	ch       chan msg
	listener eventListener
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewTxEventService is used to create an instance of tx event service, default with tx event data source.
func NewTxEventService(client nodeservice.EventServiceClient, chainID string) (*TxEventService, error) {
	return NewTxEventServiceWithSourceType(client, chainID, Tx)
}

// NewTxEventServiceWithSourceType is used to create an instance of tx event service with specified event data source.
func NewTxEventServiceWithSourceType(client nodeservice.EventServiceClient, chainID string,
	source SourceType) (*TxEventService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan msg)
	var listener eventListener
	var err error
	if source == Block {
		listener, err = newBlockListener(ctx, client, chainID, ch)
		if err != nil {
			cancel()
			return nil, errors.WithMessage(err, "new block listener error")
		}
	} else {
		listener, err = newTxListener(ctx, client, chainID, ch)
		if err != nil {
			cancel()
			return nil, errors.WithMessage(err, "new tx listener error")
		}
	}
	es := &TxEventService{
		chainID:  chainID,
		ch:       ch,
		txMap:    make(map[string]chan *common.TxResult),
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
	}
	go listener.listen()
	go es.dispatch()
	return es, nil
}

// RegisterTx is used to register transaction.
func (s *TxEventService) RegisterTx(txHash []byte) (chan *common.TxResult, error) {
	return s.listener.registerTx(txHash)
}

// Close is used to close tx event service.
func (s *TxEventService) Close() {
	s.cancel()
}

func (s *TxEventService) dispatch() {
	for {
		select {
		case m := <-s.ch:
			v, ok := m.(*registerMsg)
			if ok {
				txID := hex.EncodeToString(v.txHash)
				s.txMap[txID] = v.ch
				continue
			}
			rm, ok := m.(*resultMsg)
			if !ok {
				log.Errorf("unsupported message type: %v", reflect.TypeOf(m))
				continue
			}
			txID := hex.EncodeToString(rm.txResult.TxHash)
			ch, ok := s.txMap[txID]
			if !ok {
				continue
			}
			if ch == nil {
				log.Error("tx result chan is nil")
				continue
			}
			ch <- rm.txResult
			delete(s.txMap, txID)
		case <-s.ctx.Done():
			return
		}
	}
}

type msg interface{}

type registerMsg struct {
	ch     chan *common.TxResult
	txHash []byte
}

type resultMsg struct {
	txResult *common.TxResult
}
