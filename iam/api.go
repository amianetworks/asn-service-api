// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import (
	"net/http"
	"time"
)

type Instance interface {
	JWKSGet() (string, error)
	RenameEnabled() bool
	MfaEnforced() bool

	ServiceMfaSet(mfaRequired bool) error

	SupportedCountryCodesGet() (PhoneCountryCodeMode, []string)
	AccountPhoneSend(countryCode, number string, sendWhenExistOnly bool) (string, error)
	AccountEmailSend(email string, sendWhenExistOnly bool) (string, error)

	AccountCreate(
		username, metadata string,
		password string,
		email, emailCode string, skipEmailValidation bool,
		phone *Phone, phoneCode string, skipPhoneValidation bool,
	) (accountID string, err error)
	AccountDelete(accountID string) error
	AccountExists(accountID string) (bool, error)
	AccountGet(accountID, username, countryCode, number, email string) (*Account, error)
	AccountList(username, countryCode, number, email string) ([]*Account, error)
	AccountListByIDs(accountIDs []string) ([]*Account, error)
	AccountRename(accountID, newUsername string) error
	AccountPhoneUpdate(accountID string, phone *Phone, phoneCode string, skipPhoneValidation bool) error
	AccountEmailUpdate(accountID, email, emailCode string, skipEmailValidation bool) error
	AccountWeChatUpdate(accountID, weChatAppID, weChatCode string) error
	AccountAppleUpdate(accountID, appleIDToken string) error
	AccountGoogleUpdate(accountID, googleIDToken string) error
	AccountMetadataUpdate(accountID, metadata string) error
	AccountPasswordUpdate(accountID, oldPassword, newPassword string) error
	AccountPasswordReset(accountID, newPassword string) error
	AccountPasskeyBindChallengeGet(domain, accountID string) (sessionID, data string, err error)
	AccountPasskeyBind(domain, accountID, sessionID, deviceID, data string) error
	AccountPasskeyUnbind(accountID, id string) error
	AccountPasskeyUnbindAll(accountID string) error

	LoginMethods() (
		usernameAndPassword, emailAndPassword, phoneAndPassword, emailCode, phoneCode, weChat, apple, google, passkey bool,
		err error,
	)
	PasswordVerify(username, countryCode, number, email, password string) error
	LoginOrCreateWithPassword(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		username, countryCode, number, email, password string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
	LoginOrCreateWithPhone(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		phone *Phone, code string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
	LoginOrCreateWithEmail(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		email, code string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
	LoginOrCreateWithWeChat(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		appID, code string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
	LoginOrCreateWithApple(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		idToken string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
	LoginOrCreateWithGoogle(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		idToken string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
	AccountPasskeyLoginChallengeGet(domain string) (sessionID, data string, err error)
	AccountPasskeyAuth(
		device *DeviceInfo,
		userClaims string, durationAccess, durationRefresh time.Duration,
		domain, sessionID, data string,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)
	Logout(accountID, deviceID string) error
	AppleRedirect(w http.ResponseWriter, r *http.Request)

	TokenRefresh(userClaims string, tokenSet *TokenSet, durationAccess time.Duration) (*TokenSet, error)
	TokenVerify(accessToken string) (mfaNeeded bool, accountID, username, deviceID, userClaims string, err error)
	TokenRevoke(accessToken string) error

	DeviceLimitUpdate(accountID string, limit int) error
	DeviceInfoUpdate(device *Device) error
	DeviceDelete(accountID, deviceID string) error

	AccountEnableMFA(accountID string) error
	AccountDisableMFA(accountID string) error
	MFALoginVerify(accessToken string, method MfaType, code, domain, sessionID, data string) (*TokenSet, error)
	TotpBindConfirm(accountID, code string) error
	TotpBind(accountID string) (img, issuer, secret string, err error)
	TotpUnbind(accountID string) error

	GroupCreate(groupName, metadata string) error
	GroupDelete(groupName string) error
	GroupExists(groupName string) (bool, error)
	GroupRename(oldName, newName string) error
	GroupMetadataUpdate(groupName, metadata string) error
	GroupGet(groupName string) (*Group, error)
	GroupList() ([]*Group, error)
	GroupMemberList(groupName string) ([]*Account, error)
	AccountJoinGroup(groupName string, accountIDs []string) error
	AccountLeaveGroup(groupName string, accountIDs []string) error
	AccountGroupList(accountID string) ([]*Group, error)

	AccessCreate(name, scope, operation string, time *TimeControl) error
	AccessUpdate(name, scope, operation string, time *TimeControl) error
	AccessDelete(name string) error
	AccessExists(name string) (bool, error)
	AccessList() ([]*Access, error)
	AccountAccessList(accountID string) (map[string][]*Access, error)
	GroupAccessList(groupName string) ([]*Access, error)
	AccessGrantToGroup(groupName string, accesses []string) error
	AccessRevokeFromGroup(groupName string, accesses []string) error
}
