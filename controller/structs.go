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

	// DefaultRoot is true for the root network the controller configuration
	// names (network.default_root_network), however the root was created. A
	// root created with that name while no root is marked becomes the default
	// root at once; a root renamed to it, at the controller's next start. A
	// controller whose configuration names no existing root has none. A
	// service can neither delete nor rename the default root.
	DefaultRoot bool
}

// NetworkFilter restricts ListNetworks. Conditions are ANDed; zero values do
// not filter.
type NetworkFilter struct {
	IDs      []string // these networks only; unknown IDs are ignored
	RootOnly bool     // root networks only
	ParentID string   // direct children of this network only
}

// Node represents a service node within a network.
type Node struct {
	ID           string
	Name         string
	Type         commonapi.NodeType
	RegisteredAt time.Time
	State        commonapi.NodeState // runtime connectivity
	NetworkID    string
	NodeGroupID  string // empty if the node is not in a group
	Description  string
	// Metadata is an opaque string set by the service via UpdateNodeMetadata().
	Metadata string
	Location *commonapi.Location

	Managed bool
	Info    *commonapi.NodeInfo

	// ServiceInfo is nil if this service is not in the node's service_names (not
	// eligible). An eligible node whose plugin is not loaded yet has ServiceInfo
	// with State ServiceStateUnavailable.
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

// ChangeKind is what a NodeStateChange reports. A consumer branches on it and
// ignores kinds it does not know; the fields a kind does not list are zero.
type ChangeKind int

const (
	// ChangeKindState: the node's connectivity or this service's state on it
	// (NodeState, FrameworkError, ServiceState, ServiceError). The only kind in
	// the initial snapshot.
	ChangeKindState ChangeKind = iota
	// ChangeKindNodeDeleted: the node was deleted, by an operator or a service.
	// The host is not cleaned up by the deletion.
	ChangeKindNodeDeleted
	// ChangeKindNodeServiceRemoved: this service was removed from the node's
	// record, by an operator or by this service itself. Sent when the record
	// changes, before the host's package is purged and while the node is
	// offline too.
	ChangeKindNodeServiceRemoved
	// ChangeKindNodeMoved: the node moved from FromNetworkID to NetworkID.
	ChangeKindNodeMoved
	// ChangeKindRootNetworkDeleted: a root network (NetworkID) was deleted,
	// with every node in it; sent to every service. NodeID is empty.
	ChangeKindRootNetworkDeleted
)

// NodeStateChange is delivered on the channel returned by SubscribeNodeStateChanges().
//
// For ChangeKindState it is a snapshot of the node's two orthogonal axes at
// the time of the event: connectivity (NodeState) and service operational
// state (ServiceState). Every field carries its true current value regardless
// of which axis triggered the event. FrameworkError is non-nil on
// framework-level failures (e.g. node disconnection). ServiceError is non-nil
// when the service itself reported an error during the transition.
//
// The other kinds report a change of the record; see ChangeKind.
type NodeStateChange struct {
	Timestamp time.Time
	Kind      ChangeKind
	NodeID    string // empty for ChangeKindRootNetworkDeleted

	// ChangeKindState
	NodeState      commonapi.NodeState
	FrameworkError error

	ServiceState commonapi.ServiceState
	ServiceError error

	// ChangeKindNodeMoved: FromNetworkID -> NetworkID.
	// ChangeKindRootNetworkDeleted: NetworkID.
	NetworkID     string
	FromNetworkID string
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

// CreateNodeRequest creates a node for the calling service or, with
// AllowExisting, adds the calling service to a node that already exists with
// NodeName in the parent's root network tree.
type CreateNodeRequest struct {
	// ParentNetworkID is the network the node is placed under; the node's
	// network path derives from it. Required.
	ParentNetworkID string
	// NodeName identifies the node; it is unique across the parent's whole root
	// network tree and is the key AllowExisting matches on. Required.
	NodeName string
	// Type is the node's hardware/logical role, e.g. NodeTypeServer or
	// NodeTypeAppliance.
	Type  commonapi.NodeType
	Label string
	// AllowExisting turns a NodeName collision from an error into "add the
	// calling service to that node"; see ASNController.CreateNode.
	AllowExisting bool
	// UpdateInfo applies only when AllowExisting matched an existing node: true
	// overwrites that node's Type and Label with the request's, false leaves them
	// unchanged. Type and Label are node-level and shared by every service on the
	// node, so an overwrite affects the other services too. A newly created node
	// always takes the request's Type and Label, regardless of UpdateInfo.
	UpdateInfo bool
	// TokenTTLSeconds is the lifetime of the returned install token; see
	// ScriptToken.ExpiresAt.
	TokenTTLSeconds int64
}

// PendingReason is why a node task that changes the host is queued instead
// of sent at once.
type PendingReason string

const (
	PendingReasonOffline     PendingReason = "offline"     // the node is not connected
	PendingReasonMaintenance PendingReason = "maintenance" // the node is in Maintenance; an install waits until it is admitted normally
	PendingReasonBusy        PendingReason = "busy"        // another task runs on the node, or the controller's task limit is reached
)

// CreateNodeResult is returned by CreateNode. It carries exactly one of a
// task (TaskID, or TaskPending with PendingReason) and a Token, or none of
// them when the calling service's deb is already installed at the version
// the controller serves.
type CreateNodeResult struct {
	NodeID string
	// Created is true when a new node was created, false when AllowExisting
	// matched an existing node.
	Created bool
	// Token is the node's install token (ScriptKindInstall); nil when the node
	// can run node tasks.
	Token *ScriptToken
	// TaskID is the node task that installs the calling service's deb, when it
	// was sent at once.
	TaskID string
	// TaskPending is true when that task is queued; PendingReason says why.
	TaskPending   bool
	PendingReason PendingReason
}

// GetInstallTokenRequest asks for a fresh install token for an existing node.
type GetInstallTokenRequest struct {
	// NodeID is the existing node; the calling service must be on it. Required.
	NodeID string
	// TokenTTLSeconds is the lifetime of the returned token; see
	// ScriptToken.ExpiresAt.
	TokenTTLSeconds int64
}

// DeleteNodeRequest removes the calling service from a node; see
// ASNController.DeleteNode.
type DeleteNodeRequest struct {
	// NodeID is the node to remove the calling service from. Required.
	NodeID string
	// Reason is recorded in the audit log (e.g. "decommissioned").
	Reason string
	// DeleteEmptyNode applies only when the calling service is the node's last:
	// true destroys the node record, false keeps it serviceless. Ignored when
	// other services remain, since the node is then always kept.
	DeleteEmptyNode bool
	// TokenTTLSeconds is the lifetime of the returned uninstall token; see
	// ScriptToken.ExpiresAt.
	TokenTTLSeconds int64
}

// DeleteNodeResult is returned by DeleteNode.
type DeleteNodeResult struct {
	NodeID string
	// LastService is true when the calling service was the node's last. The
	// node's credential was then revoked, and the uninstall script also purges
	// asnsn and the node's credential.
	LastService bool
	// NodeDeleted is true when the node record was destroyed (LastService and
	// DeleteEmptyNode were both true).
	NodeDeleted bool
	// Token is the uninstall token (ScriptKindUninstall); nil when the node
	// can run node tasks and the calling service was not its last.
	Token *ScriptToken
	// TaskID is the node task that purges the calling service's deb, when it
	// was sent at once.
	TaskID string
	// TaskPending is true when that task is queued; PendingReason says why.
	TaskPending   bool
	PendingReason PendingReason
}

// UpgradeNodeRequest upgrades a node the calling service is on; see
// ASNController.UpgradeNode.
type UpgradeNodeRequest struct {
	// NodeID is the node. Required.
	NodeID string
}

// ScriptKind is what a ScriptToken, and the Script it redeems to, does on the
// host.
type ScriptKind int

const (
	ScriptKindInstall   ScriptKind = iota + 1 // install or upgrade asnsn and the calling service's deb
	ScriptKindUninstall                       // remove the calling service's deb (and asnsn when it was the last service)
)

// ScriptToken is an opaque, bearer credential that redeems to one script via
// ASNController.RedeemScriptToken. Every token has the same shape and is
// redeemed the same way, whichever call issued it.
//
// A token is bound to one node and to the service that issued it, and only that
// service can redeem it. A node may have any number of valid tokens at once;
// issuing a new one never invalidates another. A token is never consumed by
// redemption and stays valid until ExpiresAt. An install token also stops
// redeeming once the issuing service is no longer on the node.
type ScriptToken struct {
	// Token is the opaque value handed to the host and presented back for
	// redemption. Treat it as a secret: an install token lets its holder obtain
	// the node's credential until it expires, and it cannot be revoked early, so
	// keep TokenTTLSeconds short. An uninstall token carries no secret.
	Token  string
	NodeID string
	Kind   ScriptKind
	// ExpiresAt is the Unix time (seconds) after which the token no longer
	// redeems. It is issue time plus the request's TokenTTLSeconds, capped by the
	// controller's configured token lifetime (servicenode.enrollment.token_ttl),
	// which is also used when TokenTTLSeconds is zero.
	ExpiresAt int64
}

// RedeemScriptTokenRequest redeems a ScriptToken for its script.
type RedeemScriptTokenRequest struct {
	// Token is ScriptToken.Token as presented by the host. Required.
	Token string
}

// Script is a rendered host script, returned by RedeemScriptToken. The service
// serves Content to the host as-is.
type Script struct {
	Content     []byte
	ContentType string // e.g. "text/x-shellscript"
	NodeID      string
	Kind        ScriptKind
}
