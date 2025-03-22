// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// ASN Service Controller API
//
// ASN is an distributed framework for clustered services.
// An ASN Controller is the centralized control plane, which manages ASN Service Node(s).
//
// An distributed ASN Service is managed by ASN framework and controlled by A Service Controller.
// A Service Controller needs to implement the following interfaces to be loaded and started.
//
// Service Controller is used by ASN framework as a general purpose term.
// A service may use "manager", "master", or "controller" based on its implemented role.
// For example, SWAN Manager is indeed implemented as a Server Controller for the Service "SWAN".

// ASNServiceController interface is implemented by a Service Controller.
// ASN Framework will call these functions to manage the liftime of the service.
type ASNServiceController interface {
	// Init initializes the Service.
	// Before being initialized, Service should have only provided its CLI commands, which
	// don't need to be runnable unitl Init() is called.
	Init(asnc ASNController) error

	// HandleMessageFromServiceNode handles up calls from Service Nodes if needed.
	// If this functionality is not needed, a service's implementation may simply
	// ignore the message and return an error.
	HandleMessageFromServiceNode(serviceNodeId string, message []byte) error

	// GetCLICommands returns the Service's CLI commands to integrate them in ASN CLI.
	// This function must be callable BEFORE Init().
	GetCLICommands(applyCLIOps func(opScope int, opScopeList []string, opCmd, opParams string) error) []*cobra.Command

	// GetWebHandler returns a function to mount the web handler got this service controller.
	GetWebHandler(staticPath string) func(group *gin.RouterGroup) error

	// Finish the service then it could be unloaded.
	Finish()
}
