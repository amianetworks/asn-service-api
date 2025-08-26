// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
	"asn.amiasys.com/asn-service-api/v25/iam"
	"asn.amiasys.com/asn-service-api/v25/log"
)

// ASNController
//
// Provided resource and functions:
// 1. Initialization and Resource Allocation
// 2. Service Management
// 3. Networks
// 4. Nodes
// 5. Node Group
type ASNController interface {

	/*
		Initialization and Resource Allocation
	*/

	// InitLogger returns the logger for a service.
	//
	// ASN Framework manages loggers for all services, and the default log files are <servicename>-*.log.
	// SHOULD ONLY call once. Further calls will get an error.
	InitLogger() (*log.Logger, error)

	// InitDocDB returns a doc DB handle.
	//
	// The DB is connected and ready for use through the DocDBHandler upon return.
	// SHOULD ONLY call once for each name. Further calls will get an error.
	InitDocDB() (commonapi.DocDBHandler, error)

	// InitTSDB returns a connected time-series database handle.
	//
	// SHOULD ONLY call once for each name. Further calls will get an error.
	InitTSDB() (commonapi.TSDBHandler, error)

	// InitLocker returns a distributed locker for the service.
	//
	// SHOULD ONLY call once. Further calls will get an error.
	InitLocker() (Lock, error)

	// GetIAM returns the IAM instance for user and group management.
	//
	// SHOULD ONLY call once. Further calls will get an error.
	GetIAM(forceMfa bool) (iam.Instance, error)

	/*
		Service Management
	*/

	// AddServiceToNode loads the service .so into an existing node and initializes it.
	AddServiceToNode(nodeID string) error

	// DeleteServiceFromNode unloads the service .so from an existing node.
	DeleteServiceFromNode(nodeID string) error

	// StartService starts the service on the specified Service Nodes.
	StartService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// StopService stops the service on the specified Service Nodes.
	StopService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// ResetService resets the service on the specified Service Nodes.
	ResetService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// SendServiceOps sends an op command to service nodes.
	//
	// The service defines the op payload. Both service.controller and service.sn
	// should share the same structure, so they can use json.Marshal/json.Unmarshal to convert
	// between string and the struct.
	SendServiceOps(serviceScope commonapi.ServiceScope, serviceScopeList []string, opCmd, opParams string) error

	/*
		Networks
	*/

	// GetNetworks returns all networks, their info, and subnetworks in the topology.
	GetNetworks() ([]*Network, error)

	/*
		Nodes
	*/

	// GetNodeByID returns a node's info by ID that includes Metadata set by the service.
	GetNodeByID(nodeID string) (*Node, error)

	// UpdateNodeMetadata updates a service-specific metadata for the Node.
	//
	// The framework stores the Metadata for services and can be retrieved by GetNodeByID.
	UpdateNodeMetadata(nodeID, metadata string) error

	// SetConfigOfNode saves the service config for a node.
	//
	// config is expected to contain YAML (UTF-8).
	SetConfigOfNode(nodeID, config string) error

	// GetNodesOfNetwork returns all nodes of a network, and its internal and external links.
	//
	// If withService is true, only nodes currently with this service are returned.
	//
	// Links may contain two types:
	// - Internal links connect nodes within the same network; referenced nodes are included in the returned nodes slice.
	// - External links connect in-network nodes to external "To" nodes via NodeExternalLink.
	GetNodesOfNetwork(networkID string, withService bool) (nodes []*Node, links []*Link, err error)

	// SubscribeNodeStateChanges returns a receive-only channel for node state changes.
	//
	// Upon subscription, the channel first yields all initial states, then the following state changes.
	// SHOULD ONLY be called once. Further calls will get an error.
	SubscribeNodeStateChanges() (<-chan *NodeStateChange, error)

	/*
		Node Group
	*/

	// CreateNodeGroup creates a node group for *this* service.
	CreateNodeGroup(networkID, name, description, metadata string) error

	// ListNodeGroups returns all node groups for *this* service.
	ListNodeGroups(networkID string) ([]*NodeGroup, error)

	// GetNodeGroupByID returns a node group's info by ID that includes service metadata.
	GetNodeGroupByID(nodeGroupID string) (*NodeGroup, error)

	// UpdateNodeGroupMetadata updates a node group's metadata used by *this* service.
	UpdateNodeGroupMetadata(nodeGroupID, meta string) error

	// DeleteNodeGroup removes a node group for *this* service.
	DeleteNodeGroup(nodeGroupID string) error

	// SetConfigOfNodeGroup saves the service config for a node group.
	//
	// config format is private to the service.
	SetConfigOfNodeGroup(nodeGroupID, config string) error

	// AddNodesToNodeGroup adds the specified nodes to the node group identified by ID.
	AddNodesToNodeGroup(nodeGroupID string, nodeIDs []string) error

	// RemoveNodeFromNodeGroup removes the specified nodes from the node group identified by ID.
	RemoveNodeFromNodeGroup(nodeGroupID string, nodeIDs []string) error
}
