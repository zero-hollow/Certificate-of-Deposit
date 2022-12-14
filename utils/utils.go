package utils

import (
	"encoding/hex"
	"time"

	"git.huawei.com/huaweichain/proto/common"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)
import "fmt"

type myErr struct {
	code int
	msg  string
}

func (e myErr) Error() string {
	return fmt.Sprintf("code:%d,msg:%v", e.code, e.msg)
}

func ErrorNew(code int, msg string) error {
	return myErr{
		code: code,
		msg:  msg,
	}
}

func GetCode(err error) int {
	if e, ok := err.(myErr); ok {
		return e.code
	}
	return -1
}

func GetMsg(err error) string {
	if e, ok := err.(myErr); ok {
		return e.msg
	}
	return ""
}
func GetPayloadWithResp(responseMsg *common.RawMessage) ([]byte, error) {
	response := &common.Response{}
	if err := proto.Unmarshal(responseMsg.Payload, response); err != nil {
		return nil, errors.WithMessage(err, "execChainAction Unmarshal message error")
	}
	if response.Status != common.SUCCESS {
		return nil, errors.Errorf("execChainAction response.Status != common.SUCCESS, status: %s, info: %v",
			response.Status.String(), response.StatusInfo)
	}
	return response.Payload, nil
}

// Hash2str is used to convert the hash value to string for human reading.
func Hash2str(h []byte) string {
	return hex.EncodeToString(h)
}

const rfc3339Short = "2006-01-02T15:04:05Z"

// FormatTime 格式化时间为 UTC 标准时间, rfc3339Short 格式
func FormatTime(input time.Time) string {
	return input.UTC().Format(rfc3339Short)
}

// ParseNanosecond 转换纳秒时间戳
func ParseNanosecond(timeStamp int64) time.Time {
	return time.Unix(timeStamp/int64(time.Second), timeStamp%int64(time.Second)).UTC()
}
