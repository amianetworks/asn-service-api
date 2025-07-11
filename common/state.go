// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

const (
	ServiceNodeStateUnregistered = iota
	ServiceNodeStateOffline
	ServiceNodeStateOnline
	ServiceNodeStateMaintenance
)

const (
	ServiceStateUnavailable = iota
	ServiceStateUninitialized
	ServiceStateInitialized
	ServiceStateConfiguring
	ServiceStateRunning
	ServiceStateMalfunctioning
)
