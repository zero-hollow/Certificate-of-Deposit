/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

package version

const (
	platformVersion0_0 = "0.0"
	platformVersion1_0 = "1.0"
)

var (
	// PlatformVersion0_0 represents 0.0 version.
	PlatformVersion0_0 *Version

	// PlatformVersion1_0 represents 1.0 version.
	PlatformVersion1_0 *Version
)

func initVersions() {
	ver, _ := NewVersion(platformVersion0_0)
	PlatformVersion0_0 = ver
	ver, _ = NewVersion(platformVersion1_0)
	PlatformVersion1_0 = ver
}

// NewPlatformVersion parse platform version string and new a
// version object. If version string is empty, using the default
// 0.0 version.
func NewPlatformVersion(ver string) (*Version, error) {
	if ver == "" {
		return PlatformVersion0_0, nil
	}
	return NewVersion(ver)
}
