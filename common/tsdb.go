// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import "time"

type TSCollOptions struct {
	MongoTsOpts  TSMongoOptions
	TSFileDBOpts TSFileDBOptions // Required when creating a file time-series collection, otherwise ignored.
}

type TSMongoOptions struct {
	TimeField   string   // Required when creating a mongo time-series collection, otherwise ignored.
	Indexes     []string // key = "column1;column2;..."
	MetaField   string   // Optional, Set only when creating a mongo time-series collection, otherwise ignored.
	Granularity string   // Optional, Set only when creating a mongo time-series collection, otherwise ignored.
	Expired     string   // Optional, Set only when creating a mongo time-series collection, otherwise ignored.
}
type TSFileDBOptions struct {
	// Optional, when creating a file time-series collection, otherwise ignored.
	// The retention policy is disabled by default.
	Retention int
}

type TSRetentionPolicy struct {
	Duration           string // data keep, for examples, 1000s, 100m, 10h, 1d
	ShardGroupDuration string // for influxDBV1, for influxDBV2, for examples, "1000s", "100m", "10h", "1d"
	Name               string // rp name for influxDBV1
	Replicas           int    // replications number for influxDBV1
	CollName           string // coll name for mongo-tsdb
}

// TSPoint is a data point to insert into TSDB.
type TSPoint struct {
	// Measurement is only necessary when doing a concurrent insertion into multiple
	// collections/measurements via TSDBHandler.AddDataPoint(s); otherwise it is ignored.
	Measurement string
	Tags        map[string]string
	Fields      map[string]any
	// Timestamp is the time of the point, it is passed to the database as is.
	Timestamp time.Time
	// RPName is the retention policy to insert into, only for influxDBV1.
	RPName string
}

// TSQuery is the condition to query data points.
// The zero value of each field means the field is not set.
type TSQuery struct {
	// Field is the field to query; with the *WithRegex methods it can be a regular expression.
	Field string

	// Start and Stop specify the boundaries of the time range.
	Start time.Time
	Stop  time.Time
	// Step is the maximum time between two slices within the boundaries, only used in prometheus.
	// It must be a whole number of seconds.
	Step time.Duration

	// Tags matches the tag values: values of one tag are ORed, different tags are ANDed.
	// With the *WithRegex methods the values are regular expressions.
	Tags map[string][]string

	// RPName is the retention policy to query, only for influxDBV1.
	RPName string

	// Desc outputs the data in descending order, ascending by default.
	Desc bool
	// Limit limits the number of data outputs; 0 means no limit.
	Limit uint64
	// Page returns the data on page *Page (starts from 0), Limit is the amount of data per page.
	// nil means no pagination; it requires Limit > 0.
	Page *int
}

type TSDBHandler interface {
	// Name returns the actual database name, i.e. the name given to InitTSDB
	// with the framework's prefix applied, see ASNController.InitTSDB and
	// ASNServiceNode.InitTSDB.
	Name() string

	// Destroy releases this database handle.
	Destroy() error

	// FindOrCreateCollection creates or find(if already exist) measurement
	// name: name of measurement
	// options:
	//
	//	for mongodb ts, options.tsFields including {timeField, metaField, granularity}, only timeField is mandatory
	//	for influxdb, indexes are nil
	//	for prometheus, indexes are nil
	//
	// returns the collection handle corresponding to the collection name
	FindOrCreateCollection(name string, options *TSCollOptions) (TSCollection, error)

	// DeleteCollection deletes the collection / measurement
	// name: name of collection / measurement
	DeleteCollection(name string) error

	// ListCollections list all collections / measurement names in one database
	ListCollections() ([]string, error)

	AddRetentionPolicies(rps []TSRetentionPolicy) error

	DeleteRetentionPolicies(rps []TSRetentionPolicy) error

	// GetRetentionPolicies gets the options of retention policy
	GetRetentionPolicies(name string) ([]TSRetentionPolicy, error)

	// AddDataPoint adds one data point to the collection named by pt.Measurement.
	AddDataPoint(pt *TSPoint, useTimeout bool) error

	// AddDataPoints adds some data points, each to the collection named by its Measurement.
	AddDataPoints(pts []*TSPoint, useTimeout bool) error
}

type TSPivotQueryResult []map[string]any

type TSTime int64
type TSSampleValue any

// TSSamplePoint is a sampling point
// Timestamp is time to insert the point into TSDB, milliseconds from epoch time
// Value is any, but in prometheus only can be float64
type TSSamplePoint struct {
	Timestamp TSTime // this is milliseconds
	Value     TSSampleValue
}

// TSSamples is a sample stream, amount of SamplePoint with the same tags group a sample stream
type TSSamples struct {
	Tags   map[string]string
	Values []TSSamplePoint
}

// TSMatrix key: the tags group marshaled string, different tags group according to different Samples
type TSMatrix map[string]*TSSamples

// TSQueryResult is the return value after execution one time query.
// key: field name, values with the same field-name stored in the same one Matrix
type TSQueryResult struct {
	Results map[string]TSMatrix
}

type TSCollection interface {
	// AddDataPoint adds one data point to collection c.
	AddDataPoint(pt *TSPoint) error

	// AddDataPoints adds some data points to collection c.
	AddDataPoints(pts []*TSPoint) error

	// FindDataPointsBySyntax enter the original query syntax for the database.
	FindDataPointsBySyntax(syntax string) (any, error)

	FindInfluxv2PivotDataPoints(syntax string) (TSPivotQueryResult, error)

	// FindDataPoints finds data points in the collection c.
	// This function does not care how many data points are returned in the process, maybe 0, 1, or more.
	FindDataPoints(query *TSQuery) (TSQueryResult, error)

	FindDataPointsFromJson(queryJson []byte) (TSQueryResult, error)

	// FindDataPointsWithRegex finds data points in the collection c.
	// This function does not care how many data points are matched in the process, maybe 0, 1, or more.
	FindDataPointsWithRegex(query *TSQuery) (TSQueryResult, error)

	FindDataPointsWithRegexFromJson(queryJson []byte) (TSQueryResult, error)

	// FindPivotDataPoints Pivot the query result to row table with the same tags value and time stamps, as query operation in traditional db
	FindPivotDataPoints(query *TSQuery) (TSPivotQueryResult, error)

	FindPivotDataPointsFromJson(queryJson []byte) (TSPivotQueryResult, error)

	FindPivotDataPointsWithRegex(query *TSQuery) (TSPivotQueryResult, error)

	FindPivotDataPointsWithRegexFromJson(queryJson []byte) (TSPivotQueryResult, error)

	// CountDataPoints return count of data points according to the condition in collection c.
	// The input argument is the same usage as in the function FindDataPoints
	CountDataPoints(query *TSQuery) (uint64, error)

	CountDataPointsFromJson(queryJson []byte) (uint64, error)
}
