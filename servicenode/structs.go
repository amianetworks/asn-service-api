// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	"errors"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

var (
	// ErrServiceNotFound is returned when the target service is not loaded on this node.
	ErrServiceNotFound = errors.New("target service not found")
	// ErrKeyNotFound is returned when a requested shared data key is not provided by the target service.
	ErrKeyNotFound = errors.New("key not found in target service")
	// ErrRestartNeeded is returned from ASNService.Start() when the service cannot hot-reload
	// its configuration. The framework will execute Stop() → Start() with the new config.
	ErrRestartNeeded = errors.New("restart needed")

	// ErrNotMasterNode is returned by slave API methods when the node is not in cluster mode.
	ErrNotMasterNode = errors.New("not running as master node")
	// ErrSlaveNotConnected is returned when the named slave's stream is not established.
	ErrSlaveNotConnected = errors.New("slave is not connected")
	// ErrSlaveServiceNotFound is returned when the named slave has not reported loading this service.
	ErrSlaveServiceNotFound = errors.New("service not found on slave")
)

// NodeInfo is the node information available to a service running on the node.
// It embeds commonapi.NodeInfo (hardware interfaces, IPMI, management, device specs)
// and adds the node's ID and the list of active config op strings.
type NodeInfo struct {
	ID string
	commonapi.NodeInfo
	// ConfigOps is the list of active config op param strings applied to this node.
	ConfigOps []string
}

// SlaveNodeInfo contains the identifying information for a slave node managed by the master node.
type SlaveNodeInfo struct {
	// Name is the node_name reported by the slave in its SnRegistrationRequest.
	// Empty if the slave has never successfully connected.
	Name string
	// Connected reports whether the bidi stream to this slave is currently established.
	Connected bool

	Management *commonapi.Management
	DeviceInfo *commonapi.DeviceInfo
	GrpcUrl    string
}
