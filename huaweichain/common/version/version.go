/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

// Package version provides implementation of release and platform version.
package version

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	sep     = "."
	base    = 10
	bitSize = 64
)

// Version parses version string to build a version object.
// This provides method to compare to other versions.
type Version struct {
	segments []uint64
	ver      string
}

// NewVersion parses version string to build a version object.
func NewVersion(ver string) (*Version, error) {
	if len(ver) == 0 {
		return nil, errors.New("empty string")
	}
	segment, err := split(ver)
	if err != nil {
		return nil, errors.WithMessage(err, "split version string")
	}
	return &Version{
		segments: segment,
		ver:      ver,
	}, nil
}

func split(ver string) ([]uint64, error) {
	segmentsStr := strings.Split(ver, sep)
	segments := make([]uint64, len(segmentsStr))
	for i, seg := range segmentsStr {
		segment, err := strconv.ParseUint(seg, base, bitSize)
		if err != nil {
			return nil, errors.Errorf("parseUint for string [%s] failed: %s", seg, err)
		}
		segments[i] = segment
	}
	return segments, nil
}

// Compare compares to other version.
// It returns 1 if >, returns 0 if ==, and returns -1 if <.
func (v *Version) Compare(o *Version) int {
	if o == nil {
		return 1
	}

	vLen := len(v.segments)
	oLen := len(o.segments)
	min := vLen
	if vLen > oLen {
		min = oLen
	}
	for i := 0; i < min; i++ {
		if r := compare(v.segments[i], o.segments[i]); r != 0 {
			return r
		}
	}

	if vLen > oLen {
		for i := oLen; i < vLen; i++ {
			if r := compare(v.segments[i], 0); r != 0 {
				return r
			}
		}
		return 0
	}
	for i := vLen; i < oLen; i++ {
		if r := compare(0, o.segments[i]); r != 0 {
			return r
		}
	}
	return 0
}

func compare(a, b uint64) int {
	if a > b {
		return 1
	}
	if a == b {
		return 0
	}
	return -1
}

// Equal checks if this version is equal to input version.
func (v *Version) Equal(o *Version) bool {
	return v.Compare(o) == 0
}

// GreaterThan checks if this version is greater than input version.
func (v *Version) GreaterThan(o *Version) bool {
	return v.Compare(o) > 0
}

// GreaterThanOrEqual checks if this version is greater than or equal to input version.
func (v *Version) GreaterThanOrEqual(o *Version) bool {
	return v.Compare(o) >= 0
}

// LessThan checks if this version is less than input version.
func (v *Version) LessThan(o *Version) bool {
	return v.Compare(o) < 0
}

// LessThanOrEqual checks if this version is less than or equal to input version.
func (v *Version) LessThanOrEqual(o *Version) bool {
	return v.Compare(o) <= 0
}

// String returns a string of this version.
func (v *Version) String() string {
	return v.ver
}

// Bytes returns a bytes of this version.
func (v *Version) Bytes() []byte {
	return []byte(v.ver)
}
