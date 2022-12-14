module git.huawei.com/huaweichain/sdk

go 1.15

require (
	git.huawei.com/huaweichain/common v0.0.0
	git.huawei.com/huaweichain/gmssl v0.0.0
	git.huawei.com/huaweichain/proto v0.0.0
	github.com/gogo/protobuf v1.3.2
	github.com/mitchellh/mapstructure v1.3.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.3.0
	github.com/spf13/viper v1.6.2
	google.golang.org/grpc v1.28.0
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace (
	git.huawei.com/huaweichain/common => ../common
	git.huawei.com/huaweichain/gmssl => ../gmssl
	git.huawei.com/huaweichain/proto => ../proto
)
