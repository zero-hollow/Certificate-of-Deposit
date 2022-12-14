/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

// Package logger is the implementation of log.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	empty                 = ""
	logTypePretty         = "pretty"
	logTypeJSON           = "json"
	moduleNameKey         = "module"
	loggerNameKey         = "name"
	loggerModuleSeparator = "."

	defaultLevel          = zerolog.InfoLevel
	defaultSkipFrameCount = 3
	defaultLogType        = "json"
	defaultTimeformat     = "2006-01-02 15:04:05.000 -0700"

	defaultLogFileName = "wnode.log"

	initialized = iota
	notInitialized
)

var wienerLogger *WienerLogger
var moduleLevels map[string]zerolog.Level

type observer func(writer io.Writer, level zerolog.Level)

var hasInit int32 = notInitialized
var observers []observer
var obLock sync.Mutex

/// nolint
func init() {
	writer, err := getLogWriter(defaultLogType, &LogRotation{Enabled: false})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	logger := zerolog.New(writer).
		Level(defaultLevel).
		With().
		Timestamp().
		CallerWithSkipFrameCount(defaultSkipFrameCount).
		Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	wienerLogger = &WienerLogger{logger: &logger}
}

// Init is used to init log config.
func Init(config *LogCfg) error {
	writer, err := getLogWriter(config.Type, &config.Rotation)
	if err != nil {
		return err
	}
	level, err := zerolog.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		return fmt.Errorf("unknown log level: %s", config.Level)
	}
	logger := wienerLogger.logger.
		Output(writer).
		Level(level)
	wienerLogger.logger = &logger

	moduleLevels = make(map[string]zerolog.Level)
	for module, levelStr := range config.ModuleLevels {
		moduleLevel, err := zerolog.ParseLevel(strings.ToLower(levelStr))
		if err != nil {
			return fmt.Errorf("unknown module log level: module - %s, level - %s", module, levelStr)
		}
		moduleLevels[module] = moduleLevel
	}
	atomic.StoreInt32(&hasInit, initialized)
	for _, ob := range observers {
		ob(writer, level)
	}
	observers = nil
	return nil
}

// GetDefaultLogger is used to get default logger.
func GetDefaultLogger() *WienerLogger {
	return wienerLogger
}

// GetLogger is used to get logger with specified logger name.
func GetLogger(loggerName string) *WienerLogger {
	return GetModuleLogger(empty, loggerName)
}

// GetModuleLogger is used to get the logger for a specified module.
// different modules might have different log level, which can be set
// in config file.
func GetModuleLogger(moduleName, loggerName string) *WienerLogger {
	subLogger := wienerLogger.logger.
		With().
		Str(moduleNameKey, moduleName).
		Str(loggerNameKey, loggerName).
		Logger()
	lg := &WienerLogger{logger: &subLogger}

	init := atomic.LoadInt32(&hasInit)
	if init == notInitialized {
		addObserver(func(writer io.Writer, level zerolog.Level) {
			logger := subLogger.Output(writer).Level(moduleLevel(moduleName, level))
			lg.logger = &logger
		})
	} else {
		ml, ok := maybeUpdateLevel(moduleName, defaultLevel)
		if ok {
			logger := subLogger.Level(ml)
			lg.logger = &logger
		}
	}
	return lg
}

func maybeUpdateLevel(moduleName string, defaultLevel zerolog.Level) (zerolog.Level, bool) {
	l := moduleLevel(moduleName, defaultLevel)
	if l == defaultLevel {
		return defaultLevel, false
	}
	return l, true
}

func moduleLevel(moduleName string, defaultLevel zerolog.Level) zerolog.Level {
	var ms []string
	for m := range moduleLevels {
		ms = append(ms, m)
	}
	sort.Slice(ms, func(i, j int) bool {
		return ms[i] < ms[j]
	})
	l := defaultLevel
	for _, m := range ms {
		if moduleName == m || strings.HasPrefix(moduleName, m+loggerModuleSeparator) {
			l = moduleLevels[m]
		}
	}
	return l
}

func getLogWriter(logType string, rotation *LogRotation) (io.Writer, error) {
	var writer io.Writer = os.Stderr
	if rotation.Enabled {
		writer = getRotationWriter(rotation)
	}
	switch strings.ToLower(logType) {
	case logTypePretty:
		return &Writer{out: writer, timeFormat: defaultTimeformat}, nil
	case logTypeJSON:
		return writer, nil
	case empty:
		return &Writer{out: writer, timeFormat: defaultTimeformat}, nil
	}
	return nil, fmt.Errorf("unknown log type: %s", logType)
}

func getRotationWriter(rotation *LogRotation) io.Writer {
	writer := &lumberjack.Logger{}
	if len(rotation.Filename) == 0 {
		writer.Filename = filepath.Join(os.TempDir(), defaultLogFileName)
	} else {
		writer.Filename = rotation.Filename
	}
	if rotation.MaxSize > 0 {
		writer.MaxSize = rotation.MaxSize
	}
	if rotation.MaxAge > 0 {
		writer.MaxAge = rotation.MaxAge
	}
	if rotation.MaxBackups > 0 {
		writer.MaxBackups = rotation.MaxBackups
	}
	writer.LocalTime = rotation.LocalTime
	writer.Compress = rotation.Compress
	return writer
}

func addObserver(ob observer) {
	obLock.Lock()
	observers = append(observers, ob)
	obLock.Unlock()
}

// InitTestingLogger initializes default logger for unit testing.
func InitTestingLogger(level string) {
	writer := &Writer{out: os.Stderr, timeFormat: "15:04:05.000"}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	logger := zerolog.New(writer).
		Level(lvl).
		With().
		Timestamp().
		CallerWithSkipFrameCount(defaultSkipFrameCount).
		Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	wienerLogger = &WienerLogger{logger: &logger}
}

// TestingLogger gets an logger instance for unit testing.
func TestingLogger() *WienerLogger {
	return wienerLogger
}
