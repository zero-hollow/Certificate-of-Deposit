####1. utils工具介绍
- config.go 保存客户端相关配置信息，需要根据实际配置进行修改。
- block.go 存放块相关函数
- tx.go 存放块相关函数
- contract.go 存放合约操作信息，主要分为Send和Query，合约中对数据有修改的操作如插入和删除，需要通过对Send方法传参进行调用，而对数据的查询操作如查询某个键的值，需要调用Query方法。
####2. utils工具调用示例
- block.go中BlockTool方法调用示例。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
block, err := utils.QueryBlockByNumber(gatewayClient, utils.AppConfig(), "34")
if err != nil{
    fmt.Println(err)
}

blockTool := utils.BlockTool{}
block, err := blockTool.QueryBlockByNumber(gatewayClient, utils.AppConfig(), "34")
if err != nil{
    fmt.Println(err)
}
blockTool.GetBodyHash(block)
```

- tx.go中TxTool方法调用示例。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
tx, err := utils.QueryTxByTxID(gatewayClient, utils.AppConfig(), "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
if err != nil {
    fmt.Println(err)
}
txTools := utils.TxTool{}
txTools.GetTimestamp(tx)
```
- contract.go中合约相关方法调用示例。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
_, txHashID, err := utils.Send(gatewayClient, utils.AppConfig(), "insert", "f;222133")
if err != nil{
    fmt.Println(err)
}
	
value, err := utils.Query(gatewayClient, utils.AppConfig(), "query", "f")
if err != nil {
    fmt.Println(err)
}
```


####3. 合约调用示例
- 插入key-value。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
_, txHashID, err := utils.Send(gatewayClient, utils.AppConfig(), "insert", "a;huaweichain")
if err != nil {
    fmt.Println(err)
}
```
- 查询key对应的value。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
value, err := utils.Query(gatewayClient, utils.AppConfig(), "query", "a")
if err != nil {
    fmt.Println(err)
}
```
- 范围查询，查询指定key范围内的记录(key范围过滤仅支持字符串比较)，参数为"startKey;endKey"，查询区间为左闭右开[startKey, endKey)。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
_, txHashID, err := utils.Send(gatewayClient, utils.AppConfig(), "insert", "a;huaweichain_a")
if err != nil {
    fmt.Println(err)
}
_, txHashID, err = utils.Send(gatewayClient, utils.AppConfig(), "insert", "b;huaweichain_b")
if err != nil {
    fmt.Println(err)
}
_, txHashID, err = utils.Send(gatewayClient, utils.AppConfig(), "insert", "c;huaweichain_c")
if err != nil {
    fmt.Println(err)
}
value, err := utils.Query(gatewayClient, utils.AppConfig(), "queryInRange", "a;c") // key区间为[a, c)
if err != nil {
    fmt.Println(err)
}
```
- 条件查询，查询指定key范围内符合value条件的记录(key和value范围过滤仅支持字符串比较)，参数为"startKey;endKey;judge;value"，查询区间为左闭右开[startKey, endKey)，判断符judge为 >、<、!=、==，比较值value为字符串。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
_, txHashID, err := utils.Send(gatewayClient, utils.AppConfig(), "insert", "a;001")
if err != nil {
    fmt.Println(err)
}
_, txHashID, err = utils.Send(gatewayClient, utils.AppConfig(), "insert", "b;002")
if err != nil {
    fmt.Println(err)
}
_, txHashID, err = utils.Send(gatewayClient, utils.AppConfig(), "insert", "c;003")
if err != nil {
    fmt.Println(err)
}
value, err := utils.Query(gatewayClient, utils.AppConfig(), "queryInRangeMatchCondition", "a;c;!=;001") // 查询区间为[a,c)、判断符为!=、比较值为001
if err != nil {
    fmt.Println(err)
}
```

####4. 富媒体存储调用示例
- 初始化富媒体存储客户端，需指定已初始化完成的gatewayClient、链ID、共识节点。
```
gatewayClient, err := client.NewGatewayClient("D:/goDemo/cert/default-organization-sdk.yaml")
if err != nil {
    fmt.Println("init new gateway client error")
}
bsClient, err := bstore.NewBsClient(gatewayClient, utils.AppConfig().ChainID, utils.AppConfig().ConsensusNode)
if err != nil {
	fmt.Println(err)
} 
```
- 文件上链，需指定待上链文件在本地的路径，及文件在链上的名称
```
fileInfo, err := bsClient.UploadFile("D:/goDemo/example.PNG", "example.PNG")
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(fmt.Sprintf("upload file to chain success! file name: %s, file hash: %s, version: %s",
	    fileInfo.FileName, fileInfo.FileHash, fileInfo.Version))
}
```
- 文件下载，需指定文件下载到本地和路径，及待下载文件在链上的名称
```
err = bsClient.DownloadFile("D:/goDemo/example.PNG", "example.PNG", 1)
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println("download file from chain success!")
}
```
- 查询文件历史版本，需指定链上文件名
```
fileHistories, err := bsClient.GetFileHistory("example.PNG")
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println("get file history success!")
	for _, history := range fileHistories {
		fmt.Println(fmt.Sprintf("create time: %s, uploader: %s, updated time: %s, hashcode: %s, version: %d",
			history.CreatedTime, history.Uploader, history.UpdatedTime, history.HashCode, history.Version))
	}
}
```
- 查询文件操作记录，需指定链上文件名，及查询的起止时间，单位为毫秒，查询区间为左闭右开([startTime, endTime])
```
fileOperations, err := bsClient.GetFileOperation("example.PNG", "1645440974728", "1645441080861")
if err != nil {
	fmt.Println(err)
} else {
	fmt.Println("get file operation record success!")
	for _, operation := range fileOperations {
		fmt.Println(fmt.Sprintf("operator: %s, time: %s, event type: %s", operation.Operator, operation.Time, operation.EventType))
	}
}
```


