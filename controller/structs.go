// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"errors"
	"time"

	commonapi "asn.amiasys.com/asn-service-api/v25/common"
)

// Structs used between asn.controller and service.controller.

// Network is the structure for a network.
type Network struct {
	ID          string
	Name        string
	ParentID    string
	Description string
	Tiers       []string
	Location    *commonapi.Location
	Networks    []*Network
}

// Node is the structure for a node.
type Node struct {
	ID           string             // Node ID
	Name         string             // Device display name
	Type         commonapi.NodeType // Node Type
	RegisteredAt time.Time
	State        commonapi.NodeState // Node state, refer Service Node State enum
	NetworkID    string              // Network ID
	NodeGroupID  string              // Node Group ID, if in a node group
	Description  string
	Metadata     string              // metadata used by the service
	Location     *commonapi.Location // Node physical location

	Managed bool
	Info    *commonapi.NodeInfo

	ServiceInfo *ServiceInfo
}

type ServiceInfo struct {
	State        commonapi.ServiceState
	UsedConfig   string // exists if Config Source is Node, otherwise is empty, can get the config from a node group
	ConfigSource commonapi.ServiceConfigSource
}

type NodeStateChange struct {
	Timestamp time.Time
	NodeID    string

	NodeState      commonapi.NodeState
	FrameworkError FrameworkErr

	ServiceState commonapi.ServiceState
	ServiceError error
}

type OpsResponse struct {
	Timestamp time.Time
	NodeID    string

	FrameworkError FrameworkErr

	ServiceResponse string
	ServiceError    error
}

type FrameworkErr error

var FrameworkErrServiceTimeout FrameworkErr = errors.New("service timed out")
var FrameworkErrNodeDisconnected FrameworkErr = errors.New("node disconnected")

type Link struct {
	ID          string // uuid
	Description string // the name of the link can be empty
	Bandwidth   int64  // the bandwidth between two nodes, the up speed equals to the down speed

	From, To *LinkNode
}

type LinkNode struct {
	NodeID    string
	Interface string
}

type NodeGroup struct {
	ID          string
	NetworkID   string
	Name        string
	Description string
	Metadata    string // metadata used by the service
	Nodes       []string
	Config      string
}
