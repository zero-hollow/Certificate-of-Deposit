package utils

import (
	"git.huawei.com/huaweichain/sdk/config"
)

type Config struct {
	ConfigFilePath string // yaml配置文件路径
	ContractName   string // 合约名称
	EndorserNodes  string // 背书节点，若背书策略为“全部组织背书”，则背书节点为每个组织的中的任一节点即可，eg: "node-0.organization,node-0.organization1"
	ConsensusNode  string // 共识节点，默认为"node-0.organization",选择共识组织下的任一节点即可
	ChainID        string // 链ID，即实例概览页面的"链信息->链ID"
	QueryNode      string // 查询节点,可选择一个或者若干个组织内的节点(参见yaml配置文件中)，用于发出执行请求
	SignAlgorithm  string // 安全机制
}

var appConfig Config

func AppConfig() Config {
	return appConfig
}

func SetSignAlg(cfg Config) {
	clientConfig, err := config.NewClientConfig(cfg.ConfigFilePath)
	if err != nil {
		return
	}
	appConfig.SignAlgorithm = clientConfig.Client.Type
}

func init() {
	appConfig.ConfigFilePath = "C:\\Users\\HuJie\\Desktop\\Testing\\configuration\\sdk.yaml"
	appConfig.ContractName = "example01"

	//背书节点名称node-2.certificate-w0hv3kwia
	appConfig.EndorserNodes = "node-0.certificate-w0hv3kwia,node-1.certificate-w0hv3kwia,node-2.certificate-w0hv3kwia"
	//共识节点名称
	appConfig.ConsensusNode = "node-0.certificate-w0hv3kwia"
	appConfig.QueryNode = "node-0.certificate-w0hv3kwia"
	appConfig.ChainID = "bcs-2n7xka-3f771e976"
	SetSignAlg(appConfig)
}
