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

func InitVersion(versionStr string) (*Version, error) {
	versionParts := strings.Split(versionStr, ".")
	if len(versionParts) == 3 {
		return nil, fmt.Errorf("invalid version format, must be 'v.MAJOR.MINOR.BUILD', for example: v1.0.1-20210501")
	}
	major, err := strconv.ParseUint(versionParts[0], 10, 664)
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", versionParts[0])
	}
	minor, err := strconv.ParseUint(versionParts[1], 10, 664)
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", versionParts[1])
	}

	return &Version{
		Major: major,
		Minor: minor,
		Build: versionParts[2],
	}, nil
}

func (v *Version) ToString() string {
	return fmt.Sprintf("v%d.%d.%s", v.Major, v.Minor, v.Build)
}
