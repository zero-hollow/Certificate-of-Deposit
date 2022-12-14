/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package rawmessage

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/sdk/crypto"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/contract"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"github.com/gogo/protobuf/proto"
)

// LifecycleRawMessage is the definition of LifecycleRawMessage.
type LifecycleRawMessage struct {
	builder MsgBuilder
	crypto  crypto.Crypto
}

// Contract is the definition of contract.
type Contract struct {
	chainID string
	name    string
	version string
}

type voteParams struct {
	c              *Contract
	desc           string
	policy         string
	schema         string
	historySupport bool
	initRequired   bool
}

// NewLifecycleRawMessage is used to create an instance of lifecycle raw message by specifying an message builder.
func NewLifecycleRawMessage(builder MsgBuilder, crypto crypto.Crypto) *LifecycleRawMessage {
	return &LifecycleRawMessage{builder: builder, crypto: crypto}
}

// NewContract is used to create an instance of contract by specifying chain id, contract name and contract version.
func NewContract(chainID string, name string, version string) *Contract {
	return &Contract{chainID: chainID, name: name, version: version}
}

// BuildImportRawMessage is used to build import raw message for import contract.
func (msg *LifecycleRawMessage) BuildImportRawMessage(c *Contract, path string,
	sandbox string, language string) (*common.RawMessage, error) {
	if err := checkParams(c.chainID, c.name, c.version); err != nil {
		return nil, errors.WithMessage(err, "check params error")
	}
	lowerSandbox := strings.ToLower(sandbox)
	var sandboxType common.ContractRunEnv
	switch lowerSandbox {
	case "docker":
		sandboxType = common.Docker
	case "native":
		sandboxType = common.Native
	case "nativewasm":
		sandboxType = common.NativeWasm
	case "teewasm":
		sandboxType = common.TEEWasm
	default:
		return nil, errors.New("unsupported sandbox type")
	}

	codes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.WithMessage(err, "read contract code error")
	}
	externalContractInfo := &contract.ExternalContractInfo{
		SandboxType:   sandboxType,
		Code:          codes,
		SchemaVersion: c.version,
		Language:      language,
	}
	contractInfoType := &nodeservice.ContractInfo_ExternalContractInfo{ExternalContractInfo: externalContractInfo}
	contractInfo := &nodeservice.ContractInfo{ChainId: c.chainID, ContractName: c.name, Type: contractInfoType}
	return msg.buildImportRawMsg(contractInfo)
}

// BuildImportFabricRawMessage is used to build import raw message for import fabric contract.
func (msg *LifecycleRawMessage) BuildImportFabricRawMessage(c *Contract, path string,
	sandbox string, language string) (*common.RawMessage, error) {
	if err := checkParams(c.chainID, c.name, c.version); err != nil {
		return nil, errors.WithMessage(err, "check params error")
	}
	lowerSandbox := strings.ToLower(sandbox)
	var sandboxType common.ContractRunEnv
	if lowerSandbox != "docker" {
		return nil, errors.New("unsupported sandbox type")
	}
	sandboxType = common.Docker
	codes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.WithMessage(err, "read contract code error")
	}
	ext := map[string][]byte{
		"contractType": []byte("fabric"),
	}
	externalContractInfo := &contract.ExternalContractInfo{SandboxType: sandboxType, Code: codes,
		SchemaVersion: c.version, Language: language, Extensions: ext}
	contractInfoType := &nodeservice.ContractInfo_ExternalContractInfo{ExternalContractInfo: externalContractInfo}
	contractInfo := &nodeservice.ContractInfo{ChainId: c.chainID,
		ContractName: c.name, Type: contractInfoType}
	return msg.buildImportRawMsg(contractInfo)
}

// BuildUnImportRawMessage is used to build unimport raw message for unimport contract.
func (msg *LifecycleRawMessage) BuildUnImportRawMessage(c *Contract) (*common.RawMessage, error) {
	externalContractInfo := &contract.ExternalContractInfo{
		SchemaVersion: c.version,
	}
	contractInfoType := &nodeservice.ContractInfo_ExternalContractInfo{
		ExternalContractInfo: externalContractInfo,
	}
	contractInfo := &nodeservice.ContractInfo{
		ChainId:      c.chainID,
		ContractName: c.name,
		Type:         contractInfoType,
	}
	return msg.buildImportRawMsg(contractInfo)
}

// BuildVoteRawMessage is used to build vote raw message for vote contract.
func (msg *LifecycleRawMessage) BuildVoteRawMessage(c *Contract, desc string, policy string,
	historySupport bool, initRequired bool) (*TxRawMsg, error) {
	if err := checkParams(c.chainID, c.name, c.version); err != nil {
		return nil, errors.WithMessage(err, "check params error")
	}
	vp := &voteParams{
		c:              c,
		desc:           desc,
		policy:         policy,
		schema:         "",
		historySupport: historySupport,
		initRequired:   initRequired,
	}
	rawMsg, err := msg.buildVoteRawMessage(vp)
	if err != nil {
		return nil, errors.WithMessage(err, "build vote raw message error")
	}
	return rawMsg, nil
}

// BuildSQLVoteRawMessage is used to build SQL vote raw message for vote contract.
func (msg *LifecycleRawMessage) BuildSQLVoteRawMessage(c *Contract, desc string, policy string,
	schema string, historySupport bool) (*TxRawMsg, error) {
	if err := checkParams(c.chainID, c.name, c.version); err != nil {
		return nil, errors.WithMessage(err, "check params error")
	}
	voteParams := &voteParams{
		c:              c,
		desc:           desc,
		policy:         policy,
		schema:         schema,
		historySupport: historySupport,
		initRequired:   false,
	}
	rawMsg, err := msg.buildVoteRawMessage(voteParams)
	if err != nil {
		return nil, errors.WithMessage(err, "build sql vote raw message error")
	}
	return rawMsg, nil
}

func (msg *LifecycleRawMessage) buildVoteRawMessage(vp *voteParams) (*TxRawMsg, error) {
	validatorExtensions := make(map[string][]byte)
	validatorExtensions["policy"] = []byte(vp.policy)
	contractDefinition := &contract.ContractDefinition{
		ContractName:        vp.c.name,
		SchemaVersion:       vp.c.version,
		SequenceNumber:      0,
		Description:         vp.desc,
		HistorySupport:      vp.historySupport,
		RequireInit:         vp.initRequired,
		ApprovalValidator:   "default",
		ValidatorExtensions: validatorExtensions}
	if vp.schema != "" {
		contractDefinition.SqlDbSchema = vp.schema
	}

	contractDefinitionBytes, err := proto.Marshal(contractDefinition)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal contract definition error")
	}

	voteTxData := &common.VoteTxData{
		Handler: "lifecycle",
		Subject: "start",
		Payload: contractDefinitionBytes,
	}

	tx, err := msg.builder.BuildVoteTx(vp.c.chainID, voteTxData)
	if err != nil {
		return nil, errors.WithMessage(err, "build vote tx error")
	}
	return msg.builder.GetTxRawMsg(tx)
}

// BuildManageRawMessage is used to build contract manage vote raw message for vote contract.
func (msg *LifecycleRawMessage) BuildManageRawMessage(chain string, contract string,
	option string) (*TxRawMsg, error) {
	if option != freeze && option != unfreeze && option != destroy {
		return nil, errors.Errorf("support option:[%s,%s,%s]", freeze, unfreeze, destroy)
	}
	voteTxData := &common.VoteTxData{
		Handler: "lifecycle",
		Subject: option,
		Payload: []byte(contract),
	}
	tx, err := msg.builder.BuildVoteTx(chain, voteTxData)
	if err != nil {
		return nil, errors.WithMessage(err, "build vote tx error")
	}
	return msg.builder.GetTxRawMsg(tx)
}

// BuildQueryStateRawMessage is used to build query contract state raw message for import contract.
func (msg *LifecycleRawMessage) BuildQueryStateRawMessage(chain string, contract string) (*common.RawMessage, error) {
	contractInfo := &nodeservice.ContractInfo{
		ChainId:      chain,
		ContractName: contract,
	}
	payload, err := proto.Marshal(contractInfo)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal contract info error")
	}
	return msg.builder.GetRawMessage(payload)
}

func (msg *LifecycleRawMessage) buildImportRawMsg(
	contractInfo *nodeservice.ContractInfo) (*common.RawMessage, error) {
	if contractInfo.ChainId == "" {
		return nil, errors.New("contract info field chain id is empty")
	}
	payload, err := proto.Marshal(contractInfo)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal contract info error")
	}
	rawMsg, err := msg.builder.GetRawMessage(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "get raw message error")
	}
	return rawMsg, nil
}
