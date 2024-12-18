// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

type IAM interface {
	AccountCreate(username string, password []byte, info map[string]string) (res *IAMError)             // AccountCreate allows the user system to create a new account.
	AccountRemove(username string) (res *IAMError)                                                      // AccountRemove allows the user system to remove an existed account by username.
	AccountList(filter map[string]string) (userList []string, res *IAMError)                            // AccountList allows the user system to query the existed account list.
	AccountListReverse(filter map[string]string) (userList []string, res *IAMError)                     // AccountListReverse allows the user system to query the existed account list.
	AccountSelfDefineSearch(filter, reverseFilter map[string]string) (userList []string, res *IAMError) // AccountSelfDefineSearch allows the user system to query the existed account list.
	AccountInfoQuery(username string, info []string) (infoMap map[string]string, res *IAMError)         // AccountInfoQuery allows the user system to query information of an existed account.
	AccountInfosQuery(username string, info []string) (infoMap map[string][]string, res *IAMError)      // AccountInfosQuery allows the user system to query information of an existed account.
	AccountInfoUpdate(username string, info map[string]string) (res *IAMError)                          // AccountInfoUpdate allows the user system to update information of an existed account.
	AccountPasswordUpdate(username string, oldPassword, newPassword []byte) (res *IAMError)             // AccountPasswordUpdate allows the user system to update password of an existed account.
	AccountPasswordReset(username string, newPassword []byte) (res *IAMError)                           // AccountPasswordReset allows the user system to reset password of an existed account.
	AccountRecoverEmailSend(username string) (res *IAMError)                                            // AccountRecoverEmailSend allows the user system to send a verify code to user email.
	AccountRecoverByEmail(username string, newPassword []byte, code string) (res *IAMError)             // AccountRecoverByEmail allows the user system to recover account by email.

	UserGroupAdd(sysType, groupName string) (groupID string, res *IAMError)                              // UserGroupAdd allows the user system to add a new user group.
	UserGroupDelete(sysType, groupID string) (res *IAMError)                                             // UserGroupDelete allows the user system to delete a new user group.
	UserGroupRename(sysType, groupID, groupName string) (res *IAMError)                                  // UserGroupRename allows the user system to rename the user group.
	UserGroupList(sysType string, filter map[string]string) (groupList map[string]string, res *IAMError) // UserGroupList allows the user system to query the existed user group list.
	UserGroupMemberList(sysType, groupID string) (memberList []string, res *IAMError)                    // UserGroupMemberList allows the user system to query the members of the user group.
	UserJoinUserGroup(sysType string, username []string, userGroupID string) (res *IAMError)             // UserJoinUserGroup allows the user system to bind a account to a user group.
	UserLeaveUserGroup(sysType string, username []string, userGroupID string) (res *IAMError)            // UserLeaveUserGroup allows the user system to unbind a account from a user group.

	Login(sysType, username string, password []byte, entity *EntityBaseInfo) (needMFAVerify bool, qrImg, secret, accessToken, refreshToken string) // Login allows the user system to  login.
	PasswordVerify(username string, password []byte) (res *IAMError)                                                                               // PasswordVerify allows the user system to check password of user.
	MFAVerify(sysType, username string, mfaCode int, mfaType string) (res *IAMError)                                                               // MFAVerify allows the user system to perform Multi-Factor Authentication.
	MFALoginVerify(sysType, username string, mfaCode int, mfaType string) (accessToken, refreshToken string, res *IAMError)                        // MFALoginVerify allows the user system to perform Multi-Factor Authentication while login.
	AuthenticatorBindConfirm(sysType, username string, mfaCode int, secret string) (res *IAMError)                                                 // AuthenticatorBindConfirm allows the user system to perform Multi-Factor Authentication.
	AuthenticatorBind(sysType, username string, mfaCode int, secret string) (res *IAMError)                                                        // AuthenticatorBind allows the user to bind when Authenticator is needed.
	AuthenticatorBindStatus(sysType, username string) (bind bool, res *IAMError)                                                                   // AuthenticatorBindStatus allows the user to bind when Authenticator is needed.
	AuthenticatorUnbind(sysType, username string) (res *IAMError)                                                                                  // AuthenticatorUnbind allows the user to unbind when Authenticator isn't needed or rebind.
	EntityList(sysType, username string) (info map[string]*EntityBaseInfo, res *IAMError)                                                          // EntityList allows the admin to get all bound entities of target user.
	EntityNameUpdate(sysType, username, entityID, entityName string) (res *IAMError)                                                               // EntityNameUpdate allows the admin to rename entity.
	EntityAdd(sysType, username string, info *EntityBaseInfo) (res *IAMError)                                                                      // EntityAdd allows the admin to add bound entities of target user.
	EntityDelete(sysType, username, entityID string) (res *IAMError)                                                                               // EntityDelete allows the admin to delete bound entities of target user.
	Logout(sysType, username string) (res *IAMError)                                                                                               // Logout allows the user logout and redirect to login page.

	TokenRefresh(sysType, username, deviceCode, refreshToken, accessToken string) (newAccessToken string, res *IAMError)                                  // TokenRefresh allows the user system to refresh access token.
	TokenVerify(sysType, username, deviceCode, accessToken string) (res *IAMError)                                                                        // TokenVerify allows the user system to verify valid of access token.
	RoleAdd(sysType, roleName, remark string) (res *IAMError)                                                                                             // RoleAdd allows the user system to add a new role.
	RoleDelete(sysType, roleName string) (res *IAMError)                                                                                                  // RoleDelete allows the user system to delete a role.
	RoleList(sysType string, filter map[string]string) (roleList map[string]string, res *IAMError)                                                        // RoleList allows the user system to query the existed roleList.
	RoleBindUser(sysType, username, roleName string) (res *IAMError)                                                                                      // RoleBindUser allows the user system to bind a user to a role.
	RoleUnbindUser(sysType, username, roleName string) (res *IAMError)                                                                                    // RoleUnbindUser allows the user system to unbind a user to a role.
	RoleBindUserGroup(sysType, userGroupID, roleName string) (res *IAMError)                                                                              // RoleBindUserGroup allows the user system to bind a user group to a role.
	RoleUnBindUserGroup(sysType, userGroupID, roleName string) (res *IAMError)                                                                            // RoleUnBindUserGroup allows the user system to unbind a user group to a role.
	RoleMemberList(sysType, roleName string) (userList, userGroupList []string, res *IAMError)                                                            // RoleMemberList allows the user system to query the bound users and user groups of the role.
	AccessControlPolicyAdd(sysType, name, scope, operation, time string) (res *IAMError)                                                                  // AccessControlPolicyAdd allows the user system to add a new access control policy.
	AccessControlPolicyUpdate(sysType, name, scope, operation, time string) (res *IAMError)                                                               // AccessControlPolicyUpdate allows the user system to update access control policy.
	AccessControlPolicyDelete(sysType, name string) (res *IAMError)                                                                                       // AccessControlPolicyDelete allows the user system to delete an access control policy.
	AccessControlPolicyList(sysType string) (policyList []string, res *IAMError)                                                                          // AccessControlPolicyList allows the user system to query the existed access control policy list.
	AccessControlPolicyQuery(sysType, name string) (scope, operation, time string, res *IAMError)                                                         // AccessControlPolicyQuery allows the user system to query the permission of an existed access control policy.
	AccessControlGrantToUser(sysType, name, username string) (res *IAMError)                                                                              // AccessControlGrantToUser allows the user system to grant access control to user.
	AccessControlRevokeFromUser(sysType, name, username string) (res *IAMError)                                                                           // AccessControlRevokeFromUser allows the user system to cancel granted access control of user.
	AccessControlGrantToUserGroup(sysType, name, userGroupID string) (res *IAMError)                                                                      // AccessControlGrantToUserGroup allows the user system to grant access control to user group.
	AccessControlRevokeFromUserGroup(sysType, name, userGroupID string) (res *IAMError)                                                                   // AccessControlRevokeFromUserGroup allows the user system to cancel granted access control of user group.
	AccessControlGrantToRole(sysType, name, roleName string) (res *IAMError)                                                                              // AccessControlGrantToRole allows the user system to grant access control to role.
	AccessControlRevokeFromRole(sysType, name, roleName string) (res *IAMError)                                                                           // AccessControlRevokeFromRole allows the user system to cancel granted access control of role.
	UserPermissionQuery(sysType, username string) (userPermission []string, inheritedFromUserGroup, inheritedFromRole map[string][]string, res *IAMError) // UserPermissionQuery allows the user system to query identity and access control of a user.
	UserGroupPermissionQuery(sysType, userGroupID string) (userGroupPermission []string, inheritedFromRole map[string][]string, res *IAMError)            // UserGroupPermissionQuery allows the user system to query identity and access control of a user group.
	RolePermissionQuery(sysType, roleName string) (rolePermission []string, res *IAMError)                                                                // RolePermissionQuery allows the user system to query identity and access control of a role.
}

type EntityBaseInfo struct {
	ID   string
	Name string
	Type string
	OS   string
}

type IAMError struct {
	ErrCode int
	ErrInfo string
}
