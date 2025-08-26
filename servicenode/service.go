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

	// Init
	//
	// This function initializes the service
	Init(asnServiceNode ASNServiceNode) error

	// Start
	//
	// This function starts the service with the configuration.
	//
	// IMPORTANT: If this service wishes to be auto-started by ASN,
	// DO NOT rely on startResponse to report to the controller, as it will NOT be returned in all cases.
	// Instead, use SendMessageToController to communicate with the controller.
	//
	// However, if auto-start is not needed, then it is safe for startResponse to reach the controller.
	//
	// Parameters:
	//
	// 1. Config: the configuration of the service. Service MUST update this configuration to the local file.
	// When DumpConfiguration called, the service needs to return the current configuration to framework,
	//
	// 2. The return values indicate service state:
	// 	 - If err is nil, the service node will assign the state CONFIGURED to the service,
	//     send the response to the controller, and keep listening to the runtimeErrChan channel.
	// 	 - If err is NOT nil, the service node will try to init the service and reapply the config for 3 times.
	// 	   After all retries if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (10 secs by default) to process the initialization,
	// if it cannot be done within 10 secs, the service node will assign state MALFUNCTIONAL to the service.
	Start(config string) (runtimeErrChan <-chan error, err error)

	// ApplyServiceOps
	//
	// This function applies the service operations to the service.
	//
	// Service operations will not change the service status (enabled/disabled),
	// but will do some runtime operations such as: insert/delete/getXXX/setXXX
	//
	// Apply the configuration to the service, this method will be called under a go routine,
	// the return value indicate service state:
	//   - if error is nil, the service node will remain the previous state (CONFIGURED/INITIALIZED)
	//   - if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
	//     after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (10 secs by default) to process the initialization,
	// 		 	if it cannot be done within 10 secs, service node will assign the state MALFUNCTIONAL to the service
	ApplyServiceOps(opCmd, opParams string) error

	// Stop
	//
	// This function stops the service with the configuration.
	//
	// The return value to channel indicate service state:
	// 	 - If error is nil, the service node will assign the state INITIALIZED to the service
	// 	 - If error is NOT nil, the service node will try to init the service and reapply the configuration for 3 times,
	// 	   after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (10 secs by default) to process the initialization,
	// if it cannot be done within 10 secs, service node will assign the state MALFUNCTIONAL to the service
	Stop() error

	// Finish
	//
	// This function is the last call before the service node's termination. Do the necessary clean up here.
	Finish()
}
