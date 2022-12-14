// block.go 区块相关功能
package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"

	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/proto/nodeservice"
	"git.huawei.com/huaweichain/sdk/client"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// common.block struct
/*block
header       	 	// 类型: BlockHeader，区块头
	number      	// 类型: uint64，区块高度
	parent_hash 	// 类型: bytes，父区块的区块头的hash值
	body_hash   	// 类型: bytes，区块体的hash值
	timestamp   	// 类型: int64，区块打包的时间戳
body          		// 类型: bytes，由BlockBody序列化得到，区块体，包含交易列表
	tx_list     	// 类型: Tx列表，由一系列交易构成
*/

type BlockTool struct {
}

/**
 *  GetNumber
 *  @Description: 获取区块的区块号
 *  @return 区块号
 */
func (bt *BlockTool) GetNumber(block common.Block) uint64 {
	return block.Header.Number
}

/**
 *  GetTimestamp
 *  @Description: 获取交易的时间戳
 *  @return 交易的时间戳
 */
func (bt *BlockTool) GetTimestamp(block common.Block) string {
	return FormatTime(ParseNanosecond(block.Header.Timestamp))
}

/**
 *  GetParentHash
 *  @Description: 父区块的区块头的hash值
 *  @return string 父区块区块头的hash值
 */
func (bt *BlockTool) GetParentHash(block common.Block) string {
	return base64.StdEncoding.EncodeToString(block.Header.ParentHash)
}

/**
 *  GetParentHash
 *  @Description: 区块体的hash值
 *  @return string 区块体的hash值
 */
func (bt *BlockTool) GetBodyHash(block common.Block) string {
	return base64.StdEncoding.EncodeToString(block.Header.BodyHash)
}

/**
 *  GetBlockHeader
 *  @Description: 获取区块头
 *  @return *common.BlockHeader 区块头结构体
 */
func (bt *BlockTool) GetBlockHeader(block common.Block) *common.BlockHeader {
	return block.Header
}

/**
 *  GetTxList
 *  @Description: 获取区块上所有交易的ID
 *  @return []*common.Tx 交易ID集合
 */
func (bt *BlockTool) GetTxIdList(block *common.Block) ([]string, error) {
	var txIdList []string
	txList, err := bt.GetTxList(block)
	if err != nil {
		return nil, errors.WithMessage(err, "get tx id list error")
	}
	for i := range txList {
		txIdList = append(txIdList, Hash2str(txList[i].Hash))
	}
	fmt.Println("GetTxIdList ", txIdList)
	return txIdList, nil
}

/**
 *  GetTxList
 *  @Description: 获取区块上的所有交易
 *  @return []*common.Tx 交易集合
 */
func (bt *BlockTool) GetTxList(block *common.Block) ([]*common.Tx, error) {
	blockBody := common.BlockBody{}
	if err := proto.Unmarshal(block.Body, &blockBody); err != nil {
		return nil, errors.WithMessage(err, "unmarshal tx payload error")
	}
	fmt.Println("GetTxList ", len(blockBody.TxList))
	for key, value := range blockBody.TxList {
		fmt.Printf("txList[%d] %s\n", key, Hash2str(value.Hash))
	}
	return blockBody.TxList, nil
}

/**
 *  PrintBlockBasicInfo
 *  @Description: 打印区块基本信息
 */
func (bt *BlockTool) PrintBlockBasicInfo(block common.Block) {
	fmt.Println("block`s number is: ", bt.GetNumber(block))
	fmt.Println("block`s timestamp is: ", bt.GetTimestamp(block))
	fmt.Println("block`s parent hash is: ", string(bt.GetParentHash(block)))
	fmt.Println("block`s body hash is: ", string(bt.GetBodyHash(block)))
}

/**
 *  PrintBlockTxsInfo
 *  @Description: 打印区块上所有交易的数据
 */
func (bt *BlockTool) PrintBlockTxsInfo(block *common.Block) error {
	fmt.Println("PrintBlockTxsInfo")
	txList, err := bt.GetTxList(block)
	if err != nil {
		return errors.WithMessage(err, "get tx id list error")
	}
	transTool := TxTool{}
	for i := range txList {
		fmt.Println(transTool.GetTxID(*txList[i]))
		if kv, err := transTool.GetTxKeyValues(*txList[i]); err == nil {
			fmt.Println(kv)
		}
	}
	return nil
}

/**
 *  QueryLastBlockNumber
 *  @Description: 查询区块块高，查询当前最新区块的区块号
 *  @return uint64 当前最新区块的区块号
 */
func (bt *BlockTool) QueryLastBlockNumber(gatewayClient *client.GatewayClient, config Config) (uint64, error) {
	// 1.消息构建
	rawMsg, err := gatewayClient.QueryRawMessage.BuildLatestChainStateRawMessage(config.ChainID)
	if err != nil {
		return 0, errors.WithMessage(err, "build latest chain state raw message error")
	}

	// 2.获取节点对象
	nodeMap := gatewayClient.Nodes
	node, ok := nodeMap[config.QueryNode]
	if !ok {
		return 0, errors.Errorf("node not exist： %v", config.QueryNode)
	}

	// 3.消息发送
	responseMsg, err := node.QueryAction.GetLatestChainState(rawMsg)
	if err != nil {
		return 0, errors.WithMessage(err, "query action get latest chain state error")
	}

	// 4.解析response获得response.Payload
	payload, err := GetPayloadWithResp(responseMsg)
	if err != nil {
		return 0, errors.WithMessage(err, "parse response msg with struct type error")
	}

	// 5.解析payload获得nodeservice.LatestChainState{}
	latestChainState := &nodeservice.LatestChainState{}
	if err := proto.Unmarshal(payload, latestChainState); err != nil {
		return 0, errors.WithMessage(err, "unmarshal latest chain state error")
	}
	return latestChainState.Height - 1, nil
}

/**
 *  QueryBlockByNumber
 *  @Description: 查询区块详情，区块号从 0 开始计数, 通过区块号查询区块
 *  @param blockNum 区块号
 *  @return common.Block 区块信息
 */
func (bt *BlockTool) QueryBlockByNumber(gatewayClient *client.GatewayClient, config Config, blockNum string) (*common.Block, error) {
	// 1. 入参校验
	if blockNum == "" {
		return nil, errors.Errorf("please specify a block number or transaction id.")
	}
	num, err := strconv.ParseInt(blockNum, 10, 64)
	if err != nil {
		return nil, errors.WithMessage(err, "parse block number error")
	}

	// 2.获取节点对象
	nodeMap := gatewayClient.Nodes
	node, ok := nodeMap[config.QueryNode]
	if !ok {
		return nil, errors.Errorf("node not exist： %v", config.QueryNode)
	}

	// 3.消息构建
	rawMsg, err := gatewayClient.QueryRawMessage.BuildBlockRawMessage(config.ChainID, uint64(num))
	if err != nil {
		return nil, errors.WithMessage(err, "build block raw message error")
	}

	// 4.消息发送
	var responseMsg *common.RawMessage
	responseMsg, err = node.QueryAction.GetBlockByNum(rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "query action get latest chain state error")
	}

	// 5.解析response获得response.Payload
	payload, err := GetPayloadWithResp(responseMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "parse response msg with struct type error")
	}

	// 6.解析payload获得common.Block{}
	block := common.Block{}
	if err := proto.Unmarshal(payload, &block); err != nil {
		return nil, errors.WithMessage(err, "unmarshal block error")
	}

	blockBody := common.BlockBody{}
	if err := proto.Unmarshal(block.Body, &blockBody); err != nil {
		return nil, errors.WithMessage(err, "unmarshal tx payload error")
	}
	return &block, nil
}

/**
 *  QueryBlockByTxID
 *  @Description: 利用交易ID查询块详情，通过交易 ID 查询块信息
 *  @param txID 交易 ID
 *  @return common.Block 区块信息
 *  @return error
 */
func (bt *BlockTool) QueryBlockByTxID(gatewayClient *client.GatewayClient, config Config, txID string) (*common.Block, error) {
	// 1.入参处理
	txHash, err := hex.DecodeString(txID)
	if err != nil {
		return nil, errors.WithMessage(err, "decode tx id")
	}

	// 2.获取节点对象
	nodeMap := gatewayClient.Nodes
	node, ok := nodeMap[config.QueryNode]
	if !ok {
		return nil, errors.Errorf("node not exist： %v", config.QueryNode)
	}

	// 3.消息构建
	rawMsg, err := gatewayClient.QueryRawMessage.BuildTxRawMessage(config.ChainID, txHash)
	if err != nil {
		return nil, errors.WithMessage(err, "build tx raw message error")
	}

	// 4.消息发送
	responseMsg, err := node.QueryAction.GetBlockByTxHash(rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "query action get tx result by tx id error")
	}

	// 5.解析response获得response.Payload
	payload, err := GetPayloadWithResp(responseMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "parse response msg with struct type error")
	}

	// 6.解析payload获得common.Transaction{}
	block := common.Block{}
	if err := proto.Unmarshal(payload, &block); err != nil {
		return nil, errors.WithMessage(err, "unmarshal transaction error")
	}
	return &block, nil
}
