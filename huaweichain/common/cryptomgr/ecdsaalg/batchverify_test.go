/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */
package ecdsaalg

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"git.huawei.com/huaweichain/common/cryptomgr/bccryptoutil"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

// PemCertAndKey the pem format certificate and key
type pemCertAndKey struct {
	Name string

	// CertPem is the pem format cert which should not be empty
	CertPem string

	// KeyPem is the pem format cert which should not be empty
	KeyPem string
}

// CertInfo the information for certificate creation
type certInfo struct {
	Usage           int    // it can only be set to SIGN = 1 or TLS = 2;
	SignAlg         string // it can only be set to "sm2_with_sm3" or "ecdsa_with_sha256"
	NodePrefix      string // the prefix for node cert common name, which should not be empty
	ClientPrefix    string // the prefix for client cert common name, which should not be empty
	ValidationYears uint   // validation years ranges from 1 to 30

	Organization     string
	OrganizationUnit string
	Country          string
	Province         string
	Locality         string
}

const (
	// EcdsaWithSha256 : the alg is ecdsa_with_sha256
	ecdsaWithSha256 = "ecdsa_with_sha256"
	// Sm2WithSm3 : the alg is sm2_with_sm3
	sm2WithSm3 = "sm2_with_sm3"
)
const (
	// signUsage : the certificate usage is sign
	signUsage = 1
	// tlsUsage : the certificate usage is tls communication
	tlsUsage = 2
)

type signAndMsg struct {
	signInfo []byte
	msgInfo  []byte
	res      bool
}

// TestPerformanceSingle test performance for single verify
func TestPerformanceSingle(t *testing.T) {
	bccryptoutil.InitRandom()
	compareVerifySpeedSingleGoRoutine()
}

// TestPerformanceMulti test performance for multi verify
func TestPerformanceMulti(t *testing.T) {
	bccryptoutil.InitRandom()
	compareVerifySpeedMulGoRoutine()
}

func createSelfSignCACertAndKeyNew(caInfo *certInfo, commonName string) (*pemCertAndKey, error) {
	if caInfo == nil || commonName == "" {
		return nil, fmt.Errorf("the parameter is not correct")
	}

	basicInfo := &cryptomgr.CertBasicInfo{
		Organization:     caInfo.Organization,
		OrganizationUnit: caInfo.OrganizationUnit,
		Country:          caInfo.Country,
		Province:         caInfo.Province,
		Locality:         caInfo.Locality,
		CommonName:       commonName,
		ValidationYears:  caInfo.ValidationYears,
	}

	p256Key, err := CreateKeyWithCurveP256()
	if err != nil {
		return nil, errors.New("create p256 key failed")
	}

	caCert, err := GenerateSelfSignCert(basicInfo, p256Key)
	if err != nil {
		return nil, errors.New("generate self sign cert failed")
	}
	return &pemCertAndKey{
		Name:    commonName,
		CertPem: string(caCert.GetPemCertBytes()),
		KeyPem:  string(p256Key.GetPemBytes()),
	}, nil
}

func compareVerifySpeedSingleGoRoutine() {
	caInfo := &certInfo{
		Usage:            signUsage,
		SignAlg:          sm2WithSm3,
		NodePrefix:       "node",
		ClientPrefix:     "user",
		ValidationYears:  20,
		Organization:     "",
		OrganizationUnit: "",
		Country:          "",
		Province:         "",
		Locality:         "",
	}
	certAndKey, err := createSelfSignCACertAndKeyNew(caInfo, "www.example.com")
	if err != nil {
		return
	}

	key, err := GetKeyFromPem([]byte(certAndKey.KeyPem))
	if err != nil {
		return
	}

	cert, err := GetCert([]byte(certAndKey.CertPem))
	if err != nil {
		return
	}

	batchSize := 500
	msg := []byte("hello")

	var signSlice = make([][]byte, 0, batchSize)
	var msgSlice = make([][]byte, 0, batchSize)
	fmt.Printf("begin to sign %d\n", batchSize)
	begin1 := time.Now()
	for i := 0; i < batchSize; i++ {
		sig, err := key.SignForBatchVerify(msg, "")
		if err != nil {
			fmt.Printf("sign failed\n")
			return
		}
		signSlice = append(signSlice, sig)
		msgSlice = append(msgSlice, msg)
	}
	end1 := time.Since(begin1)
	fmt.Printf("sign use time: %s\n", end1.String())

	begin2 := time.Now()
	for i := 0; i < batchSize; i++ {
		err := cert.Verify(msgSlice[i], signSlice[i], "")
		if err != nil {
			fmt.Printf("verify failed\n")
			return
		}
	}
	end2 := time.Since(begin2)
	fmt.Printf("verify use time: %s\n", end2.String())

	begin3 := time.Now()
	err = cert.VerifyBatch(msgSlice, signSlice, "")
	if err != nil {
		fmt.Printf("verify batch failed\n")
	}
	end3 := time.Since(begin3)
	fmt.Printf("verify batch time:%s\n", end3.String())
}

func compareVerifySpeedMulGoRoutine() {
	caInfo := &certInfo{
		Usage:            signUsage,
		SignAlg:          sm2WithSm3,
		NodePrefix:       "node",
		ClientPrefix:     "user",
		ValidationYears:  20,
		Organization:     "",
		OrganizationUnit: "",
		Country:          "",
		Province:         "",
		Locality:         "",
	}
	certAndKey, err := createSelfSignCACertAndKeyNew(caInfo, "www.example.com")
	if err != nil {
		return
	}

	key, err := GetKeyFromPem([]byte(certAndKey.KeyPem))
	if err != nil {
		return
	}

	cert, err := GetCert([]byte(certAndKey.CertPem))
	if err != nil {
		return
	}

	var batchSize = 1000000
	var msgSlice = make([][]byte, 0, batchSize)
	var sigSlice = make([][]byte, 0, batchSize)
	now := time.Now()
	fmt.Printf("batch num:%d\n", batchSize)
	fmt.Printf("begin sign for batch is %s\n", now.String())
	msgTmp := []byte("uppppp")

	var sigAndMsgChanSlice []chan *signAndMsg
	for i := 0; i < batchSize; i++ {
		sigAndMsgCh := signFast(key, msgTmp)
		sigAndMsgChanSlice = append(sigAndMsgChanSlice, sigAndMsgCh)
	}

	for _, v := range sigAndMsgChanSlice {
		sigAndMsgTmp := <-v
		if !sigAndMsgTmp.res {
			fmt.Printf("sign failed")
			return
		}
		sigSlice = append(sigSlice, sigAndMsgTmp.signInfo)
		msgSlice = append(msgSlice, sigAndMsgTmp.msgInfo)
	}
	since := time.Since(now)
	fmt.Printf("the time for sign for batch is %s\n", since.String())

	begin2 := time.Now()
	var wg sync.WaitGroup
	var failCount int32
	for i := 0; i < batchSize; i++ {
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			err := cert.Verify(msgSlice[pos], sigSlice[pos], "")
			if err != nil {
				fmt.Printf("verify failed\n")
				return
			}
		}(i)
	}

	wg.Wait()
	if failCount != 0 {
		fmt.Printf("verify failed")
	}
	end2 := time.Since(begin2)
	fmt.Printf("the time for verify is %s\n", end2.String())

	fmt.Printf("before batch verify\n")
	begin3 := time.Now()

	verifyRes := multiVerify(sigSlice, msgSlice, cert)
	if !verifyRes {
		fmt.Printf("batch verify failed\n")
	}
	end3 := time.Since(begin3)
	fmt.Printf("the time for batch verify is %s\n", end3.String())
}

func signFast(key cryptomgr.Key, msgTmp []byte) chan *signAndMsg {
	var sigAndMsgInfoCh = make(chan *signAndMsg)

	go func() {
		var sigAndMsgInfo *signAndMsg
		sign, err := key.SignForBatchVerify([]byte(msgTmp), "")
		if err != nil {
			sigAndMsgInfo = &signAndMsg{
				signInfo: sign,
				msgInfo:  msgTmp,
				res:      false,
			}
		} else {
			sigAndMsgInfo = &signAndMsg{
				signInfo: sign,
				msgInfo:  msgTmp,
				res:      true,
			}
		}

		sigAndMsgInfoCh <- sigAndMsgInfo
	}()
	return sigAndMsgInfoCh
}

func multiVerify(signBatch [][]byte, payloadBatch [][]byte, cert cryptomgr.Cert) bool {
	var oneBatchNum = 50
	sigNum := len(signBatch)
	if sigNum < oneBatchNum {
		return false
	}

	var wg sync.WaitGroup
	var checkFailedNum int32 = 0

	loop := sigNum / oneBatchNum
	for i := 0; i < loop; i++ {
		wg.Add(1)
		go func(pos int, payloadBatchTmp [][]byte, signSliceTmp [][]byte, certTmp cryptomgr.Cert) {
			defer wg.Done()
			begin := pos * oneBatchNum
			if begin >= sigNum {
				return
			}
			end := begin + oneBatchNum
			if sigNum-end < oneBatchNum {
				end = sigNum
			}

			batchRes := certTmp.VerifyBatch(payloadBatchTmp[begin:end], signSliceTmp[begin:end], "")

			if batchRes != nil {
				atomic.AddInt32(&checkFailedNum, 1)
				return
			}
		}(i, payloadBatch, signBatch, cert)
	}

	wg.Wait()
	if int(checkFailedNum) != 0 {
		fmt.Printf("batch verify failed, checkFailedNum:%d\n", checkFailedNum)
		return false
	}
	return true
}
