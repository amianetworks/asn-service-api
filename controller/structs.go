// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"time"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

// AvailableService is a service loaded on the controller, as listed by
// ASNController.ListAvailableServices.
type AvailableService struct {
	Name    string
	Version commonapi.Version // the service controller's StaticResource.Version()
}

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

	// AutoUpgrade is the node's auto-upgrade switch; see
	// ASNController.SetNodeAutoUpgrade. It belongs to the node, not to the
	// calling service.
	AutoUpgrade bool
	// AutoUpgradeHeld is true when the last automatic upgrade for the versions
	// the controller serves now did not bring the node to them. The node is
	// not tried again until those versions change or the switch is turned off
	// and on again.
	AutoUpgradeHeld bool

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

	// PendingTask is the calling service's queued change on the node that is
	// not sent yet (an install of its deb), with PendingReason; nil when none.
	PendingTask *NodeTask
	// RunningTask is the node task in flight on the node when it restarts
	// asnsn; nil otherwise.
	RunningTask *NodeTask
	// LastRequest is the calling service's own last request on the node that
	// ended, with or without a task; nil when none since the service joined
	// the node. Only the calling service's requests overwrite it: a request is
	// over when LastRequest.RequestID equals its id.
	LastRequest *NodeTask
	// LastTask is the last node task on the node that ended and restarted
	// asnsn, of any actor (the calling service's own included); nil when none
	// since the service joined the node.
	LastTask *NodeTask
}

// NodeTaskKind is what a node task did to the calling service on the node.
type NodeTaskKind string

const (
	NodeTaskKindInstall NodeTaskKind = "install" // installed the calling service's deb
	NodeTaskKindUpgrade NodeTaskKind = "upgrade" // upgraded asnsn and the calling service's deb
	NodeTaskKindRestart NodeTaskKind = "restart" // another change that restarts asnsn and every service on the node
)

// NodeTaskStatus is how a node task, or a request that ended without one,
// ended for the calling service. Failure causes are not told apart; an ASN
// operator sees the cause.
type NodeTaskStatus int

const (
	NodeTaskStatusUnspecified NodeTaskStatus = iota // not ended
	NodeTaskStatusOK                                // the task succeeded and asnsn came back
	NodeTaskStatusFailed
)

// NodeTask is the calling service's view of one node task, or of one of its
// requests that has not been sent yet. Another actor's task carries no
// RequestID and names no actor, service, or package.
type NodeTask struct {
	// RequestID is the calling service's request (CreateNodeResult.RequestID,
	// the requestID UpgradeNode returns); empty when another actor caused the
	// task.
	RequestID string
	// TaskID is the node task; empty while the request is queued, or when it
	// ended without a task.
	TaskID string
	Kind   NodeTaskKind
	// PendingReason is set in ServiceInfo.PendingTask only.
	PendingReason PendingReason
	// Status is set in ServiceInfo.LastRequest and LastTask only.
	Status NodeTaskStatus
	// RequestedAt is when the calling service's request was made, or, for
	// another actor's task, when that task was sent.
	RequestedAt time.Time
	// EndedAt is set in ServiceInfo.LastRequest and LastTask only.
	EndedAt time.Time
}

// NodeTaskSlot names the ServiceInfo slot a ChangeKindNodeTask reports.
type NodeTaskSlot int

const (
	NodeTaskSlotUnspecified NodeTaskSlot = iota
	NodeTaskSlotPending                  // ServiceInfo.PendingTask
	NodeTaskSlotRunning                  // ServiceInfo.RunningTask
	NodeTaskSlotLastRequest              // ServiceInfo.LastRequest
	NodeTaskSlotLastTask                 // ServiceInfo.LastTask
)

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
	// ChangeKindNodeTask: one of this service's task slots on the node
	// (TaskSlot) was filled or cleared; Task is the slot's new value, nil when
	// cleared. A task that ends sends one change per slot, in the order
	// RunningTask cleared, LastTask filled, LastRequest filled.
	ChangeKindNodeTask
	// ChangeKindNodeAutoUpgrade: the node's auto-upgrade switch changed
	// (AutoUpgrade), by an ASN operator or any service on the node; sent to
	// every service on the node. It names no actor.
	ChangeKindNodeAutoUpgrade
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

	// ChangeKindNodeTask
	TaskSlot NodeTaskSlot
	Task     *NodeTask

	// ChangeKindNodeAutoUpgrade
	AutoUpgrade bool
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
	// RequestID identifies the install in ServiceInfo.PendingTask,
	// RunningTask and LastRequest, from queue to result; set whenever there is
	// a task, sent or queued. A call that joins the calling service's queued
	// install returns that install's id.
	RequestID string
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
	// redeems. It is issue time plus the requested lifetime (TokenTTLSeconds,
	// or GetInstallToken's tokenTTLSeconds), capped by the controller's
	// configured token lifetime (servicenode.enrollment.token_ttl), which is
	// also used when the requested lifetime is zero.
	ExpiresAt int64
}

// Script is a rendered host script, returned by RedeemScriptToken. The service
// serves Content to the host as-is.
type Script struct {
	Content     []byte
	ContentType string // e.g. "text/x-shellscript"
	NodeID      string
	Kind        ScriptKind
}
