package cors

import (
	"testing"
)

func TestWildcardMatcher(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		origin   string
		expected bool
	}{
		{
			name:     "exact match",
			patterns: []string{"example.com"},
			origin:   "example.com",
			expected: true,
		},
		{
			name:     "exact match with https",
			patterns: []string{"example.com"},
			origin:   "https://example.com",
			expected: true,
		},
		{
			name:     "wildcard subdomain match",
			patterns: []string{"*.junv.cc"},
			origin:   "api.junv.cc",
			expected: true,
		},
		{
			name:     "wildcard subdomain match with https",
			patterns: []string{"*.junv.cc"},
			origin:   "https://auth.junv.cc",
			expected: true,
		},
		{
			name:     "wildcard subdomain match with port",
			patterns: []string{"*.junv.cc"},
			origin:   "https://dev.junv.cc:3000",
			expected: true,
		},
		{
			name:     "wildcard base domain match",
			patterns: []string{"*.junv.cc"},
			origin:   "junv.cc",
			expected: true,
		},
		{
			name:     "wildcard no match - different domain",
			patterns: []string{"*.junv.cc"},
			origin:   "api.example.com",
			expected: false,
		},
		{
			name:     "wildcard no match - multi-level subdomain",
			patterns: []string{"*.junv.cc"},
			origin:   "api.v1.junv.cc",
			expected: false,
		},
		{
			name:     "multiple patterns - first match",
			patterns: []string{"*.junv.cc", "*.example.com"},
			origin:   "api.junv.cc",
			expected: true,
		},
		{
			name:     "multiple patterns - second match",
			patterns: []string{"*.junv.cc", "*.example.com"},
			origin:   "api.example.com",
			expected: true,
		},
		{
			name:     "no match",
			patterns: []string{"*.junv.cc", "*.example.com"},
			origin:   "api.other.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewWildcardMatcher(tt.patterns)
			result := matcher.MatchOrigin(tt.origin)
			if result != tt.expected {
				t.Errorf("MatchOrigin() = %v, expected %v for origin %s with patterns %v",
					result, tt.expected, tt.origin, tt.patterns)
			}
		})
	}
}

func TestGetAllowedOrigins(t *testing.T) {
	tests := []struct {
		name          string
		patterns      []string
		requestOrigin string
		expected      []string
	}{
		{
			name:          "wildcard match includes request origin",
			patterns:      []string{"*.junv.cc", "https://static.com"},
			requestOrigin: "https://api.junv.cc",
			expected:      []string{"https://api.junv.cc", "https://static.com"},
		},
		{
			name:          "no wildcard match returns static origins",
			patterns:      []string{"*.junv.cc", "https://static.com"},
			requestOrigin: "https://other.com",
			expected:      []string{"*.junv.cc", "https://static.com"},
		},
		{
			name:          "multiple wildcards, one matches",
			patterns:      []string{"*.junv.cc", "*.example.com"},
			requestOrigin: "https://api.junv.cc",
			expected:      []string{"https://api.junv.cc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewWildcardMatcher(tt.patterns)
			result := matcher.GetAllowedOrigins(tt.requestOrigin, tt.patterns)

			if len(result) != len(tt.expected) {
				t.Errorf("GetAllowedOrigins() returned %d origins, expected %d", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("GetAllowedOrigins()[%d] = %v, expected %v", i, result[i], expected)
				}
			}
		})
	}
}
