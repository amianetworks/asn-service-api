// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import "time"

type TimeInfo struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Access struct {
	TimeInfo   TimeInfo
	AccessName string
	Scope      string
	Operation  string
	Time       *TimeControl
}

type Account struct {
	TimeInfo   TimeInfo
	Username   string
	Email      string
	Phone      Phone
	MfaEnabled bool

	Metadata string
}

type Phone struct {
	CountryCode string
	Number      string
}

type Group struct {
	TimeInfo     TimeInfo
	GroupName    string
	GroupMembers int

	Metadata string
}

type TokenSet struct {
	AccessToken  string
	RefreshToken string
}
