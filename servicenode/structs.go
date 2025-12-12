// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

type NodeInfo struct {
	ID string
	commonapi.NodeInfo
}
