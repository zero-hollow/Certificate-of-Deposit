/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"io/ioutil"
	"path/filepath"

	"git.huawei.com/huaweichain/proto/common"
	"github.com/pkg/errors"
)

type updateConfig struct {
	Orgs []*oneOrg
}

type oneOrg struct {
	Name      string
	Option    string
	RootPath  string
	AdminPath string
	TlsPath   string // nolint:golint,stylecheck // conflict with java.
}

// GetOrgUpdates get org updates form yaml
func GetOrgUpdates(path string, decrypt DecryptFunc) (*common.ConfigSet_OrgUpdates, error) {
	v, errMsg := parseCfg(path)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "parse config file error")
	}
	config := &updateConfig{}
	errMsg = unmarshal(v, &config)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "error unmarshal config into struct")
	}

	updates := make([]*common.ConfigSet_OrgUpdate, 0, len(config.Orgs))
	for _, each := range config.Orgs {
		switch each.Option {
		case "remove":
			updates = append(updates, &common.ConfigSet_OrgUpdate{
				Operation:    common.OP_REMOVE,
				Organization: &common.Organization{Name: each.Name},
			})
		case "append":
			org, err := getOrg(each, decrypt)
			if err != nil {
				return nil, err
			}
			updates = append(updates, &common.ConfigSet_OrgUpdate{
				Operation:    common.OP_APPEND,
				Organization: org,
			})
		case "replace":
			org, err := getOrg(each, decrypt)
			if err != nil {
				return nil, err
			}
			updates = append(updates, &common.ConfigSet_OrgUpdate{
				Operation:    common.OP_REPLACE,
				Organization: org,
			})
		default:
			return nil, errors.New("unknown operation")
		}
	}
	return &common.ConfigSet_OrgUpdates{OrgUpdate: updates}, nil
}

func getOrg(org *oneOrg, decrypt DecryptFunc) (*common.Organization, error) {
	rootCert, rootErr := readFile(org.RootPath, decrypt)
	if rootErr != nil {
		return nil, errors.WithMessage(rootErr, "read root cert error")
	}
	adminCert, adminErr := readFile(org.AdminPath, decrypt)
	if adminErr != nil {
		return nil, errors.WithMessage(adminErr, "read admin cert error")
	}
	tlsCert, tlsErr := readFile(org.TlsPath, decrypt)
	if tlsErr != nil {
		return nil, errors.WithMessage(tlsErr, "read tls cert error")
	}
	return &common.Organization{
		Name:        org.Name,
		RootCert:    rootCert,
		AdminCert:   adminCert,
		TLSRootCert: tlsCert,
	}, nil
}

func readFile(path string, decrypt DecryptFunc) ([]byte, error) {
	cert, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.New("read file error")
	}
	return decrypt(cert)
}
