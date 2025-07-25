// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v25/common"
)

// ASNService interface provides the service's API for the ASN Service Node usage,
// will be implemented by service and used by ASN service node.
type ASNService interface {
	// GetVersion returns the service's version.
	GetVersion() commonapi.Version

	// Init initializes the service
	Init() error

	// Start the service with the configuration.
	//
	// Parameters:
	// 1. Config:
	// 	 the configuration of the service. Service MUST update this configuration to the local file.
	// 	 When DumpConfiguration called, the service needs to return the current configuration to framework
	// 2. The return value to channel indicate service state:
	// 	 - if error is nil, the service node will assign the state CONFIGURED to the service
	// 	 - if error is NOT nil, the service node will try to init the service and reapply the configuration for 3 times,
	// 	   after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (20s) to process the initialization,
	//   		if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	Start(config []byte) (errChan <-chan error, err error)

	// UpdateConfig passes a set of configurations to the service.
	//
	// Service should compare this set of configurations with its current configurations.
	// If the two sets of configurations are the same, nothing should be done.
	// If the two sets of configurations are different, the service has two options:
	// 1. Handle the config internally and return an error if needed by returning false, nil/err.
	// 2. Request ASN to restart the service with the new config by returning true, nil.
	//
	// Caution: if an error is returned, ASN won't try to restart the service whatsoever.
	//          It will mark the service as malfunctioning and wait for the controller to handle it.
	UpdateConfig(config []byte) (restartRequired bool, err error)

	// ApplyServiceOps applies the service operations to the service.
	// Service operations will not change the service status (enabled/disabled),
	// but will do some runtime operations such as: insert/delete/getXXX/setXXX
	// Apply the configuration to the service, this method will be called under a go routine,
	// the return value to channel indicate service state:
	// if error is nil, the service node will remain the previous state (CONFIGURED/INITIALIZED)
	// if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
	// 	 after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (20s) to process the initialization,
	// 		 	if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	ApplyServiceOps(opCmd, opParams string) (response <-chan *commonapi.Response)

	// Stop the service with the configuration.
	//
	// Parameters:
	// 1. The return value to channel indicate service state:
	// 	 - If error is nil, the service node will assign the state INITIALIZED to the service
	// 	 - If error is NOT nil, the service node will try to init the service and reapply the configuration for 3 times,
	// 	   after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (20s) to process the initialization,
	// 		 	if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	Stop() error

	// Finish is the last call before the service node's termination. Do the necessary clean up here.
	Finish()
}
