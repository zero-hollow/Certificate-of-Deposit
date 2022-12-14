package utils

import (
	"fmt"
	"strings"
	"time"

	"git.huawei.com/huaweichain/proto"
	"git.huawei.com/huaweichain/proto/common"
	"git.huawei.com/huaweichain/sdk/action/event"
	"git.huawei.com/huaweichain/sdk/client"
	"git.huawei.com/huaweichain/sdk/node"
	"git.huawei.com/huaweichain/sdk/rawmessage"
	"github.com/pkg/errors"
)

// Nodes is the definition of Nodes.
type Nodes struct {
	Endorsers     []*node.WNode
	Proposer      *node.WNode
	EventListener *node.WNode
}

const (
	WaitTime = 60
)

var txEvent *event.TxEventService

func SetTxEvent(txEventClient *event.TxEventService) {
	txEvent = txEventClient
}

func getTxEvent() *event.TxEventService {
	return txEvent
}

func Send(gatewayClient *client.GatewayClient, net *Nodes, config Config, funcName string, args string) (*common.RawMessage, string, error) {
	// 1.入参处理
	var err error
	argsSlice := strings.Split(strings.TrimSpace(args), ";")
	endorseNodes := strings.Split(config.EndorserNodes, ",")

	// 2.背书消息请求构造
	rawMsg, err := getInvokeRawMsg(gatewayClient, config, endorseNodes, funcName, argsSlice)
	if err != nil {
		return nil, "", errors.WithMessage(err, "contract raw messessage build invoke message error")
	}

	// 3.背书消息请求发送
	invokeResponses, net, err := sendInvokeRawMsg(gatewayClient, net, config, endorseNodes, rawMsg)
	if err != nil {
		return nil, "", errors.WithMessage(err, "send raw messessage error")
	}

	// 4.落盘消息构建
	txRawMsg, err := getTransactionRawMsg(gatewayClient, invokeResponses)
	if err != nil {
		return nil, "", errors.WithMessage(err, "build transaction message error")
	}

	// 5.落盘消息发送
	responseMsg, txResult, err := sendTransactionRawMsg(config, txRawMsg, net)
	if err != nil {
		return nil, "", errors.WithMessage(err, "build transaction message error")
	}

	// 6.解析response获得response.Payload
	payload, err := GetPayloadWithResp(responseMsg)
	if err != nil {
		return nil, "", errors.WithMessage(err, "parse response msg with struct type error")
	}

	// 7.解析payload获得common.RawMessage{}
	txResponse := &common.RawMessage{}
	if err := proto.Unmarshal(payload, txResponse); err != nil {
		return nil, "", errors.WithMessage(err, "unmarshal transaction response error")
	}

	if txResult.Status.String() == "VALID" {
		return txResponse, Hash2str(txResult.TxHash), nil
	}

	return txResponse, "", errors.WithMessage(err, "send transaction error")
}

func Query(gatewayClient *client.GatewayClient, net *Nodes, config Config, funcName string, args string) (string, error) {
	// 1.入参处理
	var err error
	argsSlice := strings.Split(strings.TrimSpace(args), ";")
	endorseNodes := strings.Split(config.EndorserNodes, ",")

	// 2.构造请求消息
	rawMsg, err := getInvokeRawMsg(gatewayClient, config, endorseNodes, funcName, argsSlice)
	if err != nil {
		return "", errors.WithMessage(err, "contract raw messessage build invoke message error")
	}

	// 3.发送请求消息
	invokeResponses, _, err := sendInvokeRawMsg(gatewayClient, net, config, endorseNodes, rawMsg)
	if err != nil {
		return "", errors.WithMessage(err, "send raw messessage error")
	}

	// 4.解析response获得response.Payload
	payload, err := GetPayloadWithResp(invokeResponses[0])
	if err != nil {
		return "", errors.WithMessage(err, "parse response msg with struct type error")
	}

	// 5.解析payload获得common.Transaction{}
	transaction := &common.Transaction{}
	if err := proto.Unmarshal(payload, transaction); err != nil {
		return "", errors.WithMessage(err, "unmarshal transaction error")
	}

	txPayLoad := &common.TxPayload{}
	if err := proto.Unmarshal(transaction.Payload, txPayLoad); err != nil {
		return "", errors.WithMessage(err, "unmarshal tx payload error")
	}
	txData := &common.CommonTxData{}
	if err := proto.Unmarshal(txPayLoad.Data, txData); err != nil {
		return "", errors.WithMessage(err, "unmarshal common tx data error")
	}
	fmt.Printf("query transaction success, result:%s\n", string(txData.Response.Payload))
	return string(txData.Response.Payload), nil
}

func getInvokeRawMsg(gatewayClient *client.GatewayClient, config Config, endorserNodes []string, function string, args []string) (*common.RawMessage, error) {
	return gatewayClient.ContractRawMessage.BuildInvokeMessage(config.ChainID, config.ContractName, function, args)
}

func sendInvokeRawMsg(gatewayClient *client.GatewayClient, net *Nodes, config Config, endorseNodes []string, rawMsg *common.RawMessage) ([]*common.RawMessage, *Nodes, error) {
	// 背书请求消息发送
	var invokeResponses []*common.RawMessage
	for _, node := range net.Endorsers {
		var invokeResponse *common.RawMessage
		invokeResponse, err := node.ContractAction.Invoke(rawMsg)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "invoke error")
		}
		invokeResponses = append(invokeResponses, invokeResponse)
	}
	return invokeResponses, net, nil
}

func getTransactionRawMsg(client *client.GatewayClient, transactionRawMsg []*common.RawMessage) (*rawmessage.TxRawMsg, error) {
	return client.ContractRawMessage.BuildTxRawMsg(transactionRawMsg)
}

func sendTransactionRawMsg(config Config, txRawMsg *rawmessage.TxRawMsg, net *Nodes) (*common.RawMessage, *common.TxResult, error) {
	txEventClient := getTxEvent()

	resultChan, err := txEventClient.RegisterTx(txRawMsg.Hash)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "tx event register tx id error")
	}

	transactionResponse, err := net.Proposer.ContractAction.Transaction(txRawMsg.Msg)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "invoke error")
	}

	select {
	case txResult := <-resultChan:
		return transactionResponse, txResult, nil
	case <-time.After(WaitTime * time.Second):
		return nil, nil, errors.Errorf("send transaction time out")
	}
}

func NewNodes(client *client.GatewayClient, nodeNames []string, consNode string) (*Nodes, error) {
	nodes := &Nodes{}
	nodeMap := client.Nodes
	nodes.Endorsers = make([]*node.WNode, len(nodeNames))
	for i, nodeName := range nodeNames {
		node, ok := nodeMap[nodeName]
		if !ok {
			return nil, errors.Errorf("node not exist： %v", nodeName)
		}
		nodes.Endorsers[i] = node
	}
	var ok bool
	nodes.Proposer, ok = nodeMap[consNode]
	if !ok {
		return nil, errors.Errorf("node not exist： %v", consNode)
	}
	nodes.EventListener = nodes.Endorsers[0]
	return nodes, nil
}
