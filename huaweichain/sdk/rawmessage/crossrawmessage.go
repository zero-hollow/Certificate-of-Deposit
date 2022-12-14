/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package rawmessage

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto/relayer"
)

// CrossChainRawMessage is the definition of CrossChainRawMessage.
type CrossChainRawMessage struct {
	builder MsgBuilder
}

// NewCrossChainRawMessage is used to create an instance of cross chain raw message by specifying an message builder.
func NewCrossChainRawMessage(builder MsgBuilder) *CrossChainRawMessage {
	return &CrossChainRawMessage{builder: builder}
}

// BuildRegisterChainRawMessage is used to build register chain raw message for register chain by specifying chain id.
func (msg *CrossChainRawMessage) BuildRegisterChainRawMessage(chainID string) (*relayer.RawMessage, error) {
	if err := checkChainID(chainID); err != nil {
		return nil, errors.WithMessage(err, "check chainID error")
	}
	registeredChains := &relayer.RegisterChain{
		Name:   chainID,
		Height: 0,
		Cas: &relayer.ChainCas{
			ChainCas: []*relayer.OrgCa{
				{
					OrgId:  "org1",
					CaCert: []byte("cert"),
				},
			},
		},
	}
	payload, err := registeredChains.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal registered chain error")
	}
	request := &relayer.Request{Type: relayer.Register_Chain, Payload: payload}
	return msg.buildRequestRawMessage(request)
}

// BuildQueryChainRawMessage is used to build query chain raw message from relayer.
func (msg *CrossChainRawMessage) BuildQueryChainRawMessage() (*relayer.RawMessage, error) {
	queryChainReq := &relayer.QueryChainsReq{}
	payload, err := queryChainReq.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal query chian req error")
	}
	request := &relayer.Request{Type: relayer.Query_Chains, Payload: payload}
	return msg.buildRequestRawMessage(request)
}

// BuildRegisterContractRawMessage is used to build register contract raw message for register chain by specifying
// contract name.
func (msg *CrossChainRawMessage) BuildRegisterContractRawMessage(chainID string, contractName string,
	invokedList []*relayer.CrossContract) (*relayer.RawMessage, error) {
	if err := checkCrossParams(chainID, contractName, invokedList); err != nil {
		return nil, errors.WithMessage(err, "check register contract paras error")
	}
	registeredContractInfo := &relayer.RegisterContract{Name: contractName, Chain: chainID,
		InvokedList: invokedList}
	payload, err := registeredContractInfo.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal register contract info error")
	}
	request := &relayer.Request{Type: relayer.Register_Contract, Payload: payload}
	return msg.buildRequestRawMessage(request)
}

// BuildQueryContractsRawMessage is used to build query contracts raw message from relayer.
func (msg *CrossChainRawMessage) BuildQueryContractsRawMessage(chainID string) (*relayer.RawMessage, error) {
	if err := checkChainID(chainID); err != nil {
		return nil, errors.WithMessage(err, "check chainID error")
	}
	queryContractsReq := &relayer.QueryContractsReq{Chain: chainID}
	payload, err := queryContractsReq.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal query contracts req error")
	}
	request := &relayer.Request{Type: relayer.Query_Contracts, Payload: payload}
	return msg.buildRequestRawMessage(request)
}

// BuildQueryContractInfoRawMessage is used to build query contract info raw message from relayer.
func (msg *CrossChainRawMessage) BuildQueryContractInfoRawMessage(chainID string,
	contract string) (*relayer.RawMessage, error) {
	if err := checkChainID(chainID); err != nil {
		return nil, errors.WithMessage(err, "check chainID error")
	}
	if err := checkContractName(contract); err != nil {
		return nil, errors.WithMessage(err, "check contract name error")
	}
	queryContractInfoReq := &relayer.QueryContractInfoReq{Chain: chainID, Contract: contract}
	payload, err := queryContractInfoReq.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal query contract info req error")
	}
	request := &relayer.Request{Type: relayer.Query_ContractInfo, Payload: payload}
	return msg.buildRequestRawMessage(request)
}

// BuildQueryEventInfoRawMessage is used to build query event info raw message from relayer.
func (msg *CrossChainRawMessage) BuildQueryEventInfoRawMessage(chainAddr, id string) (*relayer.RawMessage, error) {
	if err := checkChainAddr(chainAddr); err != nil {
		return nil, errors.WithMessage(err, "check chain addr error")
	}
	if err := checkTransactionID(id); err != nil {
		return nil, errors.WithMessage(err, "check contract name error")
	}
	queryEventInfoReq := &relayer.QueryCrossEventInfoReq{EventId: getEventID(chainAddr, id)}
	payload, err := proto.Marshal(queryEventInfoReq)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal query event info req error")
	}
	request := &relayer.Request{Type: relayer.Query_CrossEventInfo, Payload: payload}
	return msg.buildRequestRawMessage(request)
}

func (msg *CrossChainRawMessage) buildRequestRawMessage(request *relayer.Request) (*relayer.RawMessage, error) {
	payload, err := request.Marshal()
	if err != nil {
		return nil, errors.WithMessage(err, "marshal request error")
	}
	rawMsg, err := msg.builder.GetCrossChainRawMessage(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "get raw message error")
	}
	return rawMsg, nil
}

func getEventID(chainAddr, id string) string {
	unique := sha256.Sum256([]byte(id + chainAddr))
	b := unique[:]
	return base64.StdEncoding.EncodeToString(b)
}
