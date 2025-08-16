// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	commonapi "asn.amiasys.com/asn-service-api/v25/common"
)

// ASN Service Controller API
//
// ASN is a distributed framework for clustered services.
// An ASN Controller is the centralized control plane, which manages ASN Service Node(s).
//
// An distributed ASN Service is managed by the ASN framework and controlled by A Service Controller.
// A Service Controller needs to implement the following interfaces to be loaded and started.
//
// The ASN framework uses Service Controller as a general purpose term.
// A service may use "manager", "master", or "controller" based on its implemented role.
// For example, SWAN Manager is indeed implemented as a Server Controller for the Service "SWAN".

// ASNServiceController is the interface to be implemented by a Service Controller.
// ASN Framework will call these functions to manage the lifetime of the service.
type ASNServiceController interface {
	// GetVersion returns the service controller's version.
	GetVersion() commonapi.Version

	// Init initializes the Service.
	// Before being initialized, Service should have only provided its CLI commands, which
	// don't need to be runnable until Init() is called.
	Init(asnc ASNController) error

	// Start starts the service controller with the given config.
	// Before being started, the ASN controller will not call HandleMessageFromNode or GetMetrics.
	Start(config []byte) error

	// HandleMessageFromNode handles up calls from Service Nodes if needed.
	// If this functionality is not needed, a service's implementation may simply
	// ignore the message and return an error.
	HandleMessageFromNode(nodeID, messageType, message string) error

	// GetMetrics provides a way for the service to return a set of metrics under a network to the ASN controller.
	// The service can determine these metrics itself. Keys and values are not limited.
	// However, keep in mind that these metrics are for stat display purposes, so please design accordingly.
	// The ASN controller does not parse the metrics. It returns the metrics directly to the front-end.
	GetMetrics(networkID string) (map[string]string, error)

	// GetCLICommands returns the Service's CLI commands to integrate them in ASN CLI.
	// This function must be callable BEFORE Init().
	GetCLICommands(
		applyCLIOps func(serviceScope commonapi.ServiceScope, serviceScopeList []string, opCmd, opParams string) error,
	) []*cobra.Command

	// GetWebHandler returns a function to mount the web handler got this service controller.
	GetWebHandler(staticPath string) func(group *gin.RouterGroup) error

	// Stop stops the service controller.
	Stop() error

	// Finish the service, then it could be unloaded.
	Finish()
}
