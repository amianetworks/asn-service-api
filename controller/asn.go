// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
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
	// Call once per name in Init().
	InitDocDB(name string) (commonapi.DocDBHandler, error)

	// InitTSDB returns a connected time-series database handle.
	// Call once per name in Init().
	InitTSDB(name string) (commonapi.TSDBHandler, error)

	// InitLocker returns a cluster-wide distributed lock.
	// Call once in Init().
	InitLocker() (Lock, error)

	// GetIAM returns the IAM instance for account, group, and access management.
	// Call once in Init().
	GetIAM() (iam.Instance, error)

	// GetSubscription returns the In-App Subscription instance.
	// Call once in Init().
	GetSubscription() (subscription.Instance, error)

	// -------------------------------------------------------------------------
	// Service Lifecycle Management
	// -------------------------------------------------------------------------

	// AddServiceToNode loads this service's .so onto the target node and triggers Init().
	// The node must be online (NodeStateOnline).
	AddServiceToNode(nodeID string) error

	// DeleteServiceFromNode calls Stop() + Finish() on the node's service instance,
	// then unloads the .so. Use for permanent removal; not a substitute for StopService().
	DeleteServiceFromNode(nodeID string) error

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
	// If withService is true, only nodes that have this service loaded are returned.
	// Internal links: both endpoints within the network; the To node is included in the returned nodes slice.
	// External links: the To endpoint is outside the network and is not included in nodes.
	GetNodesOfNetwork(networkID string, withService bool) (nodes []*Node, links []*Link, err error)

	// SubscribeNodeStateChanges returns a channel for node state changes.
	// One-shot: a second call returns an error.
	// On subscription, the channel first delivers a NodeStateChange for every node's current state
	// (initial snapshot), then delivers incremental changes. The channel is never closed during
	// normal framework operation.
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
}
