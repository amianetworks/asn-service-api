// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

import (
	"time"

	"asn.amiasys.com/asn-service-api/v27/iam"
	"asn.amiasys.com/asn-service-api/v27/log"
)

// StructuredLogAPI keeps the calling service's structured logs: its entity
// log and its API log. ASN stores them for the service, apart from every
// other service's records and from the operation audit (AuditAPI), keeps
// them for the controller's configured retention, and deletes them when the
// service is removed from the controller. A service reads only its own.
//
// The operation audit stays the record of operations an operator reviews
// across services; the entity log is the service's own trail, e.g. access to
// its resources, read by the service.
type StructuredLogAPI interface {
	// WriteEntityLogs stores records, all checked before any is stored.
	//
	// Errors: shared.request.param_invalid (a record without a subject type,
	// a subject ID or device, an operation or a result; an object ID or part
	// without its type; a reserved or empty tag; a field of another type);
	// shared.system.unavailable.
	WriteEntityLogs(records []*log.EntityLog) error

	// FindEntityLogs returns the records q selects, newest first, and their
	// total before paging. page.Limit is 1–500 and page.Offset a multiple of
	// it.
	//
	// Errors: shared.request.param_invalid (a reserved tag, Stop before
	// Start, the page).
	FindEntityLogs(q log.EntityLogQuery, page iam.Page) (records []*log.EntityLog, total int, err error)

	// CountEntityLogs counts the records q selects.
	CountEntityLogs(q log.EntityLogQuery) (log.Statistic, error)

	// CountEntityLogsBy counts the records q selects by the values of tags
	// (log.Tag* or the service's own).
	CountEntityLogsBy(q log.EntityLogQuery, tags ...string) ([]log.Group, error)

	// CountEntityLogWindows counts the records of each window of [q.Start,
	// q.Stop), both required, by the values of tags. The first window starts
	// at q.Start; the last may end after q.Stop.
	CountEntityLogWindows(q log.EntityLogQuery, window time.Duration, tags ...string) ([]log.Window, error)

	// WriteAPILogs stores API records.
	//
	// Errors: shared.request.param_invalid (a nil record);
	// shared.system.unavailable.
	WriteAPILogs(records []*log.APILog) error

	// FindAPILogs returns the API records q selects, newest first, and their
	// total before paging. Page bounds as FindEntityLogs.
	FindAPILogs(q log.APILogQuery, page iam.Page) (records []*log.APILog, total int, err error)

	// CountAPILogs counts the API records q selects.
	CountAPILogs(q log.APILogQuery) (log.Statistic, error)
}
