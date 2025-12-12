// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"net/http"

	"github.com/spf13/cobra"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

// ASN Service Controller API
//
// ASN is a distributed framework for clustered services.
// An ASN Controller is the centralized plane of management and control for ASN Service Node(s).
// A Service Controller implements the following interfaces to be loaded and started.
//
// The ASN framework uses “Service Controller” as a general term.
// A service may use "manager", "master", or "controller" based on its implemented role(s).

// StaticResource groups pre-init metadata and hooks that must be available before Init():
// - Version reporting
// - CLI command registration
// - Web handler mounting
type StaticResource interface {
	// Version returns the service controller's version.
	// Safe to call before Init().
	Version() commonapi.Version

	// CLICommands returns the service's cobra commands to integrate into the ASN CLI.
	// Safe to call before Init().
	CLICommands(
		applyCLIOps func(serviceScope commonapi.ServiceScope, serviceScopeList []string, opCmd, opParams string) error,
	) []*cobra.Command

	// WebHandler returns an http.Handler for this service controller.
	// Safe to call before Init().
	// Framework is responsible for mounting the returned handler under the appropriate route/prefix.
	// Implementations may internally use Gin/Chi/etc. but must expose a standard http.Handler.
	WebHandler(staticPath string) (http.Handler, error)
}

// ASNServiceController is the interface to be implemented by a Service Controller.
// The ASN framework calls these functions to manage the service's lifetime and behaviors.
type ASNServiceController interface {
	// StaticResource returns the pre-init resources (version, CLI, and web).
	// Must be callable before Init().
	StaticResource() StaticResource

	// Init initializes the service.
	//
	// After Init, CLI commands should be runnable.
	Init(asnController ASNController) error

	// Start starts the service controller with the given configuration.
	//
	// Config format is marshall/unmarshal by the service, so its format doesn't matter.
	// Must return quickly (non-blocking); long-running work should run in background goroutines.
	// It may be called multiple times, the configuration from the last call must be effective by the end.
	Start(config string) error

	// HandleMessageFromNode handles up-calls from service nodes if needed.
	HandleMessageFromNode(nodeID, messageType, payload string) error

	// GetMetrics returns display-only metrics. Values must be marshal-able as JSON.
	GetMetrics(networkID string) (map[string]string, error)

	// Stop gracefully stops the service controller.
	//
	// Should be idempotent and return promptly.
	Stop() error

	// Finish closes the service so it can be unloaded.
	Finish()
}
