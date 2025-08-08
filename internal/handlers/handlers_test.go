package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"passkey-auth/internal/auth"
	"passkey-auth/internal/config"
	"passkey-auth/internal/database"
)

func TestAuthCheckNginx(t *testing.T) {
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

	tests := []struct {
		name           string
		authenticated  bool
		expectedStatus int
	}{
		{
			name:           "Unauthenticated user",
			authenticated:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Authenticated user",
			authenticated:  true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", "/auth/nginx", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Mock session if authenticated
			if tt.authenticated {
				session, _ := h.store.Get(req, "auth-session")
				session.Values["authenticated"] = true
				session.Values["user_id"] = 1
				session.Values["user_email"] = "test@example.com"
				
				w := httptest.NewRecorder()
				session.Save(req, w)
				
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
			h.AuthCheckNginx(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
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

func TestAuthCheckTraefik(t *testing.T) {
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

	tests := []struct {
		name           string
		authenticated  bool
		headers        map[string]string
		expectedStatus int
		expectedRedirect string
	}{
		{
			name:           "Unauthenticated without headers",
			authenticated:  false,
			headers:        map[string]string{},
			expectedStatus: http.StatusFound,
			expectedRedirect: "/login.html",
		},
		{
			name:           "Unauthenticated with forwarded headers",
			authenticated:  false,
			headers: map[string]string{
				"X-Forwarded-Host":  "example.com",
				"X-Forwarded-Uri":   "/protected/page",
				"X-Forwarded-Proto": "https",
			},
			expectedStatus: http.StatusFound,
			expectedRedirect: "/login.html?redirect=https://example.com/protected/page",
		},
		{
			name:           "Authenticated user",
			authenticated:  true,
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", "/auth/traefik", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Mock session if authenticated
			if tt.authenticated {
				session, _ := h.store.Get(req, "auth-session")
				session.Values["authenticated"] = true
				session.Values["user_id"] = 1
				session.Values["user_email"] = "test@example.com"
				
				w := httptest.NewRecorder()
				session.Save(req, w)
				
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
			h.AuthCheckTraefik(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check Location header for redirects
			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				if location != tt.expectedRedirect {
					t.Errorf("Expected Location header %s, got %s", tt.expectedRedirect, location)
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

func TestAuthCheckTraefikWithHost(t *testing.T) {
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

	// Test with Host header for full URL construction
	req, err := http.NewRequest("GET", "/auth/traefik", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Host", "auth.example.com")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "app.example.com")
	req.Header.Set("X-Forwarded-Uri", "/protected/resource")

	w := httptest.NewRecorder()
	h.AuthCheckTraefik(w, req)

	// Should return 302 with full URL
	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	expectedLocation := "https://auth.example.com/login.html?redirect=https://app.example.com/protected/resource"
	location := w.Header().Get("Location")
	if location != expectedLocation {
		t.Errorf("Expected Location header %s, got %s", expectedLocation, location)
	}
}