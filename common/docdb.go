// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

type DocCollOptions struct {
	// Indexes of the collection, key = "column1;column2;...". MongoDB creates
	// them; FileDB enforces the unique ones, on the records written from then
	// on, so that records it already holds are not rejected.
	Indexes map[string]DocIndexOptions

	// UseRawFormat stores the records in the MongoDB format (BSON) rather than
	// as JSON. It is another stored form, not an optimization: field names are
	// lowercased, times are BSON dates and binary values (a UUID among them)
	// BSON binaries, so a query on a UUID or time field never matches such a
	// record, and a record read back holds driver types and its _id. A
	// collection must be opened with the same value every time, and FileDB
	// does not support it. Use the JSON format for new collections.
	UseRawFormat bool

	// if true, indexes will be dropped and rebuilt from Indexes (mongoDB only)
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
	//	for mongodb, options.Indexes used for create collection indexes, options.RebuildIndex rebuilds them
	//	for filedb, the unique indexes of options.Indexes are enforced, options.RebuildIndex is ignored
	//	for both, options.UseRawFormat selects the stored format, see DocCollOptions
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

	// UpdateRecord updates one record in collection c.
	// q: condition that determines to match one record to do data update.
	// newRecord: json format data, including key-value pairs, replaces the old pairs in the matched record.
	// If the condition matches no record, exist will be false, but no error will be returned.
	// If the condition matches more than one record, an error will be returned.
	UpdateRecord(q DocQuery, newRecord string) (exist bool, err error)

	// UpdateRecords updates records in collection c.
	// q: condition that determines to match the records to do data update.
	// newRecord: json format data, including key-value pairs, replaces the old pairs in all the matched records.
	// This function does not care how many records are matched in the process, maybe 0, 1, or more.
	UpdateRecords(q DocQuery, newRecord string) error

	// ArrayAppend appends elements in a field whose value type is an array.
	// If the field type is not array, will get an error; a null or missing field is an empty array.
	// q: condition that determines to match the records to append data.
	// newRecord: json format data, including key-value pairs, values can be a single item, or items of an array,
	//            will be appended in each of the matched records, not replaced.
	// ignoreDuplicate: true will ignore duplicated data, otherwise will accept the duplicated data.
	ArrayAppend(q DocQuery, newRecord string, ignoreDuplicate bool) error

	// ArrayDelete deletes the specified values from a field whose type is an array.
	// If the field type is not array, will get an error; a null or missing field is left as it is.
	// q: condition that determines to match the records to delete data.
	// newRecord: json format data, including key-value pairs, values can be a single item, or items of an array,
	//            will be deleted from each of the matched records.
	ArrayDelete(q DocQuery, newRecord string) error

	// ListAllRecord returns all records of a collection
	ListAllRecord() ([]map[string]any, error)
}
