// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import "time"

// The time-series database of a service (ASNController.InitTSDB,
// ASNServiceNode.InitTSDB). Its store is PostgreSQL (TimescaleDB) on the
// controller; queries are built from TSQuery, never in a store's own
// language.

// TSCollOptions are the options of a collection, applied when it is
// created: a collection that exists is used as it is.
type TSCollOptions struct {
	// Retention is how long points are kept, 0 for ever; whole seconds. A
	// chunk is dropped once all its points are older.
	Retention time.Duration

	// Indexes are the tag combinations to index.
	Indexes []TSIndex

	// ReplaceSameTime merges the points of the same tags and time into one:
	// a point written later replaces the values of the fields it has and
	// keeps the others. Without it both points are kept.
	ReplaceSameTime bool
}

// TSIndex is an index of a combination of tags, in order.
type TSIndex struct {
	Tags []string
}

// TSPoint is a point: its tags, fields and time. Measurement names its
// collection for the handler's AddDataPoint(s); a collection's writes
// ignore it and its reads set it. A field is a number, a string or a bool.
type TSPoint struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]any
	Timestamp   time.Time
}

// TSCondOp is the operator of a TSCond.
type TSCondOp string

const (
	TSCondEq    TSCondOp = "eq"
	TSCondNe    TSCondOp = "ne"
	TSCondGt    TSCondOp = "gt" // numbers only, as Ge, Lt and Le
	TSCondGe    TSCondOp = "ge"
	TSCondLt    TSCondOp = "lt"
	TSCondLe    TSCondOp = "le"
	TSCondRegex TSCondOp = "regex" // Value is a regular expression, on strings
)

// TSCond is a condition on a tag or a field of a point; Tag or Field names
// it, one of them. A point without the tag or field, or whose value is of
// another type than Value, does not match, whatever the operator.
//
// On a tag the operators are Eq, Ne and Regex with a non-empty string. On a
// field Value is a number, a string or a bool for Eq and Ne, a number for
// Gt, Ge, Lt and Le, a string for Regex. Numbers compare by value. A regular
// expression is in Go (RE2) syntax, run by the store as a POSIX regular
// expression: keep to literals, ., [], ^, $, |, (), *, +, ?, {m,n}, \d, \w,
// \s and a leading (?i); escape user input with regexp.QuoteMeta.
type TSCond struct {
	Tag   string
	Field string
	Op    TSCondOp
	Value any
}

// TSQuery selects points.
type TSQuery struct {
	// Field keeps only the points having this field, with that field alone;
	// "" keeps every point with all its fields.
	Field string

	// Tags keeps the points whose value of each tag is one of the values:
	// the values of one tag are ORed, different tags ANDed.
	Tags map[string][]string

	// Where keeps the points matching every condition.
	Where []TSCond

	// NotTags drops the points whose value of each tag is one of the values;
	// a point without the tag is kept.
	NotTags map[string][]string

	// Start and Stop bound the range [Start, Stop); a zero time leaves that
	// end open.
	Start time.Time
	Stop  time.Time

	// Desc returns the newest points first.
	Desc bool

	// Limit returns at most Limit points, 0 for all; Page, from 0, is the
	// page of Limit points. A Page needs a Limit.
	Limit uint64
	Page  int

	// Timeout is the query's own timeout, instead of the database's.
	Timeout time.Duration
}

// TSAggFn aggregates the values of a field.
type TSAggFn string

const (
	TSAggCount TSAggFn = "count" // the points having the field, a uint64
	TSAggSum   TSAggFn = "sum"   // int64 when all integers, else float64; as Min and Max
	TSAggMin   TSAggFn = "min"
	TSAggMax   TSAggFn = "max"
	TSAggMean  TSAggFn = "mean"  // float64
	TSAggFirst TSAggFn = "first" // the value on the oldest point having it
	TSAggLast  TSAggFn = "last"  // the value on the newest point having it
)

// TSAggregation groups the points a query matches by window and tags,
// aggregates their fields, optionally aggregates the groups again (Then),
// and orders and pages the groups (Query.Limit and Query.Page).
type TSAggregation struct {
	// Query selects the points; its Field is empty and Desc false.
	Query TSQuery

	// Fields are the fields to aggregate; with TSAggCount and none, Count
	// alone counts the points.
	Fields []string

	// Fn aggregates each field in a group of the first stage.
	Fn TSAggFn

	// GroupBy are the tags whose values group the points; none for one
	// group. A point without a tag groups under "".
	GroupBy []string

	// Window, when positive, splits [Query.Start, Query.Stop) into windows
	// from Start; both ends are then required. Zero is one window.
	Window time.Duration

	// Fill, with a Window, adds a bucket for each window without points,
	// with Fill as each field's value and a Count of 0. Nil adds none.
	Fill *float64

	// Then is the second stage, none when nil.
	Then *TSStage

	// OrderBy orders the buckets; ties by window, then keys.
	OrderBy []TSOrder
}

// TSStage is the second stage of an aggregation: Fn (count, sum, mean, min,
// max) over the first stage's groups of a window, grouped by GroupBy (tags
// of the first GroupBy, or fields grouped by their value).
type TSStage struct {
	Fn      TSAggFn
	GroupBy []string
}

// TSOrderKey is what a TSOrder sorts on.
type TSOrderKey string

const (
	TSOrderValue     TSOrderKey = "value"     // the value of Field, an aggregated field
	TSOrderCount     TSOrderKey = "count"     // TSBucket.Count
	TSOrderFirstTime TSOrderKey = "firstTime" // TSBucket.FirstTime
	TSOrderLastTime  TSOrderKey = "lastTime"  // TSBucket.LastTime
	TSOrderGroup     TSOrderKey = "group"     // the value of Field, a key of the buckets
)

// TSOrder is a sort key of the buckets; a bucket without the value sorts
// last.
type TSOrder struct {
	Key   TSOrderKey
	Field string
	Desc  bool
}

// TSBucket is a group of points, or of groups of the first stage.
type TSBucket struct {
	Start     time.Time         // the window's start
	Group     map[string]string // the keys of the last stage
	Values    map[string]any    // each aggregated field's result
	Count     uint64            // points, or first-stage groups with Then
	FirstTime time.Time
	LastTime  time.Time
}

// TSCursor is a position in a collection's Find order, for FindAfter.
type TSCursor string

// TSGroup counts the points of one combination of tag values.
type TSGroup struct {
	Values map[string]string
	Count  uint64
}

// TSDBHandler is a service's time-series database.
type TSDBHandler interface {
	// Name returns the actual database name: the name given to InitTSDB with
	// the framework's prefix applied.
	Name() string

	// Destroy releases this handle.
	Destroy() error

	// FindOrCreateCollection returns the collection name, creating it with
	// options when it does not exist.
	FindOrCreateCollection(name string, options *TSCollOptions) (TSCollection, error)

	// DeleteCollection deletes a collection and its points.
	DeleteCollection(name string) error

	// ListCollections lists the collections.
	ListCollections() ([]string, error)

	// AddDataPoint writes a point to the collection its Measurement names,
	// which must have been opened by FindOrCreateCollection.
	AddDataPoint(pt *TSPoint) error

	// AddDataPoints writes points, each to the collection it names.
	AddDataPoints(pts []*TSPoint) error
}

// TSCollection is a collection of points.
type TSCollection interface {
	Name() string

	// AddDataPoint and AddDataPoints write to this collection.
	AddDataPoint(pt *TSPoint) error
	AddDataPoints(pts []*TSPoint) error

	// Find returns the points q selects, oldest first (newest with Desc).
	Find(q TSQuery) ([]*TSPoint, error)

	// Count counts the points q selects having q.Field, which is required
	// (its paging aside).
	Count(q TSQuery) (uint64, error)

	// CountBy counts the points q selects having presenceField, per
	// combination of values of tags.
	CountBy(q TSQuery, presenceField string, tags ...string) ([]TSGroup, error)

	// Scan calls fn with every point q selects, in Find's order, reading a
	// batch at a time; an error from fn stops it and is returned.
	Scan(q TSQuery, fn func(*TSPoint) error) error

	// FindAfter returns up to n points after the cursor (empty: from the
	// start), in Find's order, with the cursor of the last one; q's Limit
	// and Page are ignored.
	FindAfter(q TSQuery, after TSCursor, n int) ([]*TSPoint, TSCursor, error)

	// Aggregate runs an aggregation and returns its buckets and the number of
	// buckets before paging.
	Aggregate(a TSAggregation) (buckets []TSBucket, total uint64, err error)
}
