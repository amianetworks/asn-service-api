// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	"errors"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

var (
	ErrServiceNotFound = errors.New("target service not found")
	ErrKeyNotFound     = errors.New("key not found in target service")
)

type NodeInfo struct {
	ID string
	commonapi.NodeInfo
}
