// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

// StaticResource groups pre-init metadata and hooks that must be available before Init():
// - Version reporting
type StaticResource interface {
	// Version returns the service's version.
	// Safe to call before Init().
	Version() commonapi.Version

	// SharedData returns the shared data's keys that the service provides.
	//
	// In some cases, services need to query another service for values.
	// If the service you are implementing is a data PROVIDER, i.e., another service will ask you for data,
	// you should implement this function.
	// Otherwise, you can ignore this function by returning (nil, nil).
	//
	// `aggregated` are the data's keys that can be queried.
	// `subscribable` are the data's keys that can be subscribed to.
	SharedData() (aggregated, subscribable []string)
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
	//   those errors properly. THUS, the service must distinguish its service "internal errors" from fatal errors,
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
	// Any returned values will be forwarded to the original caller, Service Controller.
	// Any errors that are not service internally handleable should be reported through the runtimeErrChan.
	//
	// Please carefully use the returns to be compatible with the framework design. THANKS!
	ApplyServiceOps(opCmd, opParams string) (resp string, err error)

	// AddConfigOps add config operations to the service
	//
	// If return error for any reason, the service will be set as malfunction
	AddConfigOps(configParams []string) error

	// UpdateConfigOp updates a config operation
	//
	// If return error for any reason, the service will be set as malfunction
	UpdateConfigOp(oldConfigParam, newConfigParam string) error

	// DeleteConfigOps deletes config operations from the service
	//
	// If return error for any reason, the service will be set as malfunction
	DeleteConfigOps(configParams []string) error

	// OnQuerySharedData returns the shared data's value of the service based on the given keys.
	//
	// This function works in pairs with `ASNServiceNode.QueryServiceSharedData`.
	//
	// In some cases, services need to query another service for data.
	// If the service you are implementing is a data PROVIDER, i.e., another service will ask you for data,
	// you should implement this function.
	// Otherwise, you can ignore this function by returning ErrKeyNotFound.
	//
	// The keys should be negotiated between services beforehand so that other services know what you are providing.
	//
	// If one of the given keys does not exist, ErrKeyNotFound should be returned for the caller to know.
	//
	// Please carefully use the returns to be compatible with the framework design. THANKS!
	OnQuerySharedData(keys []string) (values map[string]any, err error)

	// OnSubscribeSharedData returns a channel for the shared data's value based on the given keys.
	//
	// This function works in pairs with `ASNServiceNode.SubscribeServiceSharedData`.
	//
	// In some cases, services need to subscribe to data from another service.
	// If the service you are implementing is a data PROVIDER, i.e., another service will ask you for data,
	// you should implement this function.
	// Otherwise, you can ignore this function by returning ErrKeyNotFound.
	//
	// The keys should be negotiated between services beforehand so that other services know what you are providing.
	//
	// If one of the given keys does not exist, ErrKeyNotFound should be returned for the caller to know.
	// For each key, after all values are dumped, you should ALWAYS close the returned channel to let the listener know.
	//
	// Please carefully use the returns to be compatible with the framework design. THANKS!
	OnSubscribeSharedData(keys []string) (values map[string]<-chan any, err error)

	// Stop stops the service.
	//
	// Any error returned by Stop() may trigger state change of the service.
	// Should be idempotent and return promptly.
	Stop() error

	// Finish closes the service so it can be unloaded.
	Finish()
}
