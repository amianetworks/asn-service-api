// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	commonapi "github.com/amianetworks/asn-service-api/v25/common"
)

// Service Controller API
// Service Controllers need to implement interfaces below to be loaded and started.
// Please ready

// ASN Service interface is implemented  provides the service's API for the ASN Controller usage,
// will be implemented by service and used by ASN controller.
type AService interface {
	// Initiatialize the Service
	// Before being initialized, Service should have only provide its CLI. TODO!!!
	Init()

	// Service Controller need to handle up calls from Service Nodes if needed.
	// This could be implemented by simply ignoring the msg.
	HandleMsgfromServiceNode(serviceNodeId string, msg []byte) error

	// FIXME: please shorten the names used.
	// Get Service's CLI commands to integrate them in ASN CLI.
	GetCLICommands(
		ApplyCLIOps func(opScope, opScopeList []string, opCmd, opParams string),
	) []*cobra.Command

	// FIXME: why staticPath is needed?
	//
	GetWebHandler(staticPath string) func(group *gin.RouterGroup) error

	// Finish the service then it could be unloaded.
	Finish()

	// FIXME:
	// GetVersion() is no longer used. ASN Controller should read the version from the .so file.
	//
}

// ServiceStatus Service status struct, this is a MUST-have!
// ServiceStatus.Enabled indicates the service state from the asn.controller's view
// ServiceStatus.Extra is an option for the service providing extra status
type ServiceStatus struct {
	Enabled bool
	Extra   []byte
}

// ASNService interface provides the service's API for the ASN Controller usage,
// will be implemented by service and used by ASN controller.
type ASNService interface {
	// GetVersion returns the version of the service.
	GetVersion() commonapi.Version

	// ReceivedMetadataFromServiceNode processes the received metadata from the service in the service node.
	ReceivedMetadataFromServiceNode(serviceNodeId string, metadata []byte) error

	// Finish the service when ASN Controller finishes work.
	Finish()
}

type ASNServiceAPIs interface {
	// GetCLICommands returns the cobra CLI commands used for this service controller.
	// These will be mounted to ASN Controller's CLI under this service.
	GetCLICommands(
		sendServiceApplyOpsCmdToNetwork func(opCmd, opParams string),
		sendServiceApplyOpsCmdToGroup func(group string, opCmd, opParams string),
		sendServiceApplyOpsCmdToNodes func(nodes []string, opCmd, opParams string),
	) []*cobra.Command // no need to include start/stop/show/status, only include ops

	// MountWebHandler returns a function to mount the web handler got this service controller.
	MountWebHandler(staticPath string) func(group *gin.RouterGroup) error
}
