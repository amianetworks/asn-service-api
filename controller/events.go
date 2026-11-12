// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"time"

	"asn.amiasys.com/asn-service-api/v27/iam"
)

// Events: services write points of the data items their catalog declares;
// event rules match points and raise incidents; notify rules mail the
// incidents to contact groups. Every rule, notify rule and contact group
// belongs to a namespace: a service's name, or "asn" for the framework's
// own data. A service writes its own namespace, and "asn" rules bound to
// it (Bound), which read only its nodes; a super service writes any.
//
// Names as the calling service sees them: its own data items by source
// (the catalog's CatalogItem.Name), every other item in full ("asn.host",
// "<service>.<source>"); an own item named in full is
// shared.request.param_invalid. The Service of a rule, incident, notify
// rule, contact group or send of its own namespace reads back empty;
// Service as input takes empty or the caller's own name for its own
// namespace. Bound keeps the service's name.

// Event levels, lowest first.
const (
	EventLevelInfo     = "info"
	EventLevelWarn     = "warn"
	EventLevelLow      = "low"
	EventLevelMedium   = "medium"
	EventLevelHigh     = "high"
	EventLevelCritical = "critical"
)

// Rule scopes.
const (
	EventScopeRoot     = "root"      // the rule's RootID
	EventScopeAllRoots = "all_roots" // every root
	EventScopePlatform = "platform"  // points of no root
)

// Rule device scopes.
const (
	EventDevicesAll   = "all"
	EventDevicesLabel = "label" // the nodes carrying LabelID
	EventDevicesNodes = "nodes" // NodeIDs
)

// Condition modes.
const (
	EventModeThreshold = "threshold" // the last value compared
	EventModeDuration  = "duration"  // compared true for Window
	EventModeAggregate = "aggregate" // Aggregate over Window compared
	EventModeCount     = "count"     // Count or more points within Window
	EventModeRatio     = "ratio"     // Percent of the points within Window compare true
	EventModeChange    = "change"    // Change over Window compared
	EventModeAbsence   = "absence"   // no point for Window
	EventModeInstant   = "instant"   // each point compared; a rule's only condition
)

// Handling of a rule's incidents.
const (
	EventHandleNone   = "none"   // recorded, nothing to handle
	EventHandleAuto   = "auto"   // closed when the conditions stop
	EventHandleManual = "manual" // closed by hand only
	EventHandleAny    = "any"    // either
)

// Incident states and close reasons.
const (
	IncidentOpen     = "open"
	IncidentClosed   = "closed"
	IncidentRecorded = "recorded"

	IncidentCloseCleared = "cleared"
	IncidentCloseManual  = "manual"
)

// EventPoint is a point of a declared data item. Tags and fields are those
// the item declares; ASN sets Service, RootID and NodeID ("-" for a point
// of no node). A node item's point carries the NodeID tag.
type EventPoint struct {
	Item   string // the source of an item of the caller's catalog
	Tags   map[string]string
	Fields map[string]any // numbers and strings, as declared
	Time   time.Time      // zero: when ASN takes it
}

// EventCondition is a condition of a rule. Op is gt, ge, lt, le, eq or ne;
// Values instead of Value makes it categorical on Tag or Field. Aggregate
// is avg, min, max, sum or last; Change is delta, rate or percent.
type EventCondition struct {
	Item      string              // the item read, when not the rule's; of the rule's namespace
	Tags      map[string][]string // narrows the item's points
	Mode      string
	Field     string
	Tag       string
	Op        string
	Value     float64
	Values    []string
	Window    time.Duration
	Aggregate string
	Count     int
	Percent   float64
	Change    string
	MaxGap    time.Duration
}

// EventTagFilter keeps (or with Exclude drops) the points whose tag is one
// of Values.
type EventTagFilter struct {
	Tag     string
	Values  []string
	Exclude bool
}

// EventLevelFrom takes an incident's level from a tag or field of the point
// that raised it.
type EventLevelFrom struct {
	Tag     string
	Field   string
	Map     map[string]string // value -> level
	Default string
}

// EventRule is an event rule. Service is its namespace, the caller when
// empty. Conditions all hold for an incident to start; GroupBy are the tags
// an incident is per (NodeID by default for a node item).
type EventRule struct {
	ID           string
	Name         string
	Enable       bool
	Service      string // namespace; the caller when empty, and empty when read back for the caller's own
	Scope        string
	RootID       string // EventScopeRoot
	Devices      string
	LabelID      string
	NodeIDs      []string
	Item         string // an item of the namespace: by source in the caller's own, in full in another
	Filters      []EventTagFilter
	Conditions   []EventCondition
	GroupBy      []string
	Level        string
	LevelFrom    *EventLevelFrom
	Handling     string
	ResolveAfter time.Duration
	Capture      []string
	// Bound is the service an "asn" rule reads the nodes of; set by ASN to
	// the caller for a service that is not super.
	Bound      string
	Builtin    bool // installed from a catalog's DefaultRules; not deleted
	CreateTime time.Time
	UpdateTime time.Time
}

// EventRuleFilter selects rules.
type EventRuleFilter struct {
	Service string // namespace; the caller when empty
	RootID  string // a root, or commonapi.PlatformScope; required
	Keyword string // the name contains it, case-insensitive
	Enable  *bool
	Items   []string // rules on these data items, named as EventRule.Item
	Scope   string   // EventScopeRoot, EventScopeAllRoots or EventScopePlatform
	Devices string   // EventDevicesAll, EventDevicesLabel or EventDevicesNodes
	NodeID  string   // rules whose devices include the node: all, its labels, or listed
}

// Incident is an incident of a rule.
type Incident struct {
	ID          string
	RuleID      string
	RuleName    string
	Service     string // namespace; empty for the caller's own
	RootID      string
	NodeID      string // "-" for an incident of no node
	Group       map[string]string
	Item        string // named as EventRule.Item
	Level       string
	State       string
	Handling    string
	Value       float64
	Detail      map[string]string
	StartTime   time.Time
	ClearTime   time.Time
	EndTime     time.Time
	CloseReason string
}

// IncidentFilter selects incidents.
type IncidentFilter struct {
	Service           string // namespace; the caller when empty
	RootID            string // a root, or commonapi.PlatformScope; or NetworkID
	NetworkID         string // a network: incidents of its nodes now; or RootID
	IncludeSubnetwork bool
	RuleID            string
	Levels            []string
	State             string
	Start, Stop       time.Time // when they started
	NodeIDs           []string  // incidents of these nodes
	Items             []string  // incidents of rules on these data items, named as EventRule.Item
	Keyword           string    // the rule name contains it, case-insensitive
}

// Notify rule fields.
const (
	NotifyLevelGE = "ge" // LevelOp: at or above Level
	NotifyLevelLE = "le" // at or below Level
	NotifyLevelEQ = "eq" // exactly Level

	NotifyOnOpen     = "open"     // SendOn: an incident opens
	NotifyOnClose    = "closed"   // an incident closes
	NotifyOnRecorded = "recorded" // a record-only incident is written
)

// NotifyRule mails the incidents it selects to contact groups. It selects
// either by level (Level, with LevelOp; every rule of the namespace and
// root) or by event rules (EventRuleIDs, at least one; Level and LevelOp
// empty), never both: both or neither is shared.request.param_invalid. A
// rule by event rules reads back with Level and LevelOp empty.
type NotifyRule struct {
	ID              string
	Name            string
	Enable          bool
	Service         string // namespace; the caller when empty, and empty when read back for the caller's own
	RootID          string
	LevelOp         string   // with Level: NotifyLevelGE (when empty), LE or EQ
	Level           string   // by level
	EventRuleIDs    []string // by event rules
	NodeIDs         []string // empty: every node
	SendOn          []string
	Cooldown        time.Duration
	MinInterval     time.Duration
	ContactGroupIDs []string // of the same namespace and root
	CreateTime      time.Time
	UpdateTime      time.Time
}

// ContactGroup is a group of mail recipients: address -> name.
type ContactGroup struct {
	ID         string
	Name       string
	Service    string // namespace; the caller when empty, and empty when read back for the caller's own
	RootID     string
	To, Cc     map[string]string
	CreateTime time.Time
	UpdateTime time.Time
}

// DeliveryFilter selects the sends of notify rules of a namespace and root.
type DeliveryFilter struct {
	Service     string // namespace; the caller when empty
	RootID      string // a root or commonapi.PlatformScope; required
	RuleID      string // one notify rule; empty: all
	Start, Stop time.Time
}

// Delivery statuses.
const (
	DeliverySent            = "sent"
	DeliveryFailed          = "failed"
	DeliveryDropped         = "dropped"
	DeliveryKindUnavailable = "kind_unavailable" // mail is not configured
)

// DeliveryMail is one mail of a send.
type DeliveryMail struct {
	Addresses []string
	Status    string
	ErrorCode string // why it was not sent
	Reason    string
	Time      time.Time
}

// Dispatch is one send of a notify rule.
type Dispatch struct {
	ID          string
	RuleID      string
	RuleName    string // as it was when it sent
	Service     string // namespace; empty for the caller's own
	RootID      string
	IncidentIDs []string
	Omitted     int    // incidents past the limit of one mail, counted but not listed
	Error       string // why the send failed before any mail
	Time        time.Time
	Mails       []DeliveryMail
}

// EventsAPI writes points and manages rules, incidents, notify rules and
// contact groups. A service that is not super reads and writes its own
// namespace, and in "asn" the rules bound to it and the incidents of its
// nodes. Lists are paged by page (Limit 1–500).
//
// Errors common to the methods: asn.event.namespace (another namespace);
// asn.service.not_super; shared.resource.not_found (an unknown or
// unreadable ID); asn.network.not_found; shared.request.param_invalid;
// shared.system.unavailable (events are not open).
type EventsAPI interface {
	// WriteEventPoints writes points of the calling service's declared
	// items. rootID is the root of points with no node, empty for the
	// platform scope; a node point takes its node's root.
	//
	// Errors: asn.event.undeclared_data; asn.event.type_mismatch.
	WriteEventPoints(rootID string, points []EventPoint) error

	// ListEventRules lists the rules filter selects.
	ListEventRules(filter EventRuleFilter, page iam.Page) (rules []*EventRule, total int, err error)
	// GetEventRule returns a rule.
	GetEventRule(ruleID string) (*EventRule, error)
	// CreateEventRule creates a rule and returns it with its ID.
	//
	// Errors: asn.event.undeclared_data; asn.event.type_mismatch;
	// asn.event.owner_unavailable (on-demand data whose owner is not loaded).
	CreateEventRule(rule EventRule) (*EventRule, error)
	// UpdateEventRule replaces a rule (rule.ID).
	UpdateEventRule(rule EventRule) (*EventRule, error)
	// DeleteEventRule deletes a rule.
	//
	// Errors: asn.event.rule_in_use (a notify rule lists it);
	// asn.event.rule_builtin.
	DeleteEventRule(ruleID string) error
	// SetEventRuleEnable turns a rule on or off.
	SetEventRuleEnable(ruleID string, enable bool) error

	// ListIncidents lists the incidents filter selects, newest first.
	ListIncidents(filter IncidentFilter, page iam.Page) (incidents []*Incident, total int, err error)
	// GetIncident returns an incident.
	GetIncident(incidentID string) (*Incident, error)
	// ResolveIncident closes an open incident of a rule with handling
	// manual or any.
	//
	// Errors: shared.resource.state_invalid.
	ResolveIncident(incidentID string) (*Incident, error)
	// SubscribeIncidents returns a channel of the incidents the caller reads
	// as they start, change level or close. One-shot. Delivery is best
	// effort: an incident that finds the channel full is dropped; reconcile
	// with ListIncidents.
	SubscribeIncidents() (<-chan *Incident, error)

	// ListNotifyRules lists the notify rules of a namespace (the caller when
	// empty) and root (required; commonapi.PlatformScope for the platform).
	ListNotifyRules(service, rootID string, page iam.Page) (rules []*NotifyRule, total int, err error)
	GetNotifyRule(ruleID string) (*NotifyRule, error)
	CreateNotifyRule(rule NotifyRule) (*NotifyRule, error)
	UpdateNotifyRule(rule NotifyRule) (*NotifyRule, error)
	DeleteNotifyRule(ruleID string) error
	SetNotifyRuleEnable(ruleID string, enable bool) error
	// ListDeliveries lists the sends of the notify rules filter selects,
	// newest first, in the namespaces ListNotifyRules reads.
	ListDeliveries(filter DeliveryFilter, page iam.Page) (dispatches []*Dispatch, total int, err error)

	// ListContactGroups lists the contact groups of a namespace (the caller
	// when empty) and root (required).
	ListContactGroups(service, rootID string, page iam.Page) (groups []*ContactGroup, total int, err error)
	CreateContactGroup(group ContactGroup) (*ContactGroup, error)
	UpdateContactGroup(group ContactGroup) (*ContactGroup, error)
	// DeleteContactGroup deletes a group no notify rule uses.
	//
	// Errors: shared.resource.in_use.
	DeleteContactGroup(groupID string) error
}
