// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major uint64
	Minor uint64
	Build string
}

// InitVersion converts a version string, e.g., "v25.0.1", into Version structure.
func InitVersion(versionStr string) (Version, error) {
	versionParts := strings.Split(versionStr, ".")
	if len(versionParts) != 3 {
		return Version{}, fmt.Errorf("invalid version format, must be 'v.MAJOR.MINOR.BUILD', for example: v1.0.1-196912")
	}
	major, err := strconv.ParseUint(strings.Replace(versionParts[0], "v", "", -1), 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version [%s] due to: %v", versionParts[0], err)
	}
	minor, err := strconv.ParseUint(versionParts[1], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version [%s]  due to: %v", versionParts[1], err)
	}

	return Version{
		Major: major,
		Minor: minor,
		Build: versionParts[2],
	}, nil
}

// ToString converts a Version structure back to the version string.
func (v Version) ToString() string {
	return fmt.Sprintf("v%d.%d.%s", v.Major, v.Minor, v.Build)
}

// GreaterThanAndEqualTo returns if the current version is greater than or equal to the given version.
func (v Version) GreaterThanAndEqualTo(target Version) bool {
	if v.Major > target.Major {
		return true
	} else if v.Major == target.Major && v.Minor >= target.Minor {
		return true
	} else {
		return false
	}
}
