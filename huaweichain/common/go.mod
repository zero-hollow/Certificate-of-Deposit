module git.huawei.com/huaweichain/common

go 1.15

require (
	git.huawei.com/huaweichain/gmssl v0.0.0
	github.com/golang/mock v1.4.1
	github.com/mitchellh/mapstructure v1.3.3
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.26.0
	github.com/spf13/viper v1.6.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace git.huawei.com/huaweichain/gmssl => ../gmssl
