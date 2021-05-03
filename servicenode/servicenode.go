package servicenode

import (
	"github.com/amianetworks/asn-service-api/shared"
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
		Send the metadata to controller
	*/
	SendMetadataToController(serviceName string, metadata []byte) error

	/*
		Init the ASN logger. This logger is different with the 'defaultLogger' that passed by the Init() function.
			- defaultLogger is the log system that managed by the ASN framework, which is writing log to '/var/log/asn.log'
			- By using this API, you can init a private logger that is distinguished with the defaultLogger which mean you can save the log to the service defined path
	*/
	InitASNLogger(serviceName string, logPath string) (*shared.ASNLogger, error)
}

// Network interface struct
type Netif struct {
	Data       []string
	Control    []string
	Management []string
	Other      []string
}

// Service status struct, this is a MUST have! ServiceStatus.Enabled indicates the service state from the asn.controller's view
type Status struct {
	Enabled bool
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
		1. configPath: Service must use the this configPath to load the service configurations
		2. defaultLogger:
			- If service want output the log to the asn service node's log, use this logger.
			- If the service want to maintain their own log, please init a new logger. For details, please refer to shared/logger.go
		3. the return value to channel indicate service state:
			- if error is nil, the service node will assign the state INITIALIZED to the service
			- if error is NOT nil, the service node will assign the state MALFUNCTIONAL to the service

		Caution: the service node will have a timeout context (20s) to process the initialization,
				 if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	*/
	Init func(configPath string, defaultLogger *shared.ASNLogger, c chan error)

	/*
		Apply the configuration to the service, this method will be called under a go routine, the return value to channel indicate service state:
		if error is nil, the service node will call GetStatus for assigning the proper state (CONFIGED/INITIALIZED) to the service
		if error is NOT nil, the service node will call Init function to reset the service

		Caution: the service node will have a timeout context (20s) to process the initialization,
				 if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	*/
	ApplyServiceOps func(conf []byte, c chan error)

	/*
		Read the current configuration of the service,
		this method cannot be blocked, read the service setting and return immediately
	*/
	DumpSettings func() ([]byte, error)

	/*
		Get service status
	*/
	GetStatus func() Status

	/*
		Get service stats
	*/
	GetStats func() ([]byte, error)

	/*
		Last call before the service node's termination. Do the necessary clean up here.
	*/
	Terminate func() error

	/*
		GetVersion of the service,
		share.Version provide the initializer (version parser) and a toString convert,
		for details, please refer to share/version.go
	*/
	GetVersion func() shared.Version
}
