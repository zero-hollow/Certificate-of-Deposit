/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package rawmessage

import (
	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"github.com/gogo/protobuf/proto"
)

const (
	// UpdateHandler is config update vote handler name
	UpdateHandler = "update"
	// LifecycleHandler is the lifecycle handler string
	LifecycleHandler = "lifecycle"

	start      = "start"
	freeze     = "freeze"
	unfreeze   = "unfreeze"
	destroy    = "destroy"
	revocation = "revocation"
)

// QueryRawMessage is the definition of QueryRawMessage.
type QueryRawMessage struct {
	builder MsgBuilder
}

// NewQueryRawMessage is used to create an instance of query raw message by specifying an message builder.
func NewQueryRawMessage(builder MsgBuilder) *QueryRawMessage {
	return &QueryRawMessage{builder: builder}
}

// BuildTxRawMessage is used to build tx raw message for query transaction.
func (msg *QueryRawMessage) BuildTxRawMessage(chainID string, txHash []byte) (*common.RawMessage, error) {
	request := &nodeservice.ChainServiceRequest{ChainId: chainID,
		Type: &nodeservice.ChainServiceRequest_TxHash{TxHash: txHash}}
	return msg.buildQueryRawMessage(request)
}

// BuildLatestChainStateRawMessage is used to build latest chain state raw message for query latest chain state.
func (msg *QueryRawMessage) BuildLatestChainStateRawMessage(chainID string) (*common.RawMessage, error) {
	request := &nodeservice.ChainServiceRequest{ChainId: chainID}
	return msg.buildQueryRawMessage(request)
}

// BuildBlockRawMessage is used to build block raw message for query block.
func (msg *QueryRawMessage) BuildBlockRawMessage(chainID string, blockNum uint64) (*common.RawMessage, error) {
	request := &nodeservice.ChainServiceRequest{ChainId: chainID,
		Type: &nodeservice.ChainServiceRequest_BlockNum{BlockNum: blockNum}}
	return msg.buildQueryRawMessage(request)
}

// BuildBlockAndResultRawMessage is used to build block and result raw message for query block.
func (msg *QueryRawMessage) BuildBlockAndResultRawMessage(chainID string, blockNum uint64) (*common.RawMessage,
	error) {
	return msg.BuildBlockRawMessage(chainID, blockNum)
}

// BuildContractRawMessage is used to build contract raw message for query contract.
func (msg *QueryRawMessage) BuildContractRawMessage(chainID string, contract string) (*common.RawMessage, error) {
	request := &nodeservice.ChainServiceRequest{ChainId: chainID,
		Type: &nodeservice.ChainServiceRequest_Contract{Contract: contract}}
	return msg.buildQueryRawMessage(request)
}

// BuildQueryChainUpdateVoteRawMessage is used to build query chain update vote raw message for query chain update vote.
func (msg *QueryRawMessage) BuildQueryChainUpdateVoteRawMessage(chainID string,
	subject string) (*common.RawMessage, error) {
	s, err := Subject(subject).conv()
	if err != nil {
		return nil, err
	}
	request := &nodeservice.VoteQuery{ChainId: chainID, Handler: UpdateHandler, Subject: s}
	return msg.buildQueryRawMessage(request)
}

// BuildQueryLifecycleVoteRawMessage is used to build query lifecycle vote raw message for query lifecycle vote.
func (msg *QueryRawMessage) BuildQueryLifecycleVoteRawMessage(chainID string, contract string,
	option string) (*common.RawMessage, error) {
	if option != start && option != freeze && option != unfreeze && option != destroy {
		return nil, errors.Errorf("support option:[%s,%s,%s,%s]", start, freeze, unfreeze, destroy)
	}
	request := &nodeservice.VoteQuery{
		ChainId: chainID,
		Handler: LifecycleHandler,
		Subject: option + "_" + contract,
	}
	return msg.buildQueryRawMessage(request)
}

func (msg *QueryRawMessage) buildQueryRawMessage(pb proto.Message) (*common.RawMessage, error) {
	payload, err := proto.Marshal(pb)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal vote query error")
	}
	return msg.builder.GetRawMessage(payload)
}

// Subject of vote configuration.
type Subject string

type innerSubject string

const (
	// SubjectPolicy is subject of configuration voting policy.
	SubjectPolicy Subject = "policy"
	// SubjectLifecycle is subject of lifecycle voting.
	SubjectLifecycle Subject = "lifecycle"
	// SubjectOrg is subject of organization configuration.
	SubjectOrg Subject = "org"
	// SubjectConsensus is subject of consensus configuration change voting.
	SubjectConsensus Subject = "consensus"
	// SubjectDomain is subject of domain configuration.
	SubjectDomain Subject = "domain"
	// SubjectZone is subject of zone configuration.
	SubjectZone Subject = "zone"

	subjectPolicy    innerSubject = "cfgPolicy"
	subjectLifecycle innerSubject = "lifecycle"
	subjectOrg       innerSubject = "organization"
	subjectConsensus innerSubject = "consensus_configuration_change_pending"
	subjectDomain    innerSubject = "DOMAIN"
	subjectZone      innerSubject = "ZONE"
)

func (s Subject) conv() (string, error) {
	switch s {
	case SubjectPolicy:
		return string(subjectPolicy), nil
	case SubjectLifecycle:
		return string(subjectLifecycle), nil
	case SubjectOrg:
		return string(subjectOrg), nil
	case SubjectConsensus:
		return string(subjectConsensus), nil
	case SubjectDomain:
		return string(subjectDomain), nil
	case SubjectZone:
		return string(subjectZone), nil
	default:
		return "", errors.New("unsupported subject " + string(s))
	}
}
