// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import commonapi "asn.amiasys.com/asn-service-api/v26/common"

// EnrollmentAPI is the framework's service-agnostic node onboarding surface,
// embedded in ASNController. A service uses it to create framework-owned node
// identities, mint single-use enrollment tokens, render bootstrap scripts,
// unbind nodes for re-enrollment, and permanently delete nodes.
//
// The service entry is single-service: a service enrolls nodes only for itself.
// CreateNode therefore takes no service list, and the token / script methods take
// no service name — the node's service_names (here, exactly the calling service)
// is the install set. Install specs (per-service deb coordinates and asnsn) are
// configured in the controller's asn.conf, so there is no RegisterNodeInstallSpec.
//
// The framework owns node-key minting, lazy certificate signing, and ASN-core
// rendering; the service never mints keys, signs certificates, or re-renders the
// core. It only relays the returned bytes (optionally wrapping them with its own
// service-owned config layer in the service-entry flow).
//
// Ownership and trust tier are deferred in this version: every node is
// framework-owned. When added later they become additive fields on the request
// and identity structs below.
type EnrollmentAPI interface {
	// CreateNode creates a persistent, framework-owned node identity and returns
	// it; service_names is set by the framework to the calling service. For a new
	// node it mints no key, issues no token, and renders no script — the network
	// path is derived from the parent placement and the service installs at
	// bootstrap. The node is placed under ParentNetworkID, but NodeName must be
	// unique across the entire root network tree that parent belongs to; an
	// existing node is matched by that name within the tree. If such a node
	// already exists: with AllowExisting the framework adds the calling service to
	// it (acting by runtime state — a live install of deb + .so + Init() if the
	// node is online, otherwise appended to service_names to install when the node
	// next bootstraps) and returns the existing identity; without AllowExisting a
	// name collision returns an error.
	// On the AllowExisting path UpdateInfo governs the node's shared attributes:
	// when set, the request's Type and Label overwrite the existing values (which
	// affects every service on the node); when unset they are left unchanged.
	CreateNode(req CreateNodeRequest) (*NodeIdentity, error)

	// MintEnrollmentToken issues a single-use, script-fetch token bound to an
	// EXISTING node. Allowed only when the node is EnrollmentStateUnbound (no
	// valid certificate); enrollment is non-reentrant. A fresh token supersedes a
	// prior unused token. To re-enroll a bound node, call UnbindNode first. Does
	// not create a node.
	MintEnrollmentToken(req MintTokenRequest) (*EnrollmentToken, error)

	// UnbindNode revokes the node's current certificate and cancels any
	// outstanding token, returning the node to EnrollmentStateUnbound so it can
	// enroll again. It does NOT delete the node: identity, service eligibility,
	// and service config are preserved. A bound, live node loses its session
	// immediately. Access-sensitive; audited.
	UnbindNode(req UnbindNodeRequest) (*NodeIdentity, error)

	// DeleteNode permanently destroys a framework-owned node identity. It is
	// allowed only when the calling service is the last service on the node
	// (service_names contains no other service); otherwise it returns an error,
	// so one service can never tear down a node another service still uses. The
	// calling service is torn down first (Stop() + Finish() + unload on an online
	// node, as DeleteServiceFromNode), then the identity is destroyed: certificate
	// revoked, node key deleted, node-group membership dropped, and node config
	// and any outstanding enrollment token discarded. Contrast UnbindNode, which
	// keeps the identity for re-enrollment. Access-sensitive; audited.
	DeleteNode(req DeleteNodeRequest) error

	// RenderBootstrapScript renders the FULL install script for the EXISTING node
	// bound to the token: validates and consumes the single-use token, mints the
	// (unpersisted) node key, lazily signs the node certificate, and renders asnsn
	// plus the deb of every service in the node's service_names (install specs
	// from asn.conf). The script is idempotent. The service serves the returned
	// bytes itself. Never creates a node.
	RenderBootstrapScript(req RenderScriptRequest) (*BootstrapScript, error)

	// GetEnrollmentStatus reads the current enrollment state for a node or token.
	GetEnrollmentStatus(ref EnrollmentRef) (*EnrollmentStatus, error)
}

// CreateNodeRequest creates a framework-owned node identity, or (with
// AllowExisting) adds the calling service to one that already exists with this
// NodeName in the parent's root network tree. service_names is fixed by the
// framework to the calling service (the service entry is single-service).
// Ownership and trust tier are deferred; when added they become additive fields
// here.
type CreateNodeRequest struct {
	ParentNetworkID string             // placement; the node's network path is derived from it
	NodeName        string             // authoritative; unique across the parent's whole root network tree
	Type            commonapi.NodeType // hardware/logical role, e.g. NodeTypeServer, NodeTypeAppliance
	Label           string
	// AllowExisting makes CreateNode add the calling service to a node that
	// already exists with this NodeName in the parent's root network tree instead
	// of failing on the name collision. The framework acts by runtime state: online ->
	// live install (deb + .so + Init(), as AddServiceToNode); not yet online ->
	// appended to service_names and installed at the node's next bootstrap.
	AllowExisting bool
	// UpdateInfo applies only on the AllowExisting path: when true, Type and Label
	// in this request overwrite the existing node's values; when false they are
	// ignored and the node's attributes are left unchanged. Type and Label are
	// node-level (shared across every service on the node), so an overwrite by a
	// joining service affects the others too — set it deliberately. It has no
	// effect when a new node is created (Type and Label are always taken then).
	UpdateInfo bool
}

// NodeIdentity is the persisted node identity returned by CreateNode / UnbindNode.
type NodeIdentity struct {
	NodeID          string
	ServiceNames    []string
	EnrollmentState commonapi.EnrollmentState
}

// MintTokenRequest mints a single-use enrollment token for an existing node.
type MintTokenRequest struct {
	NodeID     string // existing node the token enrolls; required
	TTLSeconds int64
	Label      string
}

// UnbindNodeRequest revokes a node's certificate and reopens it for enrollment.
type UnbindNodeRequest struct {
	NodeID string // required
	Reason string // audit reason (e.g. "machine swap", "lost key")
}

// DeleteNodeRequest permanently destroys a node identity. Allowed only when the
// calling service is the node's last service.
type DeleteNodeRequest struct {
	NodeID string // required
	Reason string // audit reason (e.g. "decommissioned")
}

// EnrollmentToken is the single-use script-fetch credential bound to a node.
type EnrollmentToken struct {
	Token     string
	TokenID   string
	NodeID    string
	ExpiresAt int64
}

// RenderScriptRequest renders the bootstrap script for the token's node.
type RenderScriptRequest struct {
	Token string // presented by the device to the service
}

// BootstrapScript is the rendered ASN-core install script (asnsn + the node's
// service debs).
type BootstrapScript struct {
	Content      []byte
	ContentType  string // e.g. "text/x-shellscript"
	NodeID       string // the existing node this script enrolls / re-keys
	CertNotAfter int64  // validity of the lazily signed certificate
}

// EnrollmentRef identifies an enrollment by node or token.
type EnrollmentRef struct {
	NodeID  string
	TokenID string
}

// EnrollmentStatus is the current enrollment + runtime view of a node.
type EnrollmentStatus struct {
	NodeID          string
	TokenID         string
	EnrollmentState commonapi.EnrollmentState
	NodeState       commonapi.NodeState // runtime connectivity
	LastEventAt     int64
}
