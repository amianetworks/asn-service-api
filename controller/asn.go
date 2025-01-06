// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/amianetworks/asn-service-api/v25/log"
)

/*
	Struct used between asn.controller and service.controller,
*/

// Node struct
type Node struct {
	Id             string
	Type           string
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
		Send CONFIG cmd to the service node with the specific service name, the configCmd is a pre-defined struct.
		Both of service.controller and service.sn has the same struct,
		so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct
	*/
	SendServiceOps(serviceNodeId string, serviceName string, serviceOps []byte) error

	/*
		Read the service COnf by network id and service name,
		The setting []byte is the config/rule/policies struct defined in service.controller,
		Use Unmarshall to converting the []byte to the Conf struct
	*/
	ReadConfOfNetwork(serviceName string) ([]byte, error)
	ReadConfOfGroup(groupName string, serviceName string) ([]byte, error)
	ReadConfOfServiceNode(serviceNodeId string, serviceName string) ([]byte, error)

	/*
		Set the service setting by network id and service name,
		the Conf []byte is Marshalled
		Write the service setting to a specific service node by ASN controller
	*/
	SaveConfOfNetwork(serviceName string, config []byte) error
	SaveConfOfGroup(groupName string, serviceName string, config []byte) error
	SaveConfOfServiceNode(serviceNodeId string, serviceName string, config []byte) error

	/*
		CRUD (Create, Read, Update, Delete) operation for the service metadata.
		The metadata []byte is Marshalled
	*/
	ReadMetadataOfNetwork(serviceName string, fileName string) ([]byte, error)
	ReadMetadataOfGroup(groupName string, serviceName string, fileName string) ([]byte, error)
	ReadMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string) ([]byte, error)

	SaveMetadataOfNetwork(serviceName string, fileName string, metadata []byte) error
	SaveMetadataOfGroup(groupName string, serviceName string, fileName string, metadata []byte) error
	SaveMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string, metadata []byte) error

	DeleteMetadataOfNetwork(serviceName string, fileName string) error
	DeleteMetadataOfGroup(groupName string, serviceName string, fileName string) error
	DeleteMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string) error

	/*
		Write the log to your service path. This is based on am.module logs
	*/
	GetLogger(serviceName string) (*log.Logger, error)
	GetIAM() (IAM, error)
}
