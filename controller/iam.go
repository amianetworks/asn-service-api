// Copyright 2024 Amiasys Corporation and/or its affiliates. All rights reserved.

package capi

type IAM interface {
	AccountCreate(username string, password []byte, info map[string]string) (err error)             // AccountCreate allows the user system to create a new account.
	AccountRemove(username string) (err error)                                                      // AccountRemove allows the user system to remove an existed account by username.
	AccountList(filter map[string]string) (userList []string, err error)                            // AccountList allows the user system to query the existed account list.
	AccountListReverse(filter map[string]string) (userList []string, err error)                     // AccountListReverse allows the user system to query the existed account list.
	AccountSelfDefineSearch(filter, reverseFilter map[string]string) (userList []string, err error) // AccountSelfDefineSearch allows the user system to query the existed account list.
	AccountInfoQuery(username string, info []string) (infoMap map[string]string, err error)         // AccountInfoQuery allows the user system to query information of an existed account.
	AccountInfosQuery(username string, info []string) (infoMap map[string][]string, err error)      // AccountInfosQuery allows the user system to query information of an existed account.
	AccountInfoUpdate(username string, info map[string]string) (err error)                          // AccountInfoUpdate allows the user system to update information of an existed account.
	AccountPasswordUpdate(username string, oldPassword, newPassword []byte) (err error)             // AccountPasswordUpdate allows the user system to update password of an existed account.
	AccountPasswordReset(username string, newPassword []byte) (err error)                           // AccountPasswordReset allows the user system to reset password of an existed account.
	AccountRecoverEmailSend(username string) (err error)                                            // AccountRecoverEmailSend allows the user system to send a verify code to user email.
	AccountRecoverByEmail(username string, newPassword []byte, code string) (err error)             // AccountRecoverByEmail allows the user system to recover account by email.

	UserGroupAdd(sysType, groupName string) (groupID string, err error)                              // UserGroupAdd allows the user system to add a new user group.
	UserGroupDelete(sysType, groupID string) (err error)                                             // UserGroupDelete allows the user system to delete a new user group.
	UserGroupRename(sysType, groupID, groupName string) (err error)                                  // UserGroupRename allows the user system to rename the user group.
	UserGroupList(sysType string, filter map[string]string) (groupList map[string]string, err error) // UserGroupList allows the user system to query the existed user group list.
	UserGroupMemberList(sysType, groupID string) (memberList []string, err error)                    // UserGroupMemberList allows the user system to query the members of the user group.
	UserJoinUserGroup(sysType string, username []string, userGroupID string) (err error)             // UserJoinUserGroup allows the user system to bind a account to a user group.
	UserLeaveUserGroup(sysType string, username []string, userGroupID string) (err error)            // UserLeaveUserGroup allows the user system to unbind a account from a user group.

	Login(sysType, username string, password []byte, entity *EntityBaseInfo) (needMFAVerify bool, qrImg, secret, accessToken, refreshToken string) // Login allows the user system to  login.
	PasswordVerify(username string, password []byte) (err error)                                                                                   // PasswordVerify allows the user system to check password of user.
	MFAVerify(sysType, username string, mfaCode int, mfaType string) (err error)                                                                   // MFAVerify allows the user system to perform Multi-Factor Authentication.
	MFALoginVerify(sysType, username string, mfaCode int, mfaType string) (accessToken, refreshToken string, err error)                            // MFALoginVerify allows the user system to perform Multi-Factor Authentication while login.
	AuthenticatorBindConfirm(sysType, username string, mfaCode int, secret string) (err error)                                                     // AuthenticatorBindConfirm allows the user system to perform Multi-Factor Authentication.
	AuthenticatorBind(sysType, username string, mfaCode int, secret string) (err error)                                                            // AuthenticatorBind allows the user to bind when Authenticator is needed.
	AuthenticatorBindStatus(sysType, username string) (bind bool, err error)                                                                       // AuthenticatorBindStatus allows the user to bind when Authenticator is needed.
	AuthenticatorUnbind(sysType, username string) (err error)                                                                                      // AuthenticatorUnbind allows the user to unbind when Authenticator isn't needed or rebind.
	EntityList(sysType, username string) (info map[string]*EntityBaseInfo, err error)                                                              // EntityList allows the admin to get all bound entities of target user.
	EntityNameUpdate(sysType, username, entityID, entityName string) (err error)                                                                   // EntityNameUpdate allows the admin to rename entity.
	EntityAdd(sysType, username string, info *EntityBaseInfo) (err error)                                                                          // EntityAdd allows the admin to add bound entities of target user.
	EntityDelete(sysType, username, entityID string) (err error)                                                                                   // EntityDelete allows the admin to delete bound entities of target user.
	Logout(sysType, username string) (err error)                                                                                                   // Logout allows the user logout and redirect to login page.

	TokenRefresh(sysType, username, deviceCode, refreshToken, accessToken string) (newAccessToken string, err error)                                  // TokenRefresh allows the user system to refresh access token.
	TokenVerify(sysType, username, deviceCode, accessToken string) (err error)                                                                        // TokenVerify allows the user system to verify valid of access token.
	RoleAdd(sysType, roleName, remark string) (err error)                                                                                             // RoleAdd allows the user system to add a new role.
	RoleDelete(sysType, roleName string) (err error)                                                                                                  // RoleDelete allows the user system to delete a role.
	RoleList(sysType string, filter map[string]string) (roleList map[string]string, err error)                                                        // RoleList allows the user system to query the existed roleList.
	RoleBindUser(sysType, username, roleName string) (err error)                                                                                      // RoleBindUser allows the user system to bind a user to a role.
	RoleUnbindUser(sysType, username, roleName string) (err error)                                                                                    // RoleUnbindUser allows the user system to unbind a user to a role.
	RoleBindUserGroup(sysType, userGroupID, roleName string) (err error)                                                                              // RoleBindUserGroup allows the user system to bind a user group to a role.
	RoleUnBindUserGroup(sysType, userGroupID, roleName string) (err error)                                                                            // RoleUnBindUserGroup allows the user system to unbind a user group to a role.
	RoleMemberList(sysType, roleName string) (userList, userGroupList []string, err error)                                                            // RoleMemberList allows the user system to query the bound users and user groups of the role.
	AccessControlPolicyAdd(sysType, name, scope, operation, time string) (err error)                                                                  // AccessControlPolicyAdd allows the user system to add a new access control policy.
	AccessControlPolicyUpdate(sysType, name, scope, operation, time string) (err error)                                                               // AccessControlPolicyUpdate allows the user system to update access control policy.
	AccessControlPolicyDelete(sysType, name string) (err error)                                                                                       // AccessControlPolicyDelete allows the user system to delete an access control policy.
	AccessControlPolicyList(sysType string) (policyList []string, err error)                                                                          // AccessControlPolicyList allows the user system to query the existed access control policy list.
	AccessControlPolicyQuery(sysType, name string) (scope, operation, time string, err error)                                                         // AccessControlPolicyQuery allows the user system to query the permission of an existed access control policy.
	AccessControlGrantToUser(sysType, name, username string) (err error)                                                                              // AccessControlGrantToUser allows the user system to grant access control to user.
	AccessControlRevokeFromUser(sysType, name, username string) (err error)                                                                           // AccessControlRevokeFromUser allows the user system to cancel granted access control of user.
	AccessControlGrantToUserGroup(sysType, name, userGroupID string) (err error)                                                                      // AccessControlGrantToUserGroup allows the user system to grant access control to user group.
	AccessControlRevokeFromUserGroup(sysType, name, userGroupID string) (err error)                                                                   // AccessControlRevokeFromUserGroup allows the user system to cancel granted access control of user group.
	AccessControlGrantToRole(sysType, name, roleName string) (err error)                                                                              // AccessControlGrantToRole allows the user system to grant access control to role.
	AccessControlRevokeFromRole(sysType, name, roleName string) (err error)                                                                           // AccessControlRevokeFromRole allows the user system to cancel granted access control of role.
	UserPermissionQuery(sysType, username string) (userPermission []string, inheritedFromUserGroup, inheritedFromRole map[string][]string, err error) // UserPermissionQuery allows the user system to query identity and access control of a user.
	UserGroupPermissionQuery(sysType, userGroupID string) (userGroupPermission []string, inheritedFromRole map[string][]string, err error)            // UserGroupPermissionQuery allows the user system to query identity and access control of a user group.
	RolePermissionQuery(sysType, roleName string) (rolePermission []string, err error)                                                                // RolePermissionQuery allows the user system to query identity and access control of a role.
}

type EntityBaseInfo struct {
	ID   string
	Name string
	Type string
	OS   string
}
