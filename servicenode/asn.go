// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package snapi

import (
	"github.com/amianetworks/asn-service-api/v25/log"
)

// Structs used for service node and service communication

// ASNServiceNode provided by ASN Service Node for Service uses
type ASNServiceNode interface {
	// GetServiceNodeNetif returns ASN managed netifs from Service node.
	GetServiceNodeNetif() (Netif, error)

	// SendMetadataToController sends the metadata to the controller.
	SendMetadataToController(metadata []byte) error

	// GetLogger returns the logger for this service.
	GetLogger() (*log.Logger, error)

	// GetServiceNodeType returns the ASN Service Node type, currently support 'server', 'appliance'.
	GetServiceNodeType() string
}

// Netif is the structure for network interface.
type Netif struct {
	Data       []string
	Control    []string
	Management []string
	Other      []string
}
