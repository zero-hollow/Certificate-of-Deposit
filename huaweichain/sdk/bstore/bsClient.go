package bstore

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/sdk/client"
)

func newTransport() *http.Transport {
	tlsConfigWithoutCertsAndSkipVerify := tls.Config{
		MinVersion:         tls.VersionTLS12,
		CipherSuites:       []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384},
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig:       &tlsConfigWithoutCertsAndSkipVerify,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 90 * time.Second}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second, // if not set this timeout, after fully write the request to tcp conn. once there has some network probelm, read will block goroutines forever
		MaxIdleConnsPerHost:   200,
		DisableKeepAlives:     true,
	}
	return transport
}

type BsClient struct {
	gatewayClient *client.GatewayClient
	httpClient    *http.Client
	msgBuilder    *BsMsgBuilderImpl
	instanceId    string
	bsServerAddr  string
	chainId       string
	endorserName  string
	consenterName string
}

func NewBsClient(gatewayClient *client.GatewayClient, chainId string, endorserName, consenterName string) (*BsClient, error) {
	bsClient := &BsClient{
		gatewayClient: gatewayClient,
		httpClient:    &http.Client{Transport: newTransport()},
	}

	msgBuilder, err := NewMsgBuilderImpl(gatewayClient.Crypto)
	if err != nil {
		return nil, errors.WithMessagef(err, "init bs msg builder error")
	}
	bsClient.msgBuilder = msgBuilder
	bsClient.chainId = chainId
	bsClient.endorserName = endorserName
	bsClient.consenterName = consenterName

	bsServerAddr, err := bsClient.getBsServerAddress()
	if err != nil {
		return nil, errors.WithMessagef(err, "get bs server addr error")
	}
	instanceId, err := bsClient.getInstanceId()
	if err != nil {
		return nil, errors.WithMessagef(err, "get instance id error")
	}
	bsClient.bsServerAddr = bsServerAddr
	bsClient.instanceId = instanceId


	return bsClient, nil
}

func (bc *BsClient) getBsServerAddress() (string, error) {
	bsAddressByte, err := queryContract(bc.gatewayClient, bc.chainId, bc.endorserName, contractName, bSAddressQueryCmd)
	if err != nil {
		if strings.Contains(err.Error(), contractNotFoundErr) {
			err = errors.New("function not support, you may have not upgrade huaweichain yet")
		}
		return "", errors.Errorf("get bs server address failed, error: %s", err)
	}

	return string(bsAddressByte), nil
}

func (bc *BsClient) getInstanceId() (string, error) {
	instanceIdByte, err := queryContract(bc.gatewayClient, bc.chainId, bc.endorserName, contractName, instanceIdQueryCmd)
	if err != nil {
		if strings.Contains(err.Error(), contractNotFoundErr) {
			err = errors.New("function not support, you may have not upgrade huaweichain yet")
		}
		return "", errors.Errorf("get instace id failed, error: %s", err)
	}

	return string(instanceIdByte), nil
}

func (bc *BsClient) UploadFile(filePath, fileName string) (*UploadFileResponse, error) {
	uploadRawMessage, err := bc.msgBuilder.BuildUploadRawMessage(bc.instanceId, filePath, fileName)
	if err != nil {
		return nil, errors.WithMessagef(err, "build upload raw message failed")
	}

	resp, err := bc.httpClient.Post(bsAddressPrefix+bc.bsServerAddr+bsAddressPortSuffix+uploadURI, uploadRawMessage.ContentType, uploadRawMessage.Body)
	if err != nil {
		return nil, errors.WithMessagef(err, "upload file error")
	}
	defer resp.Body.Close()
	resByte, err := copyRespBody(resp)
	if err != nil {
		return nil, errors.WithMessagef(err, "get bs server response error")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("upload file response status code: %d != 200/201, response body: %s",
			resp.StatusCode, string(resByte))
	}

	versionId := strings.ReplaceAll(string(resByte), "\"", "")
	fileInfo, err := bc.uploadFileToChain(bc.gatewayClient, uploadRawMessage.FileName, uploadRawMessage.FileHash, versionId)
	if err != nil {
		if strings.Contains(err.Error(), contractNotFoundErr) {
			err = errors.New("function not support, you may have not upgrade huaweichain yet")
		}
		return nil, errors.WithMessagef(err, "upload file to chain error")
	}

	return &UploadFileResponse{
		FileName: uploadRawMessage.FileName,
		FileHash: fileInfo.HashCode,
		Version:  strconv.Itoa(fileInfo.Version),
	}, nil
}

func (bc *BsClient) DownloadFile(filePath, fileName string, versionId int) error {
	internalVersion, err := bc.getInternalVersion(fileName, versionId)
	if err != nil {
		return errors.WithMessagef(err, "get file: %s version: %d error", fileName, versionId)
	}

	downloadRawMessage, err := bc.msgBuilder.BuildDownloadRawMessage(bc.instanceId, fileName, internalVersion)
	if err != nil {
		return errors.WithMessagef(err, "get download raw message error")
	}

	resp, err := bc.httpClient.Post(bsAddressPrefix+bc.bsServerAddr+bsAddressPortSuffix+downloadURI, downloadRawMessage.ContentType, downloadRawMessage.Body)
	if err != nil {
		return errors.WithMessagef(err, "download file error")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("download file response status code: %d != 200/201", resp.StatusCode)
	}

	file, err := handleDownloadResponse(resp.Body, filePath)
	defer file.Close()
	if err != nil {
		return errors.WithMessagef(err, "handle download response error")
	}

	fileHashFromChain, err := bc.getFileHashFromChain(fileName, versionId)
	if err != nil {
		return errors.WithMessagef(err, "get file hash from chain error")
	}

	fileHashByCal := bc.msgBuilder.getFileDigestBase64(file)
	if fileHashFromChain != fileHashByCal {
		return errors.Errorf("file hash inconsistent, file may be tempered, chain file hash: [%s], actual file hash [%s]", fileHashFromChain, fileHashByCal)
	}
	return nil
}

func (bc *BsClient) GetFileHistory(fileName string) ([]*FileHistory, error) {
	files, err := bc.getFileHistoryFromChain(fileName)
	if err != nil {
		if strings.Contains(err.Error(), contractNotFoundErr) {
			err = errors.New("function not support, you may have not upgrade huaweichain yet")
		}
		return nil, errors.WithMessagef(err, "get file history from chain error")
	}
	var fileHistories []*FileHistory
	for _, f := range files {
		fileHistories = append(fileHistories, &FileHistory{
			f.CreatedTime,
			f.UpdatedTime,
			f.Version,
			f.Uploader,
			f.HashCode,
		})
	}
	return fileHistories, nil
}

func (bc *BsClient) GetFileOperation(fileName, startTime, endTime string) ([]*StorageEvent, error) {
	if err := validateTimestamp(startTime, endTime); err != nil {
		return nil, errors.WithMessagef(err, "invalid timestamp")
	}
	fileOperation, err := queryContract(bc.gatewayClient, bc.chainId, bc.endorserName, contractName, fileOperationQueryCmd, fileName, startTime, endTime)
	if err != nil {
		if strings.Contains(err.Error(), contractNotFoundErr) {
			err = errors.New("function not support, you may have not upgrade huaweichain yet")
		}
		return nil, errors.WithMessagef(err, "get file operation from chain error")
	}

	var operationList []*StorageEvent
	err = json.Unmarshal(fileOperation, &operationList)
	if err != nil {
		return nil, errors.WithMessagef(err, "umarshal file operations error")
	}

	return operationList, nil
}

func (bc *BsClient) uploadFileToChain(gatewayClient *client.GatewayClient, fileName, fileHash, versionId string) (*FileInfo, error) {
	curTime := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	fileInfoByte, err := invokeContract(gatewayClient, bc.chainId, bc.endorserName, bc.consenterName, contractName, fileUploadCmd, fileName, fileHash, versionId, curTime)
	if err != nil {
		return nil, err
	}
	fileInfo := &FileInfo{}
	err = json.Unmarshal(fileInfoByte, fileInfo)
	if err != nil {
		return nil, err
	}

	return fileInfo, nil
}

func (bc *BsClient) getFileHashFromChain(fileName string, versionId int) (string, error) {
	curTime := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	res, err := invokeContract(bc.gatewayClient, bc.chainId, bc.endorserName, bc.consenterName, contractName, fileDownloadCmd, fileName, curTime)
	if err != nil {
		return "", errors.WithMessagef(err, "query contract error")
	}
	var fileInfo FileInfo
	if versionId <= 0 {
		err = json.Unmarshal(res, &fileInfo)
		if err != nil {
			return "", err
		}
	} else {
		fileHistories, err := bc.getFileHistoryFromChain(fileName)
		if err != nil {
			if strings.Contains(err.Error(), contractNotFoundErr) {
				err = errors.New("function not support, you may have not upgrade huaweichain yet")
			}
			return "", err
		}
		for _, fileHistory := range fileHistories {
			if fileHistory.Version == versionId {
				return fileHistory.HashCode, nil
			}
		}
	}

	return "", nil
}

func (bc *BsClient) getInternalVersion(fileName string, versionId int) (string, error) {
	if versionId <= 0 {
		return "", errors.Errorf("invalid version id: %d", versionId)
	}

	fileHistories, err := bc.getFileHistoryFromChain(fileName)
	if err != nil {
		if strings.Contains(err.Error(), contractNotFoundErr) {
			err = errors.New("function not support, you may have not upgrade huaweichain yet")
		}
		return "", errors.WithMessagef(err, "get file history from chain error")
	}

	if len(fileHistories) <= 0 {
		return "", errors.Errorf("failed to find file on chain, name: %s", fileName)
	}

	for _, fileHistory := range fileHistories {
		if fileHistory.Version == versionId {
			return fileHistory.InternalVersion, nil
		}
	}

	return "", errors.Errorf("version specified not found")
}

func (bc *BsClient) getFileHistoryFromChain(fileName string) ([]FileInfo, error) {
	rawFileHistories, err := queryContract(bc.gatewayClient, bc.chainId, bc.endorserName, contractName, fileHistoryQueryCmd, fileName)
	if err != nil {
		return nil, err
	}

	var fileHistories []FileInfo
	err = json.Unmarshal(rawFileHistories, &fileHistories)
	if err != nil {
		return nil, err
	}

	return fileHistories, nil
}

func (bc *BsClient) getSignedData(data string) (string, error) {
	sig, err := bc.gatewayClient.Crypto.Sign([]byte(data))
	if err != nil {

	}
	ret := base64.URLEncoding.EncodeToString(sig)
	return ret, nil
}
