// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/amianetworks/asn-service-api/v25/log"

	commonapi "github.com/amianetworks/asn-service-api/v25/common"
)

/*
	Struct used between asn.controller and service.controller,
*/

// Node struct
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

type Group struct {
	Name   string
	Remark string
	Nodes  []string
}

// ASNController API provided by ASN controller
type ASNController interface {
	/*
		Get all nodes of network
	*/
	GetNodesOfNetwork() ([]Node, error)

	/*
		Get all groups of network
	*/
	GetGroupsOfNetwork() ([]Group, error)

	/*
		Get all nodes of group
	*/
	GetNodesOfGroup(groupName string) ([]Node, error)

	/*
		Get all nodes of the parent.
	*/
	GetNodesOfParent(parentNodeId string) ([]Node, error)

	/*
		Get node by id
	*/
	GetNodeById(id string) (Node, error)

	/*
		Get group by group name
	*/
	GetGroupByName(groupName string) (Group, error)

	/*
		Send START cmd to the service node with the specific service name
		The config is a pre-defined struct. Both of service.controller and service.sn has the same struct,
		so they can easily use xxx.Marshall() and xxx.Unmarshall() to convert the struct between []byte and the struct
	*/
	StartService(serviceNodeId string, config []byte) error

	/*
		Send STOP cmd to the service node with the specific service name
	*/
	StopService(serviceNodeId string) error

	/*
		Send RESET cmd to the service node with the specific service name
	*/
	ResetService(serviceNodeId string) error

	/*
		Send CONFIG cmd to the service node with the specific service name, the configCmd is a pre-defined struct.
		Both of service.controller and service.sn has the same struct,
		so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct
	*/
	SendServiceOps(serviceNodeId, serviceOpCmd, serviceOpParams string) (serviceResponse chan *commonapi.Response, frameworkErr error)

	/*
		Set the service setting by network id and service name,
		the Conf []byte is Marshalled
		Write the service setting to a specific service node by ASN controller
	*/
	SaveDefaultClusterConfig(config []byte) error
	SaveClusterConfigOfGroup(groupName string, config []byte) error
	SaveClusterConfigOfServiceNode(serviceNodeId string, config []byte) error
	SaveInstanceConfigOfServiceNode(serviceNodeId string, config []byte) error

	/*
		Write the log to your service path. This is based on am.module logs
	*/
	GetLogger() (*log.Logger, error)
	GetIAM() (IAM, error)
	GetDBConfig(dbType string) (*DBConf, error)
	GetLock() (Lock, error)
}
