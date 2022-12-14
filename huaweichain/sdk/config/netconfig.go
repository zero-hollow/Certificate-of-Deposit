/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package config

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/proto/common"
)

const (
	keySeparator           = ":"
	allocCoordinatorNumKey = "alloc:coordinator_num"
	allocPeerNumKey        = "alloc:peer_num"
	allocTypeKey           = "alloc:type"
	blockBatchKey          = "block_batch"
	attachableKey          = "attachable"
	allocConfFanKey        = "alloc:conf:fan"
	domainKey              = "domains"
)

// SetZoneConf is used to set zone conf.
func SetZoneConf(zoneConf *common.ZoneConf, m map[string]interface{}, template *common.ZoneConf) error {
	err := setAllocType(zoneConf, m, template)
	if err != nil {
		return errors.WithMessage(err, "set alloc type error")
	}
	err = setAttachable(zoneConf, m, template)
	if err != nil {
		return errors.WithMessage(err, "set attachable error")
	}
	err = setBlockBatch(zoneConf, m, template)
	if err != nil {
		return errors.WithMessage(err, "set block batch error")
	}
	err = setCoordinatorNum(zoneConf, m, template)
	if err != nil {
		return errors.WithMessage(err, "set coordinator num error")
	}
	err = setPeerNum(zoneConf, m, template)
	if err != nil {
		return errors.WithMessage(err, "set peer num error")
	}
	err = setAlloc(zoneConf, m, template)
	if err != nil {
		return errors.WithMessage(err, "set alloc error")
	}
	return nil
}

func getValue(m map[string]interface{}, key string) (value interface{}, ok bool, err error) {
	index := strings.Index(key, keySeparator)
	if index == -1 {
		value, ok = m[key]
		return value, ok, nil
	}

	k := key[0:index]
	key = key[index+1:]
	value, ok = m[k]
	if !ok {
		return nil, false, nil
	}
	mm, ok := value.(map[string]interface{})
	if ok {
		return getValue(mm, key)
	}

	tempMap, ok := value.(map[interface{}]interface{})
	if !ok {
		return nil, false, errors.Errorf("type assert error: expected type: "+
			"map[interface{}]interface{}, but got: %v", reflect.TypeOf(value))
	}
	m, err = convert(tempMap)
	if err != nil {
		return nil, false, errors.WithMessage(err, "convert map error")
	}
	return getValue(m, key)
}

func convert(m map[interface{}]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range m {
		k, ok := key.(string)
		if !ok {
			return nil, errors.Errorf("type assert error: expected type: string, but got: %v",
				reflect.TypeOf(value))
		}
		result[k] = value
	}
	return result, nil
}

func setAllocType(zoneConf *common.ZoneConf, m map[string]interface{}, template *common.ZoneConf) error {
	typeStrValue, ok, err := getValue(m, allocTypeKey)
	if err != nil {
		return errors.WithMessagef(err, "get value [%v] error", allocTypeKey)
	}
	if !ok {
		zoneConf.Alloc.Type = template.Alloc.Type
		return nil
	}
	typeStr, ok := typeStrValue.(string)
	if !ok {
		return errors.Errorf("type assert error: expected type: string, but got: %v",
			reflect.TypeOf(typeStrValue))
	}
	zoneConf.Alloc.Type = common.AllocType(common.AllocType_value[strings.ToUpper(typeStr)])
	return nil
}

func setAttachable(zoneConf *common.ZoneConf, m map[string]interface{}, template *common.ZoneConf) error {
	attachableStrValue, ok, err := getValue(m, attachableKey)
	if err != nil {
		return errors.WithMessagef(err, "get value [%v] error", attachableKey)
	}
	if !ok {
		zoneConf.Attachable = template.Attachable
		return nil
	}
	attachable, ok := attachableStrValue.(bool)
	if !ok {
		return errors.Errorf("type assert error: expected type: bool, but got: %v",
			reflect.TypeOf(attachableStrValue))
	}
	zoneConf.Attachable = attachable
	return nil
}

func setBlockBatch(zoneConf *common.ZoneConf, m map[string]interface{}, template *common.ZoneConf) error {
	blockBatchStrValue, ok, err := getValue(m, blockBatchKey)
	if err != nil {
		return errors.WithMessagef(err, "get value [%v] error", blockBatchKey)
	}
	if !ok {
		zoneConf.BlockBatch = template.BlockBatch
		return nil
	}
	blockBatch, ok := blockBatchStrValue.(int)
	if !ok {
		return errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(blockBatchStrValue))
	}
	zoneConf.BlockBatch = uint32(blockBatch)
	return nil
}

func setCoordinatorNum(zoneConf *common.ZoneConf, m map[string]interface{}, template *common.ZoneConf) error {
	allocCoordinatorNumStrValue, ok, err := getValue(m, allocCoordinatorNumKey)
	if err != nil {
		return errors.WithMessagef(err, "get value [%v] error", allocCoordinatorNumKey)
	}
	if !ok {
		zoneConf.Alloc.MaxCoordinatorNum = template.Alloc.MaxCoordinatorNum
		return nil
	}
	allocCoordinatorNum, ok := allocCoordinatorNumStrValue.(int)
	if !ok {
		return errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(allocCoordinatorNumStrValue))
	}
	zoneConf.Alloc.MaxCoordinatorNum = uint32(allocCoordinatorNum)
	return nil
}

func setPeerNum(zoneConf *common.ZoneConf, m map[string]interface{}, template *common.ZoneConf) error {
	peerNumStrValue, ok, err := getValue(m, allocPeerNumKey)
	if err != nil {
		return errors.WithMessagef(err, "get value [%v] error", allocPeerNumKey)
	}
	if !ok {
		zoneConf.Alloc.MaxPeerNum = template.Alloc.MaxPeerNum
		return nil
	}
	peerNum, ok := peerNumStrValue.(int)
	if !ok {
		return errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(peerNumStrValue))
	}
	zoneConf.Alloc.MaxPeerNum = uint32(peerNum)
	return nil
}

func setAlloc(zoneConf *common.ZoneConf, m map[string]interface{}, template *common.ZoneConf) error {
	confFanStrValue, ok, err := getValue(m, allocConfFanKey)
	if err != nil {
		return errors.WithMessagef(err, "get value [%v] error", allocConfFanKey)
	}
	if !ok {
		zoneConf.Alloc.Alloc = template.Alloc.Alloc
		return nil
	}
	confFan, ok := confFanStrValue.(int)
	if !ok {
		return errors.Errorf("type assert error: expected type: int, but got: %v",
			reflect.TypeOf(confFanStrValue))
	}
	switch zoneConf.Alloc.Type {
	case common.BALANCE:
		zoneConf.Alloc.Alloc = &common.Allocator_Balance{Balance: &common.BalanceAlloc{Fan: uint32(confFan)}}
	default:
		break
	}
	return nil
}

// BuildZone convert Zone Properties to zone cfg
func BuildZone(z *common.ZoneProperties) (*common.Zone, error) {
	zone := common.Zone{
		Id:   z.Id,
		Conf: &common.ZoneConf{Alloc: &common.Allocator{}},
	}

	prop, err := getPropertyMap(z)
	if err != nil {
		return nil, err
	}

	if v, ok := prop[domainKey]; ok {
		zone.Domains = v.GetStrArr().Ss
	} else {
		return nil, errors.Errorf("zone[%s] has no assigned domain", zone.Id)
	}
	if v, ok := prop[attachableKey]; ok {
		zone.Conf.Attachable = v.GetBool()
	}
	if v, ok := prop[blockBatchKey]; ok {
		zone.Conf.BlockBatch = uint32(v.GetI32())
	}
	if v, ok := prop[allocTypeKey]; ok {
		zone.Conf.Alloc.Type = common.AllocType(v.GetI32())
	}
	if v, ok := prop[allocCoordinatorNumKey]; ok {
		zone.Conf.Alloc.MaxCoordinatorNum = uint32(v.GetI32())
	}
	if v, ok := prop[allocPeerNumKey]; ok {
		zone.Conf.Alloc.MaxPeerNum = uint32(v.GetI32())
	}

	switch zone.Conf.Alloc.Type {
	case common.BALANCE:
		if v, ok := prop[allocConfFanKey]; ok {
			zone.Conf.Alloc.Alloc = &common.Allocator_Balance{
				Balance: &common.BalanceAlloc{
					Fan: uint32(v.GetI32()),
				},
			}
		}
	default:
		return nil, errors.New("unsupported allocator type")
	}

	return &zone, nil
}

func getPropertyMap(z *common.ZoneProperties) (map[string]*common.Property, error) {
	prop := make(map[string]*common.Property, len(z.Properties))
	for _, p := range z.Properties {
		if _, ok := prop[p.K]; ok {
			return nil, errors.Errorf("has some duplicated key in properties, the key is %s", p.K)
		}
		prop[p.K] = p
	}
	return prop, nil
}

const (
	domainSeparator = "/"
	zoneSeparator   = "::"

	rootZone           = ""
	rootDomain         = "/"
	allSuffix          = "/*"
	allRecursiveSuffix = "/**"

	zoneIDReg        = `^(\w{1,100}){1}(::\w{1,100})*$`
	domainPathCfgReg = `^(/\w{1,100})+$`             // verify cfg domain operation, add/remove etc.
	domainPathTxReg  = `^(/\w{1,100})+(/[*]{1,2})?$` // verify tx domain

	maxStringSecNum = 100

	maxPeerNum        = 200
	minPeerNum        = 0
	minCoordinatorNum = 1
	maxCoordinatorNum = 9
)

// CheckSyncCfg verify network config param
func CheckSyncCfg(cfg *common.NetworkConfig) error {
	if err := CheckZoneCfg(cfg.ConsZone); err != nil {
		return err
	}
	if err := CheckZones(cfg.Zones); err != nil {
		return err
	}
	for i := 0; i < len(cfg.Domains); i++ {
		if err := CheckDomain(cfg.Domains[i]); err != nil {
			return err
		}
	}
	return nil
}

// CheckZones verify multiple zone config
func CheckZones(zs []*common.Zone) error {
	if err := checkRepeatZone(zs); err != nil {
		return err
	}
	for i := 0; i < len(zs); i++ {
		if err := checkZone(zs[i]); err != nil {
			return err
		}
	}
	return nil
}

func checkRepeatZone(zs []*common.Zone) error {
	m := make(map[string]bool)
	for i := 0; i < len(zs); i++ {
		if _, ok := m[zs[i].Id]; !ok {
			m[zs[i].Id] = true
		} else {
			return errors.Errorf("zone id is invalid: duplicate zone id: %v", zs[i].Id)
		}
	}
	return nil
}

// CheckZone verify zone config
func checkZone(z *common.Zone) error {
	if err := CheckZoneID(z.Id); err != nil {
		return err
	}
	if err := CheckDomainPath(z.Domains); err != nil {
		return err
	}
	if err := CheckZoneCfg(z.Conf); err != nil {
		return err
	}
	return nil
}

// CheckZoneCfg verify param in zone config
func CheckZoneCfg(z *common.ZoneConf) error {
	if z.Alloc.MaxCoordinatorNum < minCoordinatorNum || z.Alloc.MaxCoordinatorNum > maxCoordinatorNum {
		return errors.Errorf("zone cfg is invalid: coordinator num out of range[%v:%v], coordinator num: %v",
			minCoordinatorNum, maxCoordinatorNum, z.Alloc.MaxCoordinatorNum)
	}
	if z.Alloc.MaxPeerNum < minPeerNum || z.Alloc.MaxPeerNum > maxPeerNum {
		return errors.Errorf("zone cfg is invalid: max peer num out of range[%v:%v], peer num: %v",
			minPeerNum, maxPeerNum, z.Alloc.MaxPeerNum)
	}
	if z.Alloc.Type != common.BALANCE {
		return errors.Errorf("zone cfg is invalid: alloc type error, alloc type: %d", z.Alloc.Type)
	}
	return nil
}

// CheckZoneID verify zone id by regx
func CheckZoneID(s string) error {
	if s == "" {
		return nil
	}
	if err := checkLevelNum(s, zoneSeparator); err != nil {
		return errors.WithMessage(err, "zone id is invalid")
	}
	r, err := regexp.Compile(zoneIDReg)
	if err != nil {
		return errors.WithMessagef(err, "regexp compile error: %v", zoneIDReg)
	}
	if !r.Match([]byte(s)) {
		return errors.Errorf("zone id is invalid: mismatch regx, zone id: %v", s)
	}
	return nil
}

// CheckDomainPath verify domain path by regx (not for tx domain)
func CheckDomainPath(s []string) error {
	specialDomainPaths := []string{rootDomain}
	return checkDomainPath(s, domainPathCfgReg, specialDomainPaths)
}

// CheckTxDomainPath verify domain path by regx for tx, allow allRecursiveSuffix and allSuffix
func CheckTxDomainPath(s []string) error {
	specialDomainPaths := []string{allRecursiveSuffix, allSuffix, rootDomain}
	return checkDomainPath(s, domainPathTxReg, specialDomainPaths)
}

// CheckDomain verify domain cfg
func CheckDomain(d *common.Domain) error {
	return CheckDomainPath([]string{d.Path})
}

func checkLevelNum(s string, split string) error {
	secNum := len(strings.Split(s, split))
	if secNum > maxStringSecNum {
		return errors.Errorf("level should not larger than %d, level num: %v", maxStringSecNum, secNum)
	}
	return nil
}

func checkDomainPath(s []string, pathReg string, specialPaths []string) error {
	r, err := regexp.Compile(pathReg)
	if err != nil {
		return errors.WithMessagef(err, "regexp compile error: %v", pathReg)
	}
	for i := 0; i < len(s); i++ {
		for _, path := range specialPaths {
			if s[i] == path {
				return nil
			}
		}

		if err := checkLevelNum(s[i], domainSeparator); err != nil {
			return errors.WithMessage(err, "domain path is invalid")
		}
		if !r.Match([]byte(s[i])) {
			return errors.Errorf("domain path is invalid： mismatch regx, domain path: %v", s[i])
		}
	}
	return nil
}
