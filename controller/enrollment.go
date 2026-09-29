// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import commonapi "asn.amiasys.com/asn-service-api/v26/common"

// EnrollmentAPI is the framework's service-agnostic node onboarding surface,
// embedded in ASNController. A service uses it to create framework-owned node
// identities, mint enrollment tokens, render install and uninstall scripts,
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
// Binding is deferred to registration. Every install script carries a freshly
// signed candidate certificate for the node; the first candidate to complete
// registration becomes the node's one bound certificate and every other
// candidate is invalidated. A host that already holds a certificate for the node
// keeps it and ignores the candidate, so a credentialed node can re-fetch its
// script — to install a service deb added after enrollment, for instance —
// without unbinding and without rotating its key. See RenderBootstrapScript.
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

	// MintEnrollmentToken issues a script-fetch token bound to an EXISTING node,
	// in any EnrollmentState. A fresh token supersedes the node's prior token.
	// Does not create a node.
	//
	// The token is reusable until it expires: every fetch renders a script. Its
	// lifetime is TTLSeconds, capped by the controller's configured token lifetime
	// (servicenode.enrollment.token_ttl), which is also the default when
	// TTLSeconds is zero.
	//
	// A token authorizes fetching the node's install script; it never displaces a
	// running node. A candidate certificate fetched while the node already holds a
	// bound certificate can never bind, which is why minting is unrestricted.
	// Binding is non-reentrant, enforced at registration: to bind a NEW
	// certificate for a node (a lost key, a replaced machine) call UnbindNode
	// first.
	MintEnrollmentToken(req MintTokenRequest) (*EnrollmentToken, error)

	// UnbindNode revokes the node's bound certificate, invalidates every
	// unbound candidate certificate, and cancels any outstanding token, returning
	// the node to EnrollmentStateUnbound so it can enroll again. It does NOT
	// delete the node: identity, service eligibility, and service config are
	// preserved. A bound, live node loses its session immediately.
	// Access-sensitive; audited.
	//
	// This is the only way to re-key a node: only a node with no bound
	// certificate lets a candidate bind at registration. A host that still holds
	// the revoked certificate keeps it (install scripts never overwrite a local
	// certificate), so re-enrolling that same host requires removing its local
	// certificate first.
	UnbindNode(req UnbindNodeRequest) (*NodeIdentity, error)

	// DeleteNode removes the calling service from the node and, only when that
	// service was the node's last, optionally destroys the node identity. The
	// service is always torn down (Stop() + Finish() + unload on an online node,
	// as DeleteServiceFromNode). If other services remain, only the calling
	// service is removed and the node is kept — one service can never tear down a
	// node another service still uses. If the calling service was the last one,
	// DeleteEmptyNode decides the node's fate: true destroys the identity
	// (certificate revoked, node key deleted, node-group membership dropped, node
	// config and any outstanding token discarded); false keeps the now-serviceless
	// node. Contrast UnbindNode, which keeps a bound identity for re-enrollment.
	// Access-sensitive; audited.
	DeleteNode(req DeleteNodeRequest) error

	// RenderBootstrapScript renders the install/upgrade script for the EXISTING
	// node bound to the token: asnsn plus the CALLING service's deb only, at the
	// versions and apt repos configured in asn.conf (debs of other services the
	// node is eligible for belong to the ASN entry). A service that is not itself
	// eligible for the node is refused before anything is signed. The service
	// serves the returned bytes itself. Never creates a node, never consumes the
	// token.
	//
	// Every call mints a fresh (unpersisted) node key and signs a candidate
	// certificate, embeds both, and records the candidate on the node
	// (EnrollmentStateCertIssued until one binds). The first candidate to
	// complete registration binds; the others are invalidated. The number of live
	// candidates per node is capped by the controller; beyond it the call fails
	// and signs nothing.
	//
	// The script decides on the host, before it mutates anything:
	//
	//   - Credential: no local certificate -> install the embedded candidate and
	//     asn.conf; a local certificate for this node -> keep it and asn.conf,
	//     ignore the candidate; a local certificate for another node -> abort.
	//   - Packages: not installed -> install; lower version -> upgrade; equal ->
	//     leave; any installed version HIGHER than the target -> abort the whole
	//     script without changing anything.
	//
	// The script is idempotent and safe to re-run with any unexpired token.
	RenderBootstrapScript(req RenderScriptRequest) (*BootstrapScript, error)

	// RenderUninstallScript removes the CALLING service from the node's
	// service_names (as DeleteServiceFromNode, so a live node unloads it at once)
	// and renders a script that purges the service's deb on the host and restarts
	// asnsn. When that leaves service_names empty, it also unbinds the node (as
	// UnbindNode) and the script purges asnsn and removes the node certificate,
	// key, and asn.conf. The decision is made from the controller's record, never
	// by inspecting the host. The node record itself is kept; DeleteNode remains
	// the way to destroy it.
	//
	// The controller record changes when the script is rendered, not when it
	// runs; the script only brings the host in line. It carries no key,
	// certificate, or token, and before mutating anything it verifies that the
	// host holds this node's certificate (and, when the node was bound, the bound
	// serial). The calling service must be eligible for the node at call time, so
	// a retry after a successful render goes through the operator entry.
	// Access-sensitive; audited.
	RenderUninstallScript(req RenderUninstallRequest) (*UninstallScript, error)

	// GetEnrollmentStatus reads the current enrollment state for a node or token.
	GetEnrollmentStatus(ref EnrollmentRef) (*EnrollmentStatus, error)
}

// CreateNodeRequest creates a framework-owned node identity, or (with
// AllowExisting) adds the calling service to one that already exists with this
// NodeName in the parent's root network tree. service_names is fixed by the
// framework to the calling service (the service entry is single-service).
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

// MintTokenRequest mints an enrollment token for an existing node.
type MintTokenRequest struct {
	NodeID     string // existing node the token enrolls; required
	TTLSeconds int64  // 0 = controller default; capped by servicenode.enrollment.token_ttl
	Label      string
}

// UnbindNodeRequest revokes a node's certificate and reopens it for enrollment.
type UnbindNodeRequest struct {
	NodeID string // required
	Reason string // audit reason (e.g. "machine swap", "lost key")
}

// DeleteNodeRequest removes the calling service from a node and, when it was the
// node's last service, optionally destroys the node identity (see DeleteEmptyNode).
type DeleteNodeRequest struct {
	NodeID string // required
	Reason string // audit reason (e.g. "decommissioned")
	// DeleteEmptyNode applies only when the calling service is the node's last
	// service: true destroys the node identity, false keeps the now-serviceless
	// node. It is ignored when other services remain (the node is always kept).
	DeleteEmptyNode bool
}

// EnrollmentToken is the script-fetch credential bound to a node. It is never
// consumed: any number of fetches may use it until ExpiresAt.
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

// BootstrapScript is the rendered ASN-core install script (asnsn + the calling
// service's deb). It always carries a freshly signed candidate certificate and
// key; see RenderBootstrapScript.
type BootstrapScript struct {
	Content     []byte
	ContentType string // e.g. "text/x-shellscript"
	NodeID      string // the existing node this script enrolls / re-keys
	// CertNotAfter is when the candidate certificate carried in Content expires.
	// The host uses that certificate only if it has none of its own, and it binds
	// only if it is the first candidate to register.
	CertNotAfter int64
}

// RenderUninstallRequest renders the uninstall script that removes the calling
// service from a node.
type RenderUninstallRequest struct {
	NodeID string // required
}

// UninstallScript is the rendered uninstall script. It carries no secret.
type UninstallScript struct {
	Content     []byte
	ContentType string // e.g. "text/x-shellscript"
	NodeID      string
	// LastService reports that the calling service was the node's last: the node
	// was unbound and the script also purges asnsn and the node credential.
	LastService bool
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
