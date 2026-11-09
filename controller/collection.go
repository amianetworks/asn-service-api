// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"time"

	"asn.amiasys.com/asn-service-api/v27/iam"
)

// Framework data collection: every node samples its host and network
// interfaces into the "asn" items below, receives syslog, tails files with
// log parsers, and may collect sFlow. A service reads the data of the nodes
// it runs on; a super service of any node.

// Host metric items.
const (
	HostItem     = "asn.host"      // per node
	HostDiskItem = "asn.host_disk" // per node and mount (tags Path, Device)
	HostGPUItem  = "asn.host_gpu"  // per node and GPU (tags Index, Name)
	HostNetItem  = "asn.host_net"  // per node and interface (tag Iface)
	SFlowIfItem  = "asn.sflow_if"  // per node, agent and ifIndex (tags Agent, IfIndex)
	SFlowItem    = "asn.sflow"     // per node and agent (tag Agent)
)

// Series aggregates.
const (
	HostSeriesMean = "mean"
	HostSeriesMax  = "max"
	HostSeriesMin  = "min"
	HostSeriesLast = "last"
)

// HostSample is a node's latest sample of an item, one per series.
type HostSample struct {
	NodeID string
	Tags   map[string]string // the series' tags, none for asn.host
	Time   time.Time
	Fields map[string]float64
}

// HostSeriesBucket is one window of a series.
type HostSeriesBucket struct {
	Time   time.Time
	Values map[string]float64
}

// HostSeries is one series over a range.
type HostSeries struct {
	NodeID  string
	Tags    map[string]string
	Buckets []HostSeriesBucket
}

// Log items.
const (
	LogItemSyslog       = "asn.syslog"
	LogItemFile         = "asn.log"
	LogItemNGFWIPS      = "asn.ngfw_ips"
	LogItemNGFWBehavior = "asn.ngfw_behavior"
)

// LogQuery selects log records of one item.
type LogQuery struct {
	Item              string
	RootID            string
	NetworkID         string
	IncludeSubnetwork bool
	NodeIDs           []string
	Start, Stop       time.Time // both required
	Severities        []string
	Facilities        []string
	AppName           string
	Keyword           string
	ParserID          string
}

// LogRecord is one log record.
type LogRecord struct {
	Time   time.Time
	NodeID string
	Tags   map[string]string
	Fields map[string]any
}

// LogParser tails files on the nodes carrying one of LabelIDs and parses
// their lines with grok Patterns.
type LogParser struct {
	ID         string
	Name       string
	Path       string
	Patterns   []string
	LabelIDs   []string
	Enable     bool
	Builtin    bool
	CreateTime time.Time
	UpdateTime time.Time
}

// LogCollection is a node's syslog sources: Local (nil: on) and the network
// receiver (off by default).
type LogCollection struct {
	Local          *bool
	RemoteEnable   bool
	RemoteProtocol string // udp or tcp
	RemotePort     int
}

// SFlowSettings is a node's sFlow collector.
type SFlowSettings struct {
	Enable bool
	Port   int // UDP; 6343 when 0
}

// Capture backends.
const (
	CaptureBackendAFPacket = "afpacket" // a packet socket: what the host receives and sends
	CaptureBackendDCube    = "dcube"    // the dcube packetcapture cube at XDP: before XDP programs drop
)

// CaptureSettings is a node's packet capture: a window of Window every
// Interval, at most MaxPackets headers of SnapLen bytes per window and
// interface, sampled evenly over the window. Zero values take the defaults
// (afpacket, 1 s, 200 ms, 30000, 128). It runs only while Enable is set and a
// service installed on the node declares PacketCapture in its catalog.
type CaptureSettings struct {
	Enable      bool
	Backend     string
	Interfaces  []string
	Interval    time.Duration // 1 s - 1 h
	Window      time.Duration // 10 ms - Interval
	MaxPackets  int           // 1 - 100000
	SnapLen     int           // 0 - 256
	Filter      string        // a subset of the pcap filter syntax: docs/DESIGN.md §6.3
	IngressOnly bool          // dcube: the ingress cube only

	Active    bool     // read only: enabled and demanded
	Demanders []string // read only: the services on the node that declare the need
}

// CollectionAPI reads framework collected data and sets the collection of
// nodes. Errors common to the methods: asn.node.not_found;
// asn.service.not_super (a node the caller is not on);
// shared.request.param_invalid; shared.system.unavailable.
type CollectionAPI interface {
	// HostMetrics returns the latest samples of item for nodes.
	HostMetrics(nodeIDs []string, item string) ([]HostSample, error)
	// HostMetricSeries aggregates fields of item with fn over [start, stop)
	// for nodes, in windows of the returned length (a multiple of the
	// sampling interval, at most 1000 per series).
	HostMetricSeries(nodeIDs []string, item string, fields []string, fn string, start, stop time.Time) ([]HostSeries, time.Duration, error)

	// ListLogs lists the log records q selects, newest first.
	ListLogs(q LogQuery, page iam.Page) (records []*LogRecord, total int, err error)
	// ListLogParsers lists the log parsers (super services only write them).
	ListLogParsers() ([]*LogParser, error)
	// SaveLogParser creates (empty ID) or replaces a parser; super only.
	SaveLogParser(p LogParser) (*LogParser, error)
	// DeleteLogParser deletes a parser that is not built in; super only.
	DeleteLogParser(parserID string) error
	// GetNodeLogCollection and SetNodeLogCollection read and set a node's
	// syslog sources.
	GetNodeLogCollection(nodeID string) (*LogCollection, error)
	SetNodeLogCollection(nodeID string, c LogCollection) error

	// GetNodeSFlow and SetNodeSFlow read and set a node's sFlow collector.
	GetNodeSFlow(nodeID string) (*SFlowSettings, error)
	SetNodeSFlow(nodeID string, s SFlowSettings) error

	// GetNodeCapture and SetNodeCapture read and set a node's packet
	// capture; Active and Demanders are ignored on set.
	GetNodeCapture(nodeID string) (*CaptureSettings, error)
	SetNodeCapture(nodeID string, c CaptureSettings) error
}
