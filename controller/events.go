// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"io"
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
//
// The lists of rules, incidents, notify rules and contact groups read
// namespaces by Services: nil or empty is the caller's own, an element ""
// or the caller's name is its own, and repeats are dropped; a namespace the
// caller cannot read is asn.event.namespace. The namespaces are sorted,
// paged and counted as one list. ListDeliveries reads one namespace
// (DeliveryFilter.Service).
//
// A reference to a label, node, event rule or contact group is a Ref: a
// write reads its ID only; a read sets Name to the object's name now, empty
// when the object no longer exists.

// Writable, Deletable and Resolvable, on the objects a list shows with
// actions (rules, notify rules, contact groups, incidents, labels, log
// parsers), say what the calling service may do with the object now: ASN
// sets them with the checks the write itself makes (access, built in,
// references, state), as of the read, on what reads and writes return.
// They are read only and ignored on write; the write still decides, and
// Resolvable leaves out whether an auto incident's rule still judges it.
// Which of the service's own users may act is the service's to decide. A
// reference count beside them (NotifyRules, Label.Rules, Label.Bindings)
// has the name the refusal's errMeta uses, and says why an object is not
// Deletable.

// Ref names another object: its ID, and its name when read.
type Ref struct {
	ID   string
	Name string
}

// Event levels, lowest first.
const (
	EventLevelInfo     = "info"
	EventLevelWarn     = "warn"
	EventLevelLow      = "low"
	EventLevelMedium   = "medium"
	EventLevelHigh     = "high"
	EventLevelCritical = "critical"
)

// Rule device scopes.
const (
	EventDevicesAll   = "all"
	EventDevicesLabel = "label" // the nodes carrying Label
	EventDevicesNodes = "nodes" // Nodes
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

// Handling of a rule's incidents. A rule of one instant condition is
// handled none or manual (empty is none): its incident is one point, with
// no condition that stops.
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
	RootID       string // a root, or commonapi.PlatformScope for the points of no root
	Devices      string
	Label        *Ref   // EventDevicesLabel
	Nodes        []Ref  // EventDevicesNodes; nodes of RootID
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
	Bound   string
	Builtin bool // installed from a catalog's DefaultRules; not deleted
	// Writable: the caller may update, enable or disable the rule.
	// Deletable: Writable, not Builtin and NotifyRules 0. Read only.
	Writable, Deletable bool
	NotifyRules         int // the notify rules naming the rule; read only
	CreateTime          time.Time
	UpdateTime          time.Time
}

// EventRuleFilter selects rules. A service that is not super reads, in
// "asn", the rules bound to it.
type EventRuleFilter struct {
	Services []string // namespaces; see the package comment
	RootID   string   // a root, or commonapi.PlatformScope; required
	Keyword  string   // the name contains it, case-insensitive
	Enable   *bool
	Items    []string // rules on these data items, named as EventRule.Item
	NodeID   string   // rules whose devices include the node: all, its labels, or listed
	// Levels selects the rules of a fixed Level among them; a rule whose
	// level comes from its data (LevelFrom) matches none.
	Levels  []string
	Builtin *bool // nil: either
}

// Incident is an incident of a rule.
type Incident struct {
	ID          string
	RuleID      string
	RuleName    string
	Service     string // namespace; empty for the caller's own
	RootID      string
	NodeID      string // "-" for an incident of no node
	NodeName    string // the node's name now; empty for "-" or a node no longer there
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
	// Resolvable: the caller may close it by hand (ResolveIncident) and it
	// is open. Read only.
	Resolvable bool
}

// IncidentTotalMax caps the total ListIncidents counts (a total of it is
// that many or more) and the incidents ExportIncidents writes.
const IncidentTotalMax = 10000

// IncidentFilter selects incidents.
type IncidentFilter struct {
	Services          []string // namespaces; see the package comment
	RootID            string   // a root, or commonapi.PlatformScope; or NetworkID
	NetworkID         string   // a network: incidents of its nodes now; or RootID
	IncludeSubnetwork bool
	RuleID            string
	Levels            []string
	States            []string  // any of IncidentOpen, IncidentClosed, IncidentRecorded
	IDs               []string  // these incidents; at most 1000
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
// root) or by event rules (EventRules, at least one; Level and LevelOp
// empty), never both: both or neither is shared.request.param_invalid. A
// rule by event rules reads back with Level and LevelOp empty.
type NotifyRule struct {
	ID            string
	Name          string
	Enable        bool
	Service       string // namespace; the caller when empty, and empty when read back for the caller's own
	RootID        string
	LevelOp       string // with Level: NotifyLevelGE (when empty), LE or EQ
	Level         string // by level
	EventRules    []Ref  // by event rules
	Nodes         []Ref  // empty: every node
	SendOn        []string
	Cooldown      time.Duration
	MinInterval   time.Duration
	ContactGroups []Ref // of the same namespace and root
	// Writable: the caller may update, enable or disable it; Deletable is
	// Writable (nothing refers to a notify rule). Read only.
	Writable, Deletable bool
	CreateTime          time.Time
	UpdateTime          time.Time
}

// NotifyRuleFilter selects notify rules.
type NotifyRuleFilter struct {
	Services       []string // namespaces; see the package comment
	RootID         string   // a root, or commonapi.PlatformScope; required
	Keyword        string   // the name contains it, case-insensitive
	Enable         *bool
	ContactGroupID string // the rules mailing this group
}

// ContactGroup is a group of mail recipients: address -> name.
type ContactGroup struct {
	ID      string
	Name    string
	Service string // namespace; the caller when empty, and empty when read back for the caller's own
	RootID  string
	To, Cc  map[string]string
	// NotifyRules counts the notify rules naming the group; a group with any
	// cannot be deleted. Set when read and in what a write returns; ignored
	// on write.
	NotifyRules int
	// Writable: the caller may update it. Deletable: Writable and
	// NotifyRules 0. Read only.
	Writable, Deletable bool
	CreateTime          time.Time
	UpdateTime          time.Time
}

// ContactGroupFilter selects contact groups.
type ContactGroupFilter struct {
	Services []string // namespaces; see the package comment
	RootID   string   // a root, or commonapi.PlatformScope; required
	// Keyword is contained, case-insensitive, in the name or in an address
	// of To or Cc.
	Keyword string
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
// nodes, except the incidents of rules on asn.audit: GetIncident and
// ResolveIncident of one are shared.resource.not_found, and
// SubscribeIncidents does not send them. It reads no "asn" notify rules or
// contact groups. A super service reads every namespace. Lists are paged by
// page (Limit 1–500).
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
	// Errors: asn.event.rule_in_use (notify rules name it; errMeta.notify_rules
	// is their IDs, comma-separated); asn.event.rule_builtin.
	DeleteEventRule(ruleID string) error
	// SetEventRuleEnable turns a rule on or off.
	SetEventRuleEnable(ruleID string, enable bool) error

	// ListIncidents lists the incidents filter selects, newest first.
	//
	// Errors: shared.request.param_invalid (errMeta.field states or ids) for
	// an unknown state, more than 1000 IDs or one that is not an incident ID.
	ListIncidents(filter IncidentFilter, page iam.Page) (incidents []*Incident, total int, err error)
	// ExportIncidents writes the incidents filter selects to w as "csv" or
	// "jsonl", newest first, at most IncidentTotalMax; truncated reports
	// that more matched. The columns are id, startTime, endTime, level,
	// state, ruleId, ruleName, service, rootId, root, nodeId, node, item,
	// value, detail, closeReason. The export is audited. A refusal is
	// returned before anything is written.
	//
	// Errors: as ListIncidents; shared.request.param_invalid (errMeta.field
	// format) for another format.
	ExportIncidents(filter IncidentFilter, format string, w io.Writer) (rows int, truncated bool, err error)
	// GetIncident returns an incident.
	GetIncident(incidentID string) (*Incident, error)
	// ResolveIncident closes an open incident by hand. An auto incident is
	// closed only while its rule does not judge it: a condition has no
	// point to judge, or the rule was deleted, disabled or regrouped.
	//
	// Errors: shared.resource.state_invalid (not open, or an auto incident
	// its rule still judges).
	ResolveIncident(incidentID string) (*Incident, error)
	// SubscribeIncidents returns a channel of the incidents the caller reads
	// as they start, change level or close. One-shot. Delivery is best
	// effort: an incident that finds the channel full is dropped; reconcile
	// with ListIncidents.
	SubscribeIncidents() (<-chan *Incident, error)

	// ListNotifyRules lists the notify rules filter selects.
	ListNotifyRules(filter NotifyRuleFilter, page iam.Page) (rules []*NotifyRule, total int, err error)
	GetNotifyRule(ruleID string) (*NotifyRule, error)
	CreateNotifyRule(rule NotifyRule) (*NotifyRule, error)
	UpdateNotifyRule(rule NotifyRule) (*NotifyRule, error)
	DeleteNotifyRule(ruleID string) error
	SetNotifyRuleEnable(ruleID string, enable bool) error
	// ListDeliveries lists the sends of the notify rules filter selects,
	// newest first, in a namespace ListNotifyRules reads.
	ListDeliveries(filter DeliveryFilter, page iam.Page) (dispatches []*Dispatch, total int, err error)

	// ListContactGroups lists the contact groups filter selects.
	ListContactGroups(filter ContactGroupFilter, page iam.Page) (groups []*ContactGroup, total int, err error)
	CreateContactGroup(group ContactGroup) (*ContactGroup, error)
	UpdateContactGroup(group ContactGroup) (*ContactGroup, error)
	// DeleteContactGroup deletes a group no notify rule uses
	// (ContactGroup.NotifyRules is 0).
	//
	// Errors: shared.resource.in_use (errMeta.notify_rules is the IDs of the
	// notify rules naming it, comma-separated).
	DeleteContactGroup(groupID string) error
}
