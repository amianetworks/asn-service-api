// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "github.com/amianetworks/asn-service-api/v25/common"
	"github.com/amianetworks/asn-service-api/v25/log"
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

	// InitDB may be called to initialize multiple DBs with specified dbName and dbType.
	// The required DB is connected and ready for use through the dnHandler.
	InitDB(dbType string, dbName string) (commonapi.DBHandler, error)

	// Placeholder for Locker, in case it's needed.
	// Placeholder for IAM, in case it's needed.

	/*
		SendMessageToController
	*/

	// SendMessageToController
	// Service Node may send a formated message to its controller, which must have implemented
	// HandleMessageFromServiceNode() to handle the received message.
	SendMessageToController(message string) error
}
