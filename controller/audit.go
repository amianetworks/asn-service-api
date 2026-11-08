// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
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
	SubjectID    string
	SubjectTypes []commonapi.AuditSubjectType
	EventTypes   []string
	Operations   []int64
	Results      []commonapi.AuditResult
	RiskLevels   []commonapi.AuditRiskLevel
	ObjectTypes  []string
	ObjectIDs    []string
	Origins      []string // "local", or node IDs the records were reported from

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
	Trend               []AuditTrendBucket
}

// AuditCatalog names a service's audit records for display. Writes are not
// checked against it.
type AuditCatalog struct {
	EventTypes  []AuditEventType
	Operations  []AuditOperation
	ObjectTypes []AuditObjectType
}

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
	// (RootIDs with NetworkID, a range over 92 days, the page);
	// asn.network.not_found.
	ListAuditRecords(filter AuditFilter, page iam.Page) (records []*AuditRecord, total int, err error)

	// SummarizeAudit summarizes the records filter selects, with the trend in
	// buckets of bucket and the top alarm subjects, at most top. Bounds as
	// ListAuditRecords.
	SummarizeAudit(filter AuditFilter, bucket time.Duration, top int) (*AuditSummary, error)
}
