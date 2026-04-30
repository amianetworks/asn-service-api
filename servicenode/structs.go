// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	"errors"

	commonapi "asn.amiasys.com/asn-service-api/v26/common"
)

var (
	// ErrServiceNotFound is returned when the target service is not loaded on this node.
	ErrServiceNotFound = errors.New("target service not found")
	// ErrKeyNotFound is returned when a requested shared data key is not provided by the target service.
	ErrKeyNotFound = errors.New("key not found in target service")
	// ErrRestartNeeded is returned from ASNService.Start() when the service cannot hot-reload
	// its configuration. The framework will execute Stop() → Start() with the new config.
	ErrRestartNeeded = errors.New("restart needed")
)

// NodeInfo is the node information available to a service running on the node.
// It embeds commonapi.NodeInfo (hardware interfaces, IPMI, management, device specs)
// and adds the node's ID and the list of active config op strings.
type NodeInfo struct {
	ID string
	commonapi.NodeInfo
	// ConfigOps is the list of active config op param strings applied to this node.
	ConfigOps []string
}
