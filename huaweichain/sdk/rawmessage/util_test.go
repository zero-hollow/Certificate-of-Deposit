/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package rawmessage

import (
	"testing"
)

func Test_CheckChainID(t *testing.T) {
	chainID := "chain_id"
	if err := checkChainID(chainID); err != nil {
		t.Errorf("chain id: %v, error: %v", chainID, err)
	}
}

func Test_CheckChainID_Error(t *testing.T) {
	chainID := "chain_id."
	if err := checkChainID(chainID); err == nil {
		t.Errorf("unexpected chain id check result. chain id: %v", chainID)
	}
}

func Test_CheckChainID_Length(t *testing.T) {
	chainID := "chain_id_chain_id_chain_id_chain_id_chain_id_chain_id_chain_id_chain_id_chain_id_chain_id_chain_id_ch"
	if err := checkChainID(chainID); err == nil {
		t.Errorf("unexpected chain id check result. chain id length: %v, chain id: %v", len(chainID), chainID)
	}
}

func Test_CheckParams(t *testing.T) {
	chainID := "chain_id"
	contractName := "contract"
	contractVersion := "version"

	if err := checkParams(chainID, contractName, contractVersion); err != nil {
		t.Errorf("unexpected params check result; chain id: %v, contract name: %v, contract version: %v, error: %v",
			chainID, contractName, contractVersion, err)
	}
}
