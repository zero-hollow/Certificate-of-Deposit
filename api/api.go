package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"git.huawei.com/goclient/utils"
	"git.huawei.com/huaweichain/sdk/client"
	"github.com/deatil/go-cryptobin/cryptobin/crypto"
)

type Message struct {
	BlockHeight int    //区块高度
	TxHash      string //交易hash
	Value       string //上链的内容
	Timestamp   string //时间戳
}

const Key = "dfertf12dfertf12" //上链

func SaveRecode(datakey string, data string) (string, error) {
	gatewayClient, err := client.NewGatewayClient(utils.AppConfig().ConfigFilePath)
	fmt.Println(err)
	if err != nil {
		fmt.Println("init new gateway client error")
		return "", utils.ErrorNew(604, "上链失败，建议重试")
	}
	//背书节点
	endorseNodes := strings.Split(utils.AppConfig().EndorserNodes, ",")
	//创建网络
	net, err := utils.NewNodes(gatewayClient, endorseNodes, utils.AppConfig().ConsensusNode)
	if err != nil {
		fmt.Println("new nodes network error")
		return "", utils.ErrorNew(604, "上链失败，建议重试")
	}
	if err != nil {
		fmt.Println(err)
		return "", utils.ErrorNew(604, "上链失败，建议重试")
	}
	//交易
	txEvent, err := net.EventListener.EventAction.GetTxEventService(utils.AppConfig().ChainID)
	if err != nil {
		fmt.Println("event action register tx event error")
		return "", utils.ErrorNew(604, "上链失败，建议重试")
	}
	defer txEvent.Close()

	utils.SetTxEvent(txEvent)

	//创建hash值
	var txHashID string
	//_, txHashID, err = utils.Send(gatewayClient, net, utils.AppConfig(), "saveRecode",phone+" "+timestamp+";"+data)
	_, txHashID, err = utils.Send(gatewayClient, net, utils.AppConfig(), "saveRecode", datakey+";"+data)
	if err != nil {
		fmt.Println(err)
		return "", utils.ErrorNew(604, "上链失败，建议重试")
	}
	return txHashID, nil
}

// 根据手机号查询
func QueryByPhone(txHashs []string) (string, error) {
	gatewayClient, err := client.NewGatewayClient(utils.AppConfig().ConfigFilePath)
	if err != nil {
		fmt.Println("init new gateway client error")
		return "", utils.ErrorNew(604, "查询失败")
	}
	if len(txHashs) == 0 {
		return "", utils.ErrorNew(602, "手机号不存在")
	}
	result := []Message{}
	for _, txHash := range txHashs {
		//BlockHeight
		blockTool := utils.BlockTool{}
		block, err := blockTool.QueryBlockByTxID(gatewayClient, utils.AppConfig(), txHash)
		if err != nil {
			fmt.Println(err)
			return "", utils.ErrorNew(604, "查询失败")
		}
		blockHeight := blockTool.GetNumber(*block)
		// base64Hash := blockTool.GetBodyHash(*block)
		// byteHash, _ := base64.StdEncoding.DecodeString(base64Hash)
		// blockHash := hex.EncodeToString(byteHash)
		//TxHash
		//value
		txTool := utils.TxTool{}
		tx, err := txTool.QueryTxByTxID(gatewayClient, utils.AppConfig(), txHash)
		if err != nil {
			fmt.Println(err)
			return "", utils.ErrorNew(604, "查询失败")
		}
		keyValues, err := txTool.GetTxKeyValues(*tx)
		if err != nil {
			fmt.Println(err)
			return "", utils.ErrorNew(604, "查询失败")
		}

		fmt.Println(keyValues[0])
		//对数据解密
		value := strings.Split(keyValues[0], ": ")[1]
		fmt.Println(value)
		crypderesult := crypto.FromBase64String(value).SetKey(Key).Aes().ECB().PKCS7Padding().Decrypt().ToString()
		//TimeStamp
		timeStamp, err := txTool.GetTimestamp(*tx)
		if err != nil {
			fmt.Println(err)
			return "", utils.ErrorNew(604, "查询失败")
		}
		result = append(result, Message{int(blockHeight), txHash, crypderesult, timeStamp})
	}
	resultString, err := json.Marshal(result)
	if err != nil {
		return "", utils.ErrorNew(604, "查询失败")
	}
	fmt.Println(string(resultString))
	return string(resultString), nil
}

func QueryByHash(txHash string) (string, error) {
	gatewayClient, err := client.NewGatewayClient(utils.AppConfig().ConfigFilePath)
	if err != nil {
		fmt.Println("init new gateway client error")
		return "", utils.ErrorNew(604, "查询失败")
	}

	//BlockNum
	blockTool := utils.BlockTool{}
	block, err := blockTool.QueryBlockByTxID(gatewayClient, utils.AppConfig(), txHash)
	if err != nil {
		fmt.Println(err)
		return "", utils.ErrorNew(604, "查询失败")
	}
	blockHeight := blockTool.GetNumber(*block)
	// base64Hash := blockTool.GetBodyHash(*block)
	// byteHash, _ := base64.StdEncoding.DecodeString(base64Hash)
	// blockHash := hex.EncodeToString(byteHash)

	//TxHash
	//value
	txTool := utils.TxTool{}
	tx, err := txTool.QueryTxByTxID(gatewayClient, utils.AppConfig(), txHash)
	if err != nil {
		fmt.Println(err)
		return "", utils.ErrorNew(604, "查询失败")
	}
	keyValues, err := txTool.GetTxKeyValues(*tx)
	if err != nil {
		fmt.Println(err)
		return "", utils.ErrorNew(604, "查询失败")
	}

	//对数据解密
	value := strings.Split(keyValues[0], ": ")[1]
	fmt.Println(value)
	crypderesult := crypto.FromBase64String(value).SetKey(Key).Aes().ECB().PKCS7Padding().Decrypt().ToString()
	//TimeStamp
	timeStamp, err := txTool.GetTimestamp(*tx)
	if err != nil {
		fmt.Println(err)
		return "", utils.ErrorNew(604, "查询失败")
	}
	result := &Message{int(blockHeight), txHash, crypderesult, timeStamp}
	resultString, err := json.Marshal(result)
	if err != nil {
		fmt.Println("转json失败")
		return "", err
	}
	return string(resultString), nil
}

func ChangeRecode(datakey string, data string) (string, error) {
	txHashID, err := SaveRecode(datakey, data)
	if err != nil {
		return "", utils.ErrorNew(604, "修改失败")
	}
	return txHashID, nil

}
