// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// Service Controller API
// Service Controllers need to implement interfaces below to be loaded and started.
// Please ready

// ServiceStatus Service status struct, this is a MUST-have!
// ServiceStatus.Enabled indicates the service state from the asn.controller's view
// ServiceStatus.Extra is an option for the service providing extra status
type ServiceStatus struct {
	Enabled bool
	Extra   []byte
}

// ASNServiceController interface is implemented  provides the service's API for the ASN Controller usage,
// will be implemented by service and used by ASN controller.
type ASNServiceController interface {
	// Init initializes the Service.
	// Before being initialized, Service should have only provided its CLI.
	Init(asnController ASNController) error

	// HandleMessageFromServiceNode handles up calls from Service Nodes if needed.
	// This could be implemented by simply ignoring the message.
	HandleMessageFromServiceNode(serviceNodeId, message string) error

	// GetCLICommands returns the Service's CLI commands to integrate them in ASN CLI.
	// This function should be ready BEFORE Init().
	GetCLICommands(applyCLIOps func(opScope int, opScopeList []string, opCmd, opParams string) error) []*cobra.Command

	// GetWebHandler returns a function to mount the web handler got this service controller.
	GetWebHandler(staticPath string) func(group *gin.RouterGroup) error

	// Finish the service then it could be unloaded.
	Finish()
}
