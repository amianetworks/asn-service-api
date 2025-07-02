// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
	"asn.amiasys.com/asn-service-api/v25/iam"
	"asn.amiasys.com/asn-service-api/v25/log"
)

// Structs used between asn.controller and service.controller.

// Network is the structure for a network.
type Network struct {
	Id       string
	ParentID string
	ChildIDs []string // subnetworks

	Name string

	// TODO
}

// Node is the structure for a node.
type Node struct {
	Id       string
	ParentId string

	Type string
	Name string

	ServiceNodeState int
	ServiceState     int

	LinkIDs []string

	// TODO
}

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
	GetIAM() (iam.Instance, error)

	/*
		Service Management
	*/

	// StartService starts service on specified Service Nodes.
	StartService(serviceScope int, serviceScopeList []string, clusterConfig, instanceConfig []byte) error

	// StopService stops service on specified Service Nodes.
	StopService(serviceScope int, serviceScopeList []string) error

	// ResetService resets service on specified Service Nodes.
	ResetService(serviceScope int, serviceScopeList []string) error

	// SendServiceOps sends CONFIG cmd to the service node.
	// The configCmd is a pre-defined struct. Both service.controller and service.sn has the same struct,
	// so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct.
	SendServiceOps(nodeId, opCmd, opParams string) (response chan *commonapi.Response, frameworkErr error)

	/*
		Service Configuration Management
	*/

	// SaveDefaultClusterConfig saves the default cluster setting.
	SaveDefaultClusterConfig(config []byte) error

	// SaveClusterConfigOfNetwork saves the cluster setting for a network.
	SaveClusterConfigOfNetwork(networkID string, config []byte) error

	// SaveClusterConfigOfNode saves the cluster setting for a node.
	SaveClusterConfigOfNode(nodeId string, config []byte) error

	// SaveInstanceConfigOfNode saves the instance setting for a node.
	SaveInstanceConfigOfNode(nodeId string, config []byte) error

	/*
		Network, Nodes, and Groups (config)
	*/

	// GetSubnetworksOfNetwork returns all subnetworks of a network
	GetSubnetworksOfNetwork(networkID string) ([]Network, error)

	// GetNodesOfNetwork returns all nodes of a network
	GetNodesOfNetwork(networkID string) ([]Node, error)

	// GetNodeById returns node by id
	GetNodeById(id string) (Node, error)
}
