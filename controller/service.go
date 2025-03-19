// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	commonapi "github.com/amianetworks/asn-service-api/v25/common"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

/*
ServiceStatus Service status struct, this is a MUST-have!
ServiceStatus.Enabled indicates the service state from the asn.controller's view
ServiceStatus.Extra is an option for the service providing extra status
*/
type ServiceStatus struct {
	Enabled bool
	Extra   []byte
}

/*
ASNService interface provides the service's API for the ASN Controller usage,
will be implemented by service and used by ASN controller
*/
type ASNService interface {
	/*
		GetVersion of the service
	*/
	GetVersion() commonapi.Version

	/*
		Received the metadata from the service in the service node
	*/
	ReceivedMetadataFromServiceNode(serviceNodeId string, metadata []byte) error

	/*
		Finish the service when ASN Controller finishes work
	*/
	Finish()
}

type ASNServiceAPIs interface {
	GetCLICommands(
		sendServiceApplyOpsCmdToNetwork func(opCmd, opParams string),
		sendServiceApplyOpsCmdToGroup func(group string, opCmd, opParams string),
		sendServiceApplyOpsCmdToNodes func(nodes []string, opCmd, opParams string),
	) []*cobra.Command // no need to include start/stop/show/status, only include ops

	MountWebHandler(staticPath string) func(group *gin.RouterGroup) error
}
