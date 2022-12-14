package bstore

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/sdk/crypto"
)

const (
	commonParamKey = "common"
	uploadParamKey = "upload"
	fileParamKey   = "file"
)

// MsgBuilder is the interface of base raw message builder.
type BsMsgBuilder interface {
	BuildUploadRawMessage(string, string, string) (*UploadRawMessage, error)
	BuildDownloadRawMessage(string, string, string) (*DownloadRawMessage, error)
}

// MsgBuilderImpl is the definition of MsgBuilderImpl.
type BsMsgBuilderImpl struct {
	crypto crypto.Crypto
}

// NewMsgBuilderImpl is used to create an instance of MsgBuilderImpl by specifying crypto.
func NewMsgBuilderImpl(crypto crypto.Crypto) (*BsMsgBuilderImpl, error) {
	msgBuilder := &BsMsgBuilderImpl{crypto: crypto}
	return msgBuilder, nil
}

type UploadRawMessage struct {
	Body        *bytes.Buffer
	ContentType string
	FileName    string
	FileHash    string
}

type DownloadRawMessage struct {
	Body        *bytes.Buffer
	ContentType string
}

func (builder *BsMsgBuilderImpl) BuildUploadRawMessage(instanceId, filePath, fileName string) (*UploadRawMessage, error) {
	if !exists(filePath) {
		return nil, errors.Errorf("invalid file path %s not exist", filePath)
	}
	if !isFile(filePath) {
		return nil, errors.Errorf("invalid file path %s is not file", filePath)
	}
	fh, err := os.Open(filePath)
	defer fh.Close()
	if err != nil {
		return nil, errors.WithMessagef(err, "open upload file error")
	}
	fileHash := builder.getFileDigestBase64(fh)
	userCert, err := builder.crypto.GetCertificate()
	if err != nil {
		return nil, errors.WithMessagef(err, "get user cert error")
	}
	commonParam := CommonParam{
		InstanceID: instanceId,
		Cert:       string(userCert),
		Data:       builder.getIdentityData(fileHash),
		SignedData: builder.getSignedData(fileHash),
	}
	uploadParam := UploadParam{
		Filename: fileName,
	}
	commonBytes, err := json.Marshal(commonParam)
	if err != nil {
		return nil, errors.WithMessagef(err, "marshal common param error")
	}
	uploadBytes, err := json.Marshal(uploadParam)
	if err != nil {
		return nil, errors.WithMessagef(err, "marshal upload param error")
	}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	err = bodyWriter.WriteField(commonParamKey, string(commonBytes))
	if err != nil {
		return nil, errors.WithMessagef(err, "write common param error")
	}
	err = bodyWriter.WriteField(uploadParamKey, string(uploadBytes))
	if err != nil {
		return nil, errors.WithMessagef(err, "write upload param error")
	}
	fileWriter, err := bodyWriter.CreateFormFile(fileParamKey, filePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "create from file error")
	}

	_, err = fh.Seek(0, 0)
	if err != nil {
		return nil, errors.WithMessagef(err, "file header seek error")
	}
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return nil, errors.WithMessagef(err, "file io copy error")
	}
	err = bodyWriter.Close()
	if err != nil {
		return nil, errors.WithMessagef(err, "body writer close error")
	}

	return &UploadRawMessage{
		Body:        bodyBuf,
		ContentType: bodyWriter.FormDataContentType(),
		FileName:    fileName,
		FileHash:    fileHash,
	}, nil
}

func (builder *BsMsgBuilderImpl) BuildDownloadRawMessage(instanceId, fileName, internalVersion string) (*DownloadRawMessage, error) {
	verifyText := time.Now().String() + downloadURI
	userCert, err := builder.crypto.GetCertificate()
	if err != nil {
		return nil, errors.WithMessagef(err, "get user cert error")
	}
	commonParam := CommonParam{
		InstanceID: instanceId,
		Cert:       string(userCert),
		Data:       builder.getIdentityData(verifyText),
		SignedData: builder.getSignedData(verifyText),
	}
	downloadParam := DownloadParam{
		fileName,
		internalVersion,
	}
	commonBytes, err := json.Marshal(commonParam)
	if err != nil {
		return nil, errors.WithMessagef(err, "marshal common param error")
	}
	paraBytes, err := json.Marshal(downloadParam)
	if err != nil {
		return nil, errors.WithMessagef(err, "marshal download param error")
	}
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	err = bodyWriter.WriteField(commonParamKey, string(commonBytes))
	if err != nil {
		return nil, errors.WithMessagef(err, "write common param error")
	}
	err = bodyWriter.WriteField("download", string(paraBytes))
	if err != nil {
		return nil, errors.WithMessagef(err, "write download param error")
	}
	err = bodyWriter.Close()
	if err != nil {
		return nil, errors.WithMessagef(err, "body writer close error")
	}
	return &DownloadRawMessage{
		Body:        bodyBuf,
		ContentType: bodyWriter.FormDataContentType(),
	}, nil
}

func (builder *BsMsgBuilderImpl) getFileDigestBase64(f io.Reader) string {
	hash := sha256.New()
	io.Copy(hash, f)
	digestBase64 := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return digestBase64
}

func (builder *BsMsgBuilderImpl) getIdentityData(data string) string {
	return base64.URLEncoding.EncodeToString([]byte(data))
}

func (builder *BsMsgBuilderImpl) getSignedData(data string) string {
	sig, _ := builder.crypto.Sign([]byte(data))
	ret := base64.URLEncoding.EncodeToString(sig)
	return ret
}
