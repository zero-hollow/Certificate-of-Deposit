/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package rawmessage

import (
	"bytes"
	"encoding/hex"
	"strings"

	"git.huawei.com/huaweichain/common/version"

	"git.huawei.com/huaweichain/sdk/utils"
	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"github.com/gogo/protobuf/proto"
)

// ContractRawMessage is the definition of ContractRawMessage.
type ContractRawMessage struct {
	builder MsgBuilder
}

// InvokeRawMsgParas is the parameters to build invoke raw message.
type InvokeRawMsgParas struct {
	ChainID      string
	CtInvocation *common.ContractInvocation
	Domains      []string
	Targets      []*common.EndorseTarget
	Supplier     func() (uint64, error)
}

// NewContractRawMessage is used to create an instance of contract raw message by specifying an message builder.
func NewContractRawMessage(builder MsgBuilder) *ContractRawMessage {
	return &ContractRawMessage{builder: builder}
}

// BuildInvokeRawMsg is used to build invoke raw message.
func (msg *ContractRawMessage) BuildInvokeRawMsg(chainID string, contractInvocation *common.ContractInvocation,
	domains []string) (*common.RawMessage, error) {
	return msg.buildInvokeRawMsg(chainID, contractInvocation, nil, domains, common.DIRECT)
}

// BuildInvokeRawMsgWithTargets is used to build invoke raw message with targets.
func (msg *ContractRawMessage) BuildInvokeRawMsgWithTargets(chainID string,
	contractInvocation *common.ContractInvocation, domains []string,
	targets []*common.EndorseTarget) ([]*common.RawMessage, error) {
	return msg.buildInvokeRawMsgWithTargets(chainID, contractInvocation, domains, targets, nil)
}

// BuildInvokeRawMsgCombined is used to build invoke raw message.
func (msg *ContractRawMessage) BuildInvokeRawMsgCombined(paras *InvokeRawMsgParas) ([]*common.RawMessage, error) {
	return msg.buildInvokeRawMsgWithTargets(paras.ChainID, paras.CtInvocation, paras.Domains, paras.Targets,
		paras.Supplier)
}

// BuildInvokeMessage is used to build invoke raw message for send transaction proposal.
// TODO Remove
func (msg *ContractRawMessage) BuildInvokeMessage(chainID string, name string, function string,
	args []string) (*common.RawMessage, error) {
	contractInvocation := msg.BuildContractInvocation(name, function, args)
	return msg.BuildInvokeRawMsg(chainID, contractInvocation, nil)
}

// BuildTxRawMsg is used to build tx raw message for send transaction.
func (msg *ContractRawMessage) BuildTxRawMsg(rawMessages []*common.RawMessage) (*TxRawMsg, error) {
	tx, err := msg.BuildTransaction(rawMessages)
	if err != nil {
		return nil, errors.WithMessage(err, "build transaction error")
	}
	return msg.builder.GetTxRawMsg(tx)
}

// BuildTxRawMsgWithProxy is used to build tx raw message with proxy propose node.
func (msg *ContractRawMessage) BuildTxRawMsgWithProxy(rawMessages []*common.RawMessage,
	target *common.ProposeTarget) (*TxRawMsg, error) {
	if target.Proposer == "" {
		return nil, errors.New("propose target is empty")
	}
	tx, err := msg.BuildTransaction(rawMessages)
	if err != nil {
		return nil, errors.WithMessage(err, "build transaction error")
	}
	return msg.builder.GetTxRawMsgWithTarget(tx, target)
}

// BuildTxHeader is used to build tx header.
func (msg *ContractRawMessage) BuildTxHeader(chainID string, latestBlkNum uint64,
	domains []string) (*common.TxHeader, error) {
	filteredDomains := filterDomains(domains)
	valid, err := validateDomains(filteredDomains)
	if !valid || err != nil {
		return nil, errors.WithMessage(err, "validate domains error")
	}
	now := utils.GenerateTimestamp()
	nonce := utils.GenerateNonce()
	return &common.TxHeader{
		Type:        common.COMMON_TRANSACTION,
		ChainId:     chainID,
		Creator:     msg.builder.BuildIdentity(),
		Timestamp:   now,
		Nonce:       nonce,
		LatestBlock: latestBlkNum,
		Domains:     filteredDomains,
		Version:     version.PlatformVersion,
	}, nil
}

// BuildContractInvocation is used to build ContractInvocation for invoke stage.
func (msg *ContractRawMessage) BuildContractInvocation(name string, function string,
	args []string) *common.ContractInvocation {
	var argsArrays [][]byte
	for _, arg := range args {
		argsArrays = append(argsArrays, []byte(arg))
	}
	return &common.ContractInvocation{
		ContractName: name,
		FuncName:     function,
		Args:         argsArrays,
	}
}

// BuildInvocation is used to build invocation for invoke stage.
func (msg *ContractRawMessage) BuildInvocation(chainID string, contractInvocation *common.ContractInvocation,
	latestBlkNum uint64, domains []string) (*nodeservice.Invocation, error) {
	header, err := msg.BuildTxHeader(chainID, latestBlkNum, domains)
	if err != nil {
		return nil, errors.WithMessage(err, "build tx header error")
	}
	return &nodeservice.Invocation{
		Header:     header,
		Parameters: contractInvocation,
	}, nil
}

// BuildTransaction is used to build transaction.
func (msg *ContractRawMessage) BuildTransaction(rawMessages []*common.RawMessage) (*common.Transaction, error) {
	var approvals []*common.Approval
	var payload []byte
	var hash []byte
	for _, rawMsg := range rawMessages {
		response := &common.Response{}
		if err := proto.Unmarshal(rawMsg.Payload, response); err != nil {
			return nil, errors.WithMessage(err, "unmarshal raw message error")
		}
		if response.Status != common.SUCCESS && response.Status != common.HASH {
			return nil, errors.Errorf("Invoke response status is: %v, status info: %v", response.Status.String(),
				response.StatusInfo)
		}
		tx := &common.Transaction{}
		if err := proto.Unmarshal(response.Payload, tx); err != nil {
			return nil, errors.WithMessage(err, "unmarshal transaction error")
		}
		approvals = append(approvals, tx.Approvals...)
		if response.Status == common.HASH {
			if len(hash) == 0 {
				hash = tx.Payload
			} else if !bytes.Equal(hash, tx.Payload) {
				return nil, errors.New("hash of response is not the same")
			}
			continue
		}
		if payload == nil {
			payload = tx.Payload
		} else if !bytes.Equal(tx.Payload, payload) {
			return nil, errors.New("payload of transactions are not the same")
		}
	}
	if len(payload) == 0 {
		return nil, errors.New("must invoke on at least one node that's org is the same with client")
	}
	if len(hash) != 0 && !bytes.Equal(hash, msg.builder.Hash(payload)) {
		return nil, errors.New("hash and payload of transactions are not the same")
	}
	return &common.Transaction{Payload: payload, Approvals: approvals}, nil
}

// ParseQueryResult get result from rawMessage of query
func (msg *ContractRawMessage) ParseQueryResult(rawMessage *common.RawMessage) ([]byte, error) {
	response := &common.Response{}
	if err := proto.Unmarshal(rawMessage.Payload, response); err != nil {
		return nil, errors.WithMessage(err, "unmarshal invoke response error")
	}
	if response.Status != common.SUCCESS && response.Status != common.HASH {
		return nil, errors.Errorf("Status is: %s, Status info is: %s",
			response.Status.String(), response.StatusInfo)
	}
	tx := &common.Transaction{}
	if err := proto.Unmarshal(response.Payload, tx); err != nil {
		return nil, errors.WithMessage(err, "unmarshal transaction error")
	}
	if response.Status == common.HASH {
		return nil, errors.Errorf("Status is: %s. Payload is: %s",
			response.Status.String(), hex.EncodeToString(tx.Payload))
	}
	txPayLoad := &common.TxPayload{}
	if err := proto.Unmarshal(tx.Payload, txPayLoad); err != nil {
		return nil, errors.WithMessage(err, "unmarshal tx payload error")
	}
	txData := &common.CommonTxData{}
	if err := proto.Unmarshal(txPayLoad.Data, txData); err != nil {
		return nil, errors.WithMessage(err, "unmarshal common tx data error")
	}
	return txData.Response.Payload, nil
}

func (msg *ContractRawMessage) buildInvokeRawMsg(chainID string, contractInvocation *common.ContractInvocation,
	supplier func() (uint64, error), domains []string, msgType common.RawMessage_Type) (*common.RawMessage, error) {
	var latestBlkNum uint64
	var err error
	if supplier != nil {
		latestBlkNum, err = supplier()
		if err != nil {
			return nil, err
		}
	}

	dms := domains
	if len(dms) == 0 {
		dms = append(dms, "/**")
	}

	invocation, err := msg.BuildInvocation(chainID, contractInvocation, latestBlkNum, dms)
	if err != nil {
		return nil, errors.WithMessage(err, "build invocation error")
	}
	payload, err := proto.Marshal(invocation)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal invocation error")
	}
	rawMsg, err := msg.builder.GetRawMessage(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "get raw message error")
	}
	rawMsg.Type = msgType
	return rawMsg, nil
}

// BuildInvokeRawMsgCombined is used to build invoke raw message.
func (msg *ContractRawMessage) buildInvokeRawMsgWithTargets(chainID string,
	contractInvocation *common.ContractInvocation, domains []string,
	targets []*common.EndorseTarget, supplier func() (uint64, error)) ([]*common.RawMessage, error) {
	if targets == nil {
		raw, err := msg.buildInvokeRawMsg(chainID, contractInvocation, supplier,
			domains, common.DIRECT)
		if err != nil {
			return nil, errors.WithMessage(err, "build invoke raw message error")
		}
		return []*common.RawMessage{raw}, nil
	}

	msgArray := make([]*common.RawMessage, len(targets))
	rawMsg, err := msg.buildInvokeRawMsg(chainID, contractInvocation, supplier,
		domains, common.PROXY)
	if err != nil {
		return nil, errors.WithMessage(err, "build invoke raw message error")
	}
	for idx, target := range targets {
		proxyInfoBytes, err := proto.Marshal(target)
		if err != nil {
			return nil, errors.WithMessage(err, "marshal endorse target error")
		}
		rawMsg.ProxyInfo = proxyInfoBytes
		msgArray[idx] = rawMsg
	}
	return msgArray, nil
}

func filterDomains(domains []string) []string {
	var filteredDomains []string
	for _, domain := range domains {
		if domain != "" {
			filteredDomains = append(filteredDomains, domain)
		}
	}
	return filteredDomains
}

func validateDomains(domains []string) (bool, error) {
	for _, domain := range domains {
		if !strings.HasPrefix(domain, "/") {
			return false, errors.Errorf("domain has to start with \"/\", but got domain: %v", domain)
		}
	}
	return true, nil
}
