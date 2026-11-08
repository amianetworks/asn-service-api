// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import "time"

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

type NodeInfo struct {
	Mode        NodeMode
	Interfaces  map[string]*Interface
	Ipmi        *Ipmi
	Management  *Management
	DeviceInfo  *DeviceInfo
	DeviceParam *DeviceParam
	// Host is what the node reported about its host at its last
	// registration; nil for a node that reports none.
	Host *HostInfo
}

// HostInfo is a node's host information, reported at registration.
type HostInfo struct {
	Hostname        string
	OS              string
	Platform        string
	PlatformVersion string
	KernelVersion   string
	Arch            string
	BootTime        time.Time
	CPUModel        string
	CPUCores        int    // physical
	CPUThreads      int    // logical
	MemoryTotal     uint64 // bytes
	DiskTotal       uint64 // bytes
	Interfaces      []HostInterface
	ReportedAt      time.Time
}

// HostInterface is one interface of a node's host information.
type HostInterface struct {
	Name  string
	MAC   string
	Addrs []string // CIDR
	MTU   int
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

type DeviceInfo struct {
	SerialNumber string
	Vendor       string
	Model        string
}

type DeviceParam struct {
	CpuCore int64
	Memory  int64
	Disk    int64
}

type Interface struct {
	Ip   string
	Tags []NetIfType
}
