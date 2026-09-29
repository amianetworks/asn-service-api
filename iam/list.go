// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

// Page selects a window of a sorted list result. The zero value selects everything.
type Page struct {
	Offset int // must be >= 0
	Limit  int // 0 = no limit; at most 500
}

// AccountListFilter restricts AccountList. Conditions are ANDed across fields; the
// values within one list are ORed. Zero values do not filter. At most 1000 values
// per list, none of them empty.
type AccountListFilter struct {
	IDs           []string // unknown or malformed IDs are ignored
	Groups        []string // group names; member of any
	ExcludeGroups []string // group names; member of none
	// Keyword is a case-insensitive substring of the username, email, phone number,
	// or the WeChat nickname, Google name, Google email or Apple email bound under
	// the service.
	Keyword    string
	MfaEnabled *bool // nil = no filter
}

// AccountSortField is what AccountList sorts by.
type AccountSortField int

const (
	AccountSortCreatedAt AccountSortField = iota
	AccountSortUsername
)

// AccountListSort orders AccountList. Ties are always broken by ID ascending.
type AccountListSort struct {
	Field AccountSortField
	Desc  bool
}

// GroupListFilter restricts GroupList. Conditions are ANDed across fields; the
// values within one list are ORed. Names are matched literally.
type GroupListFilter struct {
	Names    []string // exact names
	Prefixes []string // name starts with
	Suffixes []string // name ends with
	Members  []string // account IDs; group contains any
}

// GroupSortField is what GroupList sorts by.
type GroupSortField int

const (
	GroupSortCreatedAt GroupSortField = iota
	GroupSortName
)

// GroupListSort orders GroupList. Ties are always broken by ID ascending.
type GroupListSort struct {
	Field GroupSortField
	Desc  bool
}
