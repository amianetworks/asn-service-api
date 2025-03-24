// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

type DocCollOptions struct {
	// Optional when creating mongoDB collection, otherwise ignored.
	Indexes      map[string]DocIndexOptions // key = "column1;column2;..."
	UseRawFormat bool                       // if true, it will use raw format for performance optimization (mongoDB: BSON, fileDB: JSON); otherwise it will use JSON format for better compatibility
}

type DocIndexOptions struct {
	Sparse bool
	Unique bool
}

type DocDBHandler interface {
	// FindOrCreateCollection creates or find(if already exist) collection
	// name: name of collection
	// options:
	//
	//	for mongodb, options.indexes used for create collection indexes
	//	for filedb, options is nil
	//
	// returns the collection handle corresponding to the collection name
	FindOrCreateCollection(name string, options *DocCollOptions) (DocCollection, error)

	// DeleteCollection deletes the collection / measurement
	// name: name of collection / measurement
	DeleteCollection(name string) error

	// ListCollections list all collections / measurement names in one database
	ListCollections() ([]string, error)

	// CreateFile creates a file
	CreateFile(fileName, metadata string, file []byte) error

	// FindFiles returns files based on queries
	FindFiles(queryString string) ([]string, [][]byte, error)
}

type DocCollection interface {
	// AddRecord adds one record to collection c.
	// record: json format data, including key-value pairs.
	AddRecord(record map[string]interface{}) error

	// AddRecords adds some records to collection c.
	// records: arrays of json format data, each data includes key-value pairs.
	AddRecords(records []map[string]interface{}) error

	// DeleteRecord deletes one record in collection c.
	// queryJson: condition of key-value pairs in json format, determine which record to be deleted.
	// If condition match no record or more than one record, will get an error
	DeleteRecord(queryJson string) error

	// DeleteRecords deletes records in collection c.
	// queryJson: condition of key-value pairs in json format, determine matched records to be deleted.
	// This function do not care how many records deleted in the process, maybe 0, 1, or more.
	DeleteRecords(queryJson string) error

	// FindRecord finds record in collection c.
	// queryJson: condition of key-value pairs in json format, determine to match one record.
	// If condition match no record or more than one record, will get an error.
	FindRecord(queryJson string, fieldFilter map[string]bool) (map[string]interface{}, error)

	// FindRecords finds records in collection c, records may be processed by sorting and pagination.
	// This function do not care how many records matched in the process, maybe 0, 1, or more.
	// queryJson: condition of key-value pairs in json format, determine to match the records.
	// page: starts from 0, which page of records to return.
	// num: starts from 1, when no pagination it means limits; or when pagination it means how many records in each page.
	// When page <= -1 && num <= 0, no pagination or limits;
	// when page <= -1 && num > 0, only limits; no pagination;
	// when page > -1 && num > 0, do pagination;
	// when page > -1 && num <= 0, return error.
	// sorting: sort specifications, can be built by Sorting.Build, includes fields and ascending or descending order on each field.
	// If soring is empty string, do not sort.
	FindRecords(queryJson string, page, num int, sorting string, fieldFilter map[string]bool) ([]map[string]interface{}, error)

	// FindRecordWithRegex finds record in collection c.
	// queryJson: condition of key-value pairs in json format with regular expression style, determine to match one record.
	// If condition match no record or more than one record, will get an error.
	FindRecordWithRegex(queryJson string, fieldFilter map[string]bool) (map[string]interface{}, error)

	// FindRecordsWithRegex finds records in collection c, records may be processed by sorting and pagination.
	// This function do not care how many records matched in the process, maybe 0, 1, or more.
	// queryJson: condition of key-value pairs in json format with regular expression style, determine to match the records.
	// page: starts from 0, which page of records to return.
	// num: starts from 1, when no pagination it means limits; or when pagination it means how many records in each page.
	// When page <= -1 && num <= 0, no pagination or limits;
	// when page <= -1 && num > 0, only limits; no pagination;
	// when page > -1 && num > 0, do pagination;
	// when page > -1 && num <= 0, return error.
	// sorting: sort specifications, can be built by Sorting.Build, includes fields and ascending or descending order on each field.
	// If soring is empty string, do not sort.
	FindRecordsWithRegex(queryJson string, page, num int, sorting string, fieldFilter map[string]bool) ([]map[string]interface{}, error)

	// UpdatedRecord updates one record in collection c.
	// queryJson: condition of key-value pairs in json format, determine to match one record to do data update.
	// newRecord: json format data, including key-value pairs, replaces the old pairs in the matched record.
	// If condition match no record or more than one record, will get an error.
	UpdatedRecord(queryJson string, newRecord string) error

	// UpdatedRecords updates records in collection c.
	// queryJson: condition of key-value pairs in json format, determine to match the records to do data update.
	// newRecord: json format data, including key-value pairs, replaces the old pairs in all the matched record.
	// This function do not care how many records matched in the process, maybe 0, 1, or more.
	UpdatedRecords(queryJson string, newRecord string) error

	// ArrayAppend appends elements in a field whose value type is array.
	// If the field type is not array, will get an error.
	// queryJson: condition of key-value pairs in json format, determine to match the records to append data.
	// newRecord: json format data, including key-value pairs, values can be a single item, or items of array, will be appended in each of the matched record, not replaced.
	// ignoreDuplicate: true will ignore duplicated data, otherwise will accept the duplicated data.
	ArrayAppend(queryJson string, newRecord string, ignoreDuplicate bool) error

	// ArrayDelete deletes the specified values from a field whose type is array.
	// If the field type is not array, will get an error.
	// queryJson: condition of key-value pairs in json format, determine to match the records to delete data.
	// newRecord: json format data, including key-value pairs, values can be a single item, or items of array, will be deleted from each of the matched record.
	ArrayDelete(queryJson string, newRecord string) error

	// ListAllRecord returns all records of a collection
	ListAllRecord() ([]map[string]interface{}, error)
}
