// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	commonapi "github.com/amianetworks/asn-service-api/v25/common"
	"github.com/amianetworks/asn-service-api/v25/log"
)

// Structs used between asn.controller and service.controller.

// Node is the structure for a network node.
type Node struct {
	Id               string
	Type             string
	ServiceNodeState int
	ServiceState     int
	ParentId         string
	Group            string
	ExternalLinked   []string
	InternalLinked   []string
}

// Group is the structure for a configuration group.
type Group struct {
	Name   string
	Remark string
	Nodes  []string
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
	GetIAM() (IAM, error)

	/*
		Service Management
	*/

	// StartServiceOnNode starts service on specified Service Node.
	StartServiceOnNode(serviceScope string, serviceScopeList []string, clusterConfig, instanceConfig []byte) error

	// StopServiceOnNode stops service on specified Service Node.
	StopServiceOnNode(serviceScope string, serviceScopeList []string) error

	// ResetServiceOnNode resets service on specified Service Node.
	ResetServiceOnNode(serviceScope string, serviceScopeList []string) error

	// SendServiceOps sends CONFIG cmd to the service node.
	// The configCmd is a pre-defined struct. Both service.controller and service.sn has the same struct,
	// so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct.
	SendServiceOps(serviceNodeId, opCmd, opParams string) (response chan *commonapi.Response, frameworkErr error)

	/*
		Service Configuration Management
	*/

	// SaveDefaultClusterConfig saves the default cluster setting.
	SaveDefaultClusterConfig(config []byte) error

	// SaveClusterConfigOfGroup saves the cluster setting for a group.
	SaveClusterConfigOfGroup(groupName string, config []byte) error

	// SaveClusterConfigOfServiceNode saves the cluster setting for a service node.
	SaveClusterConfigOfServiceNode(serviceNodeId string, config []byte) error

	// SaveInstanceConfigOfServiceNode saves the instance setting for a service node.
	SaveInstanceConfigOfServiceNode(serviceNodeId string, config []byte) error

	/*
		Network, Nodes, and Groups (config)
	*/

	// GetNodesOfNetwork returns all nodes of network
	GetNodesOfNetwork() ([]Node, error)

	// GetGroupsOfNetwork returns all groups in the network
	GetGroupsOfNetwork() ([]Group, error)

	// GetGroupByName returns group by group name
	GetGroupByName(groupName string) (Group, error)

	// GetNodesOfGroup returns all nodes of group
	GetNodesOfGroup(groupName string) ([]Node, error)

	// GetNodesOfParent returns all nodes of the parent
	GetNodesOfParent(parentNodeId string) ([]Node, error)

	// GetNodeById returns node by id
	GetNodeById(id string) (Node, error)
}
