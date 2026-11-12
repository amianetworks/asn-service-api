// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

// Label is a node label of a root network, or of the platform (RootID
// commonapi.PlatformScope).
type Label struct {
	ID     string
	Name   string
	RootID string
	// Writable: the caller may rename it (see CreateLabel). Deletable:
	// Writable, and no rule or log parser binding refers to it. Read only.
	Writable, Deletable bool
	Rules               int // the event rules naming it; read only
	Bindings            int // the log parser bindings naming it; read only
}

// LabelsAPI manages node labels.
type LabelsAPI interface {
	// ListLabels returns the labels of a root and the platform labels its
	// nodes carry; any root.
	//
	// Errors: asn.network.not_found; shared.request.param_invalid (not a root).
	ListLabels(rootID string) ([]*Label, error)

	// ListPlatformLabels returns the platform labels.
	ListPlatformLabels() ([]*Label, error)

	// CreateLabel creates a label of rootID, or a platform label when
	// rootID is empty. A root label by a service with a node in the root;
	// a platform label by a super service. Names are unique in their scope.
	//
	// Errors: asn.network.not_found; asn.service.not_super;
	// shared.auth.permission_denied (no node in the root);
	// shared.resource.already_exists.
	CreateLabel(rootID, name string) (labelID string, err error)

	// RenameLabel renames a label; who may as CreateLabel.
	//
	// Errors: asn.label.not_found; as CreateLabel.
	RenameLabel(labelID, name string) error

	// DeleteLabel deletes a label no rule or binding refers to; who may as
	// CreateLabel.
	//
	// Errors: asn.label.not_found; asn.label.in_use (meta rules, bindings:
	// counts); as CreateLabel.
	DeleteLabel(labelID string) error

	// AddNodeLabels attaches labels to the calling service's nodes (a super
	// service: any node), at most 32 labels per node; platform labels by a
	// super service only. Already attached labels are kept.
	//
	// Errors: asn.node.not_found; asn.label.not_found;
	// asn.service.not_super; shared.request.param_invalid (a root label of
	// another root, over 32 labels).
	AddNodeLabels(nodeIDs, labelIDs []string) error

	// RemoveNodeLabels detaches labels; as AddNodeLabels.
	RemoveNodeLabels(nodeIDs, labelIDs []string) error

	// GetNodesByLabel returns the nodes carrying a label: the calling
	// service's own, every node for a super service.
	//
	// Errors: asn.label.not_found.
	GetNodesByLabel(labelID string) ([]*Node, error)
}
