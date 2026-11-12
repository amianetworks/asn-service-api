// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"io"
	"time"

	commonapi "asn.amiasys.com/asn-service-api/v27/common"
	"asn.amiasys.com/asn-service-api/v27/iam"
)

// AuditFilter selects audit records; set conditions must all match, a list
// matches any of its values.
type AuditFilter struct {
	Service string // namespace acted on; "asn" for framework data
	Caller  string // calling service; "asn" for ASN Web and ASN itself

	RootIDs           []string // roots or commonapi.PlatformScope
	NetworkID         string   // a network that is not a root: records of its nodes now
	IncludeSubnetwork bool     // with NetworkID: also of the networks below it

	NodeIDs      []string
	SubjectIDs   []string // any of these subjects
	SubjectTypes []commonapi.AuditSubjectType
	// Usernames selects the records of any of these account names, as
	// recorded when the record was written; exact match.
	Usernames []string
	// EventTypes, Operations and ObjectTypes are names and codes of one
	// service's audit catalog: they need Service.
	EventTypes  []string
	Operations  []int64
	Results     []commonapi.AuditResult
	RiskLevels  []commonapi.AuditRiskLevel
	ObjectTypes []string
	ObjectIDs   []string // these objects, as AuditRecord.ObjectID

	// DeviceID selects the records of one device (AuditEvent.DeviceID).
	DeviceID string

	// Tags select by a service's own dimensions: a key matches any of its
	// values, and every key must match; a key with no values sets no
	// condition. Tags need Service, and every key must be one that service
	// declares in AuditCatalog.TagKeys; Service, Caller, RootID and NodeID
	// are refused here (they have their own fields).
	Tags map[string][]string

	Start time.Time // [Start, Stop), both required, at most 92 days
	Stop  time.Time
}

// AuditRecord is one stored audit record.
type AuditRecord struct {
	commonapi.AuditEvent
	Service string
	Caller  string
	RootID  string // a root ID or commonapi.PlatformScope
	Origin  string // "local", or the node ID it was reported from
	// Username is the subject account's name when the record was written;
	// empty for a subject that is not an account.
	Username string
	NodeName string // NodeID's name now; empty for none or a node no longer there
}

// AuditCount counts records.
type AuditCount struct {
	Events int // not alarms
	Alarms int
}

// AuditSubjectCount is a subject's alarm count.
type AuditSubjectCount struct {
	SubjectID   string
	SubjectType commonapi.AuditSubjectType // of its latest alarm
	Alarms      int
}

// AuditTrendBucket counts the records of [Start, Stop).
type AuditTrendBucket struct {
	Start  time.Time
	Stop   time.Time
	Events int // not alarms
	Alarms int
}

// AuditSummary summarizes the records a filter selects.
type AuditSummary struct {
	ByLevel             map[commonapi.AuditRiskLevel]int
	EventsByType        map[string]int
	AlarmsByType        map[string]int
	EventsBySubjectType map[commonapi.AuditSubjectType]int
	AlarmsBySubjectType map[commonapi.AuditSubjectType]int
	TopAlarmSubjects    []AuditSubjectCount
	ByService           map[string]AuditCount
	ByCaller            map[string]AuditCount
	// ByTag counts, when the filter sets Service, the records of each value
	// of each key that service declares in AuditCatalog.TagKeys; records
	// without the key are not counted under it. Empty without Service.
	ByTag map[string]map[string]AuditCount
	Trend []AuditTrendBucket
}

// AuditCatalog names a service's audit records for display and declares the
// tag keys its records are queried by. Writes are not checked against it.
type AuditCatalog struct {
	EventTypes  []AuditEventType
	Operations  []AuditOperation
	ObjectTypes []AuditObjectType

	// TagKeys are the keys of AuditEvent.Tags the service's records are
	// filtered (AuditFilter.Tags) and summed up (AuditSummary.ByTag) by, with
	// their display names and, optionally, the values to offer in a picker:
	// at most AuditTagKeysMax. They are not indexed.
	TagKeys []AuditTagKey
}

// AuditTagKey is a declared audit dimension.
type AuditTagKey struct {
	// Key is an upper-case letter followed by letters, digits and '_', none
	// of AuditReservedTagKeys, unique in the catalog.
	Key string
	// Names are the display names by locale ("zh-CN", "en-US"); at least one.
	Names map[string]string
	// Values are the values to offer, optional, at most AuditTagValuesMax,
	// each set and unique. They are for display only: neither written tags
	// nor AuditFilter.Tags are checked against them.
	Values []AuditTagValue
}

// AuditTagValue is a value of a dimension with its display names.
type AuditTagValue struct {
	Value string
	Names map[string]string // by locale; at least one
}

// AuditTagKeysMax bounds AuditCatalog.TagKeys; AuditTagValuesMax bounds
// AuditTagKey.Values.
const (
	AuditTagKeysMax   = 8
	AuditTagValuesMax = 64
)

// AuditReservedTagKeys cannot be declared in AuditCatalog.TagKeys: ASN's own
// audit tags and the audit store's record keys (am.modules v27.0.5 audit).
var AuditReservedTagKeys = []string{"Service", "Caller", "RootID", "NodeID",
	"SubjectType", "SubjectID", "DeviceID", "SubjectIP", "Operation", "ObjectType",
	"ObjectID", "ObjectPart", "EventType", "Result", "RiskLevel", "Scope", "Partition"}

// AuditEventType names an event type; it may not start with "asn.".
type AuditEventType struct {
	Type  string
	Names map[string]string // by locale: "zh-CN", "en-US"
}

// AuditOperation names an operation code.
type AuditOperation struct {
	Code      int64 // unique in the catalog
	EventType string
	Names     map[string]string
}

// AuditObjectType names an object type.
type AuditObjectType struct {
	Type  string
	Names map[string]string
}

// AuditAPI writes and reads operation audit.
type AuditAPI interface {
	// WriteAudit records an operation of the calling service. rootID is the
	// root the operation belongs to, empty for the platform scope; with
	// event.NodeID set it is empty or the node's root. Service and Caller
	// are the calling service. The record is queued.
	//
	// A root or node deleted in the last 10 minutes is still accepted, a
	// deleted node on the root it was in, so that a record of the deletion
	// written after it keeps its root.
	//
	// Errors: shared.request.param_invalid (a reserved tag, an "asn." event
	// type, a subject, result or level value, a root that is not one, a
	// root and node mismatch); asn.network.not_found; asn.node.not_found;
	// shared.system.unavailable (the queue is full).
	WriteAudit(rootID string, event commonapi.AuditEvent) error

	// ListAuditRecords lists the records filter selects, newest first.
	// page.Limit is 1–500 and page.Offset a multiple of it. A service that
	// is not super sets Service or Caller (or both) to itself.
	//
	// Errors: asn.service.not_super; shared.request.param_invalid
	// (RootIDs with NetworkID, a range over 92 days, the page; Tags without
	// Service or with a key it does not declare); asn.network.not_found.
	ListAuditRecords(filter AuditFilter, page iam.Page) (records []*AuditRecord, total int, err error)

	// SummarizeAudit summarizes the records filter selects, with the trend in
	// buckets of bucket and the top alarm subjects, at most top. Bounds as
	// ListAuditRecords.
	SummarizeAudit(filter AuditFilter, bucket time.Duration, top int) (*AuditSummary, error)

	// ExportAuditRecords writes the records filter selects to w as "csv" or
	// "jsonl", oldest first, and returns how many it wrote. Bounds as
	// ListAuditRecords; a refusal is returned before anything is written.
	ExportAuditRecords(filter AuditFilter, format string, w io.Writer) (rows int, err error)
}
