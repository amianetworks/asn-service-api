// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

type DocCollOptions struct {
	// Optional when creating mongoDB collection, otherwise ignored.
	Indexes map[string]DocIndexOptions // key = "column1;column2;..."

	// if true, it will use the raw format for performance optimization (mongoDB: BSON, fileDB: JSON);
	// otherwise it will use JSON format for better compatibility
	UseRawFormat bool
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
	//	for filedb, options are nil
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
	// queryJson: condition of key-value pairs in JSON format, determine which record to be deleted.
	// If the condition matches no record or more than one record, will get an error
	DeleteRecord(queryJson string) (bool, error)

	// DeleteRecords deletes records in collection c.
	// queryJson: condition of key-value pairs in JSON format, determine matched records to be deleted.
	// This function does not care how many records are deleted in the process, maybe 0, 1, or more.
	DeleteRecords(queryJson string) error

	// FindRecord finds record in collection c.
	// queryJson: condition of key-value pairs in JSON format, determine to match one record.
	// If the condition matches no record or more than one record, will get an error.
	FindRecord(queryJson string, fieldFilter map[string]bool) (map[string]interface{}, bool, error)

	// FindRecords finds records in collection c, records may be processed by sorting and pagination.
	// This function does not care how many records are matched in the process, maybe 0, 1, or more.
	// queryJson: condition of key-value pairs in JSON format, determine to match the records.
	// page: starts from 0, which page of records to return.
	// num: starts from 1, when no pagination it means limits; or when pagination it means how many records in each page.
	// When page <= -1 && num <= 0, no pagination or limits;
	// when page <= -1 && num > 0, only limits; no pagination;
	// when page > -1 && num > 0, do pagination;
	// when page > -1 && num <= 0, return error.
	// sorting: sort specifications, can be built by Sorting.Build, includes fields and ascending or descending order on each field.
	// If soring is an empty string, do not sort.
	FindRecords(queryJson string, page, num int, sorting string, fieldFilter map[string]bool) ([]map[string]interface{}, error)

	// FindRecordWithRegex finds a record in collection c.
	// queryJson: condition of key-value pairs in JSON format with regular expression style, determine to match one record.
	// If the condition matches no record or more than one record, will get an error.
	FindRecordWithRegex(queryJson string, fieldFilter map[string]bool) (map[string]interface{}, bool, error)

	// FindRecordsWithRegex finds records in collection c, records may be processed by sorting and pagination.
	// This function does not care how many records are matched in the process, maybe 0, 1, or more.
	// queryJson: condition of key-value pairs in JSON format with regular expression style, determine to match the records.
	// page: starts from 0, which page of records to return.
	// num: starts from 1, when no pagination it means limits; or when pagination it means how many records in each page.
	// When page <= -1 && num <= 0, no pagination or limits;
	// when page <= -1 && num > 0, only limits; no pagination;
	// when page > -1 && num > 0, do pagination;
	// when page > -1 && num <= 0, return error.
	// sorting: sort specifications, can be built by Sorting.Build, includes fields and ascending or descending order on each field.
	// If soring is an empty string, do not sort.
	FindRecordsWithRegex(queryJson string, page, num int, sorting string, fieldFilter map[string]bool) ([]map[string]interface{}, error)

	// UpdateRecord updates one record in collection c.
	// queryJson: condition of key-value pairs in JSON format, determine to match one record to do data update.
	// newRecord: json format data, including key-value pairs, replaces the old pairs in the matched record.
	// If the condition matches no record or more than one record, will get an error.
	UpdateRecord(queryJson string, newRecord string) (bool, error)

	// UpdateRecords updates records in collection c.
	// queryJson: condition of key-value pairs in JSON format, determine to match the records to do data update.
	// newRecord: json format data, including key-value pairs, replaces the old pairs in all the matched records.
	// This function does not care how many records are matched in the process, maybe 0, 1, or more.
	UpdateRecords(queryJson string, newRecord string) error

	// ArrayAppend appends elements in a field whose value type is an array.
	// If the field type is not array, will get an error.
	// queryJson: condition of key-value pairs in JSON format, determine to match the records to append data.
	// newRecord: json format data, including key-value pairs, values can be a single item, or items of an array,
	//            will be appended in each of the matched records, not replaced.
	// ignoreDuplicate: true will ignore duplicated data, otherwise will accept the duplicated data.
	ArrayAppend(queryJson string, newRecord string, ignoreDuplicate bool) error

	// ArrayDelete deletes the specified values from a field whose type is an array.
	// If the field type is not array, will get an error.
	// queryJson: condition of key-value pairs in JSON format, determine to match the records to delete data.
	// newRecord: json format data, including key-value pairs, values can be a single item, or items of an array,
	//            will be deleted from each of the matched records.
	ArrayDelete(queryJson string, newRecord string) error

	// ListAllRecord returns all records of a collection
	ListAllRecord() ([]map[string]interface{}, error)
}
