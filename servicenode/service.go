// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

// StaticResource groups pre-init metadata that must be available before Init().
// All methods are re-entrant and safe to call at any time, including before Init().
type StaticResource interface {
	// Version returns the service's version. Must be a hardcoded constant.
	Version() commonapi.Version

	// SharedData declares the data keys this service provides to other services on the same node.
	// aggregated: keys available via QueryServiceSharedData (pull model).
	// subscribable: keys available via SubscribeServiceSharedData (push/stream model).
	// Return (nil, nil) if this service does not provide shared data.
	// The framework uses these declarations to validate inbound requests from consuming services.
	SharedData() (aggregated, subscribable []string)
}

// ASNService is the interface implemented by a service running on a Service Node.
// The framework calls these methods to manage the service's lifetime and behavior.
//
// Lifecycle order: StaticResource → Init → Start → [concurrent callbacks] → Stop → Finish.
type ASNService interface {
	// StaticResource returns the pre-init resources (version, shared data keys).
	// Re-entrant; must be callable before Init().
	StaticResource() StaticResource

	// Init initializes the service.
	// Called once after .so load, before Start(). Not re-entrant.
	// All Init* calls on asnServiceNode must occur here — they are one-shot.
	// No background goroutines; defer to Start().
	// Failure leaves the service in ServiceStateUninitialized.
	Init(asnServiceNode ASNServiceNode) error

	// Start starts the service with the given configuration.
	// Called after Init() and again on each config change. Sequential; may be called multiple times.
	// Must be idempotent: each invocation must fully supersede the previous config.
	// If the service cannot hot-reload, return ErrRestartNeeded; the framework will execute
	// Stop() → Start() with the new config (no re-initialization).
	// runtimeErrChan is for reporting fatal runtime errors discovered after Start() returns.
	// Only unrecoverable conditions should be sent; the framework transitions the service to
	// ServiceStateMalfunctioning on receipt.
	// A synchronous non-ErrRestartNeeded error also → ServiceStateMalfunctioning.
	Start(config string) (runtimeErrChan <-chan error, err error)

	// ApplyServiceOps applies a service op dispatched by the controller.
	// Explicitly concurrent — the framework may call this simultaneously from multiple goroutines.
	// All shared state access must be synchronized.
	// resp is forwarded as OpsResponse.ServiceResponse; err as OpsResponse.ServiceError.
	// Neither represents a fatal error; use runtimeErrChan for unrecoverable failures.
	// The framework enforces a per-call timeout; must return promptly.
	ApplyServiceOps(opCmd, opParams string) (resp string, err error)

	// AddConfigOps applies new config ops to the service.
	// Concurrent with other callbacks; guard shared state.
	// Returning an error transitions the service to ServiceStateMalfunctioning.
	// Only return errors for genuinely unrecoverable failures.
	AddConfigOps(configParams []string) (resp string, err error)

	// UpdateConfigOp updates a single config op, identified by its current value.
	// Concurrent with other callbacks; guard shared state.
	// Returning an error transitions the service to ServiceStateMalfunctioning.
	UpdateConfigOp(oldConfigParam, newConfigParam string) (resp string, err error)

	// DeleteConfigOps removes config ops from the service.
	// Concurrent with other callbacks; guard shared state.
	// Returning an error transitions the service to ServiceStateMalfunctioning.
	DeleteConfigOps(configParams []string) (resp string, err error)

	// OnQuerySharedData returns current values for the requested keys.
	// Called when another service on the same node calls QueryServiceSharedData() targeting this service.
	// Concurrent; guard shared state.
	// Implement only if this service declares aggregated keys in StaticResource.SharedData().
	// Return ErrKeyNotFound for keys not in the declared set.
	OnQuerySharedData(keys []string) (values map[string]any, err error)

	// OnSubscribeSharedData returns a per-key channel of streaming values.
	// Called when another service calls SubscribeServiceSharedData() targeting this service.
	// Concurrent; guard shared state.
	// Implement only if this service declares subscribable keys in StaticResource.SharedData().
	// Each channel must be closed after all values are sent to signal end-of-stream.
	// Return ErrKeyNotFound for keys not in the declared set.
	OnSubscribeSharedData(keys []string) (values map[string]<-chan any, err error)

	// Stop stops the service.
	// Idempotent; must return promptly. Any returned error may influence service state.
	Stop() error

	// Finish releases all resources before the .so is unloaded.
	// Called once, after Stop(). Any goroutines still running after Finish() returns
	// and touching service-owned memory cause undefined behavior.
	Finish()
}
