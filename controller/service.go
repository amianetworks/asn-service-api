// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"net/http"
	"time"

	"github.com/spf13/cobra"

	commonapi "asn.amiasys.com/asn-service-api/v27/common"
)

// StaticResource groups pre-init metadata that must be available before Init().
// All methods are re-entrant and safe to call at any time, including before Init().
type StaticResource interface {
	// Version returns the service controller's version. Must be a hardcoded constant.
	Version() commonapi.Version

	// CLICommands returns Cobra commands to register in the ASN CLI.
	// Purely declarative — no side effects, no goroutines. cli is the
	// framework's context for the command handlers.
	CLICommands(cli CLIContext) []*cobra.Command

	// Catalog returns the service's catalog (see ServiceCatalog); nil
	// declares none. ASN validates it with ValidateServiceCatalog when the
	// plugin loads and refuses the plugin when it is invalid. Must be a
	// constant value.
	Catalog() *ServiceCatalog

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

	// HandleMessageFromNode handles upcalls from service nodes sent via ASNServiceNode.SendMessageToController().
	// Concurrent — messages from multiple nodes may arrive simultaneously; guard shared state.
	// No direct response channel exists; to reply, initiate a SendServiceOpsToNode() call.
	HandleMessageFromNode(nodeID, messageType string, payload []byte) error

	// GetMetricsTemplate returns how the service's metrics for target are
	// displayed (see MetricsTemplate). The network or node exists and the
	// caller may read it. A template may differ per target; an empty
	// template (nil with a nil error) reports nothing there. Called after
	// Start; concurrent; the framework bounds each call to 3 s. The result may
	// change over time and takes effect on the next fetch; a released key
	// keeps its type and unit. The framework does not retain the value past
	// the call. Check templates with ValidateMetricsTemplate in unit tests.
	GetMetricsTemplate(target MetricsTarget) (*MetricsTemplate, error)

	// GetMetrics returns the service's metrics for target and window, shaped
	// by GetMetricsTemplate(target); window is one the template serves (the
	// framework passes the default for a caller naming none). At a network
	// target a network-scope field covers the network and all its
	// subnetworks; a service-scope field does not follow the target. Values
	// are encoded with encoding/json in base units: nil maps, slices and
	// pointers mean "no data"; NaN and ±Inf are errors; do not tag declared
	// numeric fields omitempty. Called after Start; concurrent; bounded to
	// 3 s. Maintain figures in background goroutines. Check data with
	// ValidateMetrics in unit tests.
	GetMetrics(target MetricsTarget, window string) (map[string]any, error)

	// OnDataDemand gives the full set of on-demand data the event rules read
	// now: per item of the catalog declared CatalogDemandOnDemand, the nodes
	// (none: every node of the service) and the shortest interval wanted.
	// Called after Start and on every change of the set, sequentially; the
	// service writes those points until a later set drops them. An error is
	// logged; rules on the data then raise nothing.
	OnDataDemand(demands []DataDemand) error

	// OnLicenseChanged notifies the service that the license content or status changed.
	OnLicenseChanged(license *LicenseInfo) error

	// Stop gracefully stops the service controller.
	// Idempotent; must return promptly. Stop all background goroutines started in Start().
	// The return value is informational; the framework proceeds with shutdown regardless.
	Stop() error

	// Finish releases all resources before the .so is unloaded.
	// Called once, after Stop(). Any goroutines still running after Finish() returns
	// and touching service-owned memory cause undefined behavior.
	Finish()
}

// DataDemand is the demand for an on-demand data item.
type DataDemand struct {
	Item        string
	NodeIDs     []string // none: every node of the service
	MinInterval time.Duration
}

// CLIContext is what the ASN CLI gives a service's commands.
type CLIContext interface {
	// ApplyOps dispatches an op to the service's nodes, as
	// ASNController.SendServiceOps for the calling service.
	ApplyOps(serviceScope commonapi.ServiceScope, serviceScopeList []string, opCmd, opParams string) error

	// AccessToken is the CLI user's access token, for the commands that call
	// the service's own web API.
	AccessToken() string
}
