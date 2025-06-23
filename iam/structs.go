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
	Scope      int64
	Operation  int64
	Time       *TimeControl
}

type Account struct {
	TimeInfo   TimeInfo
	Username   string
	Email      string
	Phone      Phone
	MfaEnabled bool
}

type Phone struct {
	CountryCode string
	Number      string
}

type Group struct {
	TimeInfo     TimeInfo
	GroupName    string
	GroupMembers int
}

type TokenSet struct {
	AccessToken  string
	RefreshToken string
}

type TimeControl struct {
	TimeRanges []TimeRange

	RepeatFrequency RepeatFrequency
	RepeatEndTime   time.Time
	RepeatInterval  int
	RepeatIndexes   []int

	IgnoreLoc bool
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type RepeatFrequency int

const (
	RepeatFrequencyOnlyOnce RepeatFrequency = 0 + iota
	RepeatFrequencyDaily
	RepeatFrequencyWeekly
	RepeatFrequencyMonthly
)
