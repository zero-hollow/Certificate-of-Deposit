/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

// Package event provide the implementation of event service.
package event

import (
	"context"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/logger"
	"git.huawei.com/huaweichain/proto"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
)

var log = logger.GetModuleLogger("go sdk", "event")

// BlockEventService is the definition of block event service.
type BlockEventService struct {
	chainID string
	client  nodeservice.EventServiceClient
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewBlockEventService is used to create an instance of block event service.
func NewBlockEventService(client nodeservice.EventServiceClient, chainID string) *BlockEventService {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockEventService{
		chainID: chainID,
		client:  client,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterBlockEvent is used to register block event from latest block.
func (s *BlockEventService) RegisterBlockEvent() (*BlockIterator, error) {
	return s.registerBlockEventFrom(nodeservice.LATEST, uint64(0))
}

// RegisterBlockEventFrom is used to register block event from the specified block number.
func (s *BlockEventService) RegisterBlockEventFrom(startNum uint64) (*BlockIterator, error) {
	return s.registerBlockEventFrom(nodeservice.SPECIFIC, startNum)
}

func (s *BlockEventService) registerBlockEventFrom(startPointType nodeservice.StartPointType,
	startNum uint64) (*BlockIterator, error) {
	rawMsg, err := s.buildRawMsg(startPointType, startNum)
	if err != nil {
		return nil, errors.WithMessage(err, "build raw message error")
	}
	blockEventClient, err := s.client.RegisterBlockEvent(s.ctx, rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "block event service RegisterEvent failed")
	}
	return &BlockIterator{
		client: blockEventClient,
		ctx:    s.ctx,
	}, nil
}

// RegisterBlockResultEvent is used to register block result event from latest block.
func (s *BlockEventService) RegisterBlockResultEvent() (*BlockResultIterator, error) {
	return s.registerBlockResultEventFrom(nodeservice.LATEST, uint64(0))
}

// RegisterBlockResultEventFrom is used to register block result event from the specified block number.
func (s *BlockEventService) RegisterBlockResultEventFrom(startNum uint64) (*BlockResultIterator, error) {
	return s.registerBlockResultEventFrom(nodeservice.SPECIFIC, startNum)
}

func (s *BlockEventService) registerBlockResultEventFrom(startPointType nodeservice.StartPointType,
	startNum uint64) (*BlockResultIterator, error) {
	rawMsg, err := s.buildRawMsg(startPointType, startNum)
	if err != nil {
		return nil, errors.WithMessage(err, "build raw message error")
	}
	resultEventClient, err := s.client.RegisterResultEvent(s.ctx, rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "block event service RegisterEvent failed")
	}
	return &BlockResultIterator{
		client: resultEventClient,
		ctx:    s.ctx,
	}, nil
}

// RegisterBlockAndResultEvent is used to register block and result event from latest block.
func (s *BlockEventService) RegisterBlockAndResultEvent() (*BlockAndResultIterator, error) {
	return s.registerBlockAndResultEventFrom(nodeservice.LATEST, 0)
}

// RegisterBlockAndResultEventFrom is used to register block and result event from the specified block number.
func (s *BlockEventService) RegisterBlockAndResultEventFrom(startNum uint64) (*BlockAndResultIterator, error) {
	return s.registerBlockAndResultEventFrom(nodeservice.SPECIFIC, startNum)
}

func (s *BlockEventService) registerBlockAndResultEventFrom(startPointType nodeservice.StartPointType,
	startNum uint64) (*BlockAndResultIterator, error) {
	rawMsg, err := s.buildRawMsg(startPointType, startNum)
	if err != nil {
		return nil, errors.WithMessage(err, "build raw message error")
	}
	blockAndResultEventClient, err := s.client.RegisterBlockAndResultEvent(s.ctx, rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "block event service RegisterEvent failed")
	}
	return &BlockAndResultIterator{
		client: blockAndResultEventClient,
		ctx:    s.ctx,
	}, nil
}

func (s *BlockEventService) buildRawMsg(startPointType nodeservice.StartPointType,
	startNum uint64) (*common.RawMessage, error) {
	startPoint := &nodeservice.EventStartPoint{
		ChainId:  s.chainID,
		Type:     startPointType,
		BlockNum: startNum,
	}
	bytes, err := startPoint.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal EventStartPoint error")
	}
	return &common.RawMessage{
		Payload: bytes,
	}, nil
}

// Close is used to close block event service.
func (s *BlockEventService) Close() {
	s.cancel()
}

// BlockIterator is the definition of block iterator.
type BlockIterator struct {
	client nodeservice.EventService_RegisterBlockEventClient
	ctx    context.Context
}

// Next is used to get the next element of iterator.
func (itr *BlockIterator) Next() (*common.Block, error) {
	rawMsg, err := itr.client.Recv()
	if err != nil {
		return nil, errors.WithMessage(err, "recv block error")
	}
	select {
	case <-itr.ctx.Done():
		return nil, errors.New("grpc stream has closed")
	default:
		block := &common.Block{}
		err = proto.Unmarshal(rawMsg.Payload, block)
		if err != nil {
			return nil, errors.WithMessage(err, "unmarshal block error")
		}
		return block, nil
	}
}

// BlockResultIterator is the definition of block result iterator.
type BlockResultIterator struct {
	client nodeservice.EventService_RegisterResultEventClient
	ctx    context.Context
}

// Next is used to get the next element of iterator.
func (itr *BlockResultIterator) Next() (*common.BlockResult, error) {
	rawMsg, err := itr.client.Recv()
	if err != nil {
		return nil, errors.WithMessage(err, "recv block result error")
	}
	select {
	case <-itr.ctx.Done():
		return nil, errors.New("grpc stream has closed")
	default:
		blockResult := &common.BlockResult{}
		err = proto.Unmarshal(rawMsg.Payload, blockResult)
		if err != nil {
			return nil, errors.WithMessage(err, "unmarshal block result error")
		}
		return blockResult, nil
	}
}

// BlockAndResultIterator is the definition of block and result iterator.
type BlockAndResultIterator struct {
	client nodeservice.EventService_RegisterBlockAndResultEventClient
	ctx    context.Context
}

// Next is used to get the next element of iterator.
func (itr *BlockAndResultIterator) Next() (*common.BlockAndResult, error) {
	rawMsg, err := itr.client.Recv()
	if err != nil {
		return nil, errors.WithMessage(err, "recv block result error")
	}
	select {
	case <-itr.ctx.Done():
		return nil, errors.New("grpc stream has closed")
	default:
		br := &common.BlockAndResult{}
		err = proto.Unmarshal(rawMsg.Payload, br)
		if err != nil {
			return nil, errors.WithMessage(err, "unmarshal block and result error")
		}
		return br, nil
	}
}
