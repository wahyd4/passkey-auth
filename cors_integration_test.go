package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"passkey-auth/internal/cors"
)

func TestWildcardCORSIntegration(t *testing.T) {
	// Create a simple test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Create CORS middleware with wildcard support
	corsMiddleware := cors.WildcardCORS(cors.Config{
		AllowedOrigins:   []string{"*.junv.cc", "https://static.example.com"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Wrap the test handler
	handler := corsMiddleware(testHandler)

	tests := []struct {
		name           string
		origin         string
		expectAllowed  bool
		expectedOrigin string
	}{
		{
			name:           "wildcard subdomain match",
			origin:         "https://api.junv.cc",
			expectAllowed:  true,
			expectedOrigin: "https://api.junv.cc",
		},
		{
			name:           "wildcard base domain match",
			origin:         "https://junv.cc",
			expectAllowed:  true,
			expectedOrigin: "https://junv.cc",
		},
		{
			name:          "static domain match",
			origin:        "https://static.example.com",
			expectAllowed: true,
		},
		{
			name:          "no match",
			origin:        "https://evil.com",
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a preflight OPTIONS request
			req := httptest.NewRequest("OPTIONS", "/", nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", "POST")

			// Record the response
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Check CORS headers
			allowOriginHeader := w.Header().Get("Access-Control-Allow-Origin")

			if tt.expectAllowed {
				if allowOriginHeader == "" {
					t.Errorf("Expected Access-Control-Allow-Origin header, but got none")
				}

				// For wildcard matches, should return the specific origin
				if tt.expectedOrigin != "" && allowOriginHeader != tt.expectedOrigin {
					t.Errorf("Expected Access-Control-Allow-Origin: %s, got: %s", tt.expectedOrigin, allowOriginHeader)
				}
			} else {
				if allowOriginHeader != "" {
					t.Errorf("Expected no Access-Control-Allow-Origin header, but got: %s", allowOriginHeader)
				}
			}
		})
	}
}
