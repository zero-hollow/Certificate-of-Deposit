/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package logger

import (
	"github.com/rs/zerolog"
)

// WienerLogger struct is the wrapper of logger.
type WienerLogger struct {
	logger *zerolog.Logger
}

// Write implements io.Writer interface.
func (l *WienerLogger) Write(p []byte) (n int, err error) {
	return l.logger.Write(p)
}

// Logger returns internal zerolog logger implementation.
func (l *WienerLogger) Logger() *zerolog.Logger {
	return l.logger
}

// Trace is used to print Trace level message
func (l *WienerLogger) Trace(msg string) {
	l.logger.Trace().Msg(msg)
}

// Tracef is used to print formatted Trace level message
func (l *WienerLogger) Tracef(format string, args ...interface{}) {
	l.logger.Trace().Msgf(format, args...)
}

// Debug is used to print Debug level message
func (l *WienerLogger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf is used to print formatted Debug level message
func (l *WienerLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info is used to print Info level message
func (l *WienerLogger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof is used to print formatted Info level message
func (l *WienerLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn is used to print Warn level message
func (l *WienerLogger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf is used to print formatted Warn level message
func (l *WienerLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error is used to print Error level message
func (l *WienerLogger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf is used to print formatted Error level message
func (l *WienerLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// Fatal is used to print Fatal level message
func (l *WienerLogger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf is used to print formatted Fatal level message
func (l *WienerLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// Panic is used to print panic message with log.
func (l *WienerLogger) Panic(msg string) {
	l.logger.Panic().Msg(msg)
}

// Panicf is used to print formatted panic message with log.
func (l *WienerLogger) Panicf(format string, args ...interface{}) {
	l.logger.Panic().Msgf(format, args...)
}
