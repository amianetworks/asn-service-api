// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
)

// Node is the structure for a node.
type Node struct {
	ID   string             // Node ID
	Name string             // Device display name
	Type commonapi.NodeType // Node Type

	TopoInfo *commonapi.Info

	Config []byte
}
