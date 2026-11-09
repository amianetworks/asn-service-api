// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

// DocUpdateOp is the operator of a DocChange.
type DocUpdateOp string

const (
	DocUpSet       DocUpdateOp = "set"
	DocUpUnset     DocUpdateOp = "unset"
	DocUpInc       DocUpdateOp = "inc"
	DocUpPush      DocUpdateOp = "push"
	DocUpAddToSet  DocUpdateOp = "addToSet"
	DocUpPull      DocUpdateOp = "pull"
	DocUpPullWhere DocUpdateOp = "pullWhere"
)

// DocAllUpdateOps returns every update operator, so that a test of an adapter
// can check it carries each of them.
func DocAllUpdateOps() []DocUpdateOp {
	return []DocUpdateOp{DocUpSet, DocUpUnset, DocUpInc, DocUpPush, DocUpAddToSet, DocUpPull, DocUpPullWhere}
}

// DocChange is one change to one field of a record.
type DocChange struct {
	Op     DocUpdateOp
	Field  string    // path of the field, with "." between nested keys
	Value  any       // new value of set, amount of inc
	Values []any     // elements of push, addToSet and pull
	Where  *DocQuery // condition on the elements of pullWhere, nil for the other operators
}

// DocUpdate is a set of changes, applied to each record it updates as a
// whole: to one record, all the changes are made, or none of them.
//
// Like DocQuery, it is plain data, validated by the framework's database when
// a collection runs it; errors are returned from that call.
//
// Values are stored as AddRecord stores them, as JSON. A field is a path,
// with "." between nested keys. A missing object on the path is created; a
// numeric key indexes an array, which is padded with nulls up to it; a path
// through any other value is an error. _id cannot be changed, and a key
// cannot start with "$". Two changes of one field, or of a field and a field
// in it, are an error.
type DocUpdate struct {
	Changes []DocChange
}

// DocSet sets field to value, nil for null. value may be a scalar, a struct,
// a map or a slice.
func DocSet(field string, value any) DocUpdate {
	return DocUpdate{Changes: []DocChange{{Op: DocUpSet, Field: field, Value: value}}}
}

// DocUnset removes field. A field in an array element sets the element to
// null. A missing field is left as it is.
func DocUnset(field string) DocUpdate {
	return DocUpdate{Changes: []DocChange{{Op: DocUpUnset, Field: field}}}
}

// DocInc adds n, a number, to field. Integers add as integers, and an
// overflow is an error; with a float on either side, the sum is a float. A
// missing field is set to n; a field that is null or not a number is an error.
func DocInc(field string, n any) DocUpdate {
	return DocUpdate{Changes: []DocChange{{Op: DocUpInc, Field: field, Value: n}}}
}

// DocPush appends values to the array field, in order. A missing or null
// field is an empty array; a field of another type is an error. A value that
// is an array is an error, so that a slice given without "..." is not pushed
// as one element.
func DocPush[T any](field string, values ...T) DocUpdate {
	return DocUpdate{Changes: []DocChange{{Op: DocUpPush, Field: field, Values: toAnys(values)}}}
}

// DocAddToSet appends the values that the array field does not hold yet, the
// first of equal values only. Elements are equal as a whole, numbers by
// value. Otherwise it is DocPush.
func DocAddToSet[T any](field string, values ...T) DocUpdate {
	return DocUpdate{Changes: []DocChange{{Op: DocUpAddToSet, Field: field, Values: toAnys(values)}}}
}

// DocPull removes from the array field every element equal to one of values,
// equal as DocAddToSet takes it. A missing field is left as it is, a null
// one becomes an empty array, and a field of another type is an error.
func DocPull[T any](field string, values ...T) DocUpdate {
	return DocUpdate{Changes: []DocChange{{Op: DocUpPull, Field: field, Values: toAnys(values)}}}
}

// DocPullWhere removes from the array field every object element that q
// matches, each taken as a record: the fields of q are paths in the element.
// An element that is not an object is never removed.
func DocPullWhere(field string, q DocQuery) DocUpdate {
	return DocUpdate{Changes: []DocChange{{Op: DocUpPullWhere, Field: field, Where: &q}}}
}

// DocUpdates combines updates into one, their changes in order.
func DocUpdates(us ...DocUpdate) DocUpdate {
	var changes []DocChange
	for _, u := range us {
		changes = append(changes, u.Changes...)
	}
	return DocUpdate{Changes: changes}
}

// Set adds a DocSet to u.
func (u DocUpdate) Set(field string, value any) DocUpdate { return DocUpdates(u, DocSet(field, value)) }

// Unset adds a DocUnset to u.
func (u DocUpdate) Unset(field string) DocUpdate { return DocUpdates(u, DocUnset(field)) }

// Inc adds a DocInc to u.
func (u DocUpdate) Inc(field string, n any) DocUpdate { return DocUpdates(u, DocInc(field, n)) }
