module git.huawei.com/goclient

go 1.15

require (
	git.huawei.com/huaweichain/common v0.0.0
	git.huawei.com/huaweichain/proto v0.0.0
	git.huawei.com/huaweichain/sdk v0.0.0
	git.huawei.com/poissonsearch/wienerchain/contract/docker-container/contract-go/contractapi v0.0.0-00010101000000-0000000000
	github.com/deatil/go-cryptobin v1.0.1041
	github.com/garyburd/redigo v1.6.4
	github.com/gin-gonic/gin v1.8.1
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/gomodule/redigo v1.8.9 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/pkg/errors v0.9.1
)

replace (
	git.huawei.com/huaweichain/common => ./huaweichain/common
	git.huawei.com/huaweichain/gmssl => ./huaweichain/gmssl
	git.huawei.com/huaweichain/proto => ./huaweichain/proto
	git.huawei.com/huaweichain/sdk => ./huaweichain/sdk
	git.huawei.com/poissonsearch/wienerchain/contract/docker-container/contract-go/contractapi => ./internal
)

//GOPROXY=https://repo.huaweicloud.com/repository/goproxy/
