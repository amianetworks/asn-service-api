// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
	"asn.amiasys.com/asn-service-api/v25/log"
)

// ASNServiceNode contains the APIs provided by ASN Service Node.
//
// 1. Initialization
// 2. Node Info
// 3. SendMessageToController
type ASNServiceNode interface {
	/*
		Initialization and Resource Allocation
	*/

	// InitLogger returns a logger dedicated to the service.
	//
	// ASN Framework manages logging for all services, and the default log files are <servicename>-*.log.
	// SHOULD ONLY call once. Further calls will get an error.
	InitLogger() (*log.Logger, error)

	// InitDocDB returns a doc DB handle.
	//
	// The DB is connected and ready for use through the DocDBHandler upon return.
	// SHOULD ONLY call once for each name. Further calls will get an error.
	InitDocDB(name string) (commonapi.DocDBHandler, error)

	// InitTSDB returns a doc DB handle.
	//
	// The DB is connected and ready for use through the TSDBHandler upon return.
	// SHOULD ONLY call once for each name. Further calls will get an error.
	InitTSDB(name string) (commonapi.TSDBHandler, error)

	// Placeholder for Locker, in case it's necessary.
	// Placeholder for IAM, in case it's necessary.

	/*
		Node Info
	*/

	// GetNodeType returns the service node's type.
	GetNodeType() commonapi.NodeType

	// GetNodeInfo returns the service node's info.
	GetNodeInfo() *NodeInfo

	/*
		SendMessageToController
	*/

	// SendMessageToController
	//
	// Service Node may send a formated message to its controller, which may handle the message by
	// implementing HandleMessageFromNode(). NO DIRECT RESPONSE to the message should be expected.
	SendMessageToController(messageType, payload string) error

	// SendMessageToService sends messages to another service on this service node.
	//
	// If the target service exists on this service node, its ReceiveMessageFromService function will be called.
	// If the target service does not exist on this service node, serviceExists will be false.
	//
	// A set of messageType, payload and error should be pre-negotiated between these two services.
	// Any returned values will be forwarded back to this function directly.
	SendMessageToService(serviceName, messageType, payload string) (
		serviceExists bool, responseMessageType, responsePayload string, responseErr error)
}
