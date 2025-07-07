// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"crypto/tls"

	commonapi "asn.amiasys.com/asn-service-api/v25/common"
	"asn.amiasys.com/asn-service-api/v25/iam"
	"asn.amiasys.com/asn-service-api/v25/log"
)

// Structs used between asn.controller and service.controller.

type NetworkBasicInfo struct {
	ID          string
	Name        string
	ParentID    string
	Description string
	Tiers       []string
}

// Network is the structure for a network.
type Network struct {
	NetworkBasicInfo
	Location *Location
	Networks []*Network

	Stats *NetworkStats
}

type NetworkStats struct {
	ReceivedBits    uint64
	SentBits        uint64
	AsnReceivedBits uint64
	AsnBlockedBits  uint64
	AsnReceivedPkts uint64
	AsnBlockedPkts  uint64
	Timestamp       string
}

type Location struct {
	Description string
	Tier        string
	Address     string
	Coordinates *Coordinates
}

type Coordinates struct {
	Latitude  float32
	Longitude float32
	Altitude  float32
}

type NodeType string

const (
	NodeTypeRouter      NodeType = "router"
	NodeTypeSwitch      NodeType = "switch"
	NodeTypeAppliance   NodeType = "appliance"
	NodeTypeFirewall    NodeType = "firewall"
	NodeTypeLoadBalance NodeType = "lb"
	NodeTypeAccessPoint NodeType = "ap"
	NodeTypeEndPoint    NodeType = "ep"
	NodeTypeServer      NodeType = "server"
)

// Node is the structure for a node.
type Node struct {
	ID          string   // Node ID
	Name        string   // Device display name
	Type        NodeType // Node Type
	NetworkID   string   // Network ID
	Managed     bool
	Description string

	Location   *Location // Node physical location
	Ipmi       *Ipmi
	Management *Management
	Info       *Info
	Interfaces map[string]*Interface

	Stats *NodeStats
}

type Ipmi struct {
	Ip       string
	Username string
	Key      string
}

type Management struct {
	Hostname string
	Ip       string
}

type Info struct {
	Vendor       string
	Model        string
	SerialNumber string
}

type Interface struct {
	Ip   string
	Tags []string
}

type NodeStats struct {
	Rx                 uint64
	Tx                 uint64
	AsnReceivedPackets uint64
	AsnBlockedPackets  uint64
	AsnReceivedBits    uint64
	AsnBlockedBits     uint64
	CpuUsage           float32
}

type NetworkLink struct {
	ID          string // uuid
	Description string // the name of the link, can be empty
	Bandwidth   int64  // the bandwidth between two nodes, the up speed equals to the down speed

	From, To *NetworkLinkNode
}

type NetworkLinkNode struct {
	NetworkID string
	Interface string
}

type NodeInternalLink struct {
	ID          string // uuid
	Description string // the name of the link, can be empty
	Bandwidth   int64  // the bandwidth between two nodes, the up speed equals to the down speed

	From, To *NodeLinkNode
}

type NodeExternalLink struct {
	ID          string // uuid
	Description string // the name of the link, can be empty
	Bandwidth   int64  // the bandwidth between two nodes, the up speed equals to the down speed

	From *NodeLinkNode
	To   *Node
}

type NodeLinkNode struct {
	NodeID    string
	Interface string
}

type NodeGroup struct {
	ID            string
	Name          string
	Description   string
	Nodes         []*Node
	ClusterConfig []byte
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

	// AddServiceToNode adds a .so file to an existing node, and inits this service on that node.
	// NOTE: CURRENTLY UNSUPPORTED!!!
	AddServiceToNode(nodeID string) error

	// DeleteServiceFromNode removes this service from an existing node.
	// NOTE: CURRENTLY UNSUPPORTED!!!
	DeleteServiceFromNode(nodeID string) error

	// StartService starts service on specified Service Nodes.
	StartService(serviceScope int, serviceScopeList []string, instanceConfig []byte) error

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

	// SaveClusterConfigOfNodeGroup saves the cluster setting for a node group.
	SaveClusterConfigOfNodeGroup(nodeGroupID string, config []byte) error

	// SaveClusterConfigOfNode saves the cluster setting for a node.
	SaveClusterConfigOfNode(nodeId string, config []byte) error

	/*
		Networks, Nodes and Links
	*/

	// GetRootNetworks returns all the root networks
	GetRootNetworks() ([]*NetworkBasicInfo, error)

	// GetNetworkByID returns a network and all its subnetworks and links.
	// - locationTiers filter the networks with the given location tiers.
	// - networkTiers filter the networks with the given network tiers.
	GetNetworkByID(
		networkID string,
		locationTiers, networkTiers []string,
		includeStats bool,
	) (*Network, []*NetworkLink, error)

	// GetNodesOfNetwork returns all nodes of a network, and its internal and external links.
	// - Internal links connect the nodes within the same network, and it is included in the returned nodes array.
	//   So, only IDs are returned in this case.
	// - External links connect nodes in this network with nodes outside of this network.
	//   So, the "To" node is not included in the returned nodes array, but in the "NodeExternalLink" structure.
	GetNodesOfNetwork(networkID string, includeStats bool) ([]*Node, []*NodeInternalLink, []*NodeExternalLink, error)

	// CreateNode creates a node under a given network.
	// Note that this is only supported when ASN does not strictly verify the network topology.
	// For now, a certificate is returned for the node to register to ASN Controller.
	CreateNode(networkID, nodeName string, nodeType NodeType) (*tls.Certificate, error)

	/*
		Node Group
	*/

	// CreateNodeGroup creates a node group for this service.
	CreateNodeGroup(name, description string) error

	// ListNodeGroups returns all node groups under this service.
	ListNodeGroups() ([]*NodeGroup, error)

	// DeleteNodeGroup removes a node group under this service.
	DeleteNodeGroup(id string) error

	// AddNodesToNodeGroup adds the specified nodes to the provided node group identified by its ID.
	AddNodesToNodeGroup(nodeGroupID string, nodeIDs []string) error

	// RemoveNodeFromNodeGroup removes the specified nodes from the provided node group identified by its ID.
	RemoveNodeFromNodeGroup(nodeGroupID string, nodeIDs []string) error
}
