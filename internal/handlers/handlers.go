package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"

	"passkey-auth/internal/auth"
	"passkey-auth/internal/config"
	"passkey-auth/internal/database"
)

type Handlers struct {
	db       *database.DB
	webAuthn *auth.WebAuthn
	config   *config.Config
	store    *sessions.CookieStore
}

func New(db *database.DB, webAuthn *auth.WebAuthn, config *config.Config) *Handlers {
	webAuthn.SetDB(db)

	store := sessions.NewCookieStore([]byte(config.Auth.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Domain:   config.Auth.CookieDomain, // Share cookies across subdomains if configured
	}

	return &Handlers{
		db:       db,
		webAuthn: webAuthn,
		config:   config,
		store:    store,
	}
}

func (h *Handlers) writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		// If we can't encode the error response, log it
		// Don't try to write another response as headers are already sent
		log.Printf("Failed to encode error response: %v", err)
	}
}

func (h *Handlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, try to send a simple error response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// BeginRegistration starts the passkey registration process
func (h *Handlers) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.DisplayName == "" {
		h.writeError(w, "Email and display name are required", http.StatusBadRequest)
		return
	}

	// Check if email is allowed
	if !h.config.IsEmailAllowed(req.Email) {
		h.writeError(w, "Email address not allowed", http.StatusForbidden)
		return
	}

	// Check if user already exists
	existingUser, err := h.db.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		h.writeError(w, "User already exists", http.StatusConflict)
		return
	}

	// Create a temporary WebAuthn user for registration without saving to DB yet
	isAdmin := h.config.IsAdmin(req.Email)
	tempUser := &database.User{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Approved:    isAdmin, // Auto-approve admins
	}

	webAuthnUser := &auth.WebAuthnUser{}
	webAuthnUser.SetUser(tempUser)

	options, sessionData, err := h.webAuthn.BeginRegistration(webAuthnUser)
	if err != nil {
		logrus.Errorf("Failed to begin registration: %v", err)
		h.writeError(w, "Failed to begin registration", http.StatusInternalServerError)
		return
	}

	// Store session data with user details for later creation
	session, _ := h.store.Get(r, "webauthn-session")
	session.Values["challenge"] = sessionData.Challenge
	session.Values["pending_email"] = req.Email
	session.Values["pending_display_name"] = req.DisplayName
	session.Values["pending_is_admin"] = isAdmin
	if err := session.Save(r, w); err != nil {
		h.writeError(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Debug: log the options structure
	logrus.Debugf("WebAuthn options: %+v", options)
	logrus.Debugf("Challenge type: %T", options.Response.Challenge)
	logrus.Debugf("Challenge value: %v", options.Response.Challenge)
	logrus.Debugf("User ID type: %T", options.Response.User.ID)
	logrus.Debugf("User ID value: %v", options.Response.User.ID)

	h.writeJSON(w, options)
}

// FinishRegistration completes the passkey registration process
func (h *Handlers) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "webauthn-session")

	// Get pending user data from session instead of user_id
	pendingEmail, ok := session.Values["pending_email"].(string)
	if !ok {
		h.writeError(w, "Invalid session - no pending registration", http.StatusBadRequest)
		return
	}

	pendingDisplayName, ok := session.Values["pending_display_name"].(string)
	if !ok {
		h.writeError(w, "Invalid session - missing display name", http.StatusBadRequest)
		return
	}

	pendingIsAdmin, ok := session.Values["pending_is_admin"].(bool)
	if !ok {
		h.writeError(w, "Invalid session - missing admin flag", http.StatusBadRequest)
		return
	}

	challenge, ok := session.Values["challenge"].(string)
	if !ok {
		h.writeError(w, "Invalid session - missing challenge", http.StatusBadRequest)
		return
	}

	// Read and log the request body for debugging
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	logrus.Debugf("Received credential response body: %s", string(body))

	// Create a new reader from the body for the WebAuthn library
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Create temporary user for WebAuthn verification
	tempUser := &database.User{
		Email:       pendingEmail,
		DisplayName: pendingDisplayName,
		Approved:    pendingIsAdmin,
	}

	webAuthnUser := &auth.WebAuthnUser{}
	webAuthnUser.SetUser(tempUser)

	sessionData := webauthn.SessionData{
		Challenge: challenge,
		UserID:    webAuthnUser.WebAuthnID(),
	}

	// Log the request details for debugging
	logrus.Debugf("Finishing registration for user: %s", pendingEmail)
	logrus.Debugf("Session challenge: %s", challenge)
	logrus.Debugf("Session user ID: %v", webAuthnUser.WebAuthnID())

	credential, err := h.webAuthn.FinishRegistration(webAuthnUser, sessionData, r)
	if err != nil {
		logrus.Errorf("Failed to finish registration: %v", err)
		h.writeError(w, "Failed to finish registration", http.StatusInternalServerError)
		return
	}

	// Only NOW create the user in the database after successful passkey registration
	user, err := h.db.CreateUserWithApproval(pendingEmail, pendingDisplayName, pendingIsAdmin)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "User already exists", http.StatusConflict)
			return
		}
		logrus.Errorf("Failed to create user: %v", err)
		h.writeError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	if pendingIsAdmin {
		logrus.Infof("Admin user auto-approved: %s", pendingEmail)
	}

	// Save credential to database
	if err := h.webAuthn.SaveCredential(user.ID, credential); err != nil {
		logrus.Errorf("Failed to save credential: %v", err)
		// If we can't save the credential, we should remove the user we just created
		if deleteErr := h.db.DeleteUser(user.ID); deleteErr != nil {
			logrus.Errorf("Failed to cleanup user after credential save failure: %v", deleteErr)
		}
		h.writeError(w, "Failed to save credential", http.StatusInternalServerError)
		return
	}

	// Set authenticated session after successful registration
	authSession, _ := h.store.Get(r, "auth-session")
	authSession.Values["authenticated"] = true
	authSession.Values["user_id"] = user.ID
	authSession.Values["user_email"] = user.Email
	if err := authSession.Save(r, w); err != nil {
		h.writeError(w, "Failed to save auth session", http.StatusInternalServerError)
		return
	}

	// Clear webauthn session
	session.Values["challenge"] = nil
	session.Values["pending_email"] = nil
	session.Values["pending_display_name"] = nil
	session.Values["pending_is_admin"] = nil
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		// Don't return error here as the main operation succeeded
	}

	h.writeJSON(w, map[string]string{"status": "success"})
}

// BeginLogin starts the passkey authentication process
func (h *Handlers) BeginLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	webAuthnUser, err := h.webAuthn.GetUserByEmail(req.Email)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user is approved (if required)
	if h.config.Auth.RequireApproval && !webAuthnUser.GetUser().Approved {
		h.writeError(w, "User not approved", http.StatusForbidden)
		return
	}

	options, sessionData, err := h.webAuthn.BeginLogin(webAuthnUser)
	if err != nil {
		logrus.Errorf("Failed to begin login: %v", err)
		h.writeError(w, "Failed to begin login", http.StatusInternalServerError)
		return
	}

	// Debug: log the login options structure
	logrus.Debugf("Login options: %+v", options)
	logrus.Debugf("Login challenge type: %T", options.Response.Challenge)
	logrus.Debugf("Login challenge value: %v", options.Response.Challenge)
	if len(options.Response.AllowedCredentials) > 0 {
		logrus.Debugf("AllowedCredentials count: %d", len(options.Response.AllowedCredentials))
		for i, cred := range options.Response.AllowedCredentials {
			logrus.Debugf("Credential %d ID type: %T", i, cred.CredentialID)
			logrus.Debugf("Credential %d ID value: %v", i, cred.CredentialID)
		}
	}

	// Store session data
	session, _ := h.store.Get(r, "webauthn-session")
	session.Values["challenge"] = sessionData.Challenge
	session.Values["user_id"] = webAuthnUser.GetUser().ID
	if err := session.Save(r, w); err != nil {
		h.writeError(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, options)
}

// FinishLogin completes the passkey authentication process
func (h *Handlers) FinishLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "webauthn-session")

	userID, ok := session.Values["user_id"].(int)
	if !ok {
		h.writeError(w, "Invalid session", http.StatusBadRequest)
		return
	}

	challenge, ok := session.Values["challenge"].(string)
	if !ok {
		h.writeError(w, "Invalid session", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUser(userID)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	webAuthnUser, err := h.webAuthn.GetUserByEmail(user.Email)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	sessionData := webauthn.SessionData{
		Challenge: challenge,
		UserID:    webAuthnUser.WebAuthnID(),
	}

	credential, err := h.webAuthn.FinishLogin(webAuthnUser, sessionData, r)
	if err != nil {
		logrus.Errorf("Failed to finish login: %v", err)
		h.writeError(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Update credential sign count
	if err := h.webAuthn.UpdateCredentialSignCount(credential.ID, credential.Authenticator.SignCount); err != nil {
		logrus.Errorf("Failed to update sign count: %v", err)
	}

	// Set authenticated session
	authSession, _ := h.store.Get(r, "auth-session")
	authSession.Values["authenticated"] = true
	authSession.Values["user_id"] = user.ID
	authSession.Values["user_email"] = user.Email
	if err := authSession.Save(r, w); err != nil {
		h.writeError(w, "Failed to save auth session", http.StatusInternalServerError)
		return
	}

	// Clear webauthn session
	session.Values["challenge"] = nil
	session.Values["user_id"] = nil
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session: %v", err)
		// Don't return error here as the main operation succeeded
	}

	h.writeJSON(w, map[string]interface{}{
		"status": "success",
		"user": map[string]interface{}{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
		},
	})
}

// Logout clears the authentication session
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "auth-session")
	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	session.Values["user_email"] = nil
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Printf("Failed to save session during logout: %v", err)
		// Don't return error here as logout should still succeed
	}

	h.writeJSON(w, map[string]string{"status": "success"})
}

// AuthCheck implements the nginx auth_request protocol
func (h *Handlers) AuthCheck(w http.ResponseWriter, r *http.Request) {
	// Debug logging
	logrus.Debugf("AuthCheck request from %s", r.RemoteAddr)
	logrus.Debugf("AuthCheck headers: %+v", r.Header)
	logrus.Debugf("AuthCheck cookies: %+v", r.Cookies())

	session, err := h.store.Get(r, "auth-session")
	if err != nil {
		logrus.Errorf("Failed to get auth session: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	authenticated, ok := session.Values["authenticated"].(bool)
	logrus.Debugf("Session authenticated: %v, ok: %v", authenticated, ok)
	logrus.Debugf("Session values: %+v", session.Values)

	if !ok || !authenticated {
		logrus.Debugf("User not authenticated, returning 401")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Optional: Add user info to response headers
	if userID, ok := session.Values["user_id"].(int); ok {
		w.Header().Set("X-Auth-User-ID", strconv.Itoa(userID))
	}
	if userEmail, ok := session.Values["user_email"].(string); ok {
		w.Header().Set("X-Auth-User", userEmail)
	}

	logrus.Debugf("User authenticated, returning 200")
	w.WriteHeader(http.StatusOK)
}

// GetAuthStatus returns the current authentication status
func (h *Handlers) GetAuthStatus(w http.ResponseWriter, r *http.Request) {
	session, err := h.store.Get(r, "auth-session")
	if err != nil {
		h.writeError(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	userEmail, ok := session.Values["user_email"].(string)
	if !ok || userEmail == "" {
		h.writeError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Get user details from database
	user, err := h.db.GetUserByEmail(userEmail)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if user is approved
	if !user.Approved {
		h.writeError(w, "User not approved", http.StatusForbidden)
		return
	}

	response := map[string]interface{}{
		"authenticated": true,
		"user": map[string]interface{}{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"approved":     user.Approved,
			"is_admin":     h.config.IsAdmin(user.Email),
		},
	}

	h.writeJSON(w, response)
}

// Admin endpoints

// ListUsers returns all users (admin endpoint)
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}

	users, err := h.db.ListUsers()
	if err != nil {
		logrus.Errorf("Failed to list users: %v", err)
		h.writeError(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, users)
}

// CreateUser creates a new user (admin endpoint)
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}

	var req struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Approved    bool   `json:"approved"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if email is allowed
	if !h.config.IsEmailAllowed(req.Email) {
		h.writeError(w, "Email address not allowed", http.StatusForbidden)
		return
	}

	user, err := h.db.CreateUserWithApproval(req.Email, req.DisplayName, req.Approved)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "User already exists", http.StatusConflict)
			return
		}
		logrus.Errorf("Failed to create user: %v", err)
		h.writeError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, user)
}

// UpdateUser updates a user (admin endpoint)
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeError(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Approved *bool `json:"approved"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing user
	user, err := h.db.GetUser(id)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	// Update approval status if provided
	if req.Approved != nil && *req.Approved {
		if err := h.db.ApproveUser(id); err != nil {
			logrus.Errorf("Failed to approve user: %v", err)
			h.writeError(w, "Failed to approve user", http.StatusInternalServerError)
			return
		}
		user.Approved = true
		logrus.Infof("User approved: %s", user.Email)
	}

	h.writeJSON(w, user)
}

// DeleteUser deletes a user (admin endpoint)
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeError(w, "User ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteUser(id); err != nil {
		logrus.Errorf("Failed to delete user: %v", err)
		h.writeError(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "success"})
}

func (h *Handlers) isAdmin(r *http.Request) bool {
	session, err := h.store.Get(r, "auth-session")
	if err != nil {
		return false
	}

	userEmail, ok := session.Values["user_email"].(string)
	if !ok {
		return false
	}

	return h.config.IsAdmin(userEmail)
}

func (h *Handlers) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !h.isAdmin(r) {
		h.writeError(w, "Admin access required", http.StatusForbidden)
		return false
	}
	return true
}
