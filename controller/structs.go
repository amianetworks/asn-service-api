// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"time"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

// Network represents a network in the topology tree.
// Networks may be nested: each Network embeds a slice of child Networks,
// linked to their parent via ParentID.
type Network struct {
	ID          string
	Name        string
	ParentID    string
	Description string
	// Tiers is the subset of the location hierarchy that applies to this network.
	// Values are drawn from commonapi.LocationTiers.
	Tiers    []string
	Location *commonapi.Location
	Networks []*Network
}

// Node represents a service node within a network.
type Node struct {
	ID           string
	Name         string
	Type         commonapi.NodeType
	RegisteredAt time.Time
	State        commonapi.NodeState
	NetworkID    string
	NodeGroupID  string // empty if the node is not in a group
	Description  string
	// Metadata is an opaque string set by the service via UpdateNodeMetadata().
	Metadata string
	Location *commonapi.Location

	Managed bool
	Info    *commonapi.NodeInfo

	// ServiceInfo is nil if the service is not loaded on this node.
	ServiceInfo *ServiceInfo
}

// ServiceInfo describes the service's state on a specific node.
type ServiceInfo struct {
	Version commonapi.Version
	State   commonapi.ServiceState
	// UsedConfig is the config currently active on this node.
	// Non-empty only when ConfigSource is ServiceConfigSourceNode;
	// when inherited from a node group, retrieve the config via the group.
	UsedConfig   string
	ConfigSource commonapi.ServiceSource
	ConfigOps    []ConfigOp
}

// NodeStateChange is delivered on the channel returned by SubscribeNodeStateChanges().
// It carries both the node's connectivity state and the service's operational state
// at the time of the event.
// FrameworkError is non-nil on framework-level failures (e.g. node disconnection).
// ServiceError is non-nil when the service itself reported an error during the transition.
type NodeStateChange struct {
	Timestamp time.Time
	NodeID    string

	NodeState      commonapi.NodeState
	FrameworkError commonapi.FrameworkErr

	ServiceState commonapi.ServiceState
	ServiceError error
}

// OpsResponse is one node's response to a SendServiceOps or config op dispatch call.
// When FrameworkError != nil, ServiceResponse and ServiceError are undefined.
type OpsResponse struct {
	Timestamp time.Time
	NodeID    string

	// FrameworkError is set when the framework could not reach or invoke the service on this node.
	FrameworkError commonapi.FrameworkErr

	// ServiceResponse is the resp string returned by ASNService.ApplyServiceOps() or a config op method.
	ServiceResponse string
	// ServiceError is the error returned by the same method.
	ServiceError error
}

// Link represents a connection between two nodes.
// Bandwidth is symmetric (upload == download), expressed in bits per second.
type Link struct {
	ID          string
	Description string
	Bandwidth   int64

	From, To *LinkNode
}

// LinkNode identifies one endpoint of a Link.
type LinkNode struct {
	NodeID    string
	Interface string
}

// NodeGroup is a service-scoped collection of nodes within a network.
// Config and ConfigOps set on the group are inherited by member nodes
// unless the node has its own direct overrides.
type NodeGroup struct {
	ID          string
	NetworkID   string
	Name        string
	Description string
	// Metadata is an opaque string set by the service via UpdateNodeGroupMetadata().
	Metadata  string
	Nodes     []string
	Config    string
	ConfigOps []ConfigOp
}

// ConfigOp is a single persistent configuration directive attached to a node or node group.
// ID is framework-assigned; use it in UpdateConfigOp() and DeleteConfigOps() calls.
// ConfigParams is an opaque service-defined string.
type ConfigOp struct {
	ID           string
	ConfigParams string
	source       commonapi.ServiceSource
}
