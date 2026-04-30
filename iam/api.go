// Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

package iam

import (
	"net/http"
	"time"
)

// Instance is the IAM interface obtained via ASNController.GetIAM().
// All methods are goroutine-safe.
//
// Functional areas:
//  1. System configuration
//  2. OTP code sending
//  3. Account management
//  4. Authentication
//  5. Token lifecycle
//  6. MFA
//  7. Passkey (WebAuthn)
//  8. Device management
//  9. Groups
// 10. Accesses
type Instance interface {

	// -------------------------------------------------------------------------
	// System Configuration
	// -------------------------------------------------------------------------

	// JWKSGet returns the JSON Web Key Set document for external token verification.
	JWKSGet() (string, error)

	// RenameEnabled reports whether the IAM deployment allows username changes.
	RenameEnabled() bool

	// MfaEnforced reports whether MFA is currently enforced service-wide.
	MfaEnforced() bool

	// ServiceMfaSet enforces or relaxes MFA for all accounts in this service.
	ServiceMfaSet(mfaRequired bool) error

	// SupportedCountryCodesGet returns the phone country code filter mode and list.
	// Mode "all": no restriction. Mode "include": only listed codes accepted.
	// Mode "exclude": listed codes rejected.
	// Validate phone numbers against this before calling login or OTP-send methods.
	SupportedCountryCodesGet() (PhoneCountryCodeMode, []string)

	// -------------------------------------------------------------------------
	// OTP Code Sending
	// Call before OTP-based login or MFA verification to send a code to the user.
	// -------------------------------------------------------------------------

	// AccountPhoneSend sends an OTP code to the given phone number.
	// Returns a session token to pass to the corresponding login or MFA verify call.
	// sendWhenExistOnly: if true and no account with this number exists, silently succeeds without sending.
	AccountPhoneSend(countryCode, number string, sendWhenExistOnly bool) (string, error)

	// AccountEmailSend sends an OTP code to the given email address.
	// Returns a session token to pass to the corresponding login or MFA verify call.
	// sendWhenExistOnly: if true and no account with this email exists, silently succeeds without sending.
	AccountEmailSend(email string, sendWhenExistOnly bool) (string, error)

	// -------------------------------------------------------------------------
	// Account Management
	// -------------------------------------------------------------------------

	// AccountCreate creates a new account.
	// Credential fields (password, email+code, phone+code) are optional; pass zero values to omit.
	// skipEmailValidation / skipPhoneValidation bypass OTP pre-verification for admin-initiated creation.
	AccountCreate(
		username, metadata string,
		password string,
		email, emailCode string, skipEmailValidation bool,
		phone *Phone, phoneCode string, skipPhoneValidation bool,
	) (accountID string, err error)

	// AccountDelete permanently deletes the account and all its associated data.
	AccountDelete(accountID string) error

	// AccountExists reports whether an account with the given ID exists.
	AccountExists(accountID string) (bool, error)

	// AccountGet retrieves an account by any single non-empty lookup field;
	// the remaining fields are ignored.
	AccountGet(accountID, username, countryCode, number, email string) (*Account, error)

	// AccountList returns accounts matching all provided filter fields.
	// Empty strings match all values for that field.
	AccountList(username, countryCode, number, email string) ([]*Account, error)

	// AccountListByIDs returns accounts for the given IDs.
	AccountListByIDs(accountIDs []string) ([]*Account, error)

	// AccountRename changes the account's username.
	AccountRename(accountID, newUsername string) error

	// AccountPhoneUpdate updates the account's bound phone number, verifying the new number via OTP.
	// skipPhoneValidation bypasses OTP verification for admin-initiated updates.
	AccountPhoneUpdate(accountID string, phone *Phone, phoneCode string, skipPhoneValidation bool) error

	// AccountEmailUpdate updates the account's bound email address, verifying the new address via OTP.
	// skipEmailValidation bypasses OTP verification for admin-initiated updates.
	AccountEmailUpdate(accountID, email, emailCode string, skipEmailValidation bool) error

	// AccountWeChatUpdate binds or re-binds the account's WeChat identity using a WeChat OAuth code.
	AccountWeChatUpdate(accountID, weChatAppID, weChatCode string) error

	// AccountAppleUpdate binds or re-binds the account's Apple identity using an Apple ID token.
	AccountAppleUpdate(accountID, appleIDToken string) error

	// AccountGoogleUpdate binds or re-binds the account's Google identity using a Google ID token.
	AccountGoogleUpdate(accountID, googleIDToken string) error

	// AccountMetadataUpdate updates the service-defined opaque metadata string on the account.
	AccountMetadataUpdate(accountID, metadata string) error

	// AccountPasswordUpdate changes the account's password, requiring the current password for verification.
	// For admin-initiated resets without the old password, use AccountPasswordReset.
	AccountPasswordUpdate(accountID, oldPassword, newPassword string) error

	// AccountPasswordReset sets a new password without requiring the current password.
	// For user-initiated changes, use AccountPasswordUpdate.
	AccountPasswordReset(accountID, newPassword string) error

	// -------------------------------------------------------------------------
	// Authentication
	// All LoginOrCreate* and AccountPasskeyAuth methods share this return pattern:
	//   (account, needMfa, tokenSet, err)
	// When needMfa == true, tokenSet is a pre-MFA token; the session is not fully
	// authorized until MFALoginVerify() upgrades it.
	// device identifies the calling device and is required.
	// userClaims is an opaque service-defined string embedded in the issued token,
	// retrievable via TokenVerify().
	// createIfNotExist: if true, an account is created on first successful credential validation.
	// -------------------------------------------------------------------------

	// LoginMethods reports which authentication methods are enabled at runtime.
	LoginMethods() (
		usernameAndPassword, emailAndPassword, phoneAndPassword, emailCode, phoneCode, weChat, apple, google, passkey bool,
		err error,
	)

	// PasswordVerify validates credentials without issuing tokens. Useful for re-authentication flows.
	PasswordVerify(username, countryCode, number, email, password string) error

	// LoginOrCreateWithPassword authenticates using a password credential.
	// Pass username, countryCode+number, or email as the identifier; unused fields should be empty.
	LoginOrCreateWithPassword(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		username, countryCode, number, email, password string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)

	// LoginOrCreateWithPhone authenticates using a phone OTP.
	// Send the OTP first via AccountPhoneSend().
	LoginOrCreateWithPhone(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		phone *Phone, code string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)

	// LoginOrCreateWithEmail authenticates using an email OTP.
	// Send the OTP first via AccountEmailSend().
	LoginOrCreateWithEmail(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		email, code string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)

	// LoginOrCreateWithWeChat authenticates using a WeChat OAuth code.
	LoginOrCreateWithWeChat(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		appID, code string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)

	// LoginOrCreateWithApple authenticates using an Apple ID token.
	LoginOrCreateWithApple(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		idToken string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)

	// LoginOrCreateWithGoogle authenticates using a Google ID token.
	LoginOrCreateWithGoogle(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		idToken string,
		createIfNotExist bool,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)

	// AccountPasskeyLoginChallengeGet initiates a WebAuthn authentication ceremony.
	// Pass sessionID and data to AccountPasskeyAuth() after the client completes the assertion.
	AccountPasskeyLoginChallengeGet(domain string) (sessionID, data string, err error)

	// AccountPasskeyAuth completes a WebAuthn authentication ceremony.
	// Call AccountPasskeyLoginChallengeGet() first to obtain sessionID and data.
	AccountPasskeyAuth(
		device *DeviceInfo,
		userClaims string, durationAccess, durationRefresh time.Duration,
		domain, sessionID, data string,
	) (account *Account, needMfa bool, tokenSet *TokenSet, err error)

	// Logout invalidates the session for the given device and revokes its tokens.
	Logout(accountID, deviceID string) error

	// AppleRedirect handles the Apple OAuth redirect callback.
	// Mount this on the service's web router at the Apple-configured redirect URI.
	AppleRedirect(w http.ResponseWriter, r *http.Request)

	// -------------------------------------------------------------------------
	// Token Lifecycle
	// -------------------------------------------------------------------------

	// TokenRefresh issues a new access token from a valid refresh token.
	// userClaims is embedded in the new token.
	TokenRefresh(userClaims string, tokenSet *TokenSet, durationAccess time.Duration) (*TokenSet, error)

	// TokenVerify validates an access token and returns its claims.
	// When mfaNeeded == true, the token is a pre-MFA token; only MFA endpoints should accept it
	// until MFALoginVerify() upgrades the session.
	TokenVerify(accessToken string) (mfaNeeded bool, accountID, username, deviceID, userClaims string, err error)

	// TokenRevoke immediately invalidates the given access token.
	TokenRevoke(accessToken string) error

	// -------------------------------------------------------------------------
	// MFA
	// -------------------------------------------------------------------------

	// AccountEnableMFA enables the MFA requirement for the given account.
	AccountEnableMFA(accountID string) error

	// AccountDisableMFA disables the MFA requirement for the given account.
	AccountDisableMFA(accountID string) error

	// MFALoginVerify upgrades a pre-MFA access token to a fully-authorized session token.
	// method selects the verification type. Populate the corresponding parameters
	// (code, domain, sessionID, data) as required by the method; pass empty strings for unused fields.
	MFALoginVerify(accessToken string, method MfaType, code, domain, sessionID, data string) (*TokenSet, error)

	// TotpBind initiates TOTP enrollment for the given account.
	// Returns a QR code image (data URI), issuer name, and raw TOTP secret for display to the user.
	// Call TotpBindConfirm() with the user-provided code to complete enrollment.
	TotpBind(accountID string) (img, issuer, secret string, err error)

	// TotpBindConfirm completes TOTP enrollment by verifying the user-provided code.
	TotpBindConfirm(accountID, code string) error

	// TotpUnbind removes the TOTP authenticator from the given account.
	TotpUnbind(accountID string) error

	// -------------------------------------------------------------------------
	// Passkey (WebAuthn)
	// -------------------------------------------------------------------------

	// AccountPasskeyBindChallengeGet initiates a WebAuthn registration ceremony.
	// Pass sessionID and data to AccountPasskeyBind() after the client completes attestation.
	AccountPasskeyBindChallengeGet(domain, accountID string) (sessionID, data string, err error)

	// AccountPasskeyBind completes a WebAuthn registration ceremony and binds the passkey to the account.
	// Call AccountPasskeyBindChallengeGet() first to obtain sessionID and data.
	AccountPasskeyBind(domain, accountID, sessionID, deviceID, data string) error

	// AccountPasskeyUnbind removes a single passkey from the account, identified by passkeyID.
	AccountPasskeyUnbind(accountID, id string) error

	// AccountPasskeyUnbindAll removes all passkeys from the account.
	AccountPasskeyUnbindAll(accountID string) error

	// -------------------------------------------------------------------------
	// Device Management
	// The framework tracks one Device record per (account, device) pair.
	// -------------------------------------------------------------------------

	// DeviceLimitUpdate sets the maximum number of concurrent active devices for the account.
	DeviceLimitUpdate(accountID string, limit int) error

	// DeviceInfoUpdate updates the stored device record (push token, model, metadata, etc.).
	DeviceInfoUpdate(accountID string, device *Device) error

	// DeviceDelete removes the device record and revokes all its associated tokens.
	DeviceDelete(accountID, deviceID string) error

	// -------------------------------------------------------------------------
	// Groups
	// Groups are service-scoped collections of accounts used to assign accesses in bulk.
	// -------------------------------------------------------------------------

	// GroupCreate creates a new group within this service's namespace.
	GroupCreate(groupName, metadata string) error

	// GroupDelete deletes the group.
	GroupDelete(groupName string) error

	// GroupExists reports whether a group with the given name exists.
	GroupExists(groupName string) (bool, error)

	// GroupRename renames the group.
	GroupRename(oldName, newName string) error

	// GroupMetadataUpdate updates the group's opaque metadata string.
	GroupMetadataUpdate(groupName, metadata string) error

	// GroupGet returns the group's details.
	GroupGet(groupName string) (*Group, error)

	// GroupList returns all groups in this service's namespace.
	GroupList() ([]*Group, error)

	// GroupMemberList returns all accounts that are members of the group.
	GroupMemberList(groupName string) ([]*Account, error)

	// AccountJoinGroup adds the given accounts to the group.
	AccountJoinGroup(groupName string, accountIDs []string) error

	// AccountLeaveGroup removes the given accounts from the group.
	AccountLeaveGroup(groupName string, accountIDs []string) error

	// AccountGroupList returns all groups (within this service's namespace) the account belongs to.
	AccountGroupList(accountID string) ([]*Group, error)

	// -------------------------------------------------------------------------
	// Accesses
	// An Access defines a named, time-controlled permission rule.
	// Accesses are granted at the group level, not per account.
	// -------------------------------------------------------------------------

	// AccessCreate defines a new access rule in this service's namespace.
	AccessCreate(name, scope, operation string, time *TimeControl) error

	// AccessUpdate updates an existing access rule.
	AccessUpdate(name, scope, operation string, time *TimeControl) error

	// AccessDelete removes the access rule.
	AccessDelete(name string) error

	// AccessExists reports whether an access rule with the given name exists.
	AccessExists(name string) (bool, error)

	// AccessList returns all access rules in this service's namespace.
	AccessList() ([]*Access, error)

	// AccountAccessList returns all effective access rules for the account,
	// keyed by the namespace (service name) each access belongs to.
	AccountAccessList(accountID string) (map[string][]*Access, error)

	// GroupAccessList returns all access rules currently granted to the group.
	GroupAccessList(groupName string) ([]*Access, error)

	// AccessGrantToGroup grants the named access rules to the group.
	AccessGrantToGroup(groupName string, accesses []string) error

	// AccessRevokeFromGroup revokes the named access rules from the group.
	AccessRevokeFromGroup(groupName string, accesses []string) error
}
