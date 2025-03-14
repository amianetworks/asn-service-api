// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	commonapi "github.com/amianetworks/asn-service-api/v25/common"
	"github.com/amianetworks/asn-service-api/v25/log"
)

/*
	Struct used between asn.controller and service.controller,
*/

// Node struct
type Node struct {
	Id             string
	Type           string
	State          string
	ParentId       string
	Group          string
	ExternalLinked []string
	InternalLinked []string
	Services       map[string]bool
}

type Group struct {
	Name  string
	Nodes []string
}

// ASNController API provided by ASN controller
type ASNController interface {
	/*
		Get all nodes of network
	*/
	GetNodesOfNetwork(serviceName string) ([]Node, error)

	/*
		Get all groups of network
	*/
	GetGroupsOfNetwork(serviceName string) ([]Group, error)

	/*
		Get all nodes of group
	*/
	GetNodesOfGroup(groupName, serviceName string) ([]Node, error)

	/*
		Get all nodes of the parent.
	*/
	GetNodesOfParent(parentNodeId, serviceName string) ([]Node, error)

	/*
		Get node by id
	*/
	GetNodeById(id, serviceName string) (Node, error)

	/*
		Get group by group name
	*/
	GetGroupByName(groupName string) (Group, error)

	/*
		Send START cmd to the service node with the specific service name
		The config is a pre-defined struct. Both of service.controller and service.sn has the same struct,
		so they can easily use xxx.Marshall() and xxx.Unmarshall() to convert the struct between []byte and the struct
	*/
	StartService(serviceNodeId string, serviceName string, config []byte) error

	/*
		Send STOP cmd to the service node with the specific service name
	*/
	StopService(serviceNodeId string, serviceName string) error

	/*
		Send RESET cmd to the service node with the specific service name
	*/
	ResetService(serviceNodeId string, serviceName string) error

	/*
		Send CONFIG cmd to the service node with the specific service name, the configCmd is a pre-defined struct.
		Both of service.controller and service.sn has the same struct,
		so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct
	*/
	SendServiceOps(serviceNodeId, serviceName, serviceOpCmd, serviceOpParams string) (serviceResponse chan *commonapi.Response, frameworkErr error)

	/*
		Write the log to your service path. This is based on am.module logs
	*/
	GetLogger(serviceName string) (*log.Logger, error)
	GetIAM(serviceName string) (IAM, error)
	GetDBConfig(serviceName string, dbType string) (*DBConf, error)
	GetLock(serviceName string) (Lock, error)
}
