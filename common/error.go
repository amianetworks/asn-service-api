// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import (
	"errors"
)

type FrameworkErr error

var FrameworkErrServiceTimeout FrameworkErr = errors.New("service timed out")
var FrameworkErrNodeDisconnected FrameworkErr = errors.New("node disconnected")
var FrameworkErrServiceUnavailable FrameworkErr = errors.New("service unavailable")
var FrameworkErrServiceStateNotAllowed FrameworkErr = errors.New("service state not allowed to apply ops")
