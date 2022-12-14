/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package rawmessage

import (
	"git.huawei.com/huaweichain/sdk/utils"
	"github.com/pkg/errors"

	"github.com/gogo/protobuf/proto"

	"git.huawei.com/huaweichain/common/cryptomgr/cryptoimpl"
	"git.huawei.com/huaweichain/common/logger"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"git.huawei.com/huaweichain/sdk/config"
	"git.huawei.com/huaweichain/sdk/crypto"
	"git.huawei.com/huaweichain/sdk/genesisblock"
)

var log = logger.GetModuleLogger("go sdk", "raw message")

const (
	// HANDLER is the key value of handler.
	HANDLER = "config"
	// BlockNumber is the specified block number.
	BlockNumber = 0
)

// ChainRawMessage is the definition of ChainRawMessage.
type ChainRawMessage struct {
	builder MsgBuilder
	crypto  crypto.Crypto
}

// NewChainRawMessage is used to create an instance of chain raw message by specifying an message builder.
func NewChainRawMessage(builder MsgBuilder, crypto crypto.Crypto) *ChainRawMessage {
	return &ChainRawMessage{builder: builder, crypto: crypto}
}

// BuildGenesisBlock is used to build genesis block for create chain by specifying genesis block config file path.
func (msg *ChainRawMessage) BuildGenesisBlock(chainID string,
	genesisConfigPath string, decrypts ...func(bytes []byte) ([]byte, error)) (*common.Block, error) {
	if err := checkChainID(chainID); err != nil {
		return nil, errors.WithMessage(err, "check chain id error")
	}
	genesisConfig, err := config.NewGenesisConfig(genesisConfigPath, decrypts...)
	if err != nil {
		return nil, errors.WithMessage(err, "new genesis config error")
	}

	return msg.BuildGenesisBlockWithGenesisConfig(chainID, genesisConfig)
}

// BuildGenesisBlockWithGenesisConfig is used to build genesis block with genesis config.
func (msg *ChainRawMessage) BuildGenesisBlockWithGenesisConfig(chainID string,
	genesisConfig *config.GenesisConfig) (*common.Block, error) {
	chainConfig, err := genesisblock.GetChainConfig(genesisConfig, chainID)
	if err != nil {
		return nil, errors.WithMessage(err, "get chain config error")
	}
	tx, err := msg.getVoteTransaction(chainID, chainConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "get transaction error")
	}

	txHash := cryptoimpl.HashSha256(tx.Payload)
	t := &common.Tx{
		Hash: txHash,
		Data: &common.Tx_Full{Full: tx},
	}
	blockBody := &common.BlockBody{TxList: []*common.Tx{t}}
	bodyBytes, err := proto.Marshal(blockBody)
	if err != nil {
		return nil, errors.WithMessage(err, "block body marshal error")
	}

	header := &common.BlockHeader{
		Number:    BlockNumber,
		BodyHash:  txHash,
		Timestamp: int64(utils.GenerateTimestamp()),
		Version:   chainConfig.MinPlatformVersion,
	}
	genesisBlock := &common.Block{Header: header, Body: bodyBytes}
	return genesisBlock, nil
}

// BuildQuitChainRawMessage is used to build quit chain raw message for quit chain by specifying chain id.
func (msg *ChainRawMessage) BuildQuitChainRawMessage(chainID string) (*common.RawMessage, error) {
	deleteInfo := &nodeservice.DeleteInfo{ChainId: chainID}
	payload, err := proto.Marshal(deleteInfo)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal delete info error")
	}
	rawMsg, err := msg.builder.GetRawMessage(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "get raw message error")
	}
	return rawMsg, nil
}

// BuildJoinChainRawMessage is used to build join chain raw message for join chain by specifying genesis block bytes.
func (msg *ChainRawMessage) BuildJoinChainRawMessage(genesisBlockBytes []byte) (*common.RawMessage, error) {
	return msg.BuildJoinMsgWithLatestConf(genesisBlockBytes, nil, getDefaultInitialEntrypoint())
}

// BuildJoinMsgWithEntrypoint is used to build join chain raw message with config info.
func (msg *ChainRawMessage) BuildJoinMsgWithEntrypoint(genesisBlockBytes []byte,
	entrypoint *common.Entrypoint) (*common.RawMessage, error) {
	return msg.BuildJoinMsgWithLatestConf(genesisBlockBytes, nil, entrypoint)
}

// BuildJoinMsgWithLatestConf is used to build join chain raw message with config info and entrypoint.
func (msg *ChainRawMessage) BuildJoinMsgWithLatestConf(genesisBlockBytes []byte,
	latestConf *common.ConfigInfo, entrypoint *common.Entrypoint) (*common.RawMessage, error) {
	if entrypoint == nil {
		entrypoint = getDefaultInitialEntrypoint()
		log.Debugf("default initial entrypoint setted, default zone id: %v", entrypoint.ZoneId)
	}
	if err := config.CheckEntryPoint(entrypoint); err != nil {
		return nil, err
	}
	entrypointBytes, err := proto.Marshal(entrypoint)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal entrypoint error")
	}
	createInfo := &nodeservice.CreateInfo{
		GenesisBlock: genesisBlockBytes,
		Config:       latestConf,
		Entrypoint:   entrypointBytes,
	}
	return msg.buildJoinChainRawMessage(createInfo)
}

// BuildQueryChainRawMessage is used to build query chain raw message for query chain by specifying chain id.
func (msg *ChainRawMessage) BuildQueryChainRawMessage(chainID string) (*common.RawMessage, error) {
	return msg.buildQueryChainRawMessage(chainID)
}

// BuildQueryAllChainRawMessage is used to build query all chains raw message for query all chains.
func (msg *ChainRawMessage) BuildQueryAllChainRawMessage() (*common.RawMessage, error) {
	return msg.buildQueryChainRawMessage()
}

func (msg *ChainRawMessage) buildJoinChainRawMessage(createInfo *nodeservice.CreateInfo) (*common.RawMessage, error) {
	if createInfo.GenesisBlock == nil {
		return nil, errors.New("CreateInfo.GenesisBlock is nil")
	}
	payload, err := proto.Marshal(createInfo)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal create info error")
	}
	rawMsg, err := msg.builder.GetRawMessage(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "get raw message error")
	}
	return rawMsg, nil
}

func (msg *ChainRawMessage) buildQueryChainRawMessage(chainIDs ...string) (*common.RawMessage, error) {
	queryInfo := &nodeservice.QueryInfo{}
	if len(chainIDs) > 0 {
		queryInfo.ChainId = chainIDs[0]
	}
	payload, err := proto.Marshal(queryInfo)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal query info error")
	}
	rawMsg, err := msg.builder.GetRawMessage(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "get raw message error")
	}
	return rawMsg, nil
}

func (msg *ChainRawMessage) getVoteTransaction(chainID string,
	chainConfig *common.ChainConfig) (*common.Transaction, error) {
	if chainConfig.ChainId == "" {
		return nil, errors.New("chain id is empty")
	}
	chainConfigBytes, err := proto.Marshal(chainConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal chain config error")
	}
	voteTxData := &common.VoteTxData{
		Handler: HANDLER,
		Payload: chainConfigBytes,
	}
	return msg.builder.BuildVoteTx(chainID, voteTxData)
}

func getDefaultInitialEntrypoint() *common.Entrypoint {
	return &common.Entrypoint{
		ZoneId:        "",
		Coordinator:   true,
		InitialMaster: true,
	}
}
