/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package contractapi

// Contract 需要合约编写者实现的合约接口
type Contract interface {
	// Init 功能：合约的初始化（Init）接口，需要合约开发者实现，供合约使用者在启动合约之后调用。
	// 注意，一般将合约启动时，首先需要执行且只需要执行一次的逻辑放到此方法中
	// 入参：stub是智能合约SDK为本次合约执行交易准备的上下文对象，可以通过stub提供的API函数，获取交易请求相关信息、读写状态数据库等
	// 返回值：需要返回给合约调用者（区块链客户端）的信息，没有信息需要返回时，返回值可以为nil
	// error：初始化过程的出错信息，可由合约编写者自行设定出错逻辑
	Init(stub ContractStub) ([]byte, error)

	// Invoke 功能：合约被调用（invoke）接口，需要合约开发者实现，将主要的合约执行逻辑，放到此接口内，供合约使用者调用。
	// 入参：stub是智能合约合约SDK为本次合约执行交易准备的上下文对象，可以通过stub提供的API函数，获取交易请求相关信息、读写状态数据库等
	// 返回值：需要返回给合约调用者（区块链客户端）的信息，没有信息需要返回时，返回值可以为nil
	// error：调用过程的出错信息，可由合约编写者自行设定出错逻辑
	Invoke(stub ContractStub) ([]byte, error)
}

// Stub 公共合约stub
type Stub interface {

	// FuncName 获取智能合约请求中指定的智能合约函数名称
	// 入参：无
	// 返回值：智能合约函数名称
	FuncName() string

	// Parameters 获取请求参数
	// 入参：无
	// 返回值：用户执行智能合约逻辑时传入的多个参数，每个参数以[]byte表示
	Parameters() [][]byte

	// ChainID 功能：获取智能合约所在链ID
	// 入参：无
	// 返回值：链ID
	ChainID() string

	// ContractName 功能：获取智能合约名称
	// 入参：无
	//返回值：智能合约名称
	ContractName() string
}

// ContractStub key-value数据库的stub接口
type ContractStub interface {
	Stub

	// GetKV 功能：获取状态数据库中某个key对应的value；
	// 入参：某个键值对的key信息，只支持string类型，不可为空
	// 返回值：value值，目前只支持[]byte类型，如果接口使用者，知晓value的具体结构，可以对value进行反序列化；当key不存在时，value为nil
	// error：当网络出错，状态数据库出错，返回error信息
	GetKV(key string) ([]byte, error)

	// PutKV 功能：写状态数据库操作，此接口只是将key、value形成写集，打包到交易中，只有当交易排序、出块、并校验通过之后，
	// 才会将key-value写入到状态数据库中
	// 入参：要写入的键值对，要求key != ""，并且value != nil
	// error：入参错误
	PutKV(key string, value []byte) error

	// PutKVCommon 功能：写状态数据库操作，与PutKV功能相同；与PutKV接口的不同之处在于 value不是[]byte类型，而是一个实现了
	// Marshal(v interface{}) ([]byte, error)接口的数据，接口内部会将value通过Marshal接口序列化，然后再形成写集。
	// 入参：要写入的键值对，要求key != ""，并且value实现了Marshal接口，可以序列化为[]byte
	// error：入参错误
	PutKVCommon(key string, value interface{}) error

	// DelKV 功能：删除状态数据库中的key及其对应的value，此接口只是将待删除的key放入写集，打包到交易中，
	// 只有当交易排序、出块、并校验通过之后，才会将key删除
	// 入参：要删除的key，要求key != ""
	// error：入参错误
	DelKV(key string) error

	// GetIterator 功能：查询状态数据库中，按字典序，以startKey开头，以endKey结尾的所有状态数据，结果以迭代器的形式呈现；
	// 注意，查询范围是左闭右开的，[startKey, endKey)。例如：startKey="11"，endKey="14"，所有key都是整数的话，则查询的结果中，
	// key值包括："11","12","13"，不包括"14"
	// 入参：startKey是待查询状态数据的按字典序的起始key，startKey != ""，endKey是待查询的状态数据的按字典序的结束key，endKey!= ""；
	// 返回值：Iterator是查询结果的迭代器，可以通过此迭代器，按顺序读取查询结果
	// error：入参或网络错误
	GetIterator(startKey, endKey string) (Iterator, error)

	// GetKeyHistoryIterator 功能：查询一个key对应的所有历史的value
	// 例如，一个key的value曾经为1,2,3，当前value为4，则返回的迭代器结果中按顺序包含了1,2,3,4
	// 入参：key是待查询历史value值的key信息，key != ""
	// 返回值：HistoryIterator是按顺序包含了历史value结果的迭代器结构体变量
	// error：入参或网络错误
	GetKeyHistoryIterator(key string) (HistoryIterator, error)

	// SaveComIndex 功能：为objectKey保存索引信息，indexName_attributes_objectKey构成索引信息，注意，此处只是形成索引信息的写集，
	// 只有当含有此写集的交易经过排序、出块，并校验通过后，才会写入状态数据库。例如：存储key-value信息，key="zhangsan"，
	// value={height=175, sex="male"}，如果以sex="male"作为查询条件，查询所有的key-value，则需要反序列化所有的value，性能损耗较大，
	// 因此，应为当前key-value建立一个sex相关的索引，可调用SaveComIndex("sex", []string{"male"}, "zhangsan")。如果以sex="male",
	// height="175"作为查询条件，可调用SaveComIndex("sex/height", []string{"male", "175"}, "zhangsan")。
	//
	// 入参：indexName 索引标记，indexName != ""，attributes需要当做索引的属性，至少包含一个属性信息，objectKey 待索引的key值 ，objectKey != ""
	// error：入参错误
	SaveComIndex(indexName string, attributes []string, objectKey string) error

	// GetKVByComIndex 功能：通过索引信息，查找满足某种查询条件的key-value，key-value以迭代器的形式输出
	// 入参：indexName 索引标记，indexName != ""，attributes需要当做索引的属性，至少包含一个属性信息
	// 返回值：满足索引条件的key-value的迭代器变量
	// error：入参或网络错误
	GetKVByComIndex(indexName string, attributes []string) (Iterator, error)

	// DelComIndexOneRow 功能：删除objectKey的某个索引，indexName_attributes_objectKey构成索引信息，注意，此处只是形成索引信息的写集，
	// 只有当含有此写集的交易经过排序、出块，并校验通过后，才会写入状态数据库
	// 入参：indexName 索引标记，indexName != ""，attributes需要当做索引的属性，至少包含一个属性信息，objectKey 待索引的key值，objectKey != ""
	// error：入参错误
	DelComIndexOneRow(indexName string, attributes []string, objectKey string) error
}

// Iterator  以迭代方式获取key-value
type Iterator interface {

	// Next 检查迭代器中是否还有下一个key-value
	Next() bool

	// Key 从迭代器中获取key
	Key() string

	// Value 从迭代器中获取value
	Value() []byte

	// Close 使用完迭代器之后，需要关闭迭代器
	Close()
}

// HistoryIterator 以迭代方式获取含有某个key的写集的有效历史交易
type HistoryIterator interface {
	Iterator

	// Version 获取当前迭代位置（某笔交易）的 BlockNum 和 TxNum
	Version() (uint64, int32)

	// TxHash 获取当前迭代位置（某笔交易）的hash
	TxHash() []byte

	// IsDeleted 被查询的key，当前是否已经在状态数据库中被删除
	IsDeleted() bool

	// Timestamp 返回当前迭代位置（某笔交易）的时间戳
	Timestamp() uint64
}

type ValueSerialization interface {
	// Marshal 功能：序列化接口，需要在用户使用PutKVCommon(key string, value interface{}) error接口时，
	// 为value参数实现此接口，方便内部对value做序列化操作
	// 返回值：序列化之后的字节序列[]byte
	Marshal() ([]byte, error)
}
