// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

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
	Interfaces  *Interfaces
	Ipmi        *Ipmi
	Management  *Management
	DeviceInfo  *DeviceInfo
	DeviceParam *DeviceParam
}

type Ipmi struct {
	Verify   bool
	Ip       string
	Username string
	Key      string
}

type Management struct {
	Verify   bool
	Hostname string
	Ip       string
}

type DeviceInfo struct {
	Verify       bool
	SerialNumber string
	Vendor       string
	Model        string
}

type DeviceParam struct {
	CpuCore int64
	Memory  int64
	Disk    int64
}

type Interfaces struct {
	Verify     bool
	Interfaces map[string]*Interface
}

type Interface struct {
	Ip   string
	Tags []NetIfType
}
