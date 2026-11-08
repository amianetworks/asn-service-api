// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import commonapi "asn.amiasys.com/asn-service-api/v27/common"

// NodeAttributes are node attributes a service sets; a nil field is left
// unchanged.
type NodeAttributes struct {
	Name        *string // non-empty, unique in the root
	Description *string
	Location    *commonapi.Location
}

// Node risk levels: the highest level of the node's incidents open now or
// started within the last 24 hours.
const (
	RiskNone     = "none"
	RiskInfo     = "info"
	RiskWarn     = "warn"
	RiskLow      = "low"
	RiskMedium   = "medium"
	RiskHigh     = "high"
	RiskCritical = "critical"
)

// NodeRisk is a node's risk.
type NodeRisk struct {
	Level string // a Risk* value
	Open  int    // the node's open incidents
}

// NodeAPI sets node attributes and reads node risk.
type NodeAPI interface {
	// SetNodeAttributes sets the name, description and location of a node
	// the calling service is on (a super service: any node). Audited.
	//
	// Errors: asn.node.not_found; asn.service.not_super;
	// shared.request.param_invalid (an empty name, an invalid location);
	// asn.node.already_exists (the name is taken in the root).
	SetNodeAttributes(nodeID string, attrs NodeAttributes) error

	// NodeRisks returns the risk of nodes the calling service is on (a
	// super service: any node), by node ID, over every namespace.
	//
	// Errors: asn.node.not_found; asn.service.not_super.
	NodeRisks(nodeIDs []string) (map[string]NodeRisk, error)
}
