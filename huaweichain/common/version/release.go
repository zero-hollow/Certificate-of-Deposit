/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
 */

//go:generate go run gen.go -releaseVersion $RELEASE_VERSION -platformVersion $PLATFORM_VERSION

package version

import (
	"github.com/pkg/errors"

	"git.huawei.com/huaweichain/common/logger"
)

const notProvided = "Not provided"

var (
	// ReleaseInfo contains current release version and platform version.
	ReleaseInfo *Release
)

// nolint:gochecknoinits // no risk to use init here.
func init() {
	lg := logger.GetModuleLogger("common", "version")
	rel, err := newRelease(ReleaseVersion, PlatformVersion)
	if err != nil {
		lg.Fatalf("parse version error: %s", err)
	}
	ReleaseInfo = rel
	initVersions()
}

// Release contains current release version and platform version.
type Release struct {
	ReleaseVersion  *Version
	PlatformVersion *Version
}

func newRelease(releaseVersion, platformVersion string) (*Release, error) {
	if releaseVersion == notProvided {
		return nil, errors.New("release version is not provided when building wnode")
	}
	if platformVersion == notProvided {
		return nil, errors.New("platform version is not provided when building wnode")
	}
	relVer, err := NewVersion(releaseVersion)
	if err != nil {
		return nil, errors.WithMessagef(err, "parse release version %s", releaseVersion)
	}
	pltfVer, err := NewVersion(platformVersion)
	if err != nil {
		return nil, errors.WithMessagef(err, "parse platform version %s", platformVersion)
	}
	return &Release{
		ReleaseVersion:  relVer,
		PlatformVersion: pltfVer,
	}, nil
}
