// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
)

// StaticResource groups pre-init metadata and hooks that must be available before Init():
// - Version reporting
type StaticResource interface {
	// Version returns the service's version.
	// Safe to call before Init().
	Version() commonapi.Version
}

// ASNService interface provides the service's API for the ASN Service Node usage,
// will be implemented by service and used by ASN service node.
type ASNService interface {
	// StaticResource returns the pre-init resources (version).
	// Must be callable before Init().
	StaticResource() StaticResource

	// Init initializes the service
	Init(asnServiceNode ASNServiceNode) error

	// Start starts the service with the configuration.
	//
	// IMPORTANT: If this service wishes to be auto-started by ASN,
	// DO NOT rely on startResponse to report to the controller, as it will NOT be returned in all cases.
	// Instead, use SendMessageToController to communicate with the controller.
	//
	// However, if auto-start is not needed, then it is safe for startResponse to reach the controller.
	//
	// Parameters:
	// - Config: Configurations used to start the service.
	//   The service MUST refresh any stored configurations from last Start().
	// - runtimeErrChan: The service may report its runtime error anytime so that the framework may handle
	//   those errors properly. THUS, service must distinguish its service "internal errors" from fatal errors,
	//   only the latter should be reported to the framework through this channel.
	Start(config string) (runtimeErrChan <-chan error, err error)

	// ApplyServiceOps applies the service operations to the service.
	//
	// This function may be simultaneously called for multiple times, so the service MUST protect its internal
	// resources properly. Service operations will not directly change the service's lifetime state. But running
	// into a fatal error will eventually lead to a state change.
	//
	// The caller will handle timeout of the call, so the service should return promptly.
	//
	// Any returned values will be forwarded to the origibal caller, Service Controller.
	// Any errors which are not service internally handable should be reported through the runtimeErrChan.
	//
	// PleaseCarefully use the returns to be compitable with the framework design. THANKS!
	ApplyServiceOps(opCmd, opParams string) (resp string, err error)

	// Stop stops the service.
	//
	// Any error returned by Stop() may trigger state change of the service.
	// Should be idempotent and return promptly.
	Stop() error

	// Finish closes the service so it can be unloaded.
	Finish()
}
