/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package node

import (
	"git.huawei.com/huaweichain/sdk/action"
	"git.huawei.com/huaweichain/sdk/config"
	"github.com/pkg/errors"
)

// Relayer is the definition of Relayer.
type Relayer struct {
	*config.Node
	tls              *TLS
	CrossChainAction *action.CrossChainAction
	tlsEnable        bool
}

// NewRelayer is used to create a relayer.
func NewRelayer(node *config.Node, tlsEnable bool, tls *TLS) (*Relayer, error) {
	config, err := getConfig(node, tlsEnable, tls)
	if err != nil {
		return nil, errors.WithMessage(err, "get config error")
	}
	return &Relayer{
		Node:             node,
		tlsEnable:        tlsEnable,
		tls:              tls,
		CrossChainAction: action.NewCrossChainAction(config),
	}, nil
}
