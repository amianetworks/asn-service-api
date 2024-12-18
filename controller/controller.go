package capi

import (
	commonapi "github.com/amianetworks/asn-service-api/v25/common"
	"github.com/amianetworks/asn-service-api/v25/log"
)

/*
	Struct used between asn.controller and service.controller,
*/

// Node struct
type Node struct {
	Id             string
	Type           string
	ParentId       string
	Group          string
	ExternalLinked []string
	InternalLinked []string
	Services       map[string]bool
}

type Group struct {
	Name  string
	Nodes []string
}

/*
ServiceStatus Service status struct, this is a MUST-have!
ServiceStatus.Enabled indicates the service state from the asn.controller's view
ServiceStatus.Extra is an option for the service providing extra status
*/
type ServiceStatus struct {
	Enabled bool
	Extra   []byte
}

// ASNController API provided by ASN controller
type ASNController interface {
	/*
		Get all nodes of network
	*/
	GetNodesOfNetwork(serviceName string) ([]Node, error)

	/*
		Get all groups of network
	*/
	GetGroupsOfNetwork(serviceName string) ([]Group, error)

	/*
		Get all nodes of group
	*/
	GetNodesOfGroup(groupName, serviceName string) ([]Node, error)

	/*
		Get all nodes of the parent.
	*/
	GetNodesOfParent(parentNodeId, serviceName string) ([]Node, error)

	/*
		Get node by id
	*/
	GetNodeById(id, serviceName string) (Node, error)

	/*
		Get group by group name
	*/
	GetGroupByName(groupName string) (Group, error)

	/*
		Send START cmd to the service node with the specific service name
		The config is a pre-defined struct. Both of service.controller and service.sn has the same struct,
		so they can easily use xxx.Marshall() and xxx.Unmarshall() to convert the struct between []byte and the struct
	*/
	StartService(serviceNodeId string, serviceName string, config []byte) error

	/*
		Send STOP cmd to the service node with the specific service name
	*/
	StopService(serviceNodeId string, serviceName string) error

	/*
		Send CONFIG cmd to the service node with the specific service name, the configCmd is a pre-defined struct.
		Both of service.controller and service.sn has the same struct,
		so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct
	*/
	SendServiceOps(serviceNodeId string, serviceName string, serviceOps []byte) error

	/*
		Read the service COnf by network id and service name,
		The setting []byte is the config/rule/policies struct defined in service.controller,
		Use Unmarshall to converting the []byte to the Conf struct
	*/
	ReadConfOfNetwork(serviceName string) ([]byte, error)
	ReadConfOfGroup(groupName string, serviceName string) ([]byte, error)
	ReadConfOfServiceNode(serviceNodeId string, serviceName string) ([]byte, error)

	/*
		Set the service setting by network id and service name,
		the Conf []byte is Marshalled
		Write the service setting to a specific service node by ASN controller
	*/
	SaveConfOfNetwork(serviceName string, config []byte) error
	SaveConfOfGroup(groupName string, serviceName string, config []byte) error
	SaveConfOfServiceNode(serviceNodeId string, serviceName string, config []byte) error

	/*
		CRUD (Create, Read, Update, Delete) operation for the service metadata.
		The metadata []byte is Marshalled
	*/
	ReadMetadataOfNetwork(serviceName string, fileName string) ([]byte, error)
	ReadMetadataOfGroup(groupName string, serviceName string, fileName string) ([]byte, error)
	ReadMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string) ([]byte, error)

	SaveMetadataOfNetwork(serviceName string, fileName string, metadata []byte) error
	SaveMetadataOfGroup(groupName string, serviceName string, fileName string, metadata []byte) error
	SaveMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string, metadata []byte) error

	DeleteMetadataOfNetwork(serviceName string, fileName string) error
	DeleteMetadataOfGroup(groupName string, serviceName string, fileName string) error
	DeleteMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string) error

	/*
		Write the log to your service path. This is based on am.module logs
	*/
	GetLogger(serviceName string) (*log.Logger, error)
	GetIAM() (IAM, error)
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
		Service should return nil if no default config needed. //TODO: nil handling
	*/
	GetDefaultConf() []byte

	/*
		Get the *current* configuration of the service network/node.
		Service may have saved the configuration to DB,
		but it's safer to read current or latest configuration directly from the service controller.
	*/
	GetConfOfNetwork() []byte
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
	GetStatusOfNetwork() ServiceStatus
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
