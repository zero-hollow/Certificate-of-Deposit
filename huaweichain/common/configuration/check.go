/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package configuration is the implementation of config check for both client and node.
package configuration

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Entry contains the information and min, max, default value of the configuration
// set 0 to default value && check cfg range [min, max]
type Entry struct {
	Name       string
	Cfg        interface{} // use pointer type! in order to change origin value
	DefaultCfg interface{}
	Min        interface{}
	Max        interface{}
}

// FUNCTION
// check range: int, uint
// set default: int, uint, string
// not support: bool, []byte
func (e *Entry) checkEntry(defaultCfgLogs, errorCfgLogs *checkLogs) error {
	switch typeName(e.Cfg) {
	case "int8", "int16", "int32", "int64":
		e.setDefaultInt(defaultCfgLogs)
		e.checkOutRangeInt(errorCfgLogs)
	case "uint8", "uint16", "uint32", "uint64":
		e.setDefaultUint(defaultCfgLogs)
		e.checkOutRangeUint(errorCfgLogs)
	case "string":
		e.setDefaultString(defaultCfgLogs)
	default:
		return errors.Errorf("type of %s is not supported by param check...failed!", e.Name)
	}
	return nil
}

func (e *Entry) setDefaultInt(logs *checkLogs) {
	val := intRefValue(e.Cfg)
	defaultCfg := intValue(e.DefaultCfg)
	if val != 0 {
		return
	}
	setIntRef(e.Cfg, defaultCfg)
	logs.addLog(fmt.Sprintf("%s property value has not been set or set to zero, change to default value %v",
		e.Name, defaultCfg))
}

func (e *Entry) checkOutRangeInt(logs *checkLogs) {
	val := intRefValue(e.Cfg)
	min := intValue(e.Min)
	max := intValue(e.Max)
	if val > max {
		logs.addLog(fmt.Sprintf("%s value must less than %v, but %v", e.Name, max, val))
		return
	}
	if val < min {
		logs.addLog(fmt.Sprintf("%s value must greater than %v, but %v", e.Name, min, val))
	}
}

func (e *Entry) setDefaultUint(logs *checkLogs) {
	val := uintRefValue(e.Cfg)
	defaultCfg := uintValue(e.DefaultCfg)
	if val != 0 {
		return
	}
	setUintRef(e.Cfg, defaultCfg)
	logs.addLog(fmt.Sprintf("%s property value has not been set or set to zero, change to default value %v",
		e.Name, defaultCfg))
}

func (e *Entry) checkOutRangeUint(logs *checkLogs) {
	val := uintRefValue(e.Cfg)
	min := uintValue(e.Min)
	max := uintValue(e.Max)
	if val > max {
		logs.addLog(fmt.Sprintf("%s value must less than %v, but %v", e.Name, max, val))
		return
	}
	if val < min {
		logs.addLog(fmt.Sprintf("%s value must greater than %v, but %v", e.Name, min, val))
	}
}

func (e *Entry) setDefaultString(logs *checkLogs) {
	val := stringRefValue(e.Cfg)
	defaultCfg := stringValue(e.DefaultCfg)
	if val != "" {
		return
	}
	setStringRef(e.Cfg, defaultCfg)
	logs.addLog(fmt.Sprintf("%s is empty string...set default: %v", e.Name, defaultCfg))
}

// typeName get point type name by reflection
func typeName(i interface{}) string {
	return reflect.TypeOf(i).Elem().Name()
}

func intRefValue(valRef interface{}) int64 {
	return reflect.ValueOf(valRef).Elem().Int()
}

func intValue(val interface{}) int64 {
	return reflect.ValueOf(val).Int()
}

func setIntRef(valRef interface{}, defaultCfg int64) {
	reflect.ValueOf(valRef).Elem().SetInt(defaultCfg)
}

func uintRefValue(valRef interface{}) uint64 {
	return reflect.ValueOf(valRef).Elem().Uint()
}

func uintValue(val interface{}) uint64 {
	return uint64(reflect.ValueOf(val).Int())
}

func setUintRef(valRef interface{}, defaultCfg uint64) {
	reflect.ValueOf(valRef).Elem().SetUint(defaultCfg)
}

func stringRefValue(valRef interface{}) string {
	return reflect.ValueOf(valRef).Elem().String()
}

func stringValue(val interface{}) string {
	return reflect.ValueOf(val).String()
}

func setStringRef(valRef interface{}, defaultCfg string) {
	reflect.ValueOf(valRef).Elem().SetString(defaultCfg)
}

// checkLogs contains check param info
type checkLogs struct {
	cnt  int
	logs []string
}

// LogStr returns cnt and a joined string of check logs
func (ci *checkLogs) LogStr() (int, string) {
	logStr := strings.TrimSpace(strings.Join(ci.logs, "\n"))
	return ci.cnt, logStr
}

func (ci *checkLogs) addLog(log string) {
	ci.cnt++
	logFormat := fmt.Sprintf("[%v] %s", ci.cnt, log)
	ci.logs = append(ci.logs, logFormat)
}

// List contains all the param check info and check logs
type List struct {
	Name           string
	Entries        []Entry
	ErrorCfgLogs   checkLogs // check failed param cnt && logs
	DefaultCfgLogs checkLogs // set to default value cnt && logs
}

// CheckCfgList travel all cfg entries, to set 0 to default value && check range [min, max]
func (l *List) CheckCfgList(setDefaultInfo, outRangeInfo *checkLogs) error {
	for i := 0; i < len(l.Entries); i++ {
		err := l.Entries[i].checkEntry(setDefaultInfo, outRangeInfo)
		if err != nil {
			return err
		}
	}
	return nil
}
