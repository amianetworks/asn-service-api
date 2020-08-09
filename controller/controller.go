package controller

/*
	Struct used between asn.controller and service.controller,
*/

// Network struct TODO: the following is an example of the Network struct, NOT finalized
type Network struct {
	Id string
}

// Network interface struct
type Netif struct {
	Data       string
	Control    string
	Management string
}

// ServiceNode struct TODO: the following is an example of the Service Node struct, NOT finalized
type ServiceNode struct {
	Id    string
	Netif Netif
}

// Service status struct, this is a MUST have! ServiceStatus.Enabled indicates the service state from the asn.controller's view
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
		Get service nodes of network
	*/
	GetServiceNodes(networkId string) ([]ServiceNode, error)

	/*
		Send config cmd to the service node with the specific service name, the configCmd is a pre-defined struct.
		Both of service.controller and service.sn has the same struct,
		so they can easily use JSON.Marshall() and JSON.Unmarshall() to convert the struct between []byte and the struct
	*/
	SendServiceOps(serviceNodeId string, serviceName string, serviceOps []byte) error

	/*
		Read the service configuration by network id and service name,
		The returning []byte is the config/rule/policies struct defined in service.controller,
		Use JSON.Unmarshall to converting the []byte to the Config struct
	*/
	ReadSettings(networkId string, serviceName string) ([]byte, error)
	ReadSettingsOfServiceNode(serviceNodeId string, serviceName string) ([]byte, error)

	/*
		Set the service configuration by network id and service name,
		the config []byte is Marshalled by using JSON.Marshall()
		Write the service config to a specific service node by ASN controller
	*/
	SaveSettings(networkId string, serviceName string, config []byte) error
	SaveSettingsOfServiceNode(serviceNodeId string, serviceName string, config []byte) error
}

// This struct will be declared in service side and implemented by ASN controller
type ASNController struct {
	API API
}

/*
	This struct provides the service's API for the ASN Controller usage,
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
	Init func() error

	/*
		Get the default runtime configuration of the service.
		Service should return nil if no default config needed. //TODO: nil handling
	*/
	GetDefaultSettings func() []byte

	/*
		Get the *current* Settings of the service network/node.
		Service may have saved the settings to DB. But it's safer to read current,
		or latest, settings directly from the service controller.
	*/
	GetSettings              func(networkId string) []byte
	GetSettingsOfServiceNode func(serviceNodeId string) []byte

	/*
		Apply ServiceConfig from client(cli/dashboard), service.controller needs to parse
		ServiceCONFIG to operations for each service node. This interface has two versions:
			- ApplyConfig() applies the config on all service nodes in the network.
			- ApplyConfigToServiceNodes() only applies to a list of service nodes.
	*/
	ApplyConfig               func(networkId string, conf []byte) error
	ApplyConfigToServiceNodes func(serviceNodes []string, conf []byte) error

	/*
		Get service node's service config status, ENABLED or not.
		Service Controller must maintain this "status" of configuration and report it accordingly.
	*/
	GetServiceStatusOfServiceNode func(serviceNodeId string) ServiceStatus

	/*
		Get the applied serviceOps of the service node.
		ASN Controller may call it in the case reconfiguration is needed for a service node.
	*/
	GetServiceOpsOfServiceNode func(serviceNodeId string) []byte
}
