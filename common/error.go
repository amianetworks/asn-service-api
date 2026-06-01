// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import (
	"errors"
)

// FrameworkErr is the type for framework-level errors carried in OpsResponse.FrameworkError
// and NodeStateChange.FrameworkError. When non-nil, it indicates that the framework itself
// could not reach or invoke the service; ServiceResponse and ServiceError are undefined.
type FrameworkErr error

var (
	// FrameworkErrServiceTimeout - node is reachable, but the service did not respond within the timeout.
	FrameworkErrServiceTimeout FrameworkErr = errors.New("service timed out")

	// FrameworkErrNodeDisconnected - the target node is offline.
	FrameworkErrNodeDisconnected FrameworkErr = errors.New("node disconnected")

	// FrameworkErrServiceUnavailable - the service is not loaded on the target node.
	FrameworkErrServiceUnavailable FrameworkErr = errors.New("service unavailable")

	// FrameworkErrServiceStateNotAllowed - the service is not in ServiceStateRunning.
	FrameworkErrServiceStateNotAllowed FrameworkErr = errors.New("service state not allowed to apply ops")
)

// LicenseStatusErr is returned when the current license is missing or cannot be used.
type LicenseStatusErr error

var (
	// LicenseStatusErrNotFound indicates that no license is installed.
	LicenseStatusErrNotFound LicenseStatusErr = errors.New("license not found")

	// LicenseStatusErrInactive indicates that the installed license has not been activated.
	LicenseStatusErrInactive LicenseStatusErr = errors.New("license is inactive")

	// LicenseStatusErrExpired indicates that the installed license is outside its valid time range.
	LicenseStatusErrExpired LicenseStatusErr = errors.New("license is expired")

	// LicenseStatusErrSuspended indicates that the installed license has been suspended.
	LicenseStatusErrSuspended LicenseStatusErr = errors.New("license is suspended")

	// LicenseStatusErrAbnormal indicates that the framework cannot reach the license authority.
	LicenseStatusErrAbnormal LicenseStatusErr = errors.New("cannot reach the license server")

	// LicenseStatusErrInvalidMachineID indicates that the license is not bound to the requested machine.
	LicenseStatusErrInvalidMachineID LicenseStatusErr = errors.New("machine ID not match")
)
