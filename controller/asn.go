// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/redis/go-redis/v9"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
	"asn.amiasys.com/asn-service-api/v26/iam"
	"asn.amiasys.com/asn-service-api/v26/log"
	"asn.amiasys.com/asn-service-api/v26/subscription"
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
//  6. Node group management
//  7. Node enrollment
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

	// GetSubscription returns the In-App Subscription instance, whose state is
	// kept in the service's own document database.
	//
	// docDBName selects that database exactly as InitDocDB's name does: the
	// framework prefixes it with the service's configured docdb db_name, so an
	// empty string means <db_name> itself and any other value <db_name>_<name>.
	// docDBSubCollName is the collection holding the store subscriptions (an
	// account may have several); docDBSubRecordCollName is the collection
	// holding their per-billing-period records. Both are required (an empty name is an error), are
	// collections inside that database, and must not be used by the service for
	// anything else.
	//
	// Call once in Init().
	GetSubscription(docDBName, docDBSubCollName, docDBSubRecordCollName string) (subscription.Instance, error)

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
	// Internal links: both endpoints within the network; the To node is included in the returned nodes slice.
	// External links: the To endpoint is outside the network and is not included in nodes.
	GetNodesOfNetwork(networkID string, withService bool) (nodes []*Node, links []*Link, err error)

	// SubscribeNodeStateChanges returns a channel for node state changes.
	// One-shot: a second call returns an error.
	// On subscription, the channel first delivers a NodeStateChange for every node's current state
	// (initial snapshot), then delivers incremental changes. The channel is never closed during
	// normal framework operation.
	// Each NodeStateChange is a full snapshot of the connectivity and service
	// axes. Credential changes inside the framework (token issue, certificate
	// binding or revocation) are not service-visible and produce no event.
	SubscribeNodeStateChanges() (<-chan *NodeStateChange, error)

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
	// Node Enrollment
	// CreateNode, GetInstallToken and DeleteNode each perform their whole
	// controller-side effect at call time (the node is created, the service is
	// added, the service is removed, the node is destroyed) and return a
	// ScriptToken for the host-side work still to be done. RedeemScriptToken
	// turns any ScriptToken, whatever call issued it, into the script bytes; it
	// makes no service-visible change.
	//
	// Typical flow: the service issues a token and hands it to the operator or
	// the device; the device presents it back to the service, which redeems it
	// with RedeemScriptToken and serves the returned script bytes itself.
	//
	// These methods act only for the calling service: none takes a service name,
	// and every script covers asnsn plus the calling service's deb only. Install
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
	//     service_names (a no-op if it is already there) and an install token for
	//     that node is returned. An online node is told to load the .so and
	//     Init() the service at once, which succeeds only if the deb is already
	//     on the host; an offline node loads it at its next registration. Either
	//     way, running the install script puts the deb in place, after which
	//     asnsn restarts, re-registers and loads the service. UpdateInfo controls
	//     whether the request's Type and Label overwrite the existing node's
	//     attributes.
	//
	// The returned token is ScriptKindInstall.
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
	// join one.
	GetInstallToken(req GetInstallTokenRequest) (*ScriptToken, error)

	// DeleteNode removes the calling service from the node and returns an
	// uninstall token for cleaning up the host. Access-sensitive; audited.
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
	// The returned token is ScriptKindUninstall. It stays redeemable even when
	// the node was destroyed. Install tokens the calling service issued for the
	// node earlier no longer redeem, since the service is no longer on the node.
	// The service may discard the uninstall token if the host does not need
	// cleaning up; the removal above has already taken effect either way.
	DeleteNode(req DeleteNodeRequest) (*DeleteNodeResult, error)

	// RedeemScriptToken renders the script a token stands for. It is the single
	// redemption point for every token, whichever call issued it; the script's
	// kind follows the token's Kind. The token must have been issued to the
	// calling service and must not be expired.
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
	RedeemScriptToken(req RedeemScriptTokenRequest) (*Script, error)
}
