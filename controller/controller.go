package capi

import (
	"github.com/amianetworks/asn-service-api/common"
)

/*
	Struct used between asn.controller and service.controller,
*/

// Network struct
type Network struct {
	Id string
}

// Node struct
type Node struct {
	Id             string
	Type           string
	NetworkId      string
	ParentId       string
	ExternalLinked []string
	InternalLinked []string
}

// ServiceStatus Service status struct, this is a MUST have! ServiceStatus.Enabled indicates the service state from the asn.controller's view
type ServiceStatus struct {
	Enabled bool
}

// API provided by ASN controller
type API interface {

	/*
		Get all networks under asn controller
	*/
	GetNetworks() ([]Network, error)

	/*
		Get all nodes of network
	*/
	GetNodesOfNetwork(networkId string) ([]Node, error)

	/*
		Get all nodes of the parent.
	*/
	GetNodesOfParent(networkNodeId string) ([]Node, error)

	/*
		Get node by id
	*/
	GetNodeById(id string) (Node, error)

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
	ReadConfOfNetwork(networkId string, serviceName string) ([]byte, error)
	ReadConfOfServiceNode(serviceNodeId string, serviceName string) ([]byte, error)

	/*
		Set the service setting by network id and service name,
		the Conf []byte is Marshalled
		Write the service setting to a specific service node by ASN controller
	*/
	SaveConfOfNetwork(networkId string, serviceName string, config []byte) error
	SaveConfOfServiceNode(serviceNodeId string, serviceName string, config []byte) error

	/*
		CRUD (Create, Read, Update, Delete) operation for the service metadata.
		The metadata []byte is Marshalled
	*/
	ReadMetadataOfNetwork(networkId string, serviceName string, fileName string) ([]byte, error)
	ReadMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string) ([]byte, error)

	// SaveMetadata will create the metadata if it is not exist, otherwise will
	SaveMetadataOfNetwork(networkId string, serviceName string, fileName string, metadata []byte) error
	SaveMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string, metadata []byte) error

	DeleteMetadataOfNetwork(networkId string, serviceName string, fileName string) error
	DeleteMetadataOfServiceNode(serviceNodeId string, serviceName string, fileName string) error

	/*
		Init the ASN logger. This logger is different with the 'defaultLogger' that passed by the Init() function.
			- defaultLogger is the log system that managed by the ASN framework, which is writing log to the specific path defined by the controller
			- By using this API, you can init a private logger that is distinguished with the defaultLogger which mean you can save the log to the service defined path
	*/
	InitASNLogger(serviceName string, logPath string) (*commonapi.ASNLogger, error)
}

// ASNController struct will be declared in service side and implemented by ASN controller
type ASNController struct {
	API API
}

/*
	ASNService struct provides the service's API for the ASN Controller usage,
	will be implement by service and used by ASN controller
*/
type ASNService struct {
	/*
		Service name, it is important to have the same name with capi.ASNService.Name
	*/
	Name string

	/*
		Initialize the service
	*/
	Init func(defaultLogger *commonapi.ASNLogger) error

	/*
		Get the default runtime configuration of the service.
		Service should return nil if no default config needed. //TODO: nil handling
	*/
	GetDefaultConf func() []byte

	/*
		Get the *current* configuration of the service network/node.
		Service may have saved the configuration to DB,
		but it's safer to read current or latest configuration directly from the service controller.
	*/
	GetConfOfNetwork     func(networkId string) []byte
	GetConfOfServiceNode func(serviceNodeId string) []byte

	/*
		Apply OPERATION command from client(cli/dashboard), service.controller needs to parse
		the operations for each service node. This interface has two versions:
			- ApplyConfig() applies the config on all service nodes in the network.
			- ApplyConfigToServiceNodes() only applies to a list of service nodes.
	*/
	ApplyOpsToNetwork      func(networkId string, ops []byte) error
	ApplyOpsToServiceNodes func(serviceNodes []string, ops []byte) error

	/*
		Apply START command with the configuration from client(cli/dashboard), This interface has two versions:
			- ApplyConfig() applies the config on all service nodes in the network.
			- ApplyConfigToServiceNodes() only applies to a list of service nodes.
	*/
	ApplyStartToNetwork      func(networkId string, conf []byte) error
	ApplyStartToServiceNodes func(serviceNodes []string, conf []byte) error

	/*
		Apply STOP command from client(cli/dashboard),  This interface has two versions:
			- ApplyConfig() applies the config on all service nodes in the network.
			- ApplyConfigToServiceNodes() only applies to a list of service nodes.
	*/
	ApplyStopToNetwork      func(networkId string) error
	ApplyStopToServiceNodes func(serviceNodes []string) error

	/*
		Get service node's service config status, ENABLED or not.
		Service Controller must maintain this "status" of configuration and report it accordingly.
	*/
	GetStatusOfServiceNode func(serviceNodeId string) ServiceStatus

	/*
		Get the applied serviceOps of the service node.
		ASN Controller may call it in the case reconfiguration is needed for a service node.
	*/
	GetOpsOfServiceNode func(serviceNodeId string) []byte

	/*
		Received the metadata from the service in the service node
	*/
	ReceivedMetadataFromServiceNode func(serviceNodeId string, metadata []byte) error

	/*
		GetVersion of the service
	*/
	GetVersion func() commonapi.Version
}
