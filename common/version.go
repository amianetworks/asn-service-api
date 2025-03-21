// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	VersionCompareGreater = 1 + iota
	VersionCompareLess
	VersionCompareEqual
)

type Version struct {
	Major uint64
	Minor uint64

	HasBuild bool
	Build    string
}

// InitVersion converts a version string, e.g., "v25.0.1", into Version structure.
func InitVersion(versionStr string) (Version, error) {
	versionParts := strings.Split(versionStr, ".")
	if len(versionParts) != 2 && len(versionParts) != 3 {
		return Version{}, fmt.Errorf("invalid version format, must be 'v.MAJOR.MINOR' or 'v.MAJOR.MINOR.BUILD', for example: v1.0 or v1.0.1-196912")
	}

	major, err := strconv.ParseUint(strings.Replace(versionParts[0], "v", "", -1), 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version [%s] due to: %v", versionParts[0], err)
	}
	minor, err := strconv.ParseUint(versionParts[1], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version [%s]  due to: %v", versionParts[1], err)
	}

	version := Version{
		Major: major,
		Minor: minor,
	}
	if len(versionParts) == 3 {
		version.HasBuild = true
		version.Build = versionParts[2]
	}

	return version, nil
}

// ToString converts a Version structure back to the version string.
func (v Version) ToString() string {
	return fmt.Sprintf("v%d.%d.%s", v.Major, v.Minor, v.Build)
}

func (v Version) Compare(target Version) int {
	// compare major
	if v.Major > target.Major {
		return VersionCompareGreater
	} else if v.Major < target.Major {
		return VersionCompareLess
	}

	// compare minor
	if v.Minor > target.Minor {
		return VersionCompareGreater
	} else if v.Minor < target.Minor {
		return VersionCompareLess
	}

	// same major and minor, compare build
	if v.HasBuild && target.HasBuild {
		if v.Build > target.Build {
			return VersionCompareGreater
		} else if v.Build < target.Build {
			return VersionCompareLess
		} else {
			return VersionCompareEqual
		}
	} else if v.HasBuild {
		return VersionCompareGreater
	} else if target.HasBuild {
		return VersionCompareLess
	} else {
		return VersionCompareEqual
	}
}
