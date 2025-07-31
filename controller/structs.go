// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"time"

	commonapi "asn.amiasys.com/asn-service-api/v25/common"
)

// Structs used between asn.controller and service.controller.

type NetworkBasicInfo struct {
	ID          string
	Name        string
	ParentID    string
	Description string
	Tiers       []string
}

// Network is the structure for a network.
type Network struct {
	NetworkBasicInfo
	Location *commonapi.Location
	Networks []*Network
}

// Node is the structure for a node.
type Node struct {
	ID           string             // Node ID
	Name         string             // Device display name
	Type         commonapi.NodeType // Node Type
	RegisteredAt time.Time
	State        commonapi.ServiceNodeState // Node state, refer Service Node State enum
	NetworkID    string                     // Network ID
	NodeGroupID  string                     // Node Group ID, if in a node group
	Description  string
	Metadata     string              // metadata used by the service
	Location     *commonapi.Location // Node physical location

	Managed  bool
	TopoInfo *commonapi.Info

	ServiceInfo *ServiceInfo
}

type ServiceInfo struct {
	State        commonapi.ServiceState
	UsedConfig   []byte // exists if Config Source is Node, otherwise is empty, can get the config from a node group
	ConfigSource commonapi.ServiceConfigSource
}

type NodeStateChange struct {
	Timestamp    time.Time
	NodeID       string
	NodeState    commonapi.ServiceNodeState
	ServiceState commonapi.ServiceState
}

type NetworkLink struct {
	ID          string // uuid
	Description string // the name of the link can be empty
	Bandwidth   int64  // the bandwidth between two nodes, the up speed equals to the down speed

	From, To *NetworkLinkNode
}

type NetworkLinkNode struct {
	NetworkID string
	Interface string
}

type NodeLink struct {
	ID          string // uuid
	Description string // the name of the link can be empty
	Bandwidth   int64  // the bandwidth between two nodes, the up speed equals to the down speed

	From, To *NodeLinkNode
}

type NodeLinkNode struct {
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
	Config      []byte
}
