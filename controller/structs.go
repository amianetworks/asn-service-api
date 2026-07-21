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
	ID              string
	Name            string
	Type            commonapi.NodeType
	RegisteredAt    time.Time
	State           commonapi.NodeState       // runtime connectivity
	EnrollmentState commonapi.EnrollmentState // credential lifecycle (orthogonal to State)
	NetworkID       string
	NodeGroupID     string // empty if the node is not in a group
	Description     string
	// Metadata is an opaque string set by the service via UpdateNodeMetadata().
	Metadata string
	Location *commonapi.Location

	Managed bool
	Info    *commonapi.NodeInfo

	// Ownership axis (node-level, shared across every service on the node).
	// Set/cleared via CreateNode / SetNodeOwner; decoupled from the enrollment
	// credential. Invariant: OwnerTypeGlobal <=> OwnerID == "";
	// OwnerTypeAccount <=> OwnerID != "".
	OwnerType commonapi.OwnerType
	OwnerID   string

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
// It is a snapshot of the node's three orthogonal axes at the time of the event:
// connectivity (NodeState), service operational state (ServiceState), and
// credential lifecycle (EnrollmentState). Every field carries its true current
// value regardless of which axis triggered the event.
// FrameworkError is non-nil on framework-level failures (e.g. node disconnection).
// ServiceError is non-nil when the service itself reported an error during the transition.
type NodeStateChange struct {
	Timestamp time.Time
	NodeID    string

	NodeState      commonapi.NodeState
	FrameworkError error

	ServiceState commonapi.ServiceState
	ServiceError error

	// EnrollmentState is the node's credential-lifecycle axis at event time
	// (orthogonal to NodeState). All four values may appear. Enrollment
	// transitions do not each fire a dedicated event: losing identity
	// (-> EnrollmentStateUnbound) is signalled explicitly, and a node becoming
	// EnrollmentStateBound rides the connectivity event of its registration;
	// the intermediate provisioning states (TokenIssued, CertIssued) are not
	// separately signalled but still appear here in other events' snapshots.
	// Token/cert expiry-driven transitions are not delivered in real time.
	EnrollmentState commonapi.EnrollmentState
}

// OpsResponse is one node's response to a SendServiceOps or config op dispatch call.
// When FrameworkError != nil, ServiceResponse and ServiceError are undefined.
type OpsResponse struct {
	Timestamp time.Time
	NodeID    string

	// FrameworkError is set when the framework could not reach or invoke the service on this node.
	FrameworkError error

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
	Source       commonapi.ServiceSource
}

// LicenseInfo describes the license currently used by this service.
type LicenseInfo struct {
	LicenseKey  string
	MachineID   string
	LicenseType string
	Status      commonapi.LicenseStatus

	ValidStartTime time.Time
	ValidEndTime   time.Time

	// Content contains service-defined license payload fields.
	Contents map[string]string
}

// Node ownership (OwnerType + OwnerID) is a node-level attribute decoupled from
// the enrollment credential: the caller asserts the owner from its own
// authenticated context and the framework stores it verbatim, enforcing only the
// OwnerType/OwnerID invariant. The framework does not adjudicate owner identity,
// entitlement, or tenant isolation — those stay with the service. Ownership is
// set via CreateNode or ASNController.SetNodeOwner, read on the Node struct, and
// filtered via ASNController.GetNodesOfNetwork. See workflow/design/PrivateNode.md.

// SetNodeOwnerRequest sets, transfers, or clears a node's ownership.
type SetNodeOwnerRequest struct {
	NodeID    string
	OwnerType commonapi.OwnerType // empty => OwnerTypeGlobal
	OwnerID   string              // required for OwnerTypeAccount; must be empty otherwise
}

// NodeOwnerFilter optionally restricts GetNodesOfNetwork by ownership. The two
// axes are independent; an empty slice means that axis is not filtered, and when
// both are set a node must match both (AND).
type NodeOwnerFilter struct {
	OwnerTypes []commonapi.OwnerType // empty => no owner-type filter
	OwnerIDs   []string              // empty => no owner-id filter
}
