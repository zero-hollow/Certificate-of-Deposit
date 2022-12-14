/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package logger

// LogCfg log config struct
type LogCfg struct {
	Type         string
	Level        string
	ModuleLevels map[string]string
	Rotation     LogRotation
}

// LogRotation contains configs for log rotation.
type LogRotation struct {
	Enabled    bool
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	LocalTime  bool
	Compress   bool
}
