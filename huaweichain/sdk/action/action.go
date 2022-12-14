/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package action

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"git.huawei.com/huaweichain/common/logger"
	"git.huawei.com/huaweichain/sdk/utils"
)

var log = logger.GetModuleLogger("go sdk", "action")

// Config is the definition of action config.
type Config struct {
	hostOverride   string
	host           string
	port           int
	tlsEnable      bool
	transportCreds credentials.TransportCredentials
}

// NewConfig is used to create an config instance.
func NewConfig(hostOverride string, host string, port int, tlsEnable bool,
	transportCreds credentials.TransportCredentials) *Config {
	return &Config{
		hostOverride:   hostOverride,
		host:           host,
		port:           port,
		tlsEnable:      tlsEnable,
		transportCreds: transportCreds,
	}
}

type action struct {
	config *Config
	conn   *grpc.ClientConn
}

func (a *action) newClientConn() (*grpc.ClientConn, error) {
	ipAddr := a.config.host + ":" + strconv.Itoa(a.config.port)
	options := a.getOpts()
	ctx, cancel := context.WithTimeout(context.Background(), utils.GetTimeout()*time.Second)
	defer cancel()
	cc, err := grpc.DialContext(ctx, ipAddr, options...)
	if err != nil {
		return nil, errors.WithMessage(err, "grpc dial context error")
	}
	a.conn = cc
	return a.conn, nil
}

func (a *action) getOpts() []grpc.DialOption {
	var options []grpc.DialOption
	options = append(options, grpc.WithBlock())
	if a.config.tlsEnable {
		options = append(options, grpc.WithTransportCredentials(a.config.transportCreds))
	} else {
		options = append(options, grpc.WithInsecure())
		log.Infof("default insecure for go sdk.....")
	}

	return options
}
