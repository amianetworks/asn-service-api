// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

type IAM interface {
	/*
		Account
	*/

	// CreateAccount allows the user system to create a new account.
	CreateAccount(username string, password []byte, info map[string]string) (err error)
	// RemoveAccount allows the user system to remove an existed account by username.
	RemoveAccount(username string) (err error)
	// ListAccounts allows the user system to query the existed account list.
	ListAccounts(filter map[string]string) (userListFromASN, userListFromService []string, err error)
	// ListAccountsWithReverseQuery allows the user system to query the existed account list.
	ListAccountsWithReverseQuery(filter, reverseFilter map[string]string) (userListFromASN, userListFromService []string, err error)
	// ListAccountsWithSelfDefinedSearch allows the user system to query the existed account list.
	ListAccountsWithSelfDefinedSearch(filter string) (userListFromASN, userListFromService []string, err error)
	// QueryAccountInfo allows the user system to query information of an existed account.
	QueryAccountInfo(username string, info []string) (infoMapFromASN, infoMapFromService map[string]string, err error)
	// QueryAccountsInfo allows the user system to query information of an existed account.
	QueryAccountsInfo(username string, info []string) (infoMapFromASN, infoMapFromService map[string][]string, err error)
	// UpdateAccountInfo allows the user system to update information of an existed account.
	UpdateAccountInfo(username string, info map[string]string) (err error)
	// UpdateAccountPassword allows the user system to update password of an existed account.
	UpdateAccountPassword(username string, oldPassword, newPassword []byte) (err error)
	// ResetAccountPassword allows the user system to reset password of an existed account.
	ResetAccountPassword(username string, newPassword []byte) (err error)
	// SendAccountRecoverEmail allows the user system to send a verify code to user email.
	SendAccountRecoverEmail(username string) (err error)
	// RecoverAccountWithEmail allows the user system to recover account by email.
	RecoverAccountWithEmail(username string, newPassword []byte, code string) (err error)

	/*
		User Group
	*/

	// AddUserGroup allows the user system to add a new user group.
	AddUserGroup(groupName string) (groupID string, err error)
	// DeleteUserGroup allows the user system to delete a new user group.
	DeleteUserGroup(groupName string) (err error)
	// RenameUserGroup allows the user system to rename the user group.
	RenameUserGroup(oldName, newName string) (err error)
	// ListUserGroups allows the user system to query the existed user group list.
	ListUserGroups(filter map[string]string) (groupList map[string]string, err error)
	// ListUserGroupUsers allows the user system to query the users of the user group.
	ListUserGroupUsers(groupName string) (userList []string, err error)
	// AddUsersToUserGroup allows the user system to bind an account to a user group.
	AddUsersToUserGroup(username []string, userGroupName string) (err error)
	// RemoveUsersFromUserGroup allows the user system to unbind an account from a user group.
	RemoveUsersFromUserGroup(username []string, userGroupName string) (err error)

	/*
		Authorization
	*/

	// Login allows the user system to  login.
	Login(username string, password []byte, deviceID, userClaims string) (needMFAVerify bool, accessToken, refreshToken string, err error)
	// VerifyPassword allows the user system to check password of user.
	VerifyPassword(username string, password []byte) (err error)
	// VerifyMFA allows the user system to perform Multi-Factor Authentication.
	VerifyMFA(username string, mfaCode int32, mfaType string) (err error)
	// VerifyMFALogin allows the user system to perform Multi-Factor Authentication while login.
	VerifyMFALogin(username string, mfaCode int32, mfaType, deviceID, userClaims string) (accessToken, refreshToken string, err error)
	// ConfirmAuthenticatorBinding allows the user system to perform Multi-Factor Authentication.
	ConfirmAuthenticatorBinding(username string, mfaCode int32, secret string) (err error)
	// BindAuthenticator allows the user to bind when Authenticator is needed.
	BindAuthenticator(username string) (qrImg, secret string, err error)
	// GetAuthenticatorBindStatus allows the user to bind when Authenticator is needed.
	GetAuthenticatorBindStatus(username string) (bind bool, err error)
	// UnbindAuthenticator allows the user to unbind when Authenticator isn't needed or rebind.
	UnbindAuthenticator(username string) (err error)
	// Logout allows the user logout and redirect to login page.
	Logout(username, deviceID string) (err error)

	/*
		Authentication
	*/

	// RefreshToken allows the user system to refresh access token.
	RefreshToken(userClaims, refreshToken, accessToken string) (newAccessToken string, err error)
	// VerifyToken allows the user system to verify valid of access token.
	VerifyToken(accessToken string) (username, deviceID, userClaims string, err error)
	// RevokeToken allows the user system to revoke access token.
	RevokeToken(accessToken string) (err error)

	/*
		Role
	*/

	// AddRole allows the user system to add a new role.
	AddRole(roleName, remark string) (err error)
	// DeleteRole allows the user system to delete a role.
	DeleteRole(roleName string) (err error)
	// ListRoles allows the user system to query the existed roleList.
	ListRoles(filter map[string]string) (roleList map[string]string, err error)
	// BindRoleToUser allows the user system to bind a user to a role.
	BindRoleToUser(username, roleName string) (err error)
	// UnbindRoleFromUser allows the user system to unbind a user to a role.
	UnbindRoleFromUser(username, roleName string) (err error)
	// BindRoleToUserGroup allows the user system to bind a user group to a role.
	BindRoleToUserGroup(userGroupName, roleName string) (err error)
	// UnbindRoleFromUserGroup allows the user system to unbind a user group to a role.
	UnbindRoleFromUserGroup(userGroupName, roleName string) (err error)
	// ListRoleUsers allows the user system to query the bound users and user groups of the role.
	ListRoleUsers(roleName string) (userList, userGroupList []string, err error)

	/*
		Access Control
	*/

	// AddAccessControlPolicy allows the user system to add a new access control policy.
	AddAccessControlPolicy(name, scope, operation, time string) (err error)
	// UpdateAccessControlPolicy allows the user system to update access control policy.
	UpdateAccessControlPolicy(name, scope, operation, time string) (err error)
	// DeleteAccessControlPolicy allows the user system to delete an access control policy.
	DeleteAccessControlPolicy(name string) (err error)
	// ListAccessControlPolicies allows the user system to query the existed access control policy list.
	ListAccessControlPolicies() (policyList []string, err error)
	// QueryAccessControlPolicies allows the user system to query the permission of an existed access control policy.
	QueryAccessControlPolicies(name string) (scope, operation, time string, err error)
	// GrantAccessToUser allows the user system to grant access control to user.
	GrantAccessToUser(name, username string) (err error)
	// RevokeAccessFromUser allows the user system to cancel granted access control of user.
	RevokeAccessFromUser(name, username string) (err error)
	// GrantAccessToUserGroup allows the user system to grant access control to user group.
	GrantAccessToUserGroup(name, userGroupName string) (err error)
	// RevokeAccessFromUserGroup allows the user system to cancel granted access control of user group.
	RevokeAccessFromUserGroup(name, userGroupName string) (err error)
	// GrantAccessToRole allows the user system to grant access control to role.
	GrantAccessToRole(name, roleName string) (err error)
	// RevokeAccessFromRole allows the user system to cancel granted access control of role.
	RevokeAccessFromRole(name, roleName string) (err error)
	// QueryUserPermissions allows the user system to query identity and access control of a user.
	QueryUserPermissions(username string) (userPermission []string, inheritedFromUserGroup, inheritedFromRole map[string][]string, err error)
	// QueryUserGroupPermissions allows the user system to query identity and access control of a user group.
	QueryUserGroupPermissions(userGroupName string) (userGroupPermission []string, inheritedFromRole map[string][]string, err error)
	// QueryRolePermissions allows the user system to query identity and access control of a role.
	QueryRolePermissions(roleName string) (rolePermission []string, err error)
}
