// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package docdb

type CollOptions struct {
	// Optional when creating mongoDB collection, otherwise ignored.
	Indexes      map[string]IndexOptions // key = "column1;column2;..."
	UseRawFormat bool                    // if true, it will use raw format for performance optimization (mongoDB: BSON, fileDB: JSON); otherwise it will use JSON format for better compatibility
}

type IndexOptions struct {
	Sparse bool
	Unique bool
}

type Database interface {
	DeleteCollection(name string) error
	ListCollections() ([]string, error)

	CreateFile(fileName, metadata string, file []byte) error
	FindFiles(queryString string) ([]string, [][]byte, error)

	Destroy(useTimeout bool) error
}

type Collection[T any] interface {
	AddRecord(record *T) error
	AddRecords(records []*T) error

	DeleteRecord(queryJson string) error
	DeleteRecords(queryJson string) error

	FindRecord(queryJson string, fieldFilter map[string]bool) (*T, error)
	FindRecords(queryJson string, page, num int, sorting string, fieldFilter map[string]bool) ([]*T, error)
	FindRecordWithRegex(queryJson string, fieldFilter map[string]bool) (*T, error)
	FindRecordsWithRegex(queryJson string, page, num int, sorting string, fieldFilter map[string]bool) ([]*T, error)

	UpdatedRecord(queryJson string, newRecord string) error
	UpdatedRecords(queryJson string, newRecord string) error

	ArrayAppend(queryJson string, newRecord string, ignoreDuplicate bool) error
	ArrayDelete(queryJson string, newRecord string) error

	ListAllRecord() ([]*T, error)
}
