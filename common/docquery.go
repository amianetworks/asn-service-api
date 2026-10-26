// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import "encoding/json"

// DocOp is the operator of a DocQuery.
type DocOp string

const (
	DocOpAll    DocOp = "all"
	DocOpEq     DocOp = "eq"
	DocOpNe     DocOp = "ne"
	DocOpGt     DocOp = "gt"
	DocOpGe     DocOp = "ge"
	DocOpLt     DocOp = "lt"
	DocOpLe     DocOp = "le"
	DocOpIn     DocOp = "in"
	DocOpNin    DocOp = "nin"
	DocOpExists DocOp = "exists"
	DocOpRegex  DocOp = "regex"
	DocOpAnd    DocOp = "and"
	DocOpOr     DocOp = "or"
	DocOpNor    DocOp = "nor"
	DocOpNot    DocOp = "not"
)

// DocAllOps returns every operator, so that a test of an adapter can check it
// carries each of them.
func DocAllOps() []DocOp {
	return []DocOp{DocOpAll, DocOpEq, DocOpNe, DocOpGt, DocOpGe, DocOpLt, DocOpLe, DocOpIn, DocOpNin,
		DocOpExists, DocOpRegex, DocOpAnd, DocOpOr, DocOpNor, DocOpNot}
}

// DocQuery is the condition that selects records of a DocCollection, built
// with the functions below and combined with DocAnd, DocOr, DocNor and DocNot.
//
// It is plain data, the same fields in the same order as the query of the
// framework's database, which validates it when a collection runs it and
// returns the errors from that call; this package does not check it.
//
// Values are compared the way they are stored in JSON: they are normalized to
// a string, a number or a bool, so a UUID matches its string form. Fields are
// paths, with "." between nested keys; a path goes through arrays of objects,
// matching any element, as MongoDB does.
//
// Times stored in records are RFC 3339 strings, which do not compare in time
// order once their fraction or time zone differ, so a time is no value to
// compare: a record to be selected or sorted by time keeps it as a number as
// well, e.g. Unix microseconds.
type DocQuery struct {
	Op      DocOp
	Field   string     // path of the field, for field operators
	Value   any        // compared value of eq/ne/gt/ge/lt/le, pattern of regex, bool of exists
	Values  []any      // values of in/nin
	Options string     // options of regex
	Subs    []DocQuery // sub queries of and/or/nor/not
}

// DocAll matches every record.
func DocAll() DocQuery {
	return DocQuery{Op: DocOpAll}
}

// DocEq matches records whose field equals value, or, for an array field, holds
// an element that equals value. value is a scalar; use DocIn to match a list.
func DocEq(field string, value any) DocQuery {
	return DocQuery{Op: DocOpEq, Field: field, Value: value}
}

// DocNe matches records that DocEq does not, including those without the field.
func DocNe(field string, value any) DocQuery {
	return DocQuery{Op: DocOpNe, Field: field, Value: value}
}

// DocGt matches records whose field is greater than value. Only a number
// compares to a number and a string to a string (byte-wise); other values do
// not match.
func DocGt(field string, value any) DocQuery {
	return DocQuery{Op: DocOpGt, Field: field, Value: value}
}

// DocGe is DocGt with greater than or equal to.
func DocGe(field string, value any) DocQuery {
	return DocQuery{Op: DocOpGe, Field: field, Value: value}
}

// DocLt is DocGt with less than.
func DocLt(field string, value any) DocQuery {
	return DocQuery{Op: DocOpLt, Field: field, Value: value}
}

// DocLe is DocGt with less than or equal to.
func DocLe(field string, value any) DocQuery {
	return DocQuery{Op: DocOpLe, Field: field, Value: value}
}

// DocIn matches records whose field equals any of values. Empty values match nothing.
func DocIn[T any](field string, values []T) DocQuery {
	return DocQuery{Op: DocOpIn, Field: field, Values: toAnys(values)}
}

// DocNin matches records that DocIn does not, including those without the
// field. Empty values match everything.
func DocNin[T any](field string, values []T) DocQuery {
	return DocQuery{Op: DocOpNin, Field: field, Values: toAnys(values)}
}

// DocExists matches records that have the field (even if null) when exists is
// true, or do not have it when false.
func DocExists(field string, exists bool) DocQuery {
	return DocQuery{Op: DocOpExists, Field: field, Value: exists}
}

// DocRegex matches records whose string field matches pattern, a Go (RE2)
// regular expression. options is a combination of "i" (case-insensitive), "m"
// (multi-line) and "s" (dot matches newline); without "i" the match is
// case-sensitive.
//
// Escape user input with regexp.QuoteMeta before putting it in a pattern.
func DocRegex(field, pattern, options string) DocQuery {
	return DocQuery{Op: DocOpRegex, Field: field, Value: pattern, Options: options}
}

// DocAnd matches records that match all of qs. With no qs it is an error; use
// DocAllOf for conditions collected at run time.
func DocAnd(qs ...DocQuery) DocQuery {
	return DocQuery{Op: DocOpAnd, Subs: qs}
}

// DocAllOf is DocAnd(qs...) for conditions collected at run time: with no qs
// it matches every record, where DocAnd() is an error.
func DocAllOf(qs ...DocQuery) DocQuery {
	if len(qs) == 0 {
		return DocAll()
	}
	return DocAnd(qs...)
}

// DocOr matches records that match any of qs.
func DocOr(qs ...DocQuery) DocQuery {
	return DocQuery{Op: DocOpOr, Subs: qs}
}

// DocNor matches records that match none of qs.
func DocNor(qs ...DocQuery) DocQuery {
	return DocQuery{Op: DocOpNor, Subs: qs}
}

// DocNot matches records that q does not match.
func DocNot(q DocQuery) DocQuery {
	return DocQuery{Op: DocOpNot, Subs: []DocQuery{q}}
}

func toAnys[T any](values []T) []any {
	if values == nil {
		return nil
	}

	res := make([]any, 0, len(values))
	for _, v := range values {
		res = append(res, v)
	}
	return res
}

// DocSorting is the sort specification of DocCollection.FindRecords, the same
// as the framework's database takes.
type DocSorting struct {
	Fields []string `json:"fields"` // the fields to sort by, in order
	Asc    []bool   `json:"asc"`    // ascending (true) or descending (false), for each field
}

// Ascending sorts by field in ascending order, after the fields added before.
func (s *DocSorting) Ascending(field string) {
	s.Fields = append(s.Fields, field)
	s.Asc = append(s.Asc, true)
}

// Descending sorts by field in descending order, after the fields added before.
func (s *DocSorting) Descending(field string) {
	s.Fields = append(s.Fields, field)
	s.Asc = append(s.Asc, false)
}

// Build returns the specification as the sorting string of DocCollection.FindRecords.
func (s *DocSorting) Build() (string, error) {
	b, err := json.Marshal(s)
	return string(b), err
}
