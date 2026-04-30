// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

// NodeState is the connectivity state of a node, tracked by the framework.
// Services cannot set this directly; observe changes via ASNController.SubscribeNodeStateChanges().
type NodeState int

const (
	NodeStateUnregistered NodeState = iota // never successfully registered with the framework
	NodeStateOffline                       // registered but currently unreachable
	NodeStateOnline                        // connected and reachable
	NodeStateMaintenance                   // online but in maintenance mode
)

// ServiceState is the operational state of a service on a specific node.
// The framework tracks this independently per node.
// Services influence it only through return values and runtimeErrChan.
type ServiceState int

const (
	ServiceStateUnavailable    ServiceState = iota // .so not loaded on this node
	ServiceStateUninitialized                      // loaded; Init() not yet succeeded
	ServiceStateInitialized                        // Init() succeeded; ready for Start()
	ServiceStateConfiguring                        // applying config or config ops
	ServiceStateRunning                            // Start() succeeded; fully operational
	ServiceStateMalfunctioning                     // fatal error or config op failure
)

// NodeType identifies the hardware or logical role of a node.
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

// ServiceScope defines the targeting granularity for service management and operations dispatch.
// Used in StartService, StopService, ResetService, SendServiceOps, and config op methods.
type ServiceScope int

const (
	// ServiceScopeNetwork (1): target all nodes in the given networks.
	ServiceScopeNetwork ServiceScope = 1 + iota

	// ServiceScopeNetworkWithSubnetworks (2): target all nodes in the given networks, recursively including subnetworks.
	ServiceScopeNetworkWithSubnetworks

	// ServiceScopeNodeGroup (3): target all nodes in the given node groups.
	// Required scope for config op methods.
	ServiceScopeNodeGroup

	// ServiceScopeNode (4): target a few specific nodes by ID.
	// Required scope for config op methods.
	ServiceScopeNode
)

// ServiceSource identifies the origin of a node's active service configuration.
// Reflected in Node.ServiceInfo.ConfigSource.
type ServiceSource int

const (
	// ServiceConfigSourceNode (1): config is set directly on the node.
	ServiceConfigSourceNode ServiceSource = 1 + iota

	// ServiceConfigSourceNodeGroup (2): config is inherited from the node's group.
	ServiceConfigSourceNodeGroup
)

// NetIfType classifies the role of a network interface.
type NetIfType string

const (
	NetIfTypeData       NetIfType = "data"
	NetIfTypeControl    NetIfType = "control"
	NetIfTypeManagement NetIfType = "management"
	NetIfTypeInbound    NetIfType = "inbound"
	NetIfTypeOutbound   NetIfType = "outbound"
)

// LocationTiers defines the ordered hierarchy of physical location granularity.
// The field "Network.Tiers" contains a subset of these values.
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
