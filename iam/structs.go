// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import "time"

type CodeSendResult int

const (
	CodeSendResultSentOrIgnored CodeSendResult = 1 + iota
	CodeSendResultCodeReturned
	CodeSendResultForbidden
	CodeSendResultCooldownRequired
)

type MfaType string

const (
	MfaTypeTotp  MfaType = "totp"
	MfaTypeEmail MfaType = "email"
	MfaTypeSms   MfaType = "sms"
)

// MfaMethodInfo is one MFA option offered to the client during a login flow.
// MaskedTarget is a masked rendering of the delivery channel (e.g. "1***5678"
// or "a***@example.com") for SMS/Email, and empty for TOTP which has
// no delivery target. CountryCode is set only for SMS.
type MfaMethodInfo struct {
	Method       MfaType
	MaskedTarget string
	CountryCode  string
}

// CredentialMethod encodes the identity kind + proof kind chosen when starting a
// stateful login flow with AuthFlowInit.
type CredentialMethod string

const (
	CredentialUsernamePassword CredentialMethod = "username_password"
	CredentialPhonePassword    CredentialMethod = "phone_password"
	CredentialEmailPassword    CredentialMethod = "email_password"
	CredentialPhoneCode        CredentialMethod = "phone_code"
	CredentialEmailCode        CredentialMethod = "email_code"
	CredentialWeChat           CredentialMethod = "wechat"
	CredentialApple            CredentialMethod = "apple"
	CredentialGoogle           CredentialMethod = "google"
	CredentialPasskey          CredentialMethod = "passkey"
)

// PasswordVerifyMethod is the proof a user supplies during the password flow.
type PasswordVerifyMethod string

const (
	PasswordVerifyOldPassword PasswordVerifyMethod = "old_password"
	PasswordVerifyEmail       PasswordVerifyMethod = "email"
	PasswordVerifySMS         PasswordVerifyMethod = "sms"
)

// PasswordFlowInitState is the outcome of AccountPasswordFlowInit.
type PasswordFlowInitState int

const (
	PasswordFlowInitUnspecified PasswordFlowInitState = iota
	// PasswordFlowInitMethodRequired: flowToken and methods are valid; pick a method.
	PasswordFlowInitMethodRequired
	// PasswordFlowInitNoRecovery: account not found OR no usable verification method.
	// No flow token is issued.
	PasswordFlowInitNoRecovery
)

// PasswordFlowMethod is one verification option offered after init. MaskedTarget is
// a masked rendering of the delivery channel (e.g. "a***@example.com"); it is empty
// for PasswordVerifyOldPassword. CountryCode is set only for PasswordVerifySMS.
type PasswordFlowMethod struct {
	Method       PasswordVerifyMethod
	MaskedTarget string
	CountryCode  string
}

// PasswordFlowInspectResult is the read-only resolution of a password-flow token,
// returned by AccountPasswordFlowInspect. It reports the bound account plus the
// current flow status; the call has no side effects (attempts / pending method /
// verified state are untouched). Unlike the auth flow there is no device dimension.
type PasswordFlowInspectResult struct {
	AccountID         string
	Verified          bool                 // whether a proof has already succeeded (may proceed to Complete)
	PendingMethod     PasswordVerifyMethod // OTP method chosen by the last SendCode; empty if none
	ExpireAt          time.Time            // flow TTL
	AttemptsRemaining int

	// the bound account's identity (plaintext).
	Username string
	Email    string
	Phone    Phone

	// the verification methods the account can use, same set/masking as PasswordFlowInit.
	Methods []PasswordFlowMethod
}

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
	Version      string
	PushToken    string
	Metadata     string
}

// FlowPhase identifies which phase an auth-flow token belongs to. It decides which
// fields of AuthFlowInspectResult are populated:
//   - FlowPhaseCredential: the in-memory credential-phase handle. Method / ExpireAt /
//     the init-supplied identity + device are set; AccountID / DeviceID are empty
//     (the account may not exist yet).
//   - FlowPhaseMFA: the MFA-unverified access token. AccountID / DeviceID / the
//     account's bound identity + registered device / AvailableMfaMethods are set.
type FlowPhase int

const (
	FlowPhaseUnspecified FlowPhase = iota
	FlowPhaseCredential
	FlowPhaseMFA
)

// AuthFlowInspectResult is the read-only resolution of an auth-flow token, returned
// by AuthFlowInspect. Which fields are populated depends on Phase (see FlowPhase).
// The call has no side effects (attempts / pending method / token are untouched).
type AuthFlowInspectResult struct {
	Phase             FlowPhase
	State             LoginFlowState   // CREDENTIAL: ChallengeRequired / VerifyRequired; MFA: MFAVerify / MFASetup
	Method            CredentialMethod // CREDENTIAL phase only
	ExpireAt          time.Time        // CREDENTIAL phase (from the in-memory flow TTL)
	AttemptsRemaining int              // verification attempts left before the flow is exhausted

	// identity: CREDENTIAL phase carries the value captured at AuthFlowInit (only the
	// field matching the method is set; empty for third-party methods); MFA phase carries
	// the account's bound values.
	AccountID string // MFA phase only (the account is known)
	Username  string
	Email     string
	Phone     Phone

	// device: CREDENTIAL phase echoes the descriptor supplied at AuthFlowInit; MFA phase
	// carries the registered device record for the token's device.
	Device   *DeviceInfo
	DeviceID string // MFA phase only

	// MFA phase: the factors the account can use, each with a masked delivery target.
	AvailableMfaMethods []MfaMethodInfo
}

type LoginFlowState int

const (
	LoginFlowAuthenticated     LoginFlowState = iota // login complete; token_set is valid
	LoginFlowMFAVerify                               // credentials OK; must verify an existing MFA factor
	LoginFlowMFASetup                                // credentials OK; must bind an MFA factor first
	LoginFlowChallengeRequired                       // call AuthFlowGetChallenge (send OTP / get passkey challenge), then AuthFlowVerify
	LoginFlowVerifyRequired                          // call AuthFlowVerify with the proof (password / code / id_token / passkey assertion)
)
