// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import (
	"time"
)

type RepeatFrequency int

const (
	RepeatFrequencyOnlyOnce RepeatFrequency = 0 + iota
	RepeatFrequencyDaily
	RepeatFrequencyWeekly
	RepeatFrequencyMonthly
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

/*
TimeControl is the structure for tbrc repeat info
Support multiple time ranges in a repeat. For every time in ranges will have the same repeat.
IgnoreLoc decides whether to consider time location when comparing the time. The value cannot be changed after init
Example 1: Repeat every day
  - RepeatFrequency: RepeatDaily
  - RepeatInterval: 1
  - RepeatIndexes: [] // can always be set as nil for daily repeat.

Example 2: Repeat every two weeks
  - RepeatFrequency: RepeatWeekly
  - RepeatInterval: 2
  - RepeatIndexes: [] // repeat will happen at 2 weeks after the start date

Example 3: Repeat every week on Weekdays
  - RepeatFrequency: RepeatWeekly
  - RepeatInterval: 1
  - RepeatIndexes: [1, 2, 3, 4, 5] // repeat will happen only on weekdays. If TimeRange doesn't contain weekdays, no time will be valid

Example 4: Repeat every 2 months on 1st, 15th and 30th
  - RepeatFrequency: RepeatMonthly
  - RepeatInterval: 2
  - RepeatIndexes: [1, 15, 30]
*/
type TimeControl struct {
	TimeRanges []TimeRange

	RepeatFrequency RepeatFrequency
	RepeatEndTime   time.Time
	RepeatInterval  int
	RepeatIndexes   []int

	IgnoreLoc bool
}
