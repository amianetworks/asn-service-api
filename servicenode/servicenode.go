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
	Data       string
	Control    string
	Management string
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
		Initialize the service, the return value indicate service state:
		if error is nil, the service node will assign the state INITIALIZED to the service
		if error is NOT nil, the service node will assign the state MALFUNCTIONAL to the service
	*/
	Init func(configPath string) error

	/*
		Apply the configuration to the service, the return value indicate service state:
		the boolean true/false -> enabled/disabled, so the service node will assign the proper state (CONFIGED/INITIALIZED) to the service
		if error != nil, the service node will call Init function to reset the service
	*/
	ApplyServiceOps func(conf []byte) (bool, error)

	/*
		Read the current configuration of the service
	*/
	DumpSettings func() ([]byte, error)

	/*
		Get service status
	*/
	GetStatus func() ([]byte, error)

	/*
		Get service stats
	*/
	GetStats func() ([]byte, error)
}
