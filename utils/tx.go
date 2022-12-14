// tx.go 交易相关功能
package utils

import (
	"encoding/hex"
	"fmt"

	"git.huawei.com/huaweichain/common/cryptomgr"
	"git.huawei.com/huaweichain/common/cryptomgr/ecdsaalg"
	"git.huawei.com/huaweichain/common/cryptomgr/gmalg"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/sdk/client"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

type TxTool struct {
}

// common.Tx struct
/*
Tx
  hash             // 类型: bytes，本交易hash值
  data             // 类型: oneof{Transaction, CompactTx}，交易体
    full           // 类型: Transaction，正常交易数据
      payload      // 类型: bytes，由TxPayload序列化得到，交易信息
        header                 // 类型: TxHeader，交易头
          type                 // 类型: 枚举TxType{COMMON_TRANSACTION, VOTE_TRANSACTION}，交易类型
          chain_id             // 类型: string，链名
          creator              // 类型: Identity，交易创建者信息
            org                // 类型: string，创建者组织名称
            type               // 类型: 枚举Type{COMMON_NAME, CERT}，创建者id类型
            id                 // 类型: bytes，创建者id
          timestamp            // 类型: uint64，交易创建时间戳
          nonce                // 类型: uint64，随机值
          latest_block         // 类型: uint64，交易创建时可获取的最高区块高度
        data                   // 类型: bytes，由CommonTxData或VoteTxData序列化得到，依据header中的type指定
          contractInvocation   // 类型: bytes，由ContractInvocation序列化得到，合约调用信息
            contract_name      // 类型: string，调用合约名称
            func_name          // 类型: string，调用合约方法
            args               // 类型: bytes列表，调用合约入参
          response             // 类型: InvocationResponse，交易背书返回
            status             // 类型: 枚举Status，详见后文
            status_info        // 类型: string，status说明
            payload            // 类型: bytes，合约执行返回值
          stateUpdates         // 类型: StateUpdates列表，交易执行生成的读写集
            namespace          // 类型: string，交易读写集命名空间
            updates            // 类型: oneof{KvStateUpdates, SqlStateUpdates}，具体读写集信息
              kvUpdates        // 类型: KvStateUpdates，kv类型具体读写集信息
                key_versions   // 类型: KeyVersion列表，读集信息
                  key          // 类型: string，键
                  version      // 类型: Version，读集版本
                    block_num  // 类型: uint64，读取key的块高
                    tx_num     // 类型: int32，读取key的交易序号
                updates        // 类型: KeyValue列表，写集信息
                  key          // 类型: string，键
                  value        // 类型: bytes，写入值
                  signature    // 类型: bytes，TEE可信合约签名
                deletes        // 类型: string列表，删除key的集合
      approvals    // 类型: Approval列表，背书信息列表
        identity   // 类型: bytes，背书节点证书
        sign       // 类型: bytes，背书节点签名
        type       // 类型: 枚举ContractRunEnv{Docker,Native,NativeWasm,TEEWasm}，背书类型，决定签名验签方式
        org_name   // 类型: string，背书节点组织信息
        node_name  // 类型: string，背书节点名称
    compact        // 类型: CompactTx，子账本特性对交易的压缩表示
        height     // 类型: int32
        begin      // 类型: int32
        end        // 类型: int32
*/

/**
 *  GetTxID
 *  @Description: 获取交易的ID信息
 *  @return 交易的ID信息
 */
func (tt *TxTool) GetTxID(tx common.Tx) string {
	return Hash2str(tx.Hash)
}

/**
 *  GetTimestamp
 *  @Description: 获取交易的时间戳
 *  @return 交易的时间戳
 */
func (tt *TxTool) GetTimestamp(tx common.Tx) (string, error) {
	txPayLoadHeader, err := getTxPayloadHeader(&tx)
	if err != nil {
		return "", err
	}
	return FormatTime(ParseNanosecond(int64(txPayLoadHeader.Timestamp))), nil
}

/**
 *  GetContractName
 *  @Description: 获取交易 Tx 使用的合约的名称
 *  @return string 获取交易 Tx 使用的合约的名称
 */
func (tt *TxTool) GetContractName(tx common.Tx) (string, error) {
	txData, err := getCommonTxData(&tx)
	if err != nil {
		return "", err
	}

	contractData := &common.ContractInvocation{}
	if err := proto.Unmarshal(txData.ContractInvocation, contractData); err != nil {
		return "", errors.WithMessage(err, "unmarshal common tx data error")
	}
	return contractData.ContractName, nil
}

/**
 *  GetEndorsersOrg
 *  @Description: 获取给交易 Tx 背书的组织
 *  @return string 交易Tx的背书组织名称
 */
func (tt *TxTool) GetEndorsersOrg(tx common.Tx) ([]string, error) {
	var endorsersOrg []string
	approVals := tx.GetFull().Approvals
	for _, approval := range approVals {
		endorsersOrg = append(endorsersOrg, approval.NodeName)
	}
	return endorsersOrg, nil
}

/**
 *  GetCreateOrg
 *  @Description: 获取创建交易 Tx 的组织
 *  @return string 创建交易Tx的组织名称
 */
func (tt *TxTool) GetCreateOrg(tx common.Tx) (string, error) {
	txPayLoadHeader, err := getTxPayloadHeader(&tx)
	if err != nil {
		return "", err
	}

	if txPayLoadHeader.Creator != nil {
		return txPayLoadHeader.Creator.Org, nil
	}
	createOrgs, err := tt.GetEndorsersOrg(tx)
	if err != nil {
		return "", errors.WithMessage(err, "get endorse org error")
	}

	for i := range createOrgs {
		return createOrgs[i], nil
	}
	return "", nil
}

/**
 *  GetTxKeyValues
 *  @Description: 获取交易 Tx 中的 key 和 value
 *  @return []string 返回交易中使用的key、value
 */
func (tt *TxTool) GetTxKeyValues(tx common.Tx) ([]string, error) {
	txData, err := getCommonTxData(&tx)
	if err != nil {
		return nil, err
	}

	var keyValues []string
	for i := range txData.StateUpdates {
		kvStateUpdates := (*txData.StateUpdates[i]).GetKvUpdates()
		for j := range kvStateUpdates.Updates {
			keyValue := kvStateUpdates.Updates[j].Key + ": " + string(kvStateUpdates.Updates[j].Value)
			keyValues = append(keyValues, keyValue)
		}
	}
	return keyValues, nil
}

/**
 *  PrintTxInfo
 *  @Description: 打印交易相关的信息
 */
func (tt *TxTool) PrintTxInfo(tx common.Tx) {
	timeStamp, err := tt.GetTimestamp(tx)
	if err != nil {
		fmt.Println(err)
	}
	txID := tt.GetTxID(tx)

	contactName, err := tt.GetContractName(tx)
	if err != nil {
		fmt.Println(err)
	}

	endorsersOrg, err := tt.GetEndorsersOrg(tx)
	if err != nil {
		fmt.Println(err)
	}

	createOrg, err := tt.GetCreateOrg(tx)
	if err != nil {
		fmt.Println(err)
	}

	txKeyValues, err := tt.GetTxKeyValues(tx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("tx`s timeStamp is", timeStamp)
	fmt.Println("tx`s txID is", txID)
	fmt.Println("tx`s contactName is", contactName)
	fmt.Println("tx`s endorsersOrg is", endorsersOrg)
	fmt.Println("tx`s createOrg is", createOrg)
	fmt.Println("tx`s txKeyValues is", txKeyValues)
	fmt.Printf("\n")
}

/**
 *  QueryTxResultByTxID
 *  @Description: 查询交易执行结果，通过交易 ID 查询交易结果
 *  @return 若common.TxResult common.TxResult.Status == Valid && len(common.TxResult.TxHash)!=0，则交易执行成功
 */
func (tt *TxTool) QueryTxResultByTxID(gatewayClient *client.GatewayClient, config Config, txID string) (*common.TxResult, error) {
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
	responseMsg, err := node.QueryAction.GetTxResultByTxHash(rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "query action get tx result by tx id error")
	}

	// 5.解析response获得response.Payload
	payload, err := GetPayloadWithResp(responseMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "parse response msg with struct type error")
	}

	// 6.解析payload获得common.TxResult{}
	txResult := common.TxResult{}
	if err := proto.Unmarshal(payload, &txResult); err != nil {
		return nil, errors.WithMessage(err, "unmarshal transaction error")
	}
	return &txResult, nil
}

/**
 *  QueryTxByTxID
 *  @Description: 利用交易ID查询交易详情
 *  @return common.Tx 返回交易结构体
 */
func (tt *TxTool) QueryTxByTxID(gatewayClient *client.GatewayClient, config Config, txID string) (*common.Tx, error) {
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
	responseMsg, err := node.QueryAction.GetTxByHash(rawMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "query action get tx by tx id error")
	}

	// 5.解析response获得response.Payload
	payload, err := GetPayloadWithResp(responseMsg)
	if err != nil {
		return nil, errors.WithMessage(err, "parse response msg with struct type error")
	}

	// 6.解析payload获得common.Tx{}
	tx := common.Tx{}
	if err := proto.Unmarshal(payload, &tx); err != nil {
		return nil, errors.WithMessage(err, "unmarshal transaction error")
	}
	return &tx, nil
}

func getCommonTxData(tx *common.Tx) (*common.CommonTxData, error) {
	txPayLoad := &common.TxPayload{}
	if err := proto.Unmarshal(tx.GetFull().Payload, txPayLoad); err != nil {
		return nil, errors.WithMessage(err, "unmarshal tx payload error")
	}
	txData := &common.CommonTxData{}
	if txPayLoad.Header.Type.String() != common.COMMON_TRANSACTION.String() {
		return nil, errors.New("transaction type is not support to get commonTxData")
	}
	if err := proto.Unmarshal(txPayLoad.Data, txData); err != nil {
		return nil, errors.WithMessage(err, "unmarshal common tx data error")
	}
	return txData, nil
}

func getTxPayloadHeader(tx *common.Tx) (*common.TxHeader, error) {
	txPayLoad := &common.TxPayload{}
	if err := proto.Unmarshal(tx.GetFull().Payload, txPayLoad); err != nil {
		return nil, errors.WithMessage(err, "unmarshal tx payload error")
	}
	return txPayLoad.Header, nil
}

func getOrgsFromCert(certData []byte, signAlgorithm string) []string {
	var cert cryptomgr.Cert
	var err error

	if signAlgorithm == "ecdsa_with_sha256" {
		cert, err = ecdsaalg.GetCert(certData)
	} else if signAlgorithm == "sm2_with_sm3" {
		cert, err = gmalg.GetCert(certData)
	}

	if err != nil || cert == nil {
		return nil
	}
	return cert.GetOrganization()
}
