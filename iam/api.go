// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import "time"

type Instance interface {
	JWKSGet() (string, error)

	ServiceMfaSet(mfaRequired bool) error

	AccountPhoneSend(countryCode, number string) (string, error)
	AccountEmailSend(email string) (string, error)

	AccountCreate(
		username, metadata string,
		password string,
		email, emailCode string, skipEmailValidation bool,
		phone *Phone, phoneCode string, skipPhoneValidation bool,
		weChatAppID, weChatCode string,
		appleIDToken string,
	) (string, error)
	AccountDelete(accountID string) error
	AccountExists(accountID string) (bool, error)
	AccountGet(accountID, username, countryCode, number, email string) (*Account, error)
	AccountList(username, countryCode, number, email string) ([]*Account, error)
	AccountRenameAllowed() (bool, error)
	AccountRename(accountID, newUsername string) error
	AccountPhoneUpdate(accountID string, phone *Phone, phoneCode string, skipPhoneValidation bool) error
	AccountEmailUpdate(accountID, email, emailCode string, skipEmailValidation bool) error
	AccountWeChatUpdate(accountID, weChatAppID, weChatCode string) error
	AccountAppleUpdate(accountID, appleIDToken string) error
	AccountMetadataUpdate(accountID, metadata string) error
	AccountPasswordUpdate(accountID, oldPassword, newPassword string) error
	AccountPasswordReset(accountID, newPassword string) error

	AccountRecoverByPhone(accountID, newPassword, code string) error
	AccountRecoverByEmail(accountID, newPassword, code string) error

	LoginMethods() (
		usernameAndPassword, emailAndPassword, phoneAndPassword, emailCode, phoneCode, weChat, apple bool,
		err error,
	)
	PasswordVerify(username, countryCode, number, email, password string) error
	LoginWithPassword(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		username, countryCode, number, email, password string,
	) (accountID string, needMfa bool, tokenSet *TokenSet, err error)
	LoginWithPhone(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		phone *Phone, code string,
	) (accountID string, needMfa bool, tokenSet *TokenSet, err error)
	LoginWithEmail(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		email, code string,
	) (accountID string, needMfa bool, tokenSet *TokenSet, err error)
	LoginWithWeChat(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		appID, code string,
	) (accountID string, needMfa bool, tokenSet *TokenSet, err error)
	LoginWithApple(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		idToken string,
	) (accountID string, needMfa bool, tokenSet *TokenSet, err error)
	Logout(accountID, deviceID string) error

	TokenRefresh(userClaims string, tokenSet *TokenSet, durationAccess time.Duration) (*TokenSet, error)
	TokenVerify(accessToken string) (mfaNeeded bool, accountID, username, deviceID, userClaims string, err error)
	TokenRevoke(accessToken string) error

	AccountEnableMFA(accountID string) error
	AccountDisableMFA(accountID string) error
	MFALoginVerify(accessToken string, method MfaType, code string) (*TokenSet, error)
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
