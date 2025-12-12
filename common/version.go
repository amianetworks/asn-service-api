// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	VersionCompareGreater = 1 + iota
	VersionCompareLess
	VersionCompareEqual

	versionFormat = `^v\d+\.\d+(?:\.\d+)?(?:-\d+)?$`
)

type Version struct {
	Major uint64
	Minor uint64

	Build *Build
}

type Build struct {
	Number uint64
	Suffix uint64
}

// InitVersion converts a version string, e.g., "v25.0.1", into Version structure.
func InitVersion(versionStr string) (Version, error) {
	matched, err := regexp.MatchString(versionFormat, versionStr)
	if err != nil {
		return Version{}, err
	}
	if !matched {
		return Version{}, fmt.Errorf("invalid version format, must be 'v.MAJOR.MINOR' or 'v.MAJOR.MINOR.BUILD', for example: v1.0, v1.0.1 or v1.0.1-196912")
	}

	versionParts := strings.Split(strings.TrimPrefix(versionStr, "v"), ".")

	major, _ := strconv.ParseUint(versionParts[0], 10, 64)
	minor, _ := strconv.ParseUint(versionParts[1], 10, 64)

	var build *Build
	if len(versionParts) == 3 {
		var buildNum, buildSuffix uint64

		if strings.Contains(versionParts[2], "-") {
			buildParts := strings.Split(versionParts[2], "-")
			buildNum, _ = strconv.ParseUint(buildParts[0], 10, 64)
			buildSuffix, _ = strconv.ParseUint(buildParts[1], 10, 64)
		} else {
			buildNum, _ = strconv.ParseUint(versionParts[2], 10, 64)
		}

		build = &Build{
			Number: buildNum,
			Suffix: buildSuffix,
		}
	}

	version := Version{
		Major: major,
		Minor: minor,
		Build: build,
	}

	return version, nil
}

// ToString converts a Version structure back to the version string.
func (v Version) ToString() string {
	if v.Build == nil {
		return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
	}

	if v.Build.Suffix == 0 {
		return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Build.Number)
	}

	return fmt.Sprintf("v%d.%d.%d-%d", v.Major, v.Minor, v.Build.Number, v.Build.Suffix)
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
	if v.Build != nil && target.Build != nil {
		if v.Build.Number > target.Build.Number {
			return VersionCompareGreater
		} else if v.Build.Number < target.Build.Number {
			return VersionCompareLess
		} else {
			if v.Build.Suffix > target.Build.Suffix {
				return VersionCompareGreater
			} else if v.Build.Suffix < target.Build.Suffix {
				return VersionCompareLess
			} else {
				return VersionCompareEqual
			}
		}
	} else if v.Build != nil {
		return VersionCompareGreater
	} else if target.Build != nil {
		return VersionCompareLess
	} else {
		return VersionCompareEqual
	}
}
