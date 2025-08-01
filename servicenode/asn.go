// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
	"asn.amiasys.com/asn-service-api/v25/log"
)

// ASNServiceNode contains the APIs provided by ASN Service Node.
//
// 1. Initialization
// 2. SendMessageToController
// 3. ??
//
// ASNServiceNode provided by ASN Service Node for Service uses
type ASNServiceNode interface {
	/*
		Initialization
	*/

	// InitLogger returns a logger dedicated to the service.
	// ASN Framework manages logging for all services, and the default log files are <servicename>-*.log
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

	// Placeholder for Locker, in case it's necessary.
	// Placeholder for IAM, in case it's necessary.

	/*
		Node Info
	*/
	GetNodeInfo() *Node

	/*
		SendMessageToController
	*/

	// SendMessageToController
	// Service Node may send a formated message to its controller, which must have implemented
	// HandleMessageFromServiceNode() to handle the received message.
	SendMessageToController(message string) error
}
