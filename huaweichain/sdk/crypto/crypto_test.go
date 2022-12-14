/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package crypto

import (
	"testing"

	"git.huawei.com/huaweichain/common/cryptomgr"
)

func Test_Crypto_Ecdsa(t *testing.T) {
	alg := cryptomgr.EcdsaWithSha256
	certPath := "../../wienerchain-java-sdk/src/test/resources/security/cert/msp/sw/public.crt"
	keyPath := "../../wienerchain-java-sdk/src/test/resources/security/cert/msp/sw/private.key"
	crypto, err := NewCrypto(alg, certPath, keyPath, func(bytes []byte) ([]byte, error) {
		return bytes, nil
	})
	if err != nil {
		t.Errorf("NewCryptoWithIdentity error: %v\n", err)
	}
	t.Logf("ecdsa common name: %v\n", crypto.GetCommonName())
	t.Logf("ecdsa organization: %v\n", crypto.GetOrg())
}

func Test_Crypto_GM(t *testing.T) {
	alg := cryptomgr.Sm2WithSm3
	certPath := "../../wienerchain-java-sdk/src/test/resources/security/cert/msp/sm/public.crt"
	keyPath := "../../wienerchain-java-sdk/src/test/resources/security/cert/msp/sm/private.key"
	crypto, err := NewCrypto(alg, certPath, keyPath, func(bytes []byte) ([]byte, error) {
		return bytes, nil
	})
	if err != nil {
		t.Errorf("NewCryptoWithIdentity error: %v\n", err)
	}
	t.Logf("gm common name: %v\n", crypto.GetCommonName())
	t.Logf("gm organization: %v\n", crypto.GetOrg())
}
