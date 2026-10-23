// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/redis/go-redis/v9"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
	"asn.amiasys.com/asn-service-api/v26/iam"
	"asn.amiasys.com/asn-service-api/v26/log"
)

// ASNController is the framework-provided handle passed to ASNServiceController.Init().
// All methods are goroutine-safe after Init() unless stated otherwise.
//
// Functional areas:
//  1. Resource initialization (Init* / Get* — one-shot, call in Init())
//  2. Service lifecycle management
//  3. Ops dispatch
//  4. Config ops dispatch
//  5. Node topology
//  6. Root networks and their admin group
//  7. Node group management
//  8. Node enrollment and node tasks
type ASNController interface {

	// -------------------------------------------------------------------------
	// Resource Initialization
	// Must be called in Init(). All are one-shot; a second call returns an error.
	// -------------------------------------------------------------------------

	// InitLogger returns the logger for this service.
	// Call once in Init(). Default log files are named <servicename>-*.log.
	InitLogger() (*log.Logger, error)

	// InitDocDB returns a connected document database handle.
	// name scopes the DB instance; multiple names yield independent handles.
	// The framework prefixes it with the service's configured docdb db_name:
	// an empty name opens <db_name> itself, any other opens <db_name>_<name>.
	// Call once per name in Init().
	InitDocDB(name string) (commonapi.DocDBHandler, error)

	// InitTSDB returns a connected time-series database handle.
	// name scopes the DB instance; multiple names yield independent handles.
	// The framework prefixes it with the service's configured tsdb db_name:
	// an empty name opens <db_name> itself, any other opens <db_name>_<name>.
	// Call once per name in Init().
	InitTSDB(name string) (commonapi.TSDBHandler, error)

	// InitRedis returns a connected redis database handle.
	// Only one redis database handle is provided to each service.
	// Call once in Init().
	InitRedis() (*redis.Client, error)

	// InitLocker returns a cluster-wide distributed lock.
	// Call once in Init().
	InitLocker() (Lock, error)

	// GetIAM returns the IAM instance for account, group, and access management.
	// Call once in Init().
	GetIAM() (iam.Instance, error)

	// -------------------------------------------------------------------------
	// License Management
	// License related APIs.
	// -------------------------------------------------------------------------

	// IsLicenseValid determines whether the current license is valid for this service.
	IsLicenseValid() error

	// UseLicense binds or updates the license used by this service.
	UseLicense(licenseKey string) error

	// GetCurrentLicenseInfo returns the current machine's license snapshot.
	GetCurrentLicenseInfo() *LicenseInfo

	// -------------------------------------------------------------------------
	// Service Lifecycle Management
	// -------------------------------------------------------------------------

	// StartService triggers Start(config) on the service running on each matched node.
	// serviceScope and serviceScopeList determine the target set; see ServiceScope constants.
	StartService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// StopService triggers Stop() on the service running on each matched node.
	StopService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// ResetService triggers Stop() followed by Start() on each matched node.
	ResetService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// -------------------------------------------------------------------------
	// Ops Dispatch
	// -------------------------------------------------------------------------

	// SendServiceOps dispatches an op command to all nodes matched by serviceScope / serviceScopeList.
	// Fan-out, asynchronous.
	//
	// If paramErr != nil, the scope or scopeList is invalid; resChan is nil.
	// Otherwise, returns immediately; responses stream into resChan as nodes reply.
	// resChan is closed after all nodes have responded or timed out.
	// Check OpsResponse.FrameworkError before using ServiceResponse / ServiceError.
	//
	//	resChan, paramErr := ctrl.SendServiceOps(scope, list, cmd, params)
	//	if paramErr != nil { ... }
	//	for res := range resChan {
	//	    if res.FrameworkError != nil { ... }
	//	}
	SendServiceOps(
		serviceScope commonapi.ServiceScope, serviceScopeList []string,
		opCmd, opParams string,
	) (resChan <-chan *OpsResponse, paramErr error)

	// SendServiceOpsToNode dispatches an operation to a single node and blocks until it responds or times out.
	// If paramErr != nil, nodeID is invalid; res is nil.
	// Check res.FrameworkError before using res.ServiceResponse / res.ServiceError.
	SendServiceOpsToNode(nodeID string, opCmd, opParams string) (res *OpsResponse, paramErr error)

	// -------------------------------------------------------------------------
	// Config Ops Dispatch
	// Scope is limited to ServiceScopeNodeGroup (3) or ServiceScopeNode (4).
	// -------------------------------------------------------------------------

	// AddConfigOps persists new config ops for the given scope, then fans out to all affected nodes.
	// If paramErr != nil, scope or scopeID is invalid; resChan is nil.
	// Otherwise, returns immediately; each OpsResponse reflects the result of ASNService.AddConfigOps on that node.
	// resChan is closed after all nodes have responded or timed out.
	AddConfigOps(serviceScope commonapi.ServiceScope, scopeID string, configParams []string) (resChan <-chan *OpsResponse, paramErr error)

	// UpdateConfigOp updates a single config op identified by configOpID, persists the change,
	// and fans out to affected nodes.
	// If paramErr != nil, scope or scopeID is invalid; resChan is nil.
	UpdateConfigOp(serviceScope commonapi.ServiceScope, scopeID, configOpID, configParam string) (resChan <-chan *OpsResponse, paramErr error)

	// DeleteConfigOps removes config ops by ID for the given scope, persists, and fans out to affected nodes.
	// If paramErr != nil, scope or scopeID is invalid; resChan is nil.
	DeleteConfigOps(serviceScope commonapi.ServiceScope, scopeID string, configOpIDs []string) (resChan <-chan *OpsResponse, paramErr error)

	// ListConfigOps returns config ops directly attached to the given scope.
	// Does not traverse the group-to-node inheritance hierarchy.
	// Synchronous; does not fan out to nodes.
	ListConfigOps(serviceScope commonapi.ServiceScope, scopeID string) ([]ConfigOp, error)

	// -------------------------------------------------------------------------
	// Node Topology
	// -------------------------------------------------------------------------

	// GetNetworks returns the full network tree. Each Network embeds nested Networks (subnetworks).
	GetNetworks() ([]*Network, error)

	// ListNetworks returns the networks matching filter, a page at a time, and
	// the number of matches. The returned networks carry no nested Networks;
	// use ParentID and filter.ParentID to walk the tree.
	ListNetworks(filter NetworkFilter, page iam.Page) (networks []*Network, total int, err error)

	// GetNodeByID returns full node details: hardware info, service-defined Metadata,
	// and ServiceInfo (service state, config source, active config ops).
	GetNodeByID(nodeID string) (*Node, error)

	// UpdateNodeMetadata persists an opaque service-defined string on the node.
	// Retrievable via GetNodeByID().Metadata.
	UpdateNodeMetadata(nodeID, metadata string) error

	// SetConfigOfNode persists the service config (YAML, UTF-8) for the node.
	// Used on the next StartService() call targeting this node.
	SetConfigOfNode(nodeID, config string) error

	// GetNodesOfNetwork returns all nodes of a network and its links.
	// If withService is true, only nodes eligible for this service (this service
	// in their service_names) are returned, whether or not it is loaded yet.
	// links are all links of the network regardless of withService, so with
	// withService true an endpoint may be a node that is not in nodes.
	// Internal links: both endpoints within the network; with withService false the To node
	// is included in the returned nodes slice.
	// External links: the To endpoint is outside the network and is not included in nodes.
	GetNodesOfNetwork(networkID string, withService bool) (nodes []*Node, links []*Link, err error)

	// MoveNodeToNetwork moves a node into targetNetworkID, under any root
	// network. Allowed when the calling service is on the node. The node keeps
	// its services, session, and tasks; its network scope follows the new
	// parent at once.
	//
	// Errors: asn.node.not_found; asn.network.not_found
	// (the target); AuthPermissionDenied (the calling service is not on the
	// node); ResourceAlreadyExists (the node's name is taken in the target
	// root network tree; rename and retry); RequestNotSupported (the node has
	// links, which never cross root networks).
	MoveNodeToNetwork(nodeID, targetNetworkID string) error

	// SubscribeNodeStateChanges returns a channel for node changes.
	// One-shot: a second call returns an error.
	// On subscription, the channel first delivers a ChangeKindState change for
	// every node of this service (initial snapshot), then delivers incremental
	// changes of every kind. The channel is never closed during normal
	// framework operation.
	//
	// A consumer must branch on NodeStateChange.Kind and ignore kinds it does
	// not know: the fields a kind does not list are zero.
	//
	// Delivery is best effort: a change that finds the channel full is dropped
	// (and counted by the framework). Reconcile against GetNodeByID,
	// GetNodesOfNetwork, and ListNetworks at start and periodically. The task
	// slots in ServiceInfo and Node.AutoUpgrade are not in the initial
	// snapshot; read them with GetNodesOfNetwork or GetNodeByID when
	// subscribing.
	//
	// Changes describe the record, not the host: a node deleted, or a service
	// removed from it, may still have its packages on the host. A node task
	// that restarts asnsn changes the calling service's task slots
	// (ChangeKindNodeTask); a change of the auto-upgrade switch is
	// ChangeKindNodeAutoUpgrade. Credential changes inside the framework (token
	// issue, certificate binding or revocation) are not service-visible and
	// produce no change.
	SubscribeNodeStateChanges() (<-chan *NodeStateChange, error)

	// -------------------------------------------------------------------------
	// Root Networks And Their Admin Group
	// A service may create root networks (for example one per tenant). Each
	// root network has exactly one admin group: the controller creates it with
	// the root and deletes it with the root, and it grants view and manage on
	// the root and nothing else. Manage on a root itself comes only from its
	// admin group, so the group's members are the root's complete list of
	// administrators. Any service may change the members of any root's admin
	// group; the group is shared by every service and by ASN operators, so a
	// service adds and removes members one change at a time and never
	// replaces the list with one it keeps. ASN does not know what the service
	// uses a root for. Every method below acts on any root network, the
	// default root included, except that a service can neither delete nor
	// rename the default root (AuthPermissionDenied); an unknown network is
	// refused with asn.network.not_found and a network that is not a root
	// with RequestParamInvalid. The service decides beforehand whether its
	// own user may act. Audited as service:<name>.
	//
	// The default root (Network.DefaultRoot) is the root whose name the
	// controller configuration names. A root created with that name while no
	// root is marked becomes the default root at once (a root renamed to it,
	// at the controller's next start), and from then on no service can delete
	// or rename it, its creator included. Check Network.DefaultRoot
	// (ListNetworks) after creating a root whose name could be the configured
	// one, and again after a controller restart or after a root is deleted;
	// hand a root you cannot use as a tenant to an ASN operator.
	// -------------------------------------------------------------------------

	// CreateRootNetwork creates a root network. Root network names are unique
	// among all root networks (exact match); normalizing names and resolving
	// collisions is the service's job. A root given the configured default-root
	// name may become the default root (see above).
	//
	// The service keeps its own mapping from its tenant to the root network
	// ID. To survive a crash between this call and recording the ID, record
	// the intended name first, use a name that carries the service's own
	// tenant identifier (for example "<service>-<tenantID>"), and on
	// ResourceAlreadyExists during a retry adopt the root of that name
	// (ListNetworks with RootOnly).
	//
	// The root's admin group is created with it, empty; if the group cannot
	// be created, the root is not created either.
	//
	// Errors: ResourceAlreadyExists.
	CreateRootNetwork(name, description string) (networkID string, err error)

	// RenameRootNetwork renames a root network other than the default root.
	// Same uniqueness rule as CreateRootNetwork.
	//
	// Errors: asn.network.not_found; RequestParamInvalid (not a
	// root); AuthPermissionDenied (the default root); ResourceAlreadyExists.
	RenameRootNetwork(networkID, newName string) error

	// DeleteRootNetwork deletes a root network with the whole subtree: every
	// network and every node in it, of every service, their links and node
	// groups, and the root's admin group; other network-type groups lose the
	// deleted networks, and one left with none is deleted. Each node is told
	// to unregister; every service on a deleted node receives
	// ChangeKindNodeDeleted, and every service receives
	// ChangeKindRootNetworkDeleted. Confirm with the user
	// before calling it, listing the nodes of other services
	// (GetNodesOfNetwork with withService false). The default root
	// (Network.DefaultRoot) cannot be deleted by a service.
	//
	// Hosts are not cleaned up: asnsn, the service debs, and the node
	// credential stay on each host, as for every node deletion, and the
	// host is cleaned up by hand.
	//
	// Errors: asn.network.not_found; RequestParamInvalid (not a
	// root); AuthPermissionDenied (the default root).
	DeleteRootNetwork(networkID string) error

	// RootAdminGroupAddMembers adds accounts to the root network's admin
	// group. Only accounts already in the calling service's IAM namespace are
	// accepted (created by the service, signed up through it, or joined by
	// being a member of any network-type group). Each account also joins
	// every service configured on the controller that has an IAM namespace,
	// and stays in them after it leaves the group. Members administer the
	// networks and nodes under the root in ASN as well, including IPMI and
	// node tasks that restart every service on a node; the root itself is
	// changed only by an ASN operator.
	//
	// Errors: asn.network.not_found; RequestParamInvalid (not a root, or a
	// malformed account ID); AuthPermissionDenied (accounts outside the
	// calling service's namespace, named in the error); GroupNotFound (the
	// controller lost the root's admin group, which is a controller defect).
	RootAdminGroupAddMembers(networkID string, accountIDs []string) error

	// RootAdminGroupRemoveMembers removes accounts from the root network's
	// admin group, whoever added them. The accounts stay in the services they
	// joined.
	//
	// Errors: as RootAdminGroupAddMembers, without AuthPermissionDenied.
	RootAdminGroupRemoveMembers(networkID string, accountIDs []string) error

	// RootAdminGroupListMembers lists the members of the root network's admin
	// group, a page at a time, and the number of members. A member outside
	// the calling service's namespace is listed from its ASN record, so a
	// page always matches total.
	//
	// Errors: asn.network.not_found; RequestParamInvalid (not a root);
	// GroupNotFound (as RootAdminGroupAddMembers).
	RootAdminGroupListMembers(networkID string, page iam.Page) (accounts []*iam.Account, total int, err error)

	// -------------------------------------------------------------------------
	// Node Group Management
	// All methods are re-entrant.
	// -------------------------------------------------------------------------

	// CreateNodeGroup creates a node group scoped to this service within the given network.
	CreateNodeGroup(networkID, name, description, metadata string) error

	// ListNodeGroups returns all node groups for this service in the given network.
	ListNodeGroups(networkID string) ([]*NodeGroup, error)

	// GetNodeGroupByID returns group details including service-defined Metadata and active ConfigOps.
	GetNodeGroupByID(nodeGroupID string) (*NodeGroup, error)

	// UpdateNodeGroupMetadata persists service-defined metadata on the group.
	UpdateNodeGroupMetadata(nodeGroupID, metadata string) error

	// DeleteNodeGroup removes the node group. Member nodes are not affected.
	DeleteNodeGroup(nodeGroupID string) error

	// SetConfigOfNodeGroup persists service config for the group.
	// Member nodes inherit this config unless they have a node-level config override.
	SetConfigOfNodeGroup(nodeGroupID, config string) error

	// AddNodesToNodeGroup adds the specified nodes to the group.
	AddNodesToNodeGroup(nodeGroupID string, nodeIDs []string) error

	// RemoveNodesFromNodeGroup removes the specified nodes from the group.
	RemoveNodesFromNodeGroup(nodeGroupID string, nodeIDs []string) error

	// -------------------------------------------------------------------------
	// Node Enrollment And Node Tasks
	// CreateNode, GetInstallToken and DeleteNode each perform their whole
	// controller-side effect at call time (the node is created, the service is
	// added, the service is removed, the node is destroyed). The host-side work
	// still to be done is either a node task the framework runs on the host
	// itself, or a ScriptToken for a host that cannot run node tasks. Which of
	// the two is decided only by whether the node registered the task
	// capability; a refused or failed task never turns into a token.
	// RedeemScriptToken turns any ScriptToken, whatever call issued it, into
	// the script bytes; it makes no service-visible change.
	//
	// A node task that changes packages restarts asnsn, and so every service
	// on the node, which sees the node go offline and come back. On a node in
	// Maintenance, or in Maintenance when the task runs, the other services
	// on the node stay stopped until the node leaves Maintenance. A queued task
	// is sent when the node can take it.
	//
	// Task status: CreateNode and UpgradeNode return a RequestID, which the
	// calling service follows through ServiceInfo.PendingTask (queued),
	// RunningTask (sent) and LastRequest (ended); the request is over when
	// LastRequest.RequestID equals it. Status is OK or Failed only; the cause
	// is for an ASN operator. Another actor's task that restarts the node
	// shows in RunningTask and LastTask without a RequestID, as Kind upgrade or
	// install when it changed the calling service's deb and Kind restart
	// otherwise; it names no actor, service, or package. A Failed task without
	// a RequestID on a node in Maintenance leaves the calling service stopped
	// until an ASN operator acts; there is no call of the service's to repeat.
	// Whether the service then loads at the new version shows, as before, in
	// ServiceInfo.Version and State.
	//
	// Typical flow: the service issues a token and hands it to the operator or
	// the device; the device presents it back to the service, which redeems it
	// with RedeemScriptToken and serves the returned script bytes itself.
	//
	// These methods act only for the calling service: none takes a service name,
	// and every script covers asnsn plus the calling service's deb only. The
	// one exception is SetNodeAutoUpgrade, a node setting whose upgrades move
	// every service on the node. Install
	// specs (deb coordinates, apt repos, asnsn versions) come from the
	// controller's asn.conf.
	//
	// Node credentials are framework-internal. Key minting, certificate signing,
	// binding and revocation happen inside the framework as a side effect of
	// these calls and of the scripts running on the host; the service never sees
	// a key or certificate, cannot bind or unbind one, and cannot query a node's
	// credential state.
	//
	// A node whose host can no longer authenticate (the machine was replaced, its
	// key was lost, or its certificate expired) cannot be recovered with an
	// install token: an install script never replaces the credential of a node
	// that already has one. Recovery is one of:
	//   - an operator unbinds the node's certificate from ASN web, after which a
	//     fresh install token re-enrolls it under the same node ID;
	//   - when the calling service is the node's last, DeleteNode (which revokes
	//     the credential) followed by CreateNode: with DeleteEmptyNode false and
	//     AllowExisting the node keeps its ID, with DeleteEmptyNode true it is a
	//     new node. On the same host, run the uninstall script first so the old
	//     certificate is removed.
	// A service can never do this for a node other services are still on: only
	// ASN may revoke the credential of a node shared with other services.
	// -------------------------------------------------------------------------

	// CreateNode creates a framework-owned node for the calling service and
	// returns an install token for it.
	//
	// The node is placed under ParentNetworkID (its network path derives from
	// that placement) with service_names set to the calling service. NodeName
	// must be unique across the whole root network tree that the parent belongs
	// to, not only within the parent.
	//
	// If a node with NodeName already exists in that tree:
	//   - without AllowExisting the call fails and nothing changes;
	//   - with AllowExisting the calling service is added to the existing node's
	//     service_names (a no-op if it is already there). UpdateInfo controls
	//     whether the request's Type and Label overwrite the existing node's
	//     attributes.
	//
	// The host half:
	//   - a node that can run node tasks gets an install task for the calling
	//     service's deb (TaskID, or TaskPending with PendingReason, and a
	//     RequestID to follow it); Token is nil, which is normal. The service is
	//     recorded without loading it; the asnsn the task restarts loads it.
	//     The task restarts every service on the node; on a node with
	//     auto-upgrade on that is behind the controller, an automatic upgrade
	//     and a second restart follow. A result with no task and no token
	//     means the deb is already installed at the version the controller
	//     serves, or the call withdrew this service's queued uninstall; the service
	//     is loaded at once. Repeating CreateNode with AllowExisting for a deb
	//     that is missing or older queues a new install;
	//   - any other node gets an install token (ScriptKindInstall). An online
	//     node is told to load the .so and Init() the service at once, which
	//     succeeds only if the deb is already on the host; an offline node
	//     loads it at its next registration. Running the install script puts
	//     the deb in place, after which asnsn restarts, re-registers and loads
	//     the service.
	CreateNode(req CreateNodeRequest) (*CreateNodeResult, error)

	// GetInstallToken returns a fresh install token for an EXISTING node the
	// calling service is already on (in its service_names). Use it to (re)install
	// or upgrade asnsn and the calling service's deb on the node's host: after the
	// deb version in asn.conf is bumped, after a host was reimaged, or when the
	// token returned by CreateNode has expired.
	//
	// It changes nothing on the node itself; its only effect is issuing the
	// token. Tokens issued earlier stay valid until they expire. A node the
	// calling service is not on is refused; use CreateNode with AllowExisting to
	// join one. tokenTTLSeconds is the token's lifetime; see
	// ScriptToken.ExpiresAt.
	GetInstallToken(nodeID string, tokenTTLSeconds int64) (*ScriptToken, error)

	// DeleteNode removes the calling service from the node and starts the
	// host cleanup. Access-sensitive; audited.
	//
	// All controller-side removal happens here, not at redemption:
	//   - The calling service is torn down (Stop() + Finish() + unload of the .so
	//     on an online node) and dropped from service_names.
	//   - If other services remain on the node, nothing else changes; one service
	//     can never tear down a node another service still uses.
	//   - If the calling service was the node's last, the node's credential is
	//     revoked (a live node loses its session). DeleteEmptyNode then decides
	//     the node record: true destroys it (node-group membership and node
	//     config included); false keeps it, serviceless, and it comes back only
	//     through CreateNode with AllowExisting.
	//
	// The host half:
	//   - when other services remain on a node that can run node tasks, an
	//     uninstall task purges the calling service's deb (TaskID, or
	//     TaskPending with PendingReason); Token is nil, which is normal. The
	//     task restarts every service on the node; on a node in Maintenance the
	//     other services stay stopped until it leaves Maintenance. The removal
	//     is fire-and-forget for the calling service: it is off the node when
	//     the call returns and sees neither the task nor its outcome; an ASN
	//     operator sees and retries a failed purge;
	//   - otherwise (the calling service was the node's last, or the node
	//     cannot run node tasks) the result carries an uninstall token
	//     (ScriptKindUninstall). It stays redeemable even when the node was
	//     destroyed. The service may discard it if the host does not need
	//     cleaning up; the removal above has already taken effect either way.
	//
	// Install tokens the calling service issued for the node earlier no longer
	// redeem, since the service is no longer on the node.
	DeleteNode(req DeleteNodeRequest) (*DeleteNodeResult, error)

	// UpgradeNode upgrades asnsn and the calling service's deb on the node to
	// the versions this controller serves, from the configured APT sources,
	// and returns once the node task is sent, with the requestID to follow in
	// ServiceInfo.RunningTask and LastRequest. The node restarts asnsn, which
	// restarts every service on it, and registers again on the new versions.
	// A node below the controller's version window leaves Maintenance on
	// success.
	//
	// Refused, with nothing sent:
	//   - AuthPermissionDenied: the calling service is not on the node;
	//   - asn.node.task_unsupported: the node cannot run node tasks,
	//     or is above the controller's version window;
	//   - asn.node.connect_failed: the node is not connected;
	//   - asn.node.up_to_date: nothing would change;
	//   - asn.node.upgrade_blocked: other services on the node must
	//     move to a new line with it; an ASN operator upgrades the whole node;
	//   - asn.node.task_in_progress: a node task is in flight;
	//   - ResourceLimitReached: the controller's concurrency limit, which
	//     automatic and fleet upgrades share; repeat the call.
	UpgradeNode(nodeID string) (requestID string, err error)

	// SetNodeAutoUpgrade turns automatic upgrade of the node on or off. With it
	// on, whenever the node is admitted normally and asnsn or any service deb on
	// it is older than this controller serves on the same line, the controller
	// upgrades the whole node: asnsn restarts, and so does every service on the
	// node, including other services'. Upgrades start as soon as the controller
	// restarts with newer builds, a few nodes at a time, and also right after an
	// install on a node that is behind (a second restart). One attempt is made
	// per target version set; after a failure the node is held
	// (Node.AutoUpgradeHeld) until the target changes or the switch is turned
	// off and on again. Setting the value the node already has changes nothing.
	// Nodes on an older line are never upgraded automatically.
	//
	// The setting belongs to the node, not to the calling service: an ASN
	// operator and every service on the node can change it, and every service
	// on the node is told (ChangeKindNodeAutoUpgrade). The upgrade shows in
	// ServiceInfo as another actor's upgrade, without RequestID, also for the
	// service that turned the switch on.
	//
	// Refused:
	//   - AuthPermissionDenied: the calling service is not on the node;
	//   - asn.node.task_unsupported: turning it on for a node that cannot run
	//     node tasks.
	SetNodeAutoUpgrade(nodeID string, enabled bool) error

	// RedeemScriptToken renders the script a token stands for. It is the single
	// redemption point for every token, whichever call issued it; the script's
	// kind follows the token's Kind. token is ScriptToken.Token as presented by
	// the host; it must have been issued to the calling service and must not be
	// expired.
	//
	// Redemption makes no service-visible change: it never creates, modifies or
	// deletes a node, never changes its service_names, state or config, and never
	// consumes the token, so it may be repeated any number of times until the
	// token expires. Redeeming an install token does record framework-internal
	// credential state (each rendered script carries a freshly signed candidate
	// certificate), and the number of unregistered candidates per node is
	// capped: past the cap redemption fails with ResourceLimitReached until one
	// registers or the candidates expire.
	//
	// An install script (ScriptKindInstall) installs or upgrades asnsn and the
	// calling service's deb at the versions in asn.conf, and carries the
	// credential the node needs to register. The calling service must still be
	// on the node at redemption time. Before changing anything, the script
	// checks the host and aborts if the host already belongs to a different node,
	// or if any package is installed at a version higher than the target. A host
	// already enrolled as this node keeps its own credential, so re-running an
	// install script on a working node never disrupts it.
	//
	// An uninstall script (ScriptKindUninstall) purges the calling service's
	// deb and restarts asnsn; when the service was the node's last, it also
	// purges asnsn and removes the node's credential and asn.conf. Which of the
	// two was fixed when DeleteNode ran. Before changing anything, the script
	// checks that the host belongs to this node and aborts otherwise.
	//
	// Both scripts are idempotent and safe to re-run.
	RedeemScriptToken(token string) (*Script, error)
}
