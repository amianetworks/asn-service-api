// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import "time"

type MfaType string

const (
	MfaTypeTotp  MfaType = "totp"
	MfaTypeEmail MfaType = "email"
	MfaTypePhone MfaType = "phone"
)

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
	TimeInfo TimeInfo

	ID       string
	Username string
	Metadata string

	Password   bool
	Phone      Phone
	Email      string
	Totp       bool
	WeChat     bool
	Apple      bool
	Google     bool
	MfaEnabled bool

	ServiceAdmin bool
	Groups       []string
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
