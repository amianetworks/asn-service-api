package snapi

import (
	"github.com/amianetworks/asn-service-api/common"
)

/*
	Struct used for service node and service communication
*/

// API provided by ASN Service Node for Service uses
type API interface {
	/*
		Get ASN managed netifs from Service node
	*/
	GetServiceNodeNetif() (Netif, error)

	/*
		Send the metadata to the controller
	*/
	SendMetadataToController(serviceName string, metadata []byte) error

	/*
		Init the ASN logger. This logger is different with the 'defaultLogger' that passed by the Init() function.
			- defaultLogger is the log system that managed by the ASN framework, which is writing log to '/var/log/asn.log'
			- By using this API, you can init a private logger that is distinguished with the defaultLogger which mean you can save the log to the service defined path
	*/
	InitASNLogger(serviceName string, logPath string) (*commonapi.ASNLogger, error)

	/*
		Get ASN Service Node running Mode, currently support 'cluster', 'standalone' and 'hybrid' mode
	*/
	GetServiceNodeRunningMode() (string, error)

	/*
		Get ASN Service Node type, currently support 'server', 'appliance'
	 */
	GetServiceNodeType() string
}

// Network interface struct
type Netif struct {
	Data       []string
	Control    []string
	Management []string
	Other      []string
}

// This struct will be declared in service side and implemented by ASN Service Node
type ServiceNode struct {
	API API
}

/*
	This struct provides the service's API for the Service Node usage,
	will be implement by service and used by ASN Service Node
*/
type ASNService struct {
	/*
		Service name, it is important to have the same name with capi.ASNService.Name
	*/
	Name string

	/*
		Initialize the service, this method will be called under a go routine

		input parameters:
		 1. defaultLogger:
			- If service want output the log to the asn service node's log, use this logger.
			- If the service want to maintain their own log, please init a new logger. For details, please refer to shared/logger.go
		 2. the return value to channel indicate service state:
			- if error is nil, the service node will assign the state INITIALIZED to the service
			- if error is NOT nil, the service node will assign the state MALFUNCTIONAL to the service

		Caution: the service node will have a timeout context (20s) to process the initialization,
				 if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	*/
	Init func(defaultLogger *commonapi.ASNLogger, c chan error)

	/*
		Start the service with the configuration.

		input parameters:
		1. config:
			the configuration of the service. Service MUST update this configuration to the local file.
			When DumpConfiguration called, service need to return the current configuration to the framework
		2. the return value to channel indicate service state:
			- if error is nil, the service node will assign the state CONFIGURED to the service
			- if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
			  after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service

		Caution: the service node will have a timeout context (20s) to process the initialization,
				 if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	*/
	Start func(config []byte, c chan error)

	/*
		Apply the service operations to the service.
		Service operations will not changing the service status (enabled/disabled),
		but will do some runtime operations such as: insert/delete/getXXX/setXXX
		Apply the configuration to the service, this method will be called under a go routine, the return value to channel indicate service state:
		if error is nil, the service node will remain the previous state (CONFIGED/INITIALIZED)
		if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
			  after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service

		Caution: the service node will have a timeout context (20s) to process the initialization,
				 if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	*/
	ApplyServiceOps func(ops []byte, c chan error)

	/*
		Stop the service with the configuration.

		input parameters:
		1. the return value to channel indicate service state:
			- if error is nil, the service node will assign the state INITIALIZED to the service
			- if error is NOT nil, the service node will try to init the service and re-apply the configuration for 3 times,
			  after all retry if it is still having error, will assign the state MALFUNCTIONAL to the service

		Caution: the service node will have a timeout context (20s) to process the initialization,
				 if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	*/
	Stop func(c chan error)

	/*
		Last call before the service node's termination. Do the necessary clean up here.
	*/
	Finish func() error

	/*
		GetVersion of the service,
		share.Version provide the initializer (version parser) and a toString convert,
		for details, please refer to share/version.go
	*/
	GetVersion func() commonapi.Version
}
