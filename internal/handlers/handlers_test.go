package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"passkey-auth/internal/auth"
	"passkey-auth/internal/config"
	"passkey-auth/internal/database"
)

func TestAuthCheck(t *testing.T) {
	// Setup test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			SessionSecret: "test-secret",
		},
		WebAuthn: config.WebAuthnConfig{
			RPDisplayName: "Test Passkey Auth",
			RPID:          "localhost",
			RPOrigins:     []string{"http://localhost:8080"},
		},
	}

	// Setup test database
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup test WebAuthn (mock)
	webAuthn, err := auth.NewWebAuthn(cfg)
	if err != nil {
		t.Fatalf("Failed to create test WebAuthn: %v", err)
	}

	// Create handlers
	h := New(db, webAuthn, cfg)

	tests := []struct {
		name           string
		queryParams    string
		authenticated  bool
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "Unauthenticated without redirect param (Nginx)",
			queryParams:    "",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
			expectedHeader: "",
		},
		{
			name:           "Unauthenticated with rd param (Traefik)",
			queryParams:    "rd=https://example.com/protected",
			authenticated:  false,
			expectedStatus: http.StatusFound,
			expectedHeader: "/login.html?redirect=https://example.com/protected",
		},
		{
			name:           "Unauthenticated with redirect param (Traefik)",
			queryParams:    "redirect=https://example.com/protected",
			authenticated:  false,
			expectedStatus: http.StatusFound,
			expectedHeader: "/login.html?redirect=https://example.com/protected",
		},
		{
			name:           "Authenticated with redirect param",
			queryParams:    "rd=https://example.com/protected",
			authenticated:  true,
			expectedStatus: http.StatusOK,
			expectedHeader: "",
		},
		{
			name:           "Authenticated without redirect param",
			queryParams:    "",
			authenticated:  true,
			expectedStatus: http.StatusOK,
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", "/auth?"+tt.queryParams, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Mock session if authenticated
			if tt.authenticated {
				// Create a session cookie for testing
				session, _ := h.store.Get(req, "auth-session")
				session.Values["authenticated"] = true
				session.Values["user_id"] = 1
				session.Values["user_email"] = "test@example.com"

				// Create a response recorder to capture the session cookie
				w := httptest.NewRecorder()
				session.Save(req, w)

				// Extract the cookie and add it to the request
				for _, cookie := range w.Result().Cookies() {
					if cookie.Name == "auth-session" {
						req.AddCookie(cookie)
						break
					}
				}
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			h.AuthCheck(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check Location header for redirects
			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				if location != tt.expectedHeader {
					t.Errorf("Expected Location header %s, got %s", tt.expectedHeader, location)
				}
			}

			// Check auth headers for authenticated requests
			if tt.authenticated && tt.expectedStatus == http.StatusOK {
				userHeader := w.Header().Get("X-Auth-User")
				if userHeader == "" {
					t.Error("Expected X-Auth-User header for authenticated request")
				}
			}
		})
	}
}

func TestAuthCheckWithHost(t *testing.T) {
	// Setup test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			SessionSecret: "test-secret",
		},
		WebAuthn: config.WebAuthnConfig{
			RPDisplayName: "Test Passkey Auth",
			RPID:          "localhost",
			RPOrigins:     []string{"http://localhost:8080"},
		},
	}

	// Setup test database
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup test WebAuthn
	webAuthn, err := auth.NewWebAuthn(cfg)
	if err != nil {
		t.Fatalf("Failed to create test WebAuthn: %v", err)
	}

	// Create handlers
	h := New(db, webAuthn, cfg)

	// Test with Host header
	req, err := http.NewRequest("GET", "/auth?rd=https://example.com/protected", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Host", "auth.example.com")
	req.Header.Set("X-Forwarded-Proto", "https")

	w := httptest.NewRecorder()
	h.AuthCheck(w, req)

	// Should return 302 with full URL
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	expectedLocation := "https://auth.example.com/login.html?redirect=https://example.com/protected"
	location := w.Header().Get("Location")
	if location != expectedLocation {
		t.Errorf("Expected Location header %s, got %s", expectedLocation, location)
	}
}
