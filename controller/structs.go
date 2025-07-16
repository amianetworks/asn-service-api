// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import "time"

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

type NodeStateChange struct {
	Timestamp    time.Time
	NodeID       string
	NodeState    int
	ServiceState int
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
	State       int      // Node state, refer Service Node State enum
	NetworkID   string   // Network ID
	NodeGroupID string   // Node Group ID, if in node group
	Description string
	Meta        string // meta data used by service

	Location   *Location // Node physical location
	Ipmi       *Ipmi
	Management *Management
	Info       *Info
	SystemInfo *SystemInfo
	Interfaces map[string]*Interface

	ServiceInfo *ServiceInfo
	Stats       *NodeStats
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

type SystemInfo struct {
	MachineID string
	CpuCore   int64
	Memory    int64
	Disk      int64
}

type Interface struct {
	Ip   string
	Tags []string
}

type ServiceInfo struct {
	State        int
	UsedConfig   string // exists if Config Source is Node, otherwise is empty, can get the config in node group
	ConfigSource int
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

type NodeLink struct {
	ID          string // uuid
	Description string // the name of the link, can be empty
	Bandwidth   int64  // the bandwidth between two nodes, the up speed equals to the down speed

	From, To *NodeLinkNode
}

type NodeLinkNode struct {
	NodeID    string
	Interface string
}

type NodeGroup struct {
	ID          string
	Name        string
	Description string
	Meta        string // meta data used by service
	Nodes       []string
	Config      []byte
}
