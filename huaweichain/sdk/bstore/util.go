package bstore

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/sdk/client"
	"git.huawei.com/huaweichain/sdk/node"
	"git.huawei.com/huaweichain/sdk/rawmessage"
)

func invokeContract(gatewayClient *client.GatewayClient, chainId, endorserName, consenterName, contractName, funcName string, args ...string) ([]byte, error) {
	rawMsg, err := gatewayClient.ContractRawMessage.BuildInvokeMessage(chainId, contractName, funcName, args)
	if err != nil {
		return nil, errors.WithMessagef(err, "build invoke contract message error")
	}
	endorserNode, ok := gatewayClient.Nodes[endorserName]
	if !ok {
		return nil, errors.WithMessagef(err, "query contract error, endorser not exists")
	}
	var invokeResponses []*common.RawMessage
	invokeResponse, err := endorserNode.ContractAction.Invoke(rawMsg)
	if err != nil {
		return nil, errors.WithMessagef(err, "invoke contract error")
	}
	invokeResponses = append(invokeResponses, invokeResponse)
	// 验证背书请求
	resByte, err := processQueryResult(invokeResponses[0])
	if err != nil {
		return nil, errors.WithMessagef(err, "get invoke result error")
	}

	transactionRawMsg, err := gatewayClient.ContractRawMessage.BuildTxRawMsg(invokeResponses)
	if err != nil {
		return nil, errors.WithMessagef(err, "build transaction message error")
	}

	if err = sendInvokeResult(transactionRawMsg, chainId, consenterName, gatewayClient.Nodes); err != nil {
		return nil, errors.WithMessagef(err, "send invoke result error")
	}

	return resByte, nil
}

func queryContract(gatewayClient *client.GatewayClient, chainId, endorserName, contractName, funcName string, args ...string) ([]byte, error) {
	rawMsg, err := gatewayClient.ContractRawMessage.BuildInvokeMessage(chainId, contractName, funcName, args)
	if err != nil {
		return nil, err
	}
	endorserNode, ok := gatewayClient.Nodes[endorserName]
	if !ok {
		return nil, errors.WithMessagef(err, "query contract error, endorser not exists")
	}
	var invokeResponse *common.RawMessage
	invokeResponse, err = endorserNode.ContractAction.Invoke(rawMsg)
	if err != nil {
		return nil, errors.WithMessagef(err, "query contract error")
	}
	resByte, err := processQueryResult(invokeResponse)
	if err != nil {
		return nil, err
	}
	return resByte, nil
}

func sendInvokeResult(txRawMsg *rawmessage.TxRawMsg, chainId, consenterName string, nodes map[string]*node.WNode) error {
	var consenterNode *node.WNode
	for _, node := range nodes {
		if node.ID == consenterName {
			consenterNode = node
			break
		}
	}
	if consenterNode == nil {
		return errors.Errorf("failed to find valid consenterName node")
	}

	txEvent, err := consenterNode.EventAction.RegisterTxEvent(chainId)
	if err != nil {
		return errors.WithMessagef(err, "register tx event error")
	}

	resultChan, err := txEvent.RegisterTx(txRawMsg.Hash)
	if err != nil {
		return errors.WithMessagef(err, "register tx error")
	}

	transactionResponse, err := consenterNode.ContractAction.Transaction(txRawMsg.Msg)
	if err != nil {
		return errors.WithMessagef(err, "send transaction error")
	}
	txResponse := &common.Response{}
	if err := proto.Unmarshal(transactionResponse.Payload, txResponse); err != nil {
		return err
	}
	if txResponse.Status != common.SUCCESS {
		return errors.Errorf("transaction failed, transaction response status is: %v, status info: %v",
			txResponse.Status.String(), txResponse.StatusInfo)
	}
	select {
	case txResult := <-resultChan:
		status := txResult.Status.String()
		if status == "VALID" {
			return nil
		} else {
			return errors.Errorf("send transaction failed, status: %v", status)
		}
	case <-time.After(10 * time.Second):
		return errors.Errorf("send transaction time out")
	}
}

func processQueryResult(invokeResponse *common.RawMessage) ([]byte, error) {
	response := &common.Response{}
	if err := proto.Unmarshal(invokeResponse.Payload, response); err != nil {
		return nil, err
	}
	if response.Status == common.SUCCESS {
		tx := &common.Transaction{}
		if err := proto.Unmarshal(response.Payload, tx); err != nil {
			return nil, err
		}
		txPayLoad := &common.TxPayload{}
		if err := proto.Unmarshal(tx.Payload, txPayLoad); err != nil {
			return nil, err
		}
		txData := &common.CommonTxData{}
		if err := proto.Unmarshal(txPayLoad.Data, txData); err != nil {
			return nil, err
		}
		return txData.Response.Payload, nil
	} else {
		return nil, errors.Errorf("query transaction failed: %v", response.Status)
	}
}

func copyRespBody(res *http.Response) ([]byte, error) {
	if res == nil || res.Body == nil {
		return nil, nil
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func handleDownloadResponse(resBody io.ReadCloser, filePath string) (*os.File, error) {
	file, err := os.Create(filePath)
	defer resBody.Close()
	if err != nil {
		return nil, errors.WithMessagef(err, "create local file error")
	}
	wt := bufio.NewWriter(file)
	if _, err = io.Copy(wt, resBody); err != nil {
		return nil, errors.WithMessagef(err, "io copy file error")
	}
	wt.Flush()
	file.Seek(0, 0)
	return file, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func isFile(path string) bool {
	return !isDir(path)
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func validateTimestamp(startTimeStr, endTimeStr string) error {
	startTime, err := strconv.Atoi(startTimeStr)
	if err != nil {
		return errors.WithMessagef(err, "invalid start time: %s, please input ms timestamp, for example: 1644980664904", startTimeStr)
	}
	endTime, err := strconv.Atoi(endTimeStr)
	if err != nil {
		return errors.WithMessagef(err, "invalid end time: %s, please input ms timestamp, for example: 1644980664904", endTimeStr)
	}
	if startTime >= endTime {
		return errors.Errorf("endTime %d must be later than start time %d", endTime, startTime)
	}
	return nil
}
