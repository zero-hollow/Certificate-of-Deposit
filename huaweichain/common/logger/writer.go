/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 */

package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	maxRemainingLength = 2
	bufMapCap          = 100
)

var bufPool = newBufPool()
var allowedFields = []string{
	zerolog.TimestampFieldName,
	zerolog.LevelFieldName,
	moduleNameKey,
	loggerNameKey,
	zerolog.CallerFieldName,
	zerolog.MessageFieldName,
}

// Writer struct is the definition of Writer.
type Writer struct {
	out        io.Writer
	timeFormat string
}

// Write writes log to out
func (w *Writer) Write(in []byte) (int, error) {
	fields := make(map[string]interface{})
	decoder := json.NewDecoder(bytes.NewReader(in))
	decoder.UseNumber()
	if err := decoder.Decode(&fields); err != nil {
		return 0, errors.WithMessage(err, "cannot decode fields")
	}

	buf := bufPool.GetBuf()
	defer bufPool.PutBuf(buf)
	w.writeFields(buf, fields)
	buf.WriteByte('\n')
	_, err := buf.WriteTo(w.out)

	return len(in), err
}

func trimPath(path string) string {
	paths := strings.Split(path, "/")
	var finalPath string
	pathLen := len(paths)
	switch {
	case pathLen >= maxRemainingLength:
		finalPath = strings.Join(paths[pathLen-maxRemainingLength:], "/")
	case pathLen == 1:
		finalPath = paths[0]
	default:
		finalPath = path
	}
	return finalPath
}

func (w *Writer) writeFields(buf io.StringWriter, fields map[string]interface{}) {
	for i, allowedField := range allowedFields {
		field, ok := fields[allowedField]
		if !ok {
			continue
		}
		if i > 0 {
			writeString(buf, " | ")
		}

		switch allowedField {
		case zerolog.TimestampFieldName:
			timestamp, err := parseTimestamp(field)
			if err != nil || timestamp == nil {
				writeString(buf, "<nil>")
			} else {
				writeString(buf, timestamp.Format(w.timeFormat))
			}
		case zerolog.LevelFieldName:
			writeString(buf, strings.ToUpper(field.(string)))
		case zerolog.CallerFieldName:
			s, ok := field.(string)
			if !ok {
				fmt.Printf("type assert error: string")
				return
			}
			trimedPath := trimPath(s)
			writeString(buf, trimedPath)
		default:
			writeString(buf, field.(string))
		}
	}
}

func parseTimestamp(timeIn interface{}) (*time.Time, error) {
	switch t := timeIn.(type) {
	case string:
		timestamp, err := time.Parse(zerolog.TimeFieldFormat, t)
		if err != nil {
			return nil, err
		}
		return &timestamp, nil
	case json.Number:
		timeInteger, err := t.Int64()
		if err != nil {
			return nil, err
		}
		var sec, nsec int64 = timeInteger, 0
		switch zerolog.TimeFieldFormat {
		case zerolog.TimeFormatUnixMs:
			nsec = int64(time.Duration(timeInteger) * time.Millisecond)
			sec = 0
		case zerolog.TimeFormatUnixMicro:
			nsec = int64(time.Duration(timeInteger) * time.Microsecond)
			sec = 0
		}
		timestamp := time.Unix(sec, nsec)
		return &timestamp, nil
	default:
		return nil, nil
	}
}

func writeString(buf io.StringWriter, msg string) {
	// / nolint
	// bytes.Buffer.WriteString function will not return error but nil only.
	if _, err := buf.WriteString(msg); err != nil {
		fmt.Errorf("write string error: %v", err)
	}
}

// buffersPool struct is the definition of buffer pool.
type buffersPool struct {
	pool *sync.Pool
}

// newBufPool is used to create an instance of buffer pool.
func newBufPool() *buffersPool {
	pool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, bufMapCap))
		},
	}
	return &buffersPool{pool: pool}
}

// GetBuf is used to get buffer from buffer pool.
func (b *buffersPool) GetBuf() *bytes.Buffer {
	return b.pool.Get().(*bytes.Buffer)
}

// PutBuf is used to add buffer to buffer pool.
func (b *buffersPool) PutBuf(buf *bytes.Buffer) {
	buf.Reset()
	b.pool.Put(buf)
}
