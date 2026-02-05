// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import "time"

type MfaType string

const (
	MfaTypeTotp  MfaType = "totp"
	MfaTypeEmail MfaType = "email"
	MfaTypePhone MfaType = "phone"
)

type PhoneCountryCodeMode string

const (
	PhoneCountryCodeModeAll     PhoneCountryCodeMode = "all"
	PhoneCountryCodeModeInclude PhoneCountryCodeMode = "include"
	PhoneCountryCodeModeExclude PhoneCountryCodeMode = "exclude"
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

	ID          string
	Username    string
	Metadata    string
	DeviceLimit int

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

type DeviceCategory int

const (
	DeviceCategoryUnknown DeviceCategory = 1 + iota
	DeviceCategoryPhone
	DeviceCategoryTablet
	DeviceCategoryWearable
	DeviceCategoryBrowser
	DeviceCategoryMiniProgram
)

type DeviceType int

const (
	DeviceTypeUnknownUnknown DeviceType = 1 + iota
)

const (
	DeviceTypePhoneUnknown DeviceType = 1 + iota
	DeviceTypePhoneIPhone
	DeviceTypePhoneAndroid
)

const (
	DeviceTypeTabletUnknown DeviceType = 1 + iota
	DeviceTypeTabletIPad
	DeviceTypeTabletAndroid
)

const (
	DeviceTypeWearableUnknown DeviceType = 1 + iota
	DeviceTypeWearableAppleWatch
	DeviceTypeWearableAndroidWatch
)

const (
	DeviceTypeBrowserUnknown DeviceType = 1 + iota
	DeviceTypeBrowserChrome
	DeviceTypeBrowserSafari
	DeviceTypeBrowserFirefox
)

const (
	DeviceTypeMiniProgramUnknown DeviceType = 1 + iota
	DeviceTypeMiniProgramPhone
	DeviceTypeMiniProgramComputer
)

type DeviceOs string

const (
	DeviceOsUnknown     DeviceOs = "unknown"
	DeviceOsIos         DeviceOs = "ios"
	DeviceOsAndroid     DeviceOs = "android"
	DeviceOsWatchOS     DeviceOs = "watchos"
	DeviceOsAndroidWear DeviceOs = "android_wear"
	DeviceOsWeChat      DeviceOs = "wechat"
)

type DeviceLanguage int

const (
	DeviceLanguageEN DeviceLanguage = 1 + iota
	DeviceLanguageZH
)

type Device struct {
	ID           string // uuid
	Category     DeviceCategory
	Type         DeviceType
	OS           DeviceOs
	Language     DeviceLanguage
	Name         string
	Model        string
	SerialNumber string
	PushToken    string
	Metadata     string
}
