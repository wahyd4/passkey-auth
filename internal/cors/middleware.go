package cors

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
)

// Config holds the CORS configuration with wildcard support
type Config struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// WildcardCORS creates a CORS handler with wildcard domain support
func WildcardCORS(config Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Determine allowed origins for this request
			var allowedOrigins []string

			// Separate wildcard and static origins
			var wildcardPatterns []string
			var staticOrigins []string

			for _, configuredOrigin := range config.AllowedOrigins {
				if strings.Contains(configuredOrigin, "*") {
					wildcardPatterns = append(wildcardPatterns, configuredOrigin)
				} else {
					staticOrigins = append(staticOrigins, configuredOrigin)
				}
			}

			// Check if origin matches any wildcard pattern
			if len(wildcardPatterns) > 0 && origin != "" {
				wildcardMatcher := NewWildcardMatcher(wildcardPatterns)
				if wildcardMatcher.MatchOrigin(origin) {
					// For wildcard matches, allow the specific origin
					allowedOrigins = []string{origin}
				}
			}

			// If no wildcard match, use static origins
			if len(allowedOrigins) == 0 {
				allowedOrigins = staticOrigins
			} else {
				// If we had a wildcard match, also include static origins
				allowedOrigins = append(allowedOrigins, staticOrigins...)
			}

			// Create a new CORS instance for this request with the determined origins
			c := cors.New(cors.Options{
				AllowedOrigins:   allowedOrigins,
				AllowedMethods:   config.AllowedMethods,
				AllowedHeaders:   config.AllowedHeaders,
				AllowCredentials: config.AllowCredentials,
			})

			// Use the rs/cors handler
			c.Handler(next).ServeHTTP(w, r)
		})
	}
}
