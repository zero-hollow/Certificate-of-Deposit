package gmalg

import (
	"fmt"
	"testing"

	"git.huawei.com/huaweichain/common/cryptomgr/ecdsaalg"
)

func TestGeneratePriKey(t *testing.T) {
	fmt.Println("enter TestGeneratePriKey")
	priKey, err := GeneratePriKey()
	if err != nil {
		t.Fatalf("generate private key failed:%s", err.Error())
	}
	publicKey := priKey.GetPublicKey()
	bytes := publicKey.GetPemBytes()
	fmt.Printf("%s", string(bytes))

	ecdsaPubKey, err := ecdsaalg.GetKeyFromPem(bytes)
	if err != nil {
		t.Fatalf("ecdsaalg.GetKeyFromPem failed:%s", err.Error())
	}
	fmt.Printf("success get ecdsa pubkey:%s ", string(ecdsaPubKey.GetPemBytes()))

}
