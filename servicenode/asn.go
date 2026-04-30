// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v26/common"
	"asn.amiasys.com/asn-service-api/v26/log"
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
	// Call once per name in Init().
	InitDocDB(name string) (commonapi.DocDBHandler, error)

	// InitTSDB returns a connected time-series database handle.
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

	// -------------------------------------------------------------------------
	// Controller Communication
	// -------------------------------------------------------------------------

	// SendMessageToController sends a fire-and-forget upcall to the service controller,
	// handled by ASNServiceController.HandleMessageFromNode().
	// No response is returned. To receive a reply, the controller must initiate a
	// separate SendServiceOpsToNode() call.
	SendMessageToController(messageType, payload string) error

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
}
