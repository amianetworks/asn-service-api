// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import (
	"math"
	"reflect"
	"testing"
)

func TestDocQueryBuilders(t *testing.T) {
	cases := []struct {
		name string
		got  DocQuery
		want DocQuery
	}{
		{"in", DocIn("id", []string{"a", "b"}), DocQuery{Op: DocOpIn, Field: "id", Values: []any{"a", "b"}}},
		{"in nil", DocIn[string]("id", nil), DocQuery{Op: DocOpIn, Field: "id"}},
		{"regex", DocRegex("name", "^a", "i"), DocQuery{Op: DocOpRegex, Field: "name", Value: "^a", Options: "i"}},
		{"all of none", DocAllOf(), DocAll()},
		{"all of some", DocAllOf(DocEq("a", 1)), DocQuery{Op: DocOpAnd, Subs: []DocQuery{{Op: DocOpEq, Field: "a", Value: 1}}}},
		{"not", DocNot(DocExists("a", true)), DocQuery{Op: DocOpNot, Subs: []DocQuery{{Op: DocOpExists, Field: "a", Value: true}}}},
	}
	for _, c := range cases {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, c.got, c.want)
		}
	}

	if n := len(DocAllOps()); n != 15 {
		t.Errorf("DocAllOps: %d operators, want 15", n)
	}
}

func TestDocResolveFindOptions(t *testing.T) {
	o, err := DocResolveFindOptions(DocDesc("Seq"), DocAsc("ID"), DocPage(2, 20), DocFields("ID", "Temp"), DocOmit("_id"))
	if err != nil {
		t.Fatal(err)
	}
	want := DocFindOptions{
		Sort:   []DocSortField{{Field: "Seq", Desc: true}, {Field: "ID"}},
		Skip:   40,
		Limit:  20,
		Fields: map[string]bool{"ID": true, "Temp": true, "_id": false},
	}
	if !reflect.DeepEqual(o, want) {
		t.Errorf("got %+v, want %+v", o, want)
	}

	if o, err := DocResolveFindOptions(); err != nil || !reflect.DeepEqual(o, DocFindOptions{}) {
		t.Errorf("none: got %+v, %v", o, err)
	}
	if o, _ := DocResolveFindOptions(DocPage(math.MaxInt, 2)); o.Skip != math.MaxInt {
		t.Errorf("page past MaxInt: skip %d, want MaxInt", o.Skip)
	}
	for name, opt := range map[string]DocFindOption{
		"zero limit":    DocLimit(0),
		"negative page": DocPage(-1, 10),
		"zero size":     DocPage(0, 0),
	} {
		if _, err := DocResolveFindOptions(opt); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
