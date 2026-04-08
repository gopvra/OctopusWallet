package handlers

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/auth"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

// webauthnUser adapts AdminUser to the webauthn.User interface.
type webauthnUser struct {
	user        *models.AdminUser
	credentials []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte                         { return []byte(u.user.ID) }
func (u *webauthnUser) WebAuthnName() string                       { return u.user.Username }
func (u *webauthnUser) WebAuthnDisplayName() string                { return u.user.Username }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

type AdminWebAuthnHandler struct {
	store     store.AdminStore
	wa        *webauthn.WebAuthn
	jwtSecret string

	// Session store for in-flight challenges (short-lived, 5 min TTL)
	sessions sync.Map // key: userID -> *webauthn.SessionData
}

func NewAdminWebAuthnHandler(s store.AdminStore, wa *webauthn.WebAuthn, jwtSecret string) *AdminWebAuthnHandler {
	return &AdminWebAuthnHandler{store: s, wa: wa, jwtSecret: jwtSecret}
}

// --- Registration Flow ---

// BeginRegistration starts passkey registration for the authenticated user.
func (h *AdminWebAuthnHandler) BeginRegistration(c *gin.Context) {
	userID := c.GetString("admin_user_id")
	user, err := h.store.GetAdminUserByID(c.Request.Context(), userID)
	if err != nil {
		R.Fail(c, errcode.ErrAdminUserNotFound)
		return
	}

	// Load existing credentials
	creds, _ := h.store.GetWebAuthnCredentialsByUserID(c.Request.Context(), userID)
	waCreds := dbCredsToWebAuthn(creds)

	waUser := &webauthnUser{user: user, credentials: waCreds}

	// Build exclusion list from existing credentials
	exclusions := make([]protocol.CredentialDescriptor, len(waCreds))
	for i, cred := range waCreds {
		exclusions[i] = protocol.CredentialDescriptor{
			Type:            protocol.PublicKeyCredentialType,
			CredentialID:    cred.ID,
			Transport:       cred.Transport,
		}
	}

	options, session, err := h.wa.BeginRegistration(waUser,
		webauthn.WithExclusions(exclusions),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
	)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	// Store session with TTL
	h.sessions.Store(userID+":register", session)
	go func() {
		time.Sleep(5 * time.Minute)
		h.sessions.Delete(userID + ":register")
	}()

	R.OK(c, options)
}

// FinishRegistration completes passkey registration.
func (h *AdminWebAuthnHandler) FinishRegistration(c *gin.Context) {
	userID := c.GetString("admin_user_id")
	user, err := h.store.GetAdminUserByID(c.Request.Context(), userID)
	if err != nil {
		R.Fail(c, errcode.ErrAdminUserNotFound)
		return
	}

	sessionData, ok := h.sessions.LoadAndDelete(userID + ":register")
	if !ok {
		R.FailMsg(c, errcode.ErrBadRequest, "no registration session found, please start again")
		return
	}

	creds, _ := h.store.GetWebAuthnCredentialsByUserID(c.Request.Context(), userID)
	waCreds := dbCredsToWebAuthn(creds)
	waUser := &webauthnUser{user: user, credentials: waCreds}

	credential, err := h.wa.FinishRegistration(waUser, *sessionData.(*webauthn.SessionData), c.Request)
	if err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, "registration verification failed: "+err.Error())
		return
	}

	// Parse display name from request body
	var meta struct {
		DisplayName string `json:"display_name"`
	}
	body, _ := c.GetRawData()
	json.Unmarshal(body, &meta)
	displayName := meta.DisplayName
	if displayName == "" {
		displayName = "Passkey"
	}

	// Serialize transport
	transport := ""
	if len(credential.Transport) > 0 {
		t, _ := json.Marshal(credential.Transport)
		transport = string(t)
	}

	dbCred := &models.WebAuthnCredential{
		UserID:          userID,
		CredentialID:    credential.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       credential.Authenticator.SignCount,
		Transport:       transport,
		DisplayName:     displayName,
	}

	if err := h.store.CreateWebAuthnCredential(c.Request.Context(), dbCred); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, gin.H{"id": dbCred.ID, "display_name": displayName})
}

// --- Authentication Flow ---

// BeginLogin starts passkey authentication (no password needed).
func (h *AdminWebAuthnHandler) BeginLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		R.FailMsg(c, errcode.ErrBadRequest, err.Error())
		return
	}

	user, err := h.store.GetAdminUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		R.Fail(c, errcode.ErrAdminLoginFailed)
		return
	}

	creds, _ := h.store.GetWebAuthnCredentialsByUserID(c.Request.Context(), user.ID)
	if len(creds) == 0 {
		R.FailMsg(c, errcode.ErrBadRequest, "no passkeys registered for this user")
		return
	}

	waCreds := dbCredsToWebAuthn(creds)
	waUser := &webauthnUser{user: user, credentials: waCreds}

	options, session, err := h.wa.BeginLogin(waUser)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	h.sessions.Store(user.ID+":login", session)
	go func() {
		time.Sleep(5 * time.Minute)
		h.sessions.Delete(user.ID + ":login")
	}()

	R.OK(c, options)
}

// FinishLogin completes passkey authentication and returns JWT tokens.
func (h *AdminWebAuthnHandler) FinishLogin(c *gin.Context) {
	// We need the username to look up the session
	username := c.Query("username")
	if username == "" {
		R.FailMsg(c, errcode.ErrBadRequest, "username query parameter required")
		return
	}

	user, err := h.store.GetAdminUserByUsername(c.Request.Context(), username)
	if err != nil {
		R.Fail(c, errcode.ErrAdminLoginFailed)
		return
	}

	if !user.IsActive {
		R.Fail(c, errcode.ErrAdminUserDeactivated)
		return
	}

	sessionData, ok := h.sessions.LoadAndDelete(user.ID + ":login")
	if !ok {
		R.FailMsg(c, errcode.ErrBadRequest, "no login session found, please start again")
		return
	}

	creds, _ := h.store.GetWebAuthnCredentialsByUserID(c.Request.Context(), user.ID)
	waCreds := dbCredsToWebAuthn(creds)
	waUser := &webauthnUser{user: user, credentials: waCreds}

	credential, err := h.wa.FinishLogin(waUser, *sessionData.(*webauthn.SessionData), c.Request)
	if err != nil {
		R.Fail(c, errcode.ErrAdminLoginFailed)
		return
	}

	// Update sign count
	h.store.UpdateWebAuthnCredentialSignCount(c.Request.Context(), credential.ID, credential.Authenticator.SignCount)

	// Generate JWT token pair
	tokenPair, err := auth.GenerateTokenPair(h.jwtSecret, user.ID, user.Username, user.Role)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	R.OK(c, gin.H{
		"user":  user,
		"token": tokenPair,
	})
}

// ListCredentials returns the user's registered passkeys.
func (h *AdminWebAuthnHandler) ListCredentials(c *gin.Context) {
	userID := c.GetString("admin_user_id")
	creds, err := h.store.GetWebAuthnCredentialsByUserID(c.Request.Context(), userID)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, creds)
}

// DeleteCredential removes a registered passkey.
func (h *AdminWebAuthnHandler) DeleteCredential(c *gin.Context) {
	userID := c.GetString("admin_user_id")
	credID := c.Param("id")
	if err := h.store.DeleteWebAuthnCredential(c.Request.Context(), credID, userID); err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}
	R.OK(c, gin.H{"message": "credential deleted"})
}

// --- Helpers ---

func dbCredsToWebAuthn(creds []models.WebAuthnCredential) []webauthn.Credential {
	result := make([]webauthn.Credential, len(creds))
	for i, c := range creds {
		var transport []protocol.AuthenticatorTransport
		if c.Transport != "" {
			json.Unmarshal([]byte(c.Transport), &transport)
		}
		result[i] = webauthn.Credential{
			ID:              c.CredentialID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Transport:       transport,
			Authenticator: webauthn.Authenticator{
				AAGUID:    c.AAGUID,
				SignCount: c.SignCount,
			},
		}
	}
	return result
}
