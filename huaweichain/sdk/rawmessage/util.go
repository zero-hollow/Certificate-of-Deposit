/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package rawmessage

import (
	"regexp"

	"git.huawei.com/huaweichain/proto/relayer"

	"github.com/pkg/errors"
)

const (
	// ChainIDReg is the chain id regex format.
	ChainIDReg = "^[A-Za-z0-9_][A-Za-z0-9_-]{0,99}$"

	// ChainAddrReg is the chain addr regex format.
	ChainAddrReg = "^[A-Za-z0-9+/=]{1,100}$"

	// ContractNameReg is the contract name regex format.
	ContractNameReg = "^[A-Za-z0-9_][A-Za-z0-9_-]{0,99}$"

	// ContractVersionReg is the contract version regex format.
	ContractVersionReg = "^[A-Za-z0-9_][A-Za-z0-9_\\.]{0,99}$"

	// TransactionIDReg is the transaction id regex format.
	TransactionIDReg = "^[A-Za-z0-9]{1,100}$"
)

func checkChainID(chainID string) error {
	ok, err := checkString(ChainIDReg, chainID)
	if err != nil {
		return errors.WithMessage(err, "check string error")
	}
	if !ok {
		return errors.Errorf("check chain id not match format: %v", ChainIDReg)
	}
	return nil
}

func checkChainAddr(chainAddr string) error {
	ok, err := checkString(ChainAddrReg, chainAddr)
	if err != nil {
		return errors.WithMessage(err, "check string error")
	}
	if !ok {
		return errors.Errorf("check chain addr not match format: %v", ChainAddrReg)
	}
	return nil
}

func checkContractName(name string) error {
	ok, err := checkString(ContractNameReg, name)
	if err != nil {
		return errors.WithMessage(err, "check string error")
	}
	if !ok {
		return errors.Errorf("check contract name not match format: %v", ContractNameReg)
	}
	return nil
}

func checkContractVersion(version string) error {
	ok, err := checkString(ContractVersionReg, version)
	if err != nil {
		return errors.WithMessage(err, "check string error")
	}
	if !ok {
		return errors.Errorf("check contract version not match format: %v", ContractVersionReg)
	}
	return nil
}

func checkTransactionID(txID string) error {
	ok, err := checkString(TransactionIDReg, txID)
	if err != nil {
		return errors.WithMessage(err, "check string error")
	}
	if !ok {
		return errors.Errorf("check transaction id not match format: %v", TransactionIDReg)
	}
	return nil
}

func checkInvokedList(list []*relayer.CrossContract) error {
	if list == nil {
		return nil
	}
	for _, v := range list {
		if err := checkChainAddr(v.ChainId); err != nil {
			return errors.WithMessage(err, "check chain addr error")
		}
		if err := checkContractName(v.Contract); err != nil {
			return errors.WithMessage(err, "check contract name error")
		}
	}
	return nil
}

func checkParams(chainID string, contract string, version string) error {
	if err := checkChainID(chainID); err != nil {
		return errors.WithMessage(err, "check chain id error")
	}
	if err := checkContractName(contract); err != nil {
		return errors.WithMessage(err, "check contract name error")
	}
	if err := checkContractVersion(version); err != nil {
		return errors.WithMessage(err, "check contract version error")
	}
	return nil
}

func checkCrossParams(chainID string, contract string, invokedList []*relayer.CrossContract) error {
	if err := checkChainID(chainID); err != nil {
		return errors.WithMessage(err, "check chain id error")
	}
	if err := checkContractName(contract); err != nil {
		return errors.WithMessage(err, "check contract name error")
	}
	if err := checkInvokedList(invokedList); err != nil {
		return errors.WithMessage(err, "check invoked list error")
	}
	return nil
}

func checkString(reg string, s string) (bool, error) {
	if reg == "" {
		return false, errors.New("string is empty")
	}
	r, err := regexp.Compile(reg)
	if err != nil {
		return false, errors.WithMessage(err, "regexp compile error")
	}
	return r.MatchString(s), nil
}
