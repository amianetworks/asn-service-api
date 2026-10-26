// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package commonapi

import (
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

// The sorting string is the one the framework's database parses.
func TestDocSortingBuild(t *testing.T) {
	s := &DocSorting{}
	s.Descending("Seq")
	s.Ascending("ID")
	got, err := s.Build()
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"fields":["Seq","ID"],"asc":[false,true]}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
