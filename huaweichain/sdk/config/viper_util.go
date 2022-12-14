/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package config

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cast"

	"git.huawei.com/huaweichain/common/viperutil"

	"github.com/pkg/errors"

	"github.com/mitchellh/mapstructure"

	"github.com/spf13/viper"
)

// unmarshal is intended to unmarshal a config file into a structure
func unmarshalClientConfig(v *viper.Viper, output interface{}) error {
	leafKeys, err := viperutil.GetKeysRecursively("", v.Get, v.AllSettings())
	if err != nil {
		return errors.WithMessage(err, "get keys recursively error")
	}
	if err := processNodes(leafKeys); err != nil {
		return errors.WithMessage(err, "process nodes error")
	}
	return mapstructure.Decode(leafKeys, output)
}

// unmarshal is intended to unmarshal a config file into a structure
func unmarshalRelayerConfig(v *viper.Viper, output interface{}) error {
	leafKeys, err := viperutil.GetKeysRecursively("", v.Get, v.AllSettings())
	if err != nil {
		return errors.WithMessage(err, "get keys recursively error")
	}
	if err := processRelayers(leafKeys); err != nil {
		return errors.WithMessage(err, "process relayers error")
	}
	return mapstructure.Decode(leafKeys, output)
}

// unmarshal is intended to unmarshal a config file into a structure
func unmarshalGenesisConfig(v *viper.Viper, output interface{}) error {
	configMapTemp, err := viperutil.GetKeysRecursively("", v.Get, v.AllSettings())
	if err != nil {
		return errors.WithMessage(err, "get keys recursively error")
	}
	configMap, err := processGenesisBlock(configMapTemp)
	if err != nil {
		return errors.WithMessage(err, "process genesis block error")
	}
	return mapstructure.Decode(configMap, output)
}

// unmarshal is intended to unmarshal a config file into a structure
func unmarshal(v *viper.Viper, output interface{}) error {
	leafKeys, err := viperutil.GetKeysRecursively("", v.Get, v.AllSettings())
	if err != nil {
		return errors.WithMessage(err, "get keys recursively error")
	}
	return mapstructure.Decode(leafKeys, output)
}

func getGenesisBlock(configMap map[string]interface{}) (map[string]interface{}, error) {
	genesisBlockMapTmp := configMap["genesisblock"]
	genesisBlockMap, ok := genesisBlockMapTmp.(map[string]interface{})
	if !ok {
		return nil, errors.New("type assert error: map[string]interface{}")
	}
	return genesisBlockMap, nil
}

func processGenesisBlock(configMap map[string]interface{}) (map[string]interface{}, error) {
	genesisBlock, err := getGenesisBlock(configMap)
	if err != nil {
		return nil, errors.WithMessage(err, "process genesis block...failed")
	}
	if ok := checkGenesisBlock(genesisBlock); !ok {
		return nil, errors.New("yaml file lack of elements")
	}
	if err := convertGenesisBlock(genesisBlock); err != nil {
		return nil, errors.WithMessage(err, "process genesis block...failed")
	}
	return configMap, nil
}

func convertGenesisBlock(genesisBlock map[string]interface{}) error {
	keyNames := []string{"organizations", "consenters"}
	for _, key := range keyNames {
		err := convertKey2Array(genesisBlock, key)
		if err != nil {
			return err
		}
	}
	if minPltfVersion, exist := genesisBlock["minplatformversion"]; exist {
		verStr := cast.ToString(minPltfVersion)
		if isFloat(minPltfVersion) && !strings.Contains(verStr, ".") {
			verStr = fmt.Sprintf("%s.0", verStr)
		}
		genesisBlock["minplatformversion"] = verStr
	}
	return nil
}

func isFloat(val interface{}) bool {
	switch val.(type) {
	case float32:
		return true
	case float64:
		return true
	default:
		return false
	}
}

func convertKey2Array(m map[string]interface{}, key string) error {
	if m == nil {
		return errors.New("map is nil")
	}
	mapKey, ok := m[key].(map[string]interface{})
	if !ok {
		return errors.Errorf("type assert error: %s", key)
	}
	organizationsArray, err := convertMap2Array(mapKey)
	if err != nil {
		return errors.WithMessagef(err, "[%s] convert map to array error", key)
	}
	m[key] = organizationsArray
	return nil
}

func checkGenesisBlock(genesisBlock map[string]interface{}) bool {
	// TODO: add param check
	return true
}

func processNodes(configMap map[string]interface{}) error {
	return convertKey2Array(configMap, "nodes")
}

func convertMap2Array(m map[string]interface{}) ([]interface{}, error) {
	var keyArray []string
	for key := range m {
		keyArray = append(keyArray, key)
	}
	sort.Strings(keyArray)

	var array []interface{}
	for _, key := range keyArray {
		value := m[key]
		e, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.New("type assert error: map[string]interface{}")
		}
		e["ID"] = key
		array = append(array, e)
	}
	return array, nil
}

func processRelayers(configMap map[string]interface{}) error {
	if configMap == nil {
		return errors.New("config map is nil")
	}
	relayersMapTmp := configMap["relayers"]
	relayersMap, ok := relayersMapTmp.(map[string]interface{})
	if !ok {
		return errors.Errorf("type assert error: expected type: map[string]interface{}, but real"+
			" type: %v", reflect.TypeOf(relayersMapTmp))
	}
	relayers, err := convertMap2Array(relayersMap)
	if err != nil {
		return errors.WithMessage(err, "convert map to array error")
	}
	configMap["relayers"] = relayers
	return nil
}
