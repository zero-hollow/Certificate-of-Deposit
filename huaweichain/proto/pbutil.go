/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2021. All rights reserved.
 */

// Package proto is a wrapper of gogo protobuf.
package proto

import "github.com/gogo/protobuf/proto"

// Marshal is an adapter function for calling gogo protobuf Marshal interface.
func Marshal(m proto.Marshaler) ([]byte, error) {
	return m.Marshal()
}

// Unmarshal is an adapter function for calling gogo protobuf Unmarshal interface.
func Unmarshal(bytes []byte, um proto.Unmarshaler) error {
	return um.Unmarshal(bytes)
}

// Marshaler is the interface representing objects that can marshal themselves.
type Marshaler interface {
	Marshal() ([]byte, error)
}

// Unmarshaler is the interface representing objects that can
// unmarshal themselves.  The argument points to data that may be
// overwritten, so implementations should not keep references to the
// buffer.
// Unmarshal implementations should not clear the receiver.
// Any unmarshaled data should be merged into the receiver.
// Callers of Unmarshal that do not want to retain existing data
// should Reset the receiver before calling Unmarshal.
type Unmarshaler interface {
	Unmarshal([]byte) error
}

// Message is implemented by generated protocol buffer messages.
type Message interface {
	Marshaler
	Unmarshaler
	Reset()
	String() string
	ProtoMessage()
}
