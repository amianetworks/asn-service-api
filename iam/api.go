// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

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
	AccountDelete(username string) error
	AccountExists(username string) (bool, error)
	AccountGet(username string) (*Account, error)
	AccountList(usernameFuzzy string) ([]*Account, error)
	AccountRename(username, newUsername string) error
	AccountPhoneUpdate(username string, phone *Phone, phoneCode string, skipPhoneValidation bool) error
	AccountEmailUpdate(username, email, emailCode string, skipEmailValidation bool) error
	AccountWeChatUpdate(username, weChatAppID, weChatCode string) error
	AccountAppleUpdate(username, appleIDToken string) error
	AccountMetadataUpdate(username, metadata string) error
	AccountPasswordUpdate(username, oldPassword, newPassword string) error
	AccountPasswordReset(username, newPassword string) error

	AccountRecoverByPhone(username, newPassword, code string) error
	AccountRecoverByEmail(username, newPassword, code string) error

	LoginMethods() (
		usernameAndPassword, emailAndPassword, phoneAndPassword, emailCode, phoneCode, weChat, apple bool,
		err error,
	)
	PasswordVerify(username, countryCode, number, email, password string) error
	LoginWithPassword(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		username, countryCode, number, email, password string,
	) (needMfa bool, tokenSet *TokenSet, err error)
	LoginWithPhone(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		phone *Phone, code string,
	) (needMfa bool, tokenSet *TokenSet, err error)
	LoginWithEmail(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		email, code string,
	) (needMfa bool, tokenSet *TokenSet, err error)
	LoginWithWeChat(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		appID, code string,
	) (needMfa bool, tokenSet *TokenSet, err error)
	LoginWithApple(
		deviceID, userClaims string, durationAccess, durationRefresh time.Duration,
		idToken string,
	) (needMfa bool, tokenSet *TokenSet, err error)
	Logout(username, deviceID string) error

	TokenRefresh(userClaims string, tokenSet *TokenSet, durationAccess time.Duration) (*TokenSet, error)
	TokenVerify(accessToken string) (mfaNeeded bool, username, deviceID, userClaims string, err error)
	TokenRevoke(accessToken string) error

	AccountEnableMFA(username string) error
	AccountDisableMFA(username string) error
	MFALoginVerify(accessToken string, method MfaType, code string) (*TokenSet, error)
	TotpBindConfirm(username, code string) error
	TotpBind(username string) (img, issuer, secret string, err error)
	TotpUnbind(username string) error

	GroupCreate(groupName, metadata string) error
	GroupDelete(groupName string) error
	GroupExists(groupName string) (bool, error)
	GroupRename(oldName, newName string) error
	GroupMetadataUpdate(groupName, metadata string) error
	GroupGet(groupName string) (*Group, error)
	GroupList() ([]*Group, error)
	GroupMemberList(groupName string) ([]*Account, error)
	AccountJoinGroup(groupName string, usernames []string) error
	AccountLeaveGroup(groupName string, usernames []string) error
	AccountGroupList(username string) ([]*Group, error)

	AccessCreate(name, scope, operation string, time *TimeControl) error
	AccessUpdate(name, scope, operation string, time *TimeControl) error
	AccessDelete(name string) error
	AccessExists(name string) (bool, error)
	AccessList() ([]*Access, error)
	AccountAccessList(username string) (map[string][]*Access, error)
	GroupAccessList(groupName string) ([]*Access, error)
	AccessGrantToGroup(groupName string, accesses []string) error
	AccessRevokeFromGroup(groupName string, accesses []string) error
}
