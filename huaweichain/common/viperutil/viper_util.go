/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package viperutil for parsing config data
package viperutil

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type viperGetter func(key string) interface{}

// GetKeysRecursively is used to parse configs.
func GetKeysRecursively(base string, getKey viperGetter, nodeKeys map[string]interface{}) (
	map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key := range nodeKeys {
		fqKey := base + key
		val := getKey(fqKey)

		switch v := val.(type) {
		case map[string]interface{}:
			value, err := GetKeysRecursively(fqKey+".", getKey, v)
			if err != nil {
				return nil, err
			}
			result[key] = value
		case map[interface{}]interface{}: // FIXME: no case match
			msi, err := toMapStringInterface(v)
			if err != nil {
				return nil, err
			}
			value, err := GetKeysRecursively(fqKey+".", getKey, msi)
			if err != nil {
				return nil, err
			}
			result[key] = value
		case nil:
			if fileVal := getKey(fqKey + ".File"); fileVal != nil {
				result[key] = map[string]interface{}{"File": fileVal}
			}
		default:
			result[key] = v
		}
	}
	return result, nil
}

func toMapStringInterface(mii map[interface{}]interface{}) (map[string]interface{}, error) {
	msi := make(map[string]interface{})
	for k, v := range mii {
		s, ok := k.(string)
		if !ok {
			return nil, errors.New("non string key-entry")
		}
		msi[s] = v
	}
	return msi, nil
}

// Unmarshal is intended to unmarshal a config file into a structure
func Unmarshal(v *viper.Viper, output interface{}) error {
	baseKeys := v.AllSettings()
	leafKeys, err := GetKeysRecursively("", v.Get, baseKeys)
	if err != nil {
		return err
	}
	if err = mapstructure.Decode(leafKeys, output); err != nil {
		return errors.Wrap(err, "failed to decode structure")
	}
	return nil
}
