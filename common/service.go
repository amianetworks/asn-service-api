// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

const (
	CLIOpScopeNetwork = 1 + iota
	CLIOpScopeGroup
	CLIOpScopeNode
)

type Response struct {
	Response string
	Error    error
}
