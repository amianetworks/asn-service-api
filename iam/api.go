// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import "time"

type Instance interface {
	AccountCreate(username, password, email string, phone *Phone) error
	AccountDelete(username string) error
	AccountGet(username string) (*Account, error)
	AccountList() ([]*Account, error)
	AccountInfoUpdate(username, email string, phone *Phone) error
	AccountMetadataUpdate(username, metadata string) error
	AccountPasswordUpdate(username, oldPassword, newPassword string) error
	AccountPasswordReset(username, newPassword string) error
	AccountRecoveryEmailSend(username string) error
	AccountRecoverByEmail(username, newPassword, code string) error

	GroupCreate(groupName string) error
	GroupDelete(groupName string) error
	GroupRename(oldName, newName string) error
	GroupMetadataUpdate(groupName, metadata string) error
	GroupGet(groupName string) (*Group, error)
	GroupList() ([]*Group, error)
	GroupMemberList(groupName string) ([]*Account, error)
	AccountJoinGroup(groupName string, usernames []string) error
	AccountLeaveGroup(groupName string, usernames []string) error

	Login(username, password, deviceID, userClaims string, durationAccess, durationRefresh time.Duration) (needMfa bool, tokenSet *TokenSet, err error)
	PasswordVerify(username, password string) error
	Logout(username, deviceID string) error
	TokenRefresh(userClaims string, tokenSet *TokenSet, durationAccess time.Duration) (*TokenSet, error)
	TokenVerify(accessToken string) (username, deviceID, userClaims string, err error)
	TokenRevoke(accessToken string) error

	MFAVerify(username string, code int32) error
	MFALoginVerify(username string, code int32, deviceID, userClaims string, durationAccess, durationRefresh time.Duration) (*TokenSet, error)
	AuthenticatorBindConfirm(username string, code int32) error
	AuthenticatorBind(username string) (string, error)
	AuthenticatorUnbind(username string) error

	AccessCreate(name string, scope, operation int64, time *TimeControl) error
	AccessUpdate(name string, scope, operation int64, time *TimeControl) error
	AccessDelete(name string) error
	AccessList() ([]*Access, error)
	AccessGrantToGroup(groupName string, accesses []string) error
	AccessRevokeFromGroup(groupName string, accesses []string) error
	AccountAccessList(username string) ([]*Access, error)
	GroupAccessList(groupName string) ([]*Access, error)
}
