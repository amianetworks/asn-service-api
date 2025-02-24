// Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

type IAM interface {
	AccountCreate(username string, password []byte, info map[string]string) (err error)                                    // AccountCreate allows the user system to create a new account.
	AccountRemove(username string) (err error)                                                                             // AccountRemove allows the user system to remove an existed account by username.
	AccountList(filter map[string]string) (userListFromASN, userListFromService []string, err error)                       // AccountList allows the user system to query the existed account list.
	AccountListReverse(filter, reverseFilter map[string]string) (userListFromASN, userListFromService []string, err error) // AccountListReverse allows the user system to query the existed account list.
	AccountSelfDefineSearch(filter string) (userListFromASN, userListFromService []string, err error)                      // AccountSelfDefineSearch allows the user system to query the existed account list.
	AccountInfoQuery(username string, info []string) (infoMapFromASN, infoMapFromService map[string]string, err error)     // AccountInfoQuery allows the user system to query information of an existed account.
	AccountInfosQuery(username string, info []string) (infoMapFromASN, infoMapFromService map[string][]string, err error)  // AccountInfosQuery allows the user system to query information of an existed account.
	AccountInfoUpdate(username string, info map[string]string) (err error)                                                 // AccountInfoUpdate allows the user system to update information of an existed account.
	AccountPasswordUpdate(username string, oldPassword, newPassword []byte) (err error)                                    // AccountPasswordUpdate allows the user system to update password of an existed account.
	AccountPasswordReset(username string, newPassword []byte) (err error)                                                  // AccountPasswordReset allows the user system to reset password of an existed account.
	AccountRecoverEmailSend(username string) (err error)                                                                   // AccountRecoverEmailSend allows the user system to send a verify code to user email.
	AccountRecoverByEmail(username string, newPassword []byte, code string) (err error)                                    // AccountRecoverByEmail allows the user system to recover account by email.

	UserGroupAdd(groupName string) (groupID string, err error)                       // UserGroupAdd allows the user system to add a new user group.
	UserGroupDelete(groupName string) (err error)                                    // UserGroupDelete allows the user system to delete a new user group.
	UserGroupRename(oldName, newName string) (err error)                             // UserGroupRename allows the user system to rename the user group.
	UserGroupList(filter map[string]string) (groupList map[string]string, err error) // UserGroupList allows the user system to query the existed user group list.
	UserGroupMemberList(groupName string) (memberList []string, err error)           // UserGroupMemberList allows the user system to query the members of the user group.
	UserJoinUserGroup(username []string, userGroupName string) (err error)           // UserJoinUserGroup allows the user system to bind an account to a user group.
	UserLeaveUserGroup(username []string, userGroupName string) (err error)          // UserLeaveUserGroup allows the user system to unbind an account from a user group.

	Login(username string, password []byte, deviceID, userClaims string) (needMFAVerify bool, accessToken, refreshToken string, err error) // Login allows the user system to  login.
	PasswordVerify(username string, password []byte) (err error)                                                                           // PasswordVerify allows the user system to check password of user.
	MFAVerify(username string, mfaCode int32, mfaType string) (err error)                                                                  // MFAVerify allows the user system to perform Multi-Factor Authentication.
	MFALoginVerify(username string, mfaCode int32, mfaType, deviceID, userClaims string) (accessToken, refreshToken string, err error)     // MFALoginVerify allows the user system to perform Multi-Factor Authentication while login.
	AuthenticatorBindConfirm(username string, mfaCode int32, secret string) (err error)                                                    // AuthenticatorBindConfirm allows the user system to perform Multi-Factor Authentication.
	AuthenticatorBind(username string) (qrImg, secret string, err error)                                                                   // AuthenticatorBind allows the user to bind when Authenticator is needed.
	AuthenticatorBindStatus(username string) (bind bool, err error)                                                                        // AuthenticatorBindStatus allows the user to bind when Authenticator is needed.
	AuthenticatorUnbind(username string) (err error)                                                                                       // AuthenticatorUnbind allows the user to unbind when Authenticator isn't needed or rebind.
	Logout(username, deviceID string) (err error)                                                                                          // Logout allows the user logout and redirect to login page.

	TokenRefresh(userClaims, refreshToken, accessToken string) (newAccessToken string, err error)                                            // TokenRefresh allows the user system to refresh access token.
	TokenVerify(accessToken string) (username, deviceID, userClaims string, err error)                                                       // TokenVerify allows the user system to verify valid of access token.
	TokenRevoke(accessToken string) (err error)                                                                                              // TokenRefresh allows the user system to revoke access token.
	RoleAdd(roleName, remark string) (err error)                                                                                             // RoleAdd allows the user system to add a new role.
	RoleDelete(roleName string) (err error)                                                                                                  // RoleDelete allows the user system to delete a role.
	RoleList(filter map[string]string) (roleList map[string]string, err error)                                                               // RoleList allows the user system to query the existed roleList.
	RoleBindUser(username, roleName string) (err error)                                                                                      // RoleBindUser allows the user system to bind a user to a role.
	RoleUnbindUser(username, roleName string) (err error)                                                                                    // RoleUnbindUser allows the user system to unbind a user to a role.
	RoleBindUserGroup(userGroupName, roleName string) (err error)                                                                            // RoleBindUserGroup allows the user system to bind a user group to a role.
	RoleUnBindUserGroup(userGroupName, roleName string) (err error)                                                                          // RoleUnBindUserGroup allows the user system to unbind a user group to a role.
	RoleMemberList(roleName string) (userList, userGroupList []string, err error)                                                            // RoleMemberList allows the user system to query the bound users and user groups of the role.
	AccessControlPolicyAdd(name, scope, operation, time string) (err error)                                                                  // AccessControlPolicyAdd allows the user system to add a new access control policy.
	AccessControlPolicyUpdate(name, scope, operation, time string) (err error)                                                               // AccessControlPolicyUpdate allows the user system to update access control policy.
	AccessControlPolicyDelete(name string) (err error)                                                                                       // AccessControlPolicyDelete allows the user system to delete an access control policy.
	AccessControlPolicyList() (policyList []string, err error)                                                                               // AccessControlPolicyList allows the user system to query the existed access control policy list.
	AccessControlPolicyQuery(name string) (scope, operation, time string, err error)                                                         // AccessControlPolicyQuery allows the user system to query the permission of an existed access control policy.
	AccessControlGrantToUser(name, username string) (err error)                                                                              // AccessControlGrantToUser allows the user system to grant access control to user.
	AccessControlRevokeFromUser(name, username string) (err error)                                                                           // AccessControlRevokeFromUser allows the user system to cancel granted access control of user.
	AccessControlGrantToUserGroup(name, userGroupName string) (err error)                                                                    // AccessControlGrantToUserGroup allows the user system to grant access control to user group.
	AccessControlRevokeFromUserGroup(name, userGroupName string) (err error)                                                                 // AccessControlRevokeFromUserGroup allows the user system to cancel granted access control of user group.
	AccessControlGrantToRole(name, roleName string) (err error)                                                                              // AccessControlGrantToRole allows the user system to grant access control to role.
	AccessControlRevokeFromRole(name, roleName string) (err error)                                                                           // AccessControlRevokeFromRole allows the user system to cancel granted access control of role.
	UserPermissionQuery(username string) (userPermission []string, inheritedFromUserGroup, inheritedFromRole map[string][]string, err error) // UserPermissionQuery allows the user system to query identity and access control of a user.
	UserGroupPermissionQuery(userGroupName string) (userGroupPermission []string, inheritedFromRole map[string][]string, err error)          // UserGroupPermissionQuery allows the user system to query identity and access control of a user group.
	RolePermissionQuery(roleName string) (rolePermission []string, err error)                                                                // RolePermissionQuery allows the user system to query identity and access control of a role.
}
