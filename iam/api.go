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
//
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

	// AccountPhoneCodeSend sends an OTP code to the given phone number.
	// Returns a session token to pass to the corresponding login or MFA verify call.
	// sendWhenExistOnly: if true and no account with this number exists, silently succeeds without sending.
	AccountPhoneCodeSend(countryCode, number string, sendWhenExistOnly bool) (result CodeSendResult, code string, nextAllowed time.Duration, err error)

	// AccountEmailCodeSend sends an OTP code to the given email address.
	// Returns a session token to pass to the corresponding login or MFA verify call.
	// sendWhenExistOnly: if true and no account with this email exists, silently succeeds without sending.
	AccountEmailCodeSend(email string, sendWhenExistOnly bool) (result CodeSendResult, code string, nextAllowed time.Duration, err error)

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

	// AccountPasswordSet sets the account's password directly, without proof of the old password.
	// allowOverwrite: if false, the call fails when a password is already set; if true, any
	// existing password is replaced. For user-initiated changes that require verification
	// (old password or OTP), use the password flow (AccountPasswordFlow*) instead.
	AccountPasswordSet(accountID, newPassword string, allowOverwrite bool) error

	// -------------------------------------------------------------------------
	// Password Flow
	// Stateful flow covering both forgot-password recovery and in-session change.
	// Steps:
	//  1. AccountPasswordFlowInit — identify the account, obtain a flow token and methods.
	//  2. (OTP methods only) AccountPasswordFlowSendCode to deliver the code; skip for the
	//     old-password method.
	//  3. AccountPasswordFlowVerify — submit the chosen proof to verify the flow token.
	//  4. AccountPasswordFlowComplete — set the new password for the verified flow token.
	// -------------------------------------------------------------------------

	// AccountPasswordFlowInit is step 1: identify the account by exactly one of
	// accountID/username/email/phone and return the available verification methods.
	// A non-existent account or one with no usable method yields PasswordFlowInitNoRecovery
	// with an empty flow token.
	AccountPasswordFlowInit(accountID, username, email, countryCode, number string) (state PasswordFlowInitState, flowToken string, methods []PasswordFlowMethod, err error)

	// AccountPasswordFlowSendCode is step 2 for OTP methods: send (or resend, rate-limited)
	// a code to the account's bound email/phone. The delivery target is always resolved from
	// the flow's bound account, never caller-supplied, so a stolen flow token cannot redirect
	// the code. method must be PasswordVerifyEmail or PasswordVerifySMS. Returns the code
	// in-band only in dev mode. The masked target is reported by AccountPasswordFlowInit.
	AccountPasswordFlowSendCode(flowToken string, method PasswordVerifyMethod) (result CodeSendResult, nextAllowed time.Duration, code string, err error)

	// AccountPasswordFlowVerify is step 3: submit the chosen proof (old password or OTP).
	// The method is resolved from the flow, not re-sent: a non-empty oldPassword means the
	// old-password method; otherwise the OTP method chosen by the last AccountPasswordFlowSendCode.
	// A nil error means the flow token is now verified and may complete.
	AccountPasswordFlowVerify(flowToken, code, oldPassword string) error

	// AccountPasswordFlowComplete is step 4: set the new password for a verified flow token.
	AccountPasswordFlowComplete(flowToken, newPassword string) error

	// -------------------------------------------------------------------------
	// Authentication
	// Login is a single stateful flow regardless of credential kind:
	//   AuthFlowInit -> (optional) AuthFlowGetChallenge -> AuthFlowVerify
	// then, if AuthFlowVerify returns LoginFlowMFAVerify / LoginFlowMFASetup,
	// continue via the Auth Flow MFA APIs (AuthFlowMfaSetupInitiate / Confirm / Verify).
	// device identifies the calling device and is required at AuthFlowInit.
	// userClaims is an opaque service-defined string embedded in the issued token,
	// retrievable via TokenVerify().
	// createIfNotExist: if true, an account is created on first successful credential validation.
	// -------------------------------------------------------------------------

	// LoginMethods reports which authentication methods are enabled at runtime.
	LoginMethods() ([]CredentialMethod, error)

	// AuthFlowInit is step 1 of the stateful login flow: pick a credential method and
	// provide the matching identity, creating the flow. The token material needed to mint
	// the final token pair is collected up front here.
	// Fill the identity matching the method's identity kind: username for *_USERNAME_*,
	// phone for PHONE_*, email for EMAIL_*. Third-party methods (WeChat/Apple/Google) and
	// Passkey leave all identity fields empty — selecting the method starts the flow.
	// The returned state is either LoginFlowChallengeRequired (call AuthFlowGetChallenge
	// next, e.g. to send an OTP or fetch a passkey challenge) or LoginFlowVerifyRequired
	// (call AuthFlowVerify directly with the proof).
	AuthFlowInit(
		device *DeviceInfo, userClaims string, durationAccess, durationRefresh time.Duration,
		createIfNotExist bool,
		method CredentialMethod,
		username string, phone *Phone, email string,
	) (state LoginFlowState, flowToken string, err error)

	// AuthFlowGetChallenge is step 2 of the login flow: prepare the verify step. It
	// dispatches on the flow's method:
	//   - phone/email code: sends (or resends, rate-limited) an OTP; returns the code
	//     in-band only in dev mode. The masked target is reported by AuthFlowVerify's methods.
	//   - passkey: returns a WebAuthn challenge (data) to sign and submit via AuthFlowVerify;
	//     domain is the relying-party domain. The session is recorded on the flow, so it is
	//     not returned — AuthFlowVerify resolves it from the flow token.
	// mfaMethod is only used in the MFA phase to pick the factor; in the credential phase
	// it is ignored.
	AuthFlowGetChallenge(
		flowToken string, mfaMethod MfaType, domain string,
	) (result CodeSendResult, nextAllowed time.Duration, code, data string, err error)

	// AuthFlowVerify is step 3 of the login flow: submit the proof matching the flow's
	// method. Fill the field for the method's proof kind: password for *_PASSWORD, code for
	// *_CODE, weChat for WeChat, appleIDToken/googleIDToken for Apple/Google, and
	// data for the passkey assertion. For passkey, the session id and RP domain were captured
	// by AuthFlowGetChallenge and are resolved from the flow, so they are not re-sent.
	// On success the returned state is LoginFlowAuthenticated (tokenSet is valid), or
	// LoginFlowMFAVerify / LoginFlowMFASetup to enter the MFA flow.
	AuthFlowVerify(
		flowToken string,
		password, code string,
		weChatAppID, weChatCode string,
		appleIDToken, googleIDToken string,
		data string,
	) (account *Account, state LoginFlowState, tokenSet *TokenSet, flowToken_ string, availableMfaMethods []MfaMethodInfo, availableSetupMethods []MfaType, err error)

	// AuthFlowResume re-enters the MFA flow when TokenVerify reports that MFA verification
	// is still needed. Pass the MFA-unverified access token; it returns a result with
	// LoginFlowMFAVerify / LoginFlowMFASetup, a flow token, and the available methods,
	// exactly as a fresh AuthFlowVerify would.
	AuthFlowResume(
		accessToken string,
	) (account *Account, state LoginFlowState, tokenSet *TokenSet, flowToken string, availableMfaMethods []MfaMethodInfo, availableSetupMethods []MfaType, err error)

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
	// until the MFA flow (re-entered via AuthFlowResume) upgrades the session.
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

	// TotpBind initiates TOTP enrollment for the given account.
	// Returns a QR code image (data URI), issuer name, and raw TOTP secret for display to the user.
	// Call TotpBindConfirm() with the user-provided code to complete enrollment.
	TotpBind(accountID string) (img, issuer, secret string, err error)

	// TotpBindConfirm completes TOTP enrollment by verifying the user-provided code.
	TotpBindConfirm(accountID, code string) error

	// TotpUnbind removes the TOTP authenticator from the given account.
	TotpUnbind(accountID string) error

	// -------------------------------------------------------------------------
	// Auth Flow MFA APIs (inline MFA setup/verify continuing a login flow)
	// -------------------------------------------------------------------------

	// AuthFlowMfaSetupInitiate begins inline MFA enrollment during a login flow that returned
	// LoginFlowMFASetup. flowToken is the temporary, MFA-unverified token from the login result.
	// It dispatches on the chosen factor:
	//   - TOTP:      returns totpImage (QR code data URI), totpIssuer, and totpSecret for display.
	//   - SMS/Email: supply the new delivery target to bind via phone/email; an OTP is sent there
	//                and the delivery result (result/nextAllowed, plus the code in dev mode) is
	//                returned. Pass the same target to AuthFlowMfaSetupConfirm.
	// Complete enrollment by calling AuthFlowMfaSetupConfirm with the resulting material.
	AuthFlowMfaSetupInitiate(flowToken string, method MfaType, phone *Phone, email string) (totpImage, totpIssuer, totpSecret string, result CodeSendResult, nextAllowed time.Duration, code string, err error)

	// AuthFlowMfaSetupConfirm finishes inline MFA enrollment for the LoginFlowMFASetup path,
	// binding the chosen method to the account and promoting the flow token to a full session.
	// Submit the proof only: code for TOTP/SMS/Email.
	// The method and (for SMS/Email) the target being bound were captured by
	// AuthFlowMfaSetupInitiate and are resolved from the flow — re-initiate to switch
	// method/target before confirming.
	// On success returns the authenticated account and tokenSet; the flow token is then invalidated.
	// Each failed attempt increments the flow token's attempt counter; exhausting it forces re-login.
	AuthFlowMfaSetupConfirm(inputFlowToken, code string) (
		account *Account, state LoginFlowState, tokenSet *TokenSet, flowToken string, availableMfaMethods []MfaMethodInfo, availableSetupMethods []MfaType, err error)

	// AuthFlowMfaVerify verifies an already-bound MFA factor during a login flow that returned
	// LoginFlowMFAVerify, promoting the flow token to a full session on success.
	// Submit the proof only: code for TOTP/SMS/Email.
	// The method was captured by AuthFlowGetChallenge and is resolved
	// from the flow — call AuthFlowGetChallenge first
	// (also for TOTP) and re-call it to switch method before verifying.
	// On success returns the authenticated account and tokenSet; the flow token is then invalidated.
	// Each failed attempt increments the flow token's attempt counter; exhausting it forces re-login.
	AuthFlowMfaVerify(inputFlowToken, code string) (
		account *Account, state LoginFlowState, tokenSet *TokenSet, flowToken string, availableMfaMethods []MfaMethodInfo, availableSetupMethods []MfaType, err error)

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

	// AccountAccessList returns all effective access rules for the account.
	// The map key is the service namespace name; in practice each service only receives
	// accesses under its own namespace. The map structure aligns with the IAM interface for UI reuse.
	AccountAccessList(accountID string) (map[string][]*Access, error)

	// GroupAccessList returns all access rules currently granted to the group.
	GroupAccessList(groupName string) ([]*Access, error)

	// AccessGrantToGroup grants the named access rules to the group.
	AccessGrantToGroup(groupName string, accesses []string) error

	// AccessRevokeFromGroup revokes the named access rules from the group.
	AccessRevokeFromGroup(groupName string, accesses []string) error
}
