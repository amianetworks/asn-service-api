// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import "time"

type MfaType string

const (
	MfaTypeTotp    MfaType = "totp"
	MfaTypeEmail   MfaType = "email"
	MfaTypeSms     MfaType = "sms"
	MfaTypePasskey MfaType = "passkey"
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

	ID               string
	Username         string
	UsernameModified bool
	Metadata         string
	DeviceLimit      int

	Password   bool
	Phone      Phone
	Email      string
	Totp       bool
	WeChat     *AccountWeChatInfo
	Apple      *AccountAppleInfo
	Google     *AccountGoogleInfo
	Passkeys   []AccountPasskey
	MfaEnabled bool

	// ServiceAdmin is true for accounts managed by ASN Controller rather than by the service itself.
	// The service cannot create, delete, or modify these accounts, but must grant them full access.
	// These accounts appear in AccountList and AccountListByIDs results.
	ServiceAdmin bool
	Groups       []string

	Devices map[string]*Device
}

type AccountPasskey struct {
	ID           string
	DeviceID     string
	CredentialID string
}

type Phone struct {
	CountryCode string
	Number      string
}

type AccountWeChatInfo struct {
	Bound      bool
	Nickname   string
	HeadImgURL string
}

type AccountAppleInfo struct {
	Bound          bool
	Email          string
	EmailVerified  bool
	IsPrivateEmail bool
}

type AccountGoogleInfo struct {
	Bound         bool
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
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
	DeviceCategoryPC
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

const (
	DeviceTypePCUnknown DeviceType = 1 + iota
	DeviceTypePCWindows
	DeviceTypePCMacOS
	DeviceTypePCLinux
)

type DeviceOs string

const (
	DeviceOsUnknown     DeviceOs = "unknown"
	DeviceOsIos         DeviceOs = "ios"
	DeviceOsAndroid     DeviceOs = "android"
	DeviceOsWatchOS     DeviceOs = "watchos"
	DeviceOsAndroidWear DeviceOs = "android_wear"
	DeviceOsWeChat      DeviceOs = "wechat"
	DeviceOsWindows     DeviceOs = "windows"
	DeviceOsMacOS       DeviceOs = "macos"
	DeviceOsLinux       DeviceOs = "linux"
)

type DeviceLanguage int

const (
	DeviceLanguageEN DeviceLanguage = 1 + iota
	DeviceLanguageZH
)

type Device struct {
	ID string // uuid
	DeviceInfo
}

type DeviceInfo struct {
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
