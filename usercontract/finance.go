package usercontract

import (
	"git.huawei.com/poissonsearch/wienerchain/contract/docker-container/contract-go/contractapi"
	"github.com/pkg/errors"
)

// 声明合约的结构体
type FinanceInterface struct {
}

// 创建合约
func NewSmartContract() contractapi.Contract {
	return &FinanceInterface{}
}

// 合约的初始化接口，合约启动的时候，首先执行只需要执行一次的逻辑方法放入到此方法中
func (f *FinanceInterface) Init(_ contractapi.ContractStub) ([]byte, error) {
	return nil, nil
}

// 合约被调用的接口，将主要的合约执行逻辑，放到此方法中
func (f *FinanceInterface) Invoke(stub contractapi.ContractStub) ([]byte, error) {
	// 1.获取函数名和需要传入对应函数的参数
	funcName, args := stub.FuncName(), stub.Parameters()
	//跳转到对应的函数去
	switch funcName {
	case "saveRecode":
		return saveRecode(stub, args)
	}
	return nil, errors.Errorf("func name is not correct, the function name is %s ", funcName)
}

// 保存数据
func saveRecode(stub contractapi.ContractStub, args [][]byte) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("the argNum for saveRecode is not correct,expected 2")
	}
	key := args[0]
	value := args[1]
	

	err := stub.PutKV(string(key), value)
	if err != nil {
		return nil, errors.New("saveRecode failed!")
	}
	return nil, nil
}
