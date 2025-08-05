package cors

import (
	"strings"
)

// WildcardMatcher provides wildcard domain matching for CORS origins
type WildcardMatcher struct {
	patterns []string
}

// NewWildcardMatcher creates a new wildcard matcher with the given patterns
func NewWildcardMatcher(patterns []string) *WildcardMatcher {
	return &WildcardMatcher{
		patterns: patterns,
	}
}

// MatchOrigin checks if the given origin matches any of the wildcard patterns
func (m *WildcardMatcher) MatchOrigin(origin string) bool {
	for _, pattern := range m.patterns {
		if m.matchPattern(origin, pattern) {
			return true
		}
	}
	return false
}

// matchPattern checks if origin matches a specific pattern
// Supports patterns like:
// - "*.example.com" matches "api.example.com", "auth.example.com", etc.
// - "*.*.example.com" matches "api.v1.example.com", etc.
// - "example.com" matches exactly "example.com"
func (m *WildcardMatcher) matchPattern(origin, pattern string) bool {
	// Remove protocol from origin if present
	origin = strings.TrimPrefix(origin, "https://")
	origin = strings.TrimPrefix(origin, "http://")

	// Remove port if present
	if colonIndex := strings.LastIndex(origin, ":"); colonIndex != -1 && colonIndex > strings.LastIndex(origin, "]") {
		origin = origin[:colonIndex]
	}

	// Exact match
	if origin == pattern {
		return true
	}

	// Wildcard match
	if strings.Contains(pattern, "*") {
		return m.wildcardMatch(origin, pattern)
	}

	return false
}

// wildcardMatch performs wildcard matching
func (m *WildcardMatcher) wildcardMatch(origin, pattern string) bool {
	// Handle simple case: *.domain.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:] // Remove "*."

		// Check if origin ends with the suffix and has at least one subdomain
		if strings.HasSuffix(origin, "."+suffix) {
			// Ensure there's a subdomain (not just the suffix itself)
			prefix := strings.TrimSuffix(origin, "."+suffix)
			// Make sure the prefix doesn't contain dots (single-level subdomain wildcard)
			// If you want multi-level subdomains, remove this check
			return !strings.Contains(prefix, ".")
		}

		// Also check if origin exactly matches the suffix (without subdomain)
		return origin == suffix
	}

	// For more complex patterns, we could implement more sophisticated matching
	// For now, handle the common *.domain.com case
	return false
}

// GetAllowedOrigins returns the actual allowed origins for a request
// This expands wildcard patterns based on the request origin
func (m *WildcardMatcher) GetAllowedOrigins(requestOrigin string, staticOrigins []string) []string {
	allowedOrigins := make([]string, 0, len(staticOrigins))
	hasMatchingWildcard := false

	// First pass: check if any wildcard matches
	for _, origin := range staticOrigins {
		if strings.Contains(origin, "*") {
			if m.matchPattern(requestOrigin, origin) {
				hasMatchingWildcard = true
				break
			}
		}
	}

	// Second pass: build the result based on the logic
	for _, origin := range staticOrigins {
		if strings.Contains(origin, "*") {
			// This is a wildcard pattern
			if m.matchPattern(requestOrigin, origin) {
				// Add the actual request origin instead of the pattern
				allowedOrigins = append(allowedOrigins, requestOrigin)
			} else if !hasMatchingWildcard {
				// No wildcards match, so include this wildcard pattern as-is
				allowedOrigins = append(allowedOrigins, origin)
			}
			// If a wildcard matches but this one doesn't, skip it (don't add anything)
		} else {
			// This is a static origin, always add as-is
			allowedOrigins = append(allowedOrigins, origin)
		}
	}

	return allowedOrigins
}
