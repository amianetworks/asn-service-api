// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

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

type TSDBHandler interface {
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
}

type TSPoint interface {
	// AddTag adds a tag to point, will overwrite the value if set the same key repeatedly.
	// And now, do not limit the number of Tags
	AddTag(tagName, tagValue string) error

	// DeleteTag deletes a tag from tags
	DeleteTag(tagName string) error

	// AddTags adds some tags to a point, will overwrite the value if set the same key repeatedly.
	// And now, do not limit the number of Tags
	AddTags(tags map[string]string) error

	// DeleteTags deletes some tags
	DeleteTags(tags []string) error

	// AddField adds a field to point, will overwrite the value if set the same key repeatedly.
	// And now, do not limit the number of Fields
	AddField(fieldName string, fieldValue interface{}) error

	// DeleteField deletes a field from fields
	DeleteField(fieldName string) error

	// AddFields adds a field to point, will overwrite the value if set the same key repeatedly.
	// And now, do not limit the number of Fields
	AddFields(fields map[string]interface{}) error

	// DeleteFields deletes some fields
	DeleteFields(fields []string) error

	// SetTime sets timestamp to point.
	SetTime(t time.Time) error

	// ResetTime resets time to init time
	ResetTime() error

	// SetRPName sets rp for the point to insert.
	SetRPName(name string) error

	// ResetRPName resets rp for point
	ResetRPName() error
}

type TSPivotQueryResult []map[string]interface{}

type TSQuery interface {
	// AddTag adds a tag to the query, will overwrite the value if set the same key repeatedly.
	// And now, do not limit the number of Tags
	AddTag(tagName string, tagValue interface{}) error

	// DeleteTag deletes a tag from tags
	DeleteTag(tagName string) error

	// AddTags adds some tags to a point, will overwrite the value if set the same key repeatedly.
	// And now, do not limit the number of Tags
	AddTags(tagNames []string, tagValues []interface{}) error

	// DeleteTags deletes some tags
	DeleteTags(tags []string) error

	// SetField adds a field to Query.
	SetField(fieldName string)

	// SetStart functions can operate the TimeRange structure.
	SetStart(start time.Time)

	ResetStart()

	SetStop(stop time.Time)

	ResetStop()

	// SetStep sets the query step, will change input to time.Duration.
	// Now support "1d", "20h", "20m", "20s" when to set step 1 day, 20 hours, 20 minutes, or 20 seconds
	SetStep(step string) error

	// ResetStep clear the step setting
	ResetStep()

	SetRPName(name string) error

	ResetRPName() error

	// SetDesc determines whether the data is outputted in ascending or descending,
	// with ascending order being the default.
	SetDesc()

	// SetLimit sets the number of data outputs.
	SetLimit(n uint64)

	// SetPage determines whether the data do pagination.
	// This option must specify a limit value, which represents the amount of data per page.
	SetPage(page int) error
}

type TSTime int64
type TSSampleValue interface{}

// TSSamplePoint is a sampling point
// Timestamp is time to insert the point into TSDB, milliseconds from epoch time
// Value is interface{}, but in prometheus only can be float64
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
	NewPoint() TSPoint
	NewQuery() TSQuery

	// AddDataPoint adds one data point to collection c.
	AddDataPoint(pt TSPoint) error

	// AddDataPoints adds some data points to collection c.
	AddDataPoints(pts []TSPoint) error

	// FindDataPointsBySyntax enter the original query syntax for the database.
	FindDataPointsBySyntax(syntax string) (interface{}, error)

	FindInfluxv2PivotDataPoints(syntax string) (TSPivotQueryResult, error)

	// FindDataPoints finds data points in the collection c.
	// This function does not care how many data points are returned in the process, maybe 0, 1, or more.
	FindDataPoints(query TSQuery) (TSQueryResult, error)

	FindDataPointsFromJson(queryJson []byte) (TSQueryResult, error)

	// FindDataPointsWithRegex finds data points in the collection c.
	// This function does not care how many data points are matched in the process, maybe 0, 1, or more.
	FindDataPointsWithRegex(query TSQuery) (TSQueryResult, error)

	FindDataPointsWithRegexFromJson(queryJson []byte) (TSQueryResult, error)

	// FindPivotDataPoints Pivot the query result to row table with the same tags value and time stamps, as query operation in traditional db
	FindPivotDataPoints(query TSQuery) (TSPivotQueryResult, error)

	FindPivotDataPointsFromJson(queryJson []byte) (TSPivotQueryResult, error)

	FindPivotDataPointsWithRegex(query TSQuery) (TSPivotQueryResult, error)

	FindPivotDataPointsWithRegexFromJson(queryJson []byte) (TSPivotQueryResult, error)

	// CountDataPoints return count of data points according to the condition in collection c.
	// The input argument is the same usage as in the function FindDataPoints
	CountDataPoints(query TSQuery) (uint64, error)

	CountDataPointsFromJson(queryJson []byte) (uint64, error)
}
