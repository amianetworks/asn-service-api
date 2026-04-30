// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"net/http"

	"github.com/spf13/cobra"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

// StaticResource groups pre-init metadata that must be available before Init().
// All methods are re-entrant and safe to call at any time, including before Init().
type StaticResource interface {
	// Version returns the service controller's version. Must be a hardcoded constant.
	Version() commonapi.Version

	// CLICommands returns Cobra commands to register in the ASN CLI.
	// Purely declarative — no side effects, no goroutines.
	// applyCLIOps is the framework-provided dispatcher; call it from command handlers
	// to fan ops to service nodes. Its semantics are identical to ASNController.SendServiceOps.
	CLICommands(
		applyCLIOps func(serviceScope commonapi.ServiceScope, serviceScopeList []string, opCmd, opParams string) error,
	) []*cobra.Command

	// WebHandler returns the HTTP handler for this service controller.
	// The framework mounts it under a service-specific path prefix.
	// staticPath is the filesystem path to the service's static asset directory.
	// No resource acquisition here; defer to Init().
	WebHandler(staticPath string) (http.Handler, error)
}

// ASNServiceController is the interface implemented by a Service Controller.
// The framework calls these methods to manage the service's lifetime and behavior.
//
// Lifecycle order: StaticResource → Init → Start → [concurrent callbacks] → Stop → Finish.
type ASNServiceController interface {
	// StaticResource returns the pre-init resources (version, CLI, web handler).
	// Re-entrant; must be callable before Init().
	StaticResource() StaticResource

	// Init initializes the service controller.
	// Called once after .so load, before Start(). Not re-entrant.
	// All Init* / Get* calls on asnController must occur here — they are one-shot.
	// On success, CLI commands registered via CLICommands() must be immediately runnable.
	// No background goroutines; defer to Start().
	// Failure leaves the service in ServiceStateUninitialized.
	Init(asnController ASNController) error

	// Start starts the service controller with the given configuration.
	// Called after Init() and again on each config change. Sequential; may be called multiple times.
	// Must return promptly; start background goroutines as needed.
	// Each invocation must fully supersede the previous config.
	Start(config string) error

	// AddConfigOps is called after the framework persists and dispatches config ops to affected
	// service nodes. Implement controller-side bookkeeping here (routing, policy, in-memory state).
	// Concurrent with other callbacks; guard shared state.
	// serviceScope is ServiceScopeNodeGroup (3) or ServiceScopeNode (4); scopeID is the corresponding ID.
	AddConfigOps(serviceScope commonapi.ServiceScope, scopeID string, configParams []string) error

	// UpdateConfigOp is called after the framework persists the config op update and notifies nodes.
	// configOpID identifies the existing op being replaced.
	// Concurrent with other callbacks; guard shared state.
	// serviceScope is ServiceScopeNodeGroup (3) or ServiceScopeNode (4).
	UpdateConfigOp(serviceScope commonapi.ServiceScope, scopeID, configOpID, configParam string) error

	// DeleteConfigOps is called after the framework removes the ops from storage and notifies nodes.
	// Concurrent with other callbacks; guard shared state.
	// serviceScope is ServiceScopeNodeGroup (3) or ServiceScopeNode (4).
	DeleteConfigOps(serviceScope commonapi.ServiceScope, scopeID string, configOpIDs []string) error

	// HandleMessageFromNode handles upcalls from service nodes sent via ASNServiceNode.SendMessageToController().
	// Concurrent — messages from multiple nodes may arrive simultaneously; guard shared state.
	// No direct response channel exists; to reply, initiate a SendServiceOpsToNode() call.
	HandleMessageFromNode(nodeID, messageType, payload string) error

	// GetMetrics returns display-only metrics scoped to the given network.
	// Concurrent; must return promptly. Values must be JSON-serializable.
	// Maintain snapshots in background goroutines; do not compute on the call path.
	GetMetrics(networkID string) (map[string]string, error)

	// Stop gracefully stops the service controller.
	// Idempotent; must return promptly. Stop all background goroutines started in Start().
	// The return value is informational; the framework proceeds with shutdown regardless.
	Stop() error

	// Finish releases all resources before the .so is unloaded.
	// Called once, after Stop(). Any goroutines still running after Finish() returns
	// and touching service-owned memory cause undefined behavior.
	Finish()
}
