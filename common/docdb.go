// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

type DocCollOptions struct {
	// Indexes of the collection, key = "column1;column2;...", added to those
	// it has: an index of the same fields as one it has must have the same
	// options. A unique index the records already break is
	// shared.resource.already_exists.
	Indexes map[string]DocIndexOptions

	// RebuildIndex drops the indexes and builds them again from Indexes.
	RebuildIndex bool
}

type DocIndexOptions struct {
	Sparse bool
	Unique bool
}

type DocDBHandler interface {
	// Name returns the actual database name, i.e. the name given to InitDocDB
	// with the framework's prefix applied, see ASNController.InitDocDB and
	// ASNServiceNode.InitDocDB.
	Name() string

	// Destroy releases this database handle.
	Destroy() error

	// CollectionExists reports whether the collection exists.
	CollectionExists(name string) (bool, error)

	// FindOrCreateCollection creates or find(if already exist) collection
	// name: name of collection
	// options:
	//
	//	options.Indexes are the collection's indexes, options.RebuildIndex rebuilds them
	//
	// A collection opened again takes no options, or the same ones.
	// returns the collection handle corresponding to the collection name
	FindOrCreateCollection(name string, options *DocCollOptions) (DocCollection, error)

	// DeleteCollection deletes the collection / measurement
	// name: name of collection / measurement
	DeleteCollection(name string) error

	// ListCollections list all collections / measurement names in one database
	ListCollections() ([]string, error)
}

// DocCollection is a collection of records.
//
// A query selects records, see DocQuery. A query that is not valid, the zero
// DocQuery or an empty DocAnd() among them, is an error: to act on every
// record, pass DocAll().
//
// A write that would give two records the same key of a unique index is an
// error, and so is DeleteRecord, FindRecord or UpdateRecord with a query that
// matches more than one record.
type DocCollection interface {
	// AddRecord adds one record to collection c.
	// record: json format data, including key-value pairs.
	AddRecord(record map[string]any) error

	// AddRecords adds some records to collection c.
	// records: arrays of json format data, each data includes key-value pairs.
	AddRecords(records []map[string]any) error

	// DeleteRecord deletes one record in collection c.
	// q: condition that determines which record to be deleted.
	// If the condition matches no record, exist will be false, but no error will be returned.
	// If the condition matches more than one record, an error will be returned.
	DeleteRecord(q DocQuery) (exist bool, err error)

	// DeleteRecords deletes records in collection c.
	// q: condition that determines matched records to be deleted, DocAll() to delete all records.
	// This function does not care how many records are deleted in the process, maybe 0, 1, or more.
	DeleteRecords(q DocQuery) error

	// FindRecord finds record in collection c.
	// q: condition that determines to match one record.
	// opts: DocFields or DocOmit to project the record; other options are an error.
	// If the condition matches no record, exist will be false, but no error will be returned.
	// If the condition matches more than one record, an error will be returned.
	FindRecord(q DocQuery, opts ...DocFindOption) (record map[string]any, exist bool, err error)

	// FindRecords finds records in collection c.
	// This function does not care how many records are matched in the process, maybe 0, 1, or more.
	// q: condition that determines to match the records.
	// opts: DocAsc and DocDesc to sort, DocSkip, DocLimit and DocPage to page, DocFields and DocOmit to project;
	// none returns every matched record, in the order they were added. See DocFindOptions.
	// A bad option, such as an empty or repeated sort field or a non-positive limit, is an error.
	FindRecords(q DocQuery, opts ...DocFindOption) ([]map[string]any, error)

	// CountRecords returns the number of records q matches, DocAll() to count every record.
	CountRecords(q DocQuery) (int, error)

	// UpdateRecord applies u to the one record q matches, as a whole: all its
	// changes are made, or none of them. A record that no longer matches q
	// when it is written is not updated, so that q may hold a condition, such
	// as DocGt("seats", 0) with DocInc("seats", -1).
	// If the condition matches no record, exist will be false, but no error will be returned.
	// If the condition matches more than one record, an error will be returned, and no record updated.
	UpdateRecord(q DocQuery, u DocUpdate) (exist bool, err error)

	// UpdateRecords applies u to each record q matches, DocAll() for every
	// record. An error may leave some updated.
	UpdateRecords(q DocQuery, u DocUpdate) error

	// ReplaceRecord replaces the one record q matches with record: the fields
	// record does not have are removed.
	// If the condition matches no record, exist will be false, but no error will be returned.
	// If the condition matches more than one record, an error will be returned, and no record replaced.
	ReplaceRecord(q DocQuery, record map[string]any) (exist bool, err error)

	// UpsertRecord applies u to the one record q matches, or, when q matches
	// none, adds record as it is, without applying u: one or the other,
	// never both. inserted tells which.
	//
	// record must match q, or the next upsert would not find it; this is
	// checked. Two upserts at once may both find no record and both add one,
	// unless a unique index of the fields of q rejects the second, which is
	// then retried as an update. So give the fields of q a unique index.
	// If the condition matches more than one record, an error will be returned, and no record changed.
	UpsertRecord(q DocQuery, u DocUpdate, record map[string]any) (inserted bool, err error)

	// ReplaceOrAddRecord replaces the one record q matches with record, as
	// ReplaceRecord does, or adds record when q matches none. record and q
	// are as for UpsertRecord.
	ReplaceOrAddRecord(q DocQuery, record map[string]any) (inserted bool, err error)
}
