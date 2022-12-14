/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"strings"

	"github.com/pkg/errors"
)

// NewDomainConfig is used to new an instance of DomainConfig.
func NewDomainConfig(domainConfigPath string) (*DomainConfig, error) {
	v, errMsg := parseCfg(domainConfigPath)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "parse config file error")
	}

	domainConfig := &DomainConfig{}

	errMsg = unmarshal(v, &domainConfig)
	if errMsg != nil {
		return nil, errors.WithMessage(errMsg, "error unmarshaling config into struct")
	}
	return domainConfig, nil
}

// DomainConfig is the definition of domain config.
type DomainConfig struct {
	Targets []TargetDomain
}

// TargetDomain is the definition of target domain.
type TargetDomain struct {
	Parent  string
	Domains []string
}

// BuildDomainArray is used to build all domains into an array.
func (c *DomainConfig) BuildDomainArray(targets []TargetDomain) []string {
	domains := make(map[string]struct{})
	for i := 0; i < len(targets); i++ {
		extractFullDomainPath(domains, targets[i])
	}
	return removePrefix(domains)
}

func extractFullDomainPath(domains map[string]struct{}, targetDomain TargetDomain) {
	if domains == nil {
		return
	}
	for _, d := range targetDomain.Domains {
		domain := targetDomain.Parent + d
		domains[domain] = struct{}{}
	}
}

func removePrefix(domainMap map[string]struct{}) []string {
	var filteredDomains []string
	for domain := range domainMap {
		if !isPrefix(domain, domainMap) {
			filteredDomains = append(filteredDomains, domain)
		}
	}
	return filteredDomains
}

func isPrefix(domain string, domains map[string]struct{}) bool {
	for d := range domains {
		if d != domain && strings.HasPrefix(d, domain) {
			return true
		}
	}
	return false
}
