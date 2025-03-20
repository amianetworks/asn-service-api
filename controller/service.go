// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	commonapi "github.com/amianetworks/asn-service-api/v25/common"
)

// ServiceStatus Service status struct, this is a MUST-have!
// ServiceStatus.Enabled indicates the service state from the asn.controller's view
// ServiceStatus.Extra is an option for the service providing extra status
type ServiceStatus struct {
	Enabled bool
	Extra   []byte
}

// ASNService interface provides the service's API for the ASN Controller usage,
// will be implemented by service and used by ASN controller.
type ASNService interface {
	// GetVersion returns the version of the service.
	GetVersion() commonapi.Version

	// ReceivedMetadataFromServiceNode processes the received metadata from the service in the service node.
	ReceivedMetadataFromServiceNode(serviceNodeId string, metadata []byte) error

	// Finish the service when ASN Controller finishes work.
	Finish()
}

type ASNServiceAPIs interface {
	// GetCLICommands returns the cobra CLI commands used for this service controller.
	// These will be mounted to ASN Controller's CLI under this service.
	GetCLICommands(
		sendServiceApplyOpsCmdToNetwork func(opCmd, opParams string),
		sendServiceApplyOpsCmdToGroup func(group string, opCmd, opParams string),
		sendServiceApplyOpsCmdToNodes func(nodes []string, opCmd, opParams string),
	) []*cobra.Command // no need to include start/stop/show/status, only include ops

	// MountWebHandler returns a function to mount the web handler got this service controller.
	MountWebHandler(staticPath string) func(group *gin.RouterGroup) error
}
