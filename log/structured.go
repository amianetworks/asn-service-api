// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package log

import "time"

// Structured records a service keeps through the controller
// (capi.StructuredLogAPI): the entity log, what a subject did to an object,
// and the API log, one call of an API. ASN stores each service's records
// apart from every other service's and from the operation audit.

// Subject is who did something: its type and ID, and the device and address
// it acted from.
type Subject struct {
	Type   string // required: account, system, service ...
	ID     string // ID or Device, at least one
	Device string
	IP     string
}

// Object is what was done to: its type, its ID, and the part of it, e.g. an
// account's calendar.
type Object struct {
	Type string // required when ID or Part is set
	ID   string // empty for an operation on the type as a whole
	Part string
}

// Result is the outcome of an entity record.
type Result string

const (
	ResultSuccess Result = "success"
	ResultFailed  Result = "failed"
)

// EntityLog is one thing a subject did, or tried, to an object.
type EntityLog struct {
	Time      time.Time // zero: when ASN takes it
	Subject   Subject   // required
	Operation string    // required
	Object    Object    // zero: no object, e.g. a login
	EventType string    // the service's category of the record
	Result    Result    // required
	Message   string    // e.g. why it failed

	Tags   map[string]string // the service's indexed tags; values non-empty; reserved keys refused
	Fields map[string]any    // details, not indexed: string, int64, float64 or bool
}

// The tags an entity record is stored under, also the names CountEntityLogsBy
// and CountEntityLogWindows take besides the service's own tags. These and
// "message" and "Partition" are reserved: Tags and Fields cannot use them.
const (
	TagSubjectType = "SubjectType"
	TagSubjectID   = "SubjectID"
	TagDeviceID    = "DeviceID"
	TagSubjectIP   = "SubjectIP"
	TagOperation   = "Operation"
	TagObjectType  = "ObjectType"
	TagObjectID    = "ObjectID"
	TagObjectPart  = "ObjectPart"
	TagEventType   = "EventType"
	TagResult      = "Result"
)

// EntityLogQuery selects entity records: a list matches any of its values,
// and the conditions set must all match, exactly.
type EntityLogQuery struct {
	SubjectTypes []string
	SubjectIDs   []string
	DeviceIDs    []string
	SubjectIPs   []string
	Operations   []string
	ObjectTypes  []string
	ObjectIDs    []string
	ObjectParts  []string
	EventTypes   []string
	Results      []Result
	Tags         map[string][]string // the service's tags

	Start, Stop time.Time // [Start, Stop); zero: open
}

// APILog is one call of an API.
type APILog struct {
	Time time.Time // zero: when ASN takes it

	// Indexed.
	ApiName string
	Account string
	Entity  string

	// Not indexed.
	Success    string
	ErrInfo    string
	CallerName string
	CallerIP   string
	Reserved   string
}

// APILogQuery selects API records: a list matches any of its values, and the
// conditions set must all match, exactly.
type APILogQuery struct {
	ApiNames []string
	Accounts []string
	Entities []string

	Start, Stop time.Time // [Start, Stop); zero: open
}

// Statistic counts records over a range, and how often they came.
type Statistic struct {
	Total     uint64
	PerHour   float64 // zero without both ends of the range
	PerMinute float64
	PerSecond float64
}

// Group counts the records with the same values of the grouping tags.
type Group struct {
	Values map[string]string // tag -> value; a record without the tag has ""
	Count  uint64
}

// Window holds the groups of the records of [Start, Start+window).
type Window struct {
	Start  time.Time
	Groups []Group // none for a window without records
}
