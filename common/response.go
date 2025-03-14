// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

type Response struct {
	Response   string
	Error      error
	ErrorFatal bool
}
