// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import "time"

// PlatformScope is the root value of records, points and rules that belong
// to no root network. Queries accept it wherever they take a root; writes
// take an empty root instead.
const PlatformScope = "platform"

// AuditSubjectType is who did an audited operation.
type AuditSubjectType string

const (
	AuditSubjectAccount AuditSubjectType = "account" // SubjectID: the account ID
	AuditSubjectSystem  AuditSubjectType = "system"  // SubjectID empty
	AuditSubjectService AuditSubjectType = "service" // SubjectID empty: ASN sets the caller
)

// AuditResult is the outcome of an audited operation.
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailed  AuditResult = "failed"
)

// AuditRiskLevel is the risk of an audited operation; low and above are
// alarms.
type AuditRiskLevel string

const (
	AuditRiskInfo     AuditRiskLevel = "info"
	AuditRiskWarn     AuditRiskLevel = "warn"
	AuditRiskLow      AuditRiskLevel = "low"
	AuditRiskMedium   AuditRiskLevel = "medium"
	AuditRiskHigh     AuditRiskLevel = "high"
	AuditRiskCritical AuditRiskLevel = "critical"
)

// AuditEvent is one operation a service records.
type AuditEvent struct {
	Time        time.Time        // zero: when ASN takes it
	SubjectType AuditSubjectType // required
	SubjectID   string
	DeviceID    string // IAM device ID, accounts only
	SubjectIP   string
	EventType   string         // required; not starting with "asn."
	Operation   int64          // a code of the service's audit catalog
	Result      AuditResult    // required
	RiskLevel   AuditRiskLevel // empty: info
	ObjectType  string         // required when ObjectID or ObjectPart is set
	ObjectID    string
	ObjectPart  string
	NodeID      string // the node the record concerns; empty: none
	Message     string
	Tags        map[string]string // the service's own; keys of capi.AuditReservedTagKeys refused
	Fields      map[string]string
}
