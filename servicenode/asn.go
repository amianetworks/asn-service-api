package snapi

import (
	"github.com/amianetworks/asn-service-api/v25/log"
)

/*
	Struct used for service node and service communication
*/

// ASNServiceNode provided by ASN Service Node for Service uses
type ASNServiceNode interface {
	/*
		Get ASN managed netifs from Service node
	*/
	GetServiceNodeNetif() (Netif, error)

	/*
		Send the metadata to the controller
	*/
	SendMetadataToController(serviceName string, metadata []byte) error

	/*
		Write the log to your service path. This is based on am.module logs
	*/
	GetLogger(serviceName string) (*log.Logger, error)

	/*
		Get ASN Service Node running Mode, currently support 'cluster', 'standalone' and 'hybrid' mode
	*/
	GetServiceNodeRunningMode() (string, error)

	/*
		Get ASN Service Node type, currently support 'server', 'appliance'
	*/
	GetServiceNodeType() string
}

/*
Netif Network interface struct
*/
type Netif struct {
	Data       []string
	Control    []string
	Management []string
	Other      []string
}
