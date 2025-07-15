// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

const (
	OpScopeNetwork = 1 + iota
	OpScopeNetworkWithSubnetworks
	OpScopeGroup
	OpScopeNode
)

const (
	ServiceConfigSourceNode = 1 + iota
	ServiceConfigSourceNodeGroup
)

type Response struct {
	Response string
	Error    error
}
