// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/amianetworks/asn-service-api/v25/log"

	commonapi "github.com/amianetworks/asn-service-api/v25/common"
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

// ASNController contains the APIs provided by ASN controller.
type ASNController interface {
	// GetNodesOfNetwork returns all nodes of network
	GetNodesOfNetwork() ([]Node, error)

	// GetGroupsOfNetwork returns all groups of network
	GetGroupsOfNetwork() ([]Group, error)

	// GetNodesOfGroup returns all nodes of group
	GetNodesOfGroup(groupName string) ([]Node, error)

	// GetNodesOfParent returns all nodes of the parent
	GetNodesOfParent(parentNodeId string) ([]Node, error)

	// GetNodeById returns node by id
	GetNodeById(id string) (Node, error)

	// GetGroupByName returns group by group name
	GetGroupByName(groupName string) (Group, error)

	// StartService sends START cmd to the service node.
	// The config is a pre-defined struct. Both service.controller and service.sn has the same struct,
	// so they can easily use xxx.Marshall() and xxx.Unmarshall() to convert the struct between []byte and the struct.
	StartService(serviceNodeId string, config []byte) error

	// StopService sends STOP cmd to the service node.
	StopService(serviceNodeId string) error

	// ResetService sends RESET cmd to the service node.
	ResetService(serviceNodeId string) error

	// SendServiceOps sends CONFIG cmd to the service node.
	// The configCmd is a pre-defined struct. Both service.controller and service.sn has the same struct,
	// so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct.
	SendServiceOps(serviceNodeId, serviceOpCmd, serviceOpParams string) (serviceResponse chan *commonapi.Response, frameworkErr error)

	// SaveDefaultClusterConfig saves the default cluster setting.
	SaveDefaultClusterConfig(config []byte) error

	// SaveClusterConfigOfGroup saves the cluster setting for a group.
	SaveClusterConfigOfGroup(groupName string, config []byte) error

	// SaveClusterConfigOfServiceNode saves the cluster setting for a service node.
	SaveClusterConfigOfServiceNode(serviceNodeId string, config []byte) error

	// SaveInstanceConfigOfServiceNode saves the instance setting for a service node.
	SaveInstanceConfigOfServiceNode(serviceNodeId string, config []byte) error

	// GetLogger returns the logger for a service.
	GetLogger() (*log.Logger, error)

	// GetIAM returns the IAM instance for a service.
	GetIAM() (IAM, error)

	// GetDBConfig returns the DB config for a service.
	GetDBConfig(dbType string) (*DBConf, error)

	// GetLock returns the locker for a service.
	GetLock() (Lock, error)
}
