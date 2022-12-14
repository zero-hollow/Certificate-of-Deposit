/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

// Package rawmessage is the implementation of raw message, which is used to build raw message for grpc request.
package rawmessage

import (
	"git.huawei.com/huaweichain/common/version"
	"git.huawei.com/huaweichain/sdk/utils"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto/relayer"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/sdk/crypto"
)

// MsgBuilder is the interface of base raw message builder.
type MsgBuilder interface {
	GetRawMessage(payload []byte) (*common.RawMessage, error)
	GetCrossChainRawMessage(payload []byte) (*relayer.RawMessage, error)
	BuildVoteTx(chainID string, voteTxData *common.VoteTxData) (*common.Transaction, error)
	GetTxRawMsg(*common.Transaction) (*TxRawMsg, error)
	GetTxRawMsgWithTarget(tx *common.Transaction, target *common.ProposeTarget) (*TxRawMsg, error)
	BuildIdentity() *common.Identity
	Hash(data []byte) []byte
}

// MsgBuilderImpl is the definition of MsgBuilderImpl.
type MsgBuilderImpl struct {
	crypto crypto.Crypto
}

// NewMsgBuilderImpl is used to create an instance of MsgBuilderImpl by specifying crypto.
func NewMsgBuilderImpl(crypto crypto.Crypto) (*MsgBuilderImpl, error) {
	msgBuilder := &MsgBuilderImpl{crypto: crypto}
	return msgBuilder, nil
}

// GetRawMessage is used to generate raw message for the payload.
func (builder *MsgBuilderImpl) GetRawMessage(payload []byte) (*common.RawMessage, error) {
	cert, err := builder.crypto.GetCertificate()
	if err != nil {
		return nil, errors.WithMessage(err, "get certificate error")
	}

	sign, err := builder.crypto.Sign(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "sign payload error")
	}
	rawMessage := &common.RawMessage{
		Signature: &common.RawMessage_Signature{
			Cert: cert,
			Sign: sign,
		},
		Payload:         payload,
		PlatformVersion: version.PlatformVersion,
	}

	return rawMessage, nil
}

// GetCrossChainRawMessage is used to generate cross chain raw message for the payload.
func (builder *MsgBuilderImpl) GetCrossChainRawMessage(payload []byte) (*relayer.RawMessage, error) {
	cert, err := builder.crypto.GetCertificate()
	if err != nil {
		return nil, errors.WithMessage(err, "get certificate error")
	}

	sign, err := builder.crypto.Sign(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "sign payload error")
	}
	rawMessage := &relayer.RawMessage{
		Signature: &relayer.RawMessage_Signature{
			PubKey: cert,
			Sign:   sign,
		},
		Payload: payload,
	}

	return rawMessage, nil
}

// BuildVoteTx is used to build vote transaction.
func (builder *MsgBuilderImpl) BuildVoteTx(chainID string, voteTxData *common.VoteTxData) (*common.Transaction,
	error) {
	voteTxDataByte, err := proto.Marshal(voteTxData)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal vote tx data error")
	}

	txHeader := &common.TxHeader{
		ChainId:   chainID,
		Type:      common.VOTE_TRANSACTION,
		Timestamp: utils.GenerateTimestamp(),
		Nonce:     utils.GenerateNonce(),
		Creator:   builder.BuildIdentity(),
		Version:   version.PlatformVersion,
	}
	txPayload := &common.TxPayload{
		Header: txHeader,
		Data:   voteTxDataByte,
	}
	txPayloadBytes, err := proto.Marshal(txPayload)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal tx payload error")
	}

	certBytes, err := builder.crypto.GetCertificate()
	if err != nil {
		return nil, errors.WithMessage(err, "get certificate error")
	}
	sign, err := builder.crypto.Sign(txPayloadBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "sign error")
	}
	approval := &common.Approval{Identity: certBytes, Sign: sign}
	var approvals []*common.Approval
	approvals = append(approvals, approval)
	return &common.Transaction{Payload: txPayloadBytes, Approvals: approvals}, nil
}

// GetTxRawMsg is used to get tx raw message.
func (builder *MsgBuilderImpl) GetTxRawMsg(tx *common.Transaction) (*TxRawMsg, error) {
	return builder.GetTxRawMsgWithTarget(tx, nil)
}

// GetTxRawMsgWithTarget is used to get tx raw message with propose target.
func (builder *MsgBuilderImpl) GetTxRawMsgWithTarget(tx *common.Transaction,
	target *common.ProposeTarget) (*TxRawMsg, error) {
	payload, err := proto.Marshal(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal transaction error")
	}
	rawMsg, err := builder.GetRawMessage(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "get raw message error")
	}
	rawMsg.Type = common.DIRECT
	if target != nil {
		bytes, err := target.Marshal()
		if err != nil {
			return nil, errors.WithMessage(err, "common.ProposeTarget marshal error")
		}
		rawMsg.ProxyInfo = bytes
		rawMsg.Type = common.PROXY
	}

	return &TxRawMsg{
		Hash: builder.crypto.Hash(tx.Payload),
		Msg:  rawMsg,
	}, nil
}

// BuildIdentity is used to build identity.
func (builder *MsgBuilderImpl) BuildIdentity() *common.Identity {
	return &common.Identity{
		Org:  builder.crypto.GetOrg(),
		Type: common.COMMON_NAME,
		Id:   []byte(builder.crypto.GetCommonName()),
	}
}

// Hash is the function to compute hash. It could be sha256 or sm3 depends on the identity
// algorithm config.
func (builder *MsgBuilderImpl) Hash(data []byte) []byte {
	return builder.crypto.Hash(data)
}

// TxRawMsg is a wrapper msg for tx raw msg which contains
// tx hash and its raw message
type TxRawMsg struct {
	Hash []byte
	Msg  *common.RawMessage
}
