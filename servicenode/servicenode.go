package servicenode

/*
	Struct used for service node and service communication
*/

// API provided by ASN Service Node
type API interface {
	GetServiceNodeNetif() (Netif, error)
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
		Initialize the service, this method will be called under a go routine, the return value to channel indicate service state:
		if error is nil, the service node will assign the state INITIALIZED to the service
		if error is NOT nil, the service node will assign the state MALFUNCTIONAL to the service

		Caution: the service node will have a timeout context (20s) to process the initialization,
				 if it cannot be done within 20s, service node will assign the state MALFUNCTIONAL to the service
	*/
	Init func(configPath string, c chan error)

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
		Service node is terminated, do the necessary clean up here
	*/
	CleanUp func()
}
