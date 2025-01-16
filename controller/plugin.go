package capi

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	commonapi "github.com/amianetworks/asn-service-api/v25/common"
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
		Service name, it is important to have the same name with capi.ASNService.Name
	*/
	GetName() string

	/*
		Get the default runtime configuration of the service.
		Service should return nil if no default config needed.
	*/
	GetDefaultConf() ([]byte, error)

	/*
		Get the *current* configuration of the service network/node.
		Service may have saved the configuration to DB,
		but it's safer to read current or latest configuration directly from the service controller.
	*/
	GetConfOfNetwork() ([]byte, error)
	GetConfOfGroup(groupName string) ([]byte, error)
	GetConfOfServiceNode(serviceNodeId string) ([]byte, error)

	/*
		Apply OPERATION command from client(cli/dashboard), service.controller needs to parse
		the operations for each service node. This interface has two versions:
			- ApplyConfig() applies the config on all service nodes in the network.
			- ApplyConfigToServiceNodes() only applies to a list of service nodes.
	*/
	ApplyOpsToNetwork(ops []byte) error
	ApplyOpsToGroup(groupName string, ops []byte) error
	ApplyOpsToServiceNodes(serviceNodes []string, ops []byte) error

	/*
		Apply START command with the configuration from client(cli/dashboard), This interface has two versions:
			- ApplyConfig() applies the config on all service nodes in the network.
			- ApplyConfigToServiceNodes() only applies to a list of service nodes.
	*/
	ApplyStartToNetwork(conf []byte) error
	ApplyStartToGroup(groupName string, conf []byte) error
	ApplyStartToServiceNodes(serviceNodes []string, conf []byte) error

	/*
		Apply STOP command from client(cli/dashboard),  This interface has two versions:
			- ApplyConfig() applies the config on all service nodes in the network.
			- ApplyConfigToServiceNodes() only applies to a list of service nodes.
	*/
	ApplyStopToNetwork() error
	ApplyStopToGroup(groupName string) error
	ApplyStopToServiceNodes(serviceNodes []string) error

	/*
		Get service node's service config status, ENABLED or not.
		Service Controller must maintain this "status" of configuration and report it accordingly.
	*/
	GetStatusOfNetwork() (ServiceStatus, error)
	GetStatusOfGroup(groupName string) (ServiceStatus, error)
	GetStatusOfServiceNode(serviceNodeId string) (ServiceStatus, error)

	/*
		Get the applied serviceOps of the service node.
		ASN Controller may call it in the case reconfiguration is needed for a service node.
	*/
	GetOpsOfServiceNode(serviceNodeId string) ([]byte, error)

	/*
		Received the metadata from the service in the service node
	*/
	ReceivedMetadataFromServiceNode(serviceNodeId string, metadata []byte) error

	/*
		GetVersion of the service
	*/
	GetVersion() commonapi.Version
}

type ASNServiceAPIs interface {
	GetCLICommands(
		sendServiceApplyOpsCmdToNetwork func(ops []byte),
		sendServiceApplyOpsCmdToGroup func(group string, ops []byte),
		sendServiceApplyOpsCmdToNodes func(nodes []string, ops []byte),
	) []*cobra.Command // no need to include start/stop/show/status, only include ops

	MountWebHandler() func(group *gin.RouterGroup) error
}
