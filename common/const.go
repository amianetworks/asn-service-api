// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

type NodeState int

const (
	NodeStateUnregistered NodeState = iota
	NodeStateOffline
	NodeStateOnline
	NodeStateMaintenance
)

type ServiceState int

const (
	ServiceStateUnavailable ServiceState = iota
	ServiceStateUninitialized
	ServiceStateInitialized
	ServiceStateConfiguring
	ServiceStateRunning
	ServiceStateMalfunctioning
)

type NodeType string

const (
	NodeTypeRouter       NodeType = "router"
	NodeTypeSwitch       NodeType = "switch"
	NodeTypeAppliance    NodeType = "appliance"
	NodeTypeFirewall     NodeType = "firewall"
	NodeTypeLoadBalancer NodeType = "lb"
	NodeTypeAccessPoint  NodeType = "ap"
	NodeTypeDevice       NodeType = "device"
	NodeTypeServer       NodeType = "server"
)

type ServiceScope int

const (
	ServiceScopeNetwork ServiceScope = 1 + iota
	ServiceScopeNetworkWithSubnetworks
	ServiceScopeGroup
	ServiceScopeNode
)

type ServiceConfigSource int

const (
	ServiceConfigSourceNode ServiceConfigSource = 1 + iota
	ServiceConfigSourceNodeGroup
)

type NetIfType string

const (
	NetIfTypeData       NetIfType = "data"
	NetIfTypeControl    NetIfType = "control"
	NetIfTypeManagement NetIfType = "management"
)

var LocationTiers = []string{
	"world",
	"country",
	"state",
	"city",
	"district",
	"campus",
	"building",
	"floor",
	"room",
	"row",
	"rack",
	"unit",
}
