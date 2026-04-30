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
