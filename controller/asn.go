// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
	"asn.amiasys.com/asn-service-api/v25/iam"
	"asn.amiasys.com/asn-service-api/v25/log"
)

// ASNController
//
// 1. Initialization and resource allocation.
// 2. Service
// 3. Service Configuration Management
// 4. Network and Network Nodes
type ASNController interface {

	/*
		Initialization
	*/

	// InitLogger returns the logger for a service.
	// ASN Framework manages loggers for all services, and the default log files are <servicename>-*.log
	// Only one logger is allocated if called multiple times.
	InitLogger() (*log.Logger, error)

	// InitDocDB ASN Controller will return a doc DB handle.
	// The DB is connected and ready for use through the DocDBHandler upon return.
	//
	// A Service may call InitDocDB() multiple time forDBs for different uses.
	InitDocDB() (commonapi.DocDBHandler, error)

	// InitTSDB ASN Controller will return a doc DB handle.
	// The DB is connected and ready for use through the TSDBHandler upon return.
	//
	// A Service may call InitTSDB() multiple time forDBs for different uses.
	InitTSDB() (commonapi.TSDBHandler, error)

	// InitLocker returns the locker for a service.
	InitLocker() (Lock, error)

	// GetIAM is different from DB or logger.
	GetIAM(forceMfa bool) (iam.Instance, error)

	/*
		Service Management
	*/

	// AddServiceToNode adds a .so file to an existing node, and inits this service on that node.
	// NOTE: Load service.so for a service node
	AddServiceToNode(nodeID string) error

	// DeleteServiceFromNode removes this service from an existing node.
	// NOTE: Unload service.so for a service node
	DeleteServiceFromNode(nodeID string) error

	// StartService starts service on specified Service Nodes.
	StartService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// StopService stops service on specified Service Nodes.
	StopService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// ResetService resets service on specified Service Nodes.
	ResetService(serviceScope commonapi.ServiceScope, serviceScopeList []string) error

	// SendServiceOps sends CONFIG cmd to the service node.
	// The configCmd is a pre-defined struct. Both service.controller and service.sn have the same struct,
	// so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct.
	SendServiceOps(serviceScope commonapi.ServiceScope, serviceScopeList []string, opCmd, opParams string) error

	/*
		Networks
	*/

	// GetNetworks returns all networks, their info and subnetworks in the topo.
	GetNetworks() ([]*Network, error)

	/*
		Nodes
	*/

	// CreateNode creates a node under a given network.
	// Note that this is only supported when ASN does not strictly verify the network topology.
	// For now, a certificate is returned for the node to register to ASN Controller.
	CreateNode(networkID, nodeName string, nodeType commonapi.NodeType, metadata string) (string, error)

	UpdateNodeMetadata(nodeID, meta string) error

	// SetConfigOfNode saves the cluster setting for a node.
	SetConfigOfNode(nodeId string, config []byte) error

	GetNodeByID(nodeID string) (*Node, error)

	// GetNodesOfNetwork returns all nodes of a network, and its internal and external links.
	// - filterUnavailable will just return the service nodes that have the service if true
	// - Links may contain two different types:
	//   - Internal links connect the nodes within the same network, and it is included in the returned nodes array.
	//     So, only IDs are returned in this case.
	//   - External links connect nodes in this network with nodes outside of this network.
	//     So, the "To" node is not included in the returned nodes array, but in the "NodeExternalLink" structure.
	GetNodesOfNetwork(networkID string, filterUnavailable bool) (nodes []*Node, links []*Link, err error)

	// SubscribeNodeStateChanges returns a channel for a service to subscribe to all nodes' state changes.
	// By listening to this channel, the service will first receive all init states of the nodes,
	// then start to receive messages when the state of a node changes.
	//
	// CAUTION: This function should only be called once. Multiple calling towards this function will return an error.
	SubscribeNodeStateChanges() (<-chan *NodeStateChange, error)

	/*
		Node Group
	*/

	// CreateNodeGroup creates a node group for this service.
	CreateNodeGroup(networkID, name, description, metadata string) error

	// ListNodeGroups returns all node groups under this service.
	ListNodeGroups(networkID string) ([]*NodeGroup, error)

	GetNodeGroupByID(nodeGroupID string) (*NodeGroup, error)

	// DeleteNodeGroup removes a node group under this service.
	DeleteNodeGroup(id string) error

	// SetConfigOfNodeGroup saves the cluster setting for a node group.
	SetConfigOfNodeGroup(nodeGroupID string, config []byte) error

	UpdateNodeGroupMetadata(nodeGroupID, meta string) error

	// AddNodesToNodeGroup adds the specified nodes to the provided node group identified by its ID.
	AddNodesToNodeGroup(nodeGroupID string, nodeIDs []string) error

	// RemoveNodeFromNodeGroup removes the specified nodes from the provided node group identified by its ID.
	RemoveNodeFromNodeGroup(nodeGroupID string, nodeIDs []string) error
}
