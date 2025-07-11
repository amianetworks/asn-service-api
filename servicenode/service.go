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
	Init()

	// Start the service with the configuration.
	//
	// input parameters:
	// 1. config:
	// 	 the configuration of the service. Service MUST update this configuration to the local file.
	// 	 When DumpConfiguration called, service need to return the current configuration to the framework
	// 2. the return value to channel indicate service state:
	// 	 - if error is nil, the service node will assign the state CONFIGURED to the service
	// 	 - if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
	// 	   after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (20s) to process the initialization,
	//   		if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	Start(clusterConfig, instanceConfig []byte) (errChan chan error, err error)

	// ApplyServiceOps applies the service operations to the service.
	// Service operations will not change the service status (enabled/disabled),
	// but will do some runtime operations such as: insert/delete/getXXX/setXXX
	// Apply the configuration to the service, this method will be called under a go routine, the return value to channel indicate service state:
	// if error is nil, the service node will remain the previous state (CONFIGURED/INITIALIZED)
	// if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
	// 	 after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (20s) to process the initialization,
	// 		 	if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	ApplyServiceOps(opCmd, opParams string, response chan *commonapi.Response)

	// Stop the service with the configuration.
	//
	// input parameters:
	// 1. the return value to channel indicate service state:
	// 	 - if error is nil, the service node will assign the state INITIALIZED to the service
	// 	 - if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
	// 	   after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service
	//
	// Caution: the service node will have a timeout context (20s) to process the initialization,
	// 		 	if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	Stop() error

	// Finish is the last call before the service node's termination. Do the necessary clean up here.
	Finish()
}
