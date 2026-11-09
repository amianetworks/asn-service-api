// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	"net/netip"
	"time"

	commonapi "asn.amiasys.com/asn-service-api/v27/common"
	"asn.amiasys.com/asn-service-api/v27/log"
)

// ASNServiceNode is the framework-provided handle passed to ASNService.Init().
// All methods are goroutine-safe after Init() unless stated otherwise.
//
// Functional areas:
//  1. Resource initialization (Init* — one-shot, call in Init())
//  2. Node information
//  3. Controller communication
//  4. Cross-service data access
type ASNServiceNode interface {

	// -------------------------------------------------------------------------
	// Resource Initialization
	// Must be called in Init(). All are one-shot; a second call returns an error.
	// -------------------------------------------------------------------------

	// InitLogger returns the logger for this service.
	// Call once in Init(). Default log files are named <servicename>-*.log.
	InitLogger() (*log.Logger, error)

	// InitDocDB returns a connected document database handle.
	// name scopes the DB instance; multiple names yield independent handles.
	// The framework prefixes it with the service name: an empty name opens
	// <service_name> itself, any other opens <service_name>_<name>.
	// Returns an error if docdb is not enabled on this node.
	// Call once per name in Init().
	InitDocDB(name string) (commonapi.DocDBHandler, error)

	// InitTSDB returns a connected time-series database handle.
	// name scopes the DB instance; multiple names yield independent handles.
	// The framework prefixes it with the service name: an empty name opens
	// <service_name> itself, any other opens <service_name>_<name>.
	// Returns an error if tsdb is not enabled on this node.
	// Call once per name in Init().
	InitTSDB(name string) (commonapi.TSDBHandler, error)

	// -------------------------------------------------------------------------
	// Node Information
	// -------------------------------------------------------------------------

	// GetNodeType returns the hardware/role type of this node.
	GetNodeType() commonapi.NodeType

	// GetNodeInfo returns this node's hardware info, management addresses,
	// and the active ConfigOps string list.
	GetNodeInfo() *NodeInfo

	// GetSlaveNodes returns all slave nodes configured for this master node AND
	// have reported loading this service in their registration,
	// each annotated with its current connection state.
	// Returns nil if this node is not in cluster mode or has no slaves configured.
	GetSlaveNodes() []*SlaveNodeInfo

	// -------------------------------------------------------------------------
	// Controller Communication
	// -------------------------------------------------------------------------

	// SendMessageToController sends a fire-and-forget upcall to the service controller,
	// handled by ASNServiceController.HandleMessageFromNode().
	// No response is returned. To receive a reply, the controller must initiate a
	// separate SendServiceOpsToNode() call.
	SendMessageToController(messageType string, payload []byte) error

	// -------------------------------------------------------------------------
	// Slave Communication
	// Only meaningful in cluster (master) mode. All methods return
	// ErrNotMasterNode in standalone mode, ErrSlaveNotConnected when the named
	// slave's stream is not established, and ErrSlaveServiceNotFound when the
	// slave has not reported loading this service.
	//
	// Timeouts are enforced by the framework using the service's configured
	// Timeout value; no context is exposed to the caller.
	// -------------------------------------------------------------------------

	// SendStartToSlave sends a start command for this service to the named slave.
	// Returns only a transport-level error; the execution outcome is delivered
	// asynchronously via ClusterMasterService.OnSlaveServiceState.
	SendStartToSlave(slaveName, config string) error

	// SendStopToSlave sends a stop command for this service to the named slave.
	// Returns only a transport-level error; the execution outcome is delivered
	// asynchronously via ClusterMasterService.OnSlaveServiceState.
	SendStopToSlave(slaveName string) error

	// SendApplyOpsToSlave sends an ops command to the named slave and blocks
	// until the slave replies or the call times out.
	SendApplyOpsToSlave(slaveName, opCmd, opParams string) (resp string, err error)

	// SendAddConfigOpsToSlave sends add-config-ops to the named slave and blocks
	// until the slave acknowledges or the call times out.
	SendAddConfigOpsToSlave(slaveName string, params []string) (resp string, err error)

	// SendUpdateConfigOpToSlave sends an update-config-op to the named slave and
	// blocks until the slave acknowledges or the call times out.
	SendUpdateConfigOpToSlave(slaveName string, oldParam, newParam string) (resp string, err error)

	// SendDeleteConfigOpsToSlave sends delete-config-ops to the named slave and
	// blocks until the slave acknowledges or the call times out.
	SendDeleteConfigOpsToSlave(slaveName string, params []string) (resp string, err error)

	// -------------------------------------------------------------------------
	// Cross-Service Data Access
	// Enables data exchange between services co-located on the same node.
	// -------------------------------------------------------------------------

	// GetSharedData returns the key sets advertised by another service on this node.
	// aggregated: keys queryable via QueryServiceSharedData (pull model).
	// subscribable: keys subscribable via SubscribeServiceSharedData (push model).
	// Returns ErrServiceNotFound if the named service is not loaded.
	GetSharedData(serviceName string) (aggregated, subscribable []string, err error)

	// QueryServiceSharedData fetches current values for the given keys from another service.
	// Pairs with ASNService.OnQuerySharedData on the provider side.
	// Key names and value types must be agreed upon out-of-band between service teams.
	// Returns ErrServiceNotFound or ErrKeyNotFound on failure.
	QueryServiceSharedData(serviceName string, keys []string) (values map[string]any, err error)

	// SubscribeServiceSharedData subscribes to a stream of values for the given keys from another service.
	// Pairs with ASNService.OnSubscribeSharedData on the provider side.
	// Each returned channel is closed by the provider when the stream ends.
	// Constraint: only one active subscription per key is permitted at a time.
	// Opening a second subscription before the first channel closes is undefined behavior.
	// Returns ErrServiceNotFound or ErrKeyNotFound on failure.
	SubscribeServiceSharedData(serviceName string, keys []string) (values map[string]<-chan any, err error)

	// WriteAudit records an operation of the service on this node.
	// event.NodeID must be empty: ASN sets this node, and the record's root
	// is the node's. Queued on the node and sent to the controller; a full
	// queue fails with shared.system.unavailable.
	WriteAudit(event commonapi.AuditEvent) error

	// HostMetrics returns this node's latest framework host sample (nil
	// before the first) and the samples of the last hour, oldest first, from
	// memory. Fields are those of the controller's asn.host item.
	HostMetrics() (latest *HostSample, lastHour []HostSample)

	// SubscribeCapture returns a channel of this node's capture windows, one
	// per window and interface, and a function that ends the subscription.
	// Every subscriber gets the same read-only window: do not modify it. A
	// window that finds the channel full is lost to this subscriber
	// (counted in CaptureState), never delayed. It succeeds while capture is
	// inactive and delivers once it runs. Once the node's settings arrive, the
	// subscriptions of a service that does not declare PacketCapture end and
	// their channels close.
	//
	// Errors: asn.capture.not_declared (the catalog does not declare
	// PacketCapture); shared.request.param_invalid (buffer < 1).
	SubscribeCapture(buffer int) (<-chan *CaptureWindow, func(), error)
	// CaptureState is the state of this node's capture.
	CaptureState() CaptureState

	// SubscribeNetCounters returns a channel of this node's interface
	// counters read every interval (1 s - 1 h), and a function that ends the
	// subscription. A reading that finds the channel full is dropped.
	SubscribeNetCounters(interval time.Duration, buffer int) (<-chan []NetCounters, func(), error)
}

// PacketMeta is one captured packet: value fields only, so a window holds
// its packets in one slice.
type PacketMeta struct {
	TimeUnixNano     int64
	Length           uint32 // on the wire
	VLAN             uint16 // 0: none
	EtherType        uint16
	SrcMAC, DstMAC   [6]byte
	SrcIP, DstIP     netip.Addr // invalid when not IP
	Protocol         uint8      // IP protocol
	SrcPort, DstPort uint16     // TCP and UDP
	TCPFlags         uint8
	Ingress          bool
	HeaderOff        uint32 // the packet's bytes are Headers[HeaderOff : HeaderOff+HeaderLen]
	HeaderLen        uint32
}

// CaptureWindow is one window of one interface. SampleRate N means one
// packet in N was kept, so counts scale by N; NICPackets and NICBytes are
// exact, from the interface counters over the window.
type CaptureWindow struct {
	Interface  string
	Start      time.Time
	Window     time.Duration
	SampleRate uint32
	NICPackets uint64
	NICBytes   uint64
	Packets    []PacketMeta
	Headers    []byte
	Lost       uint64 // packets lost in the kernel or the ring in this window
}

// CaptureState is the state of a node's capture.
type CaptureState struct {
	Active  bool
	Backend string
	Error   string // the last start error; empty while running
	Windows uint64 // windows delivered to the calling service
	Lost    uint64 // windows the calling service lost to a full channel
}

// NetCounters is one reading of an interface's counters.
type NetCounters struct {
	Time                 time.Time
	Interface            string
	RxBytes, TxBytes     uint64
	RxPackets, TxPackets uint64
	RxErrors, TxErrors   uint64
	RxDrops, TxDrops     uint64
}

// HostSample is one framework host sample of this node.
type HostSample struct {
	Time   time.Time
	Fields map[string]float64
}
