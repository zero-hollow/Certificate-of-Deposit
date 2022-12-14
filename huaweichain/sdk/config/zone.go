/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"reflect"
	"sort"
	"strings"

	"git.huawei.com/huaweichain/proto/common"
	"github.com/pkg/errors"
)

// ZoneList is the definition of zone list.
type ZoneList struct {
	Zones []map[string]interface{}
}

// NewZoneList is used to create an instance of zone list.
func NewZoneList(zonePath string) (*ZoneList, error) {
	v, err := parseCfg(zonePath)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config file error")
	}
	zoneList := &ZoneList{}
	err = unmarshal(v, zoneList)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal error")
	}
	return zoneList, nil
}

// BuildProperties is used to build properties.
func BuildProperties(m map[string]interface{}) ([]*common.Property, error) {
	var ps []*common.Property
	p, err := buildAttachable(m)
	if err != nil {
		return nil, errors.WithMessage(err, "build attachable error")
	}
	ps = append(ps, p)
	p, err = buildBlockBatch(m)
	if err != nil {
		return nil, errors.WithMessage(err, "build block batch error")
	}
	ps = append(ps, p)
	p, err = buildDomains(m)
	if err != nil {
		return nil, errors.WithMessage(err, "build domains error")
	}
	ps = append(ps, p)
	p, err = buildCoordinatorNum(m)
	if err != nil {
		return nil, errors.WithMessage(err, "build coordinator num error")
	}
	ps = append(ps, p)

	p, err = buildPeerNum(m)
	if err != nil {
		return nil, errors.WithMessage(err, "build peer num error")
	}
	ps = append(ps, p)
	p, err = buildAllocType(m)
	if err != nil {
		return nil, errors.WithMessage(err, "build alloc type error")
	}
	ps = append(ps, p)
	p, err = buildConfFan(m)
	if err != nil {
		return nil, errors.WithMessage(err, "build conf fan error")
	}
	ps = append(ps, p)
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].K < ps[j].K
	})
	return ps, nil
}

func buildAttachable(m map[string]interface{}) (*common.Property, error) {
	attachableStrValue, ok, err := getValue(m, attachableKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "get value [%v] error", attachableKey)
	}
	if !ok {
		return &common.Property{K: attachableKey, V: &common.Property_Bool{Bool: false}}, nil
	}
	attachable, ok := attachableStrValue.(bool)
	if !ok {
		return nil, errors.Errorf("type assert error: expected type: bool, but got: %v",
			reflect.TypeOf(attachableStrValue))
	}

	return &common.Property{K: attachableKey, V: &common.Property_Bool{Bool: attachable}}, nil
}

func buildBlockBatch(m map[string]interface{}) (*common.Property, error) {
	blockBatchStrValue, ok, err := getValue(m, blockBatchKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "get value [%v] error", blockBatchKey)
	}
	if !ok {
		return &common.Property{K: blockBatchKey}, nil
	}
	blockBatch, ok := blockBatchStrValue.(int)
	if !ok {
		return nil, errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(blockBatchStrValue))
	}
	return &common.Property{K: blockBatchKey, V: &common.Property_I32{I32: int32(blockBatch)}}, nil
}

func buildDomains(m map[string]interface{}) (*common.Property, error) {
	domainStrsValue, ok, err := getValue(m, domainKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "get value [%v] error", domainKey)
	}
	if !ok {
		return nil, errors.New("failed to build zone configuration, because zone domains are not assigned")
	}
	ds, ok := domainStrsValue.([]interface{})
	if !ok {
		return nil, errors.Errorf("type assert error: expected type: []interface{}, but real type: %v",
			reflect.TypeOf(domainStrsValue))
	}
	var domains []string
	for _, d := range ds {
		domain, ok := d.(string)
		if !ok {
			return nil, errors.Errorf("type assert error: expected type: string, but real type: %v",
				reflect.TypeOf(d))
		}
		domains = append(domains, domain)
	}
	ss := &common.Property_StrArr{Ss: domains}
	return &common.Property{K: domainKey, V: &common.Property_StrArr_{StrArr: ss}}, nil
}

func buildCoordinatorNum(m map[string]interface{}) (*common.Property, error) {
	allocCoordinatorNumStrValue, ok, err := getValue(m, allocCoordinatorNumKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "get value [%v] error", allocCoordinatorNumKey)
	}
	if !ok {
		return &common.Property{K: allocCoordinatorNumKey}, nil
	}
	allocCoordinatorNum, ok := allocCoordinatorNumStrValue.(int)
	if !ok {
		return nil, errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(allocCoordinatorNumStrValue))
	}
	return &common.Property{K: allocCoordinatorNumKey, V: &common.Property_I32{I32: int32(allocCoordinatorNum)}}, nil
}

func buildPeerNum(m map[string]interface{}) (*common.Property, error) {
	peerNumStrValue, ok, err := getValue(m, allocPeerNumKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "get value [%v] error", allocPeerNumKey)
	}
	if !ok {
		return &common.Property{K: allocPeerNumKey}, nil
	}
	peerNum, ok := peerNumStrValue.(int)
	if !ok {
		return nil, errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(peerNumStrValue))
	}
	return &common.Property{K: allocPeerNumKey, V: &common.Property_I32{I32: int32(peerNum)}}, nil
}

func buildAllocType(m map[string]interface{}) (*common.Property, error) {
	typeStrValue, ok, err := getValue(m, allocTypeKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "get value [%v] error", allocTypeKey)
	}
	if !ok {
		return &common.Property{K: allocTypeKey, V: &common.Property_I32{I32: int32(common.BALANCE)}}, nil
	}
	typeStr, ok := typeStrValue.(string)
	if !ok {
		return nil, errors.Errorf("type assert error: expected type: string, but got: %v",
			reflect.TypeOf(typeStrValue))
	}
	switch typeStr {
	case strings.ToLower(common.BALANCE.String()):
		return &common.Property{K: allocTypeKey, V: &common.Property_Str{Str: typeStr}}, nil
	default:
		return nil, errors.New("unsupported allocator type")
	}
}

func buildConfFan(m map[string]interface{}) (*common.Property, error) {
	confFanStrValue, ok, err := getValue(m, allocConfFanKey)
	if err != nil {
		return nil, errors.WithMessagef(err, "get value [%v] error", allocConfFanKey)
	}
	if !ok {
		return &common.Property{K: allocConfFanKey}, nil
	}
	confFan, ok := confFanStrValue.(int)
	if !ok {
		return nil, errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(confFanStrValue))
	}
	return &common.Property{K: allocConfFanKey, V: &common.Property_I32{I32: int32(confFan)}}, nil
}
