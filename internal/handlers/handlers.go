package handlers

import (
	"encoding/json"
	"io"
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
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *Handlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
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

	// Create new user
	user, err := h.db.CreateUser(req.Email, req.DisplayName)
	if err != nil {
		logrus.Errorf("Failed to create user: %v", err)
		h.writeError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	webAuthnUser := &auth.WebAuthnUser{}
	webAuthnUser.SetUser(user)

	options, sessionData, err := h.webAuthn.BeginRegistration(webAuthnUser)
	if err != nil {
		logrus.Errorf("Failed to begin registration: %v", err)
		h.writeError(w, "Failed to begin registration", http.StatusInternalServerError)
		return
	}

	// Store session data
	session, _ := h.store.Get(r, "webauthn-session")
	session.Values["challenge"] = sessionData.Challenge
	session.Values["user_id"] = user.ID
	session.Save(r, w)

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

	// Read and log the request body for debugging
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	logrus.Debugf("Received credential response body: %s", string(body))

	// Create a new reader from the body for the WebAuthn library
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	user, err := h.db.GetUser(userID)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	webAuthnUser := &auth.WebAuthnUser{}
	webAuthnUser.SetUser(user)

	sessionData := webauthn.SessionData{
		Challenge: challenge,
		UserID:    webAuthnUser.WebAuthnID(),
	}

	// Log the request details for debugging
	logrus.Debugf("Finishing registration for user: %s", user.Email)
	logrus.Debugf("Session challenge: %s", challenge)
	logrus.Debugf("Session user ID: %v", webAuthnUser.WebAuthnID())

	credential, err := h.webAuthn.FinishRegistration(webAuthnUser, sessionData, r)
	if err != nil {
		logrus.Errorf("Failed to finish registration: %v", err)
		h.writeError(w, "Failed to finish registration", http.StatusInternalServerError)
		return
	}

	// Save credential to database
	if err := h.webAuthn.SaveCredential(user.ID, credential); err != nil {
		logrus.Errorf("Failed to save credential: %v", err)
		h.writeError(w, "Failed to save credential", http.StatusInternalServerError)
		return
	}

	// Clear session
	session.Values["challenge"] = nil
	session.Values["user_id"] = nil
	session.Save(r, w)

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
	session.Save(r, w)

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
	authSession.Save(r, w)

	// Clear webauthn session
	session.Values["challenge"] = nil
	session.Values["user_id"] = nil
	session.Save(r, w)

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
	session.Save(r, w)

	h.writeJSON(w, map[string]string{"status": "success"})
}

// AuthCheck implements the nginx auth_request protocol
func (h *Handlers) AuthCheck(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "auth-session")

	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok || !authenticated {
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

	w.WriteHeader(http.StatusOK)
}

// Admin endpoints

// ListUsers returns all users (admin endpoint)
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin authentication check
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
	// TODO: Add admin authentication check
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

	user, err := h.db.CreateUser(req.Email, req.DisplayName)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "User already exists", http.StatusConflict)
			return
		}
		logrus.Errorf("Failed to create user: %v", err)
		h.writeError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	if req.Approved {
		if err := h.db.ApproveUser(user.ID); err != nil {
			logrus.Errorf("Failed to approve user: %v", err)
		}
	}

	h.writeJSON(w, user)
}

// UpdateUser updates a user (admin endpoint)
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin authentication check
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
	// TODO: Add admin authentication check
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
