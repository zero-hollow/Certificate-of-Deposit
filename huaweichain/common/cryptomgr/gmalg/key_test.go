package gmalg

import (
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParseSM2P256v1PrivateKey(t *testing.T) {
	keyfiles := []string{"testdata/sm2_01.key", "testdata/sm2_02.key"}
	for _, keyfile := range keyfiles {
		priPem, err := ioutil.ReadFile(keyfile)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		priDerBlock, _ := pem.Decode(priPem)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		derHex := hex.EncodeToString(priDerBlock.Bytes)
		println("pkcs8 der = " + derHex)

		// priKey, err := x509.ParsePKCS8PrivateKey(priDerBlock.Bytes)
		priKey, err := ExtractPrivateKeyFromPKCS8(priDerBlock.Bytes)
		if err != nil {
			panic(err)
		}
		println("serialized private key = " + hex.EncodeToString(priKey))
	}
}
