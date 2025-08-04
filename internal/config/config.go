package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	WebAuthn WebAuthnConfig `yaml:"webauthn"`
	Database DatabaseConfig `yaml:"database"`
	CORS     CORSConfig     `yaml:"cors"`
	Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type WebAuthnConfig struct {
	RPDisplayName string   `yaml:"rp_display_name"`
	RPID          string   `yaml:"rp_id"`
	RPOrigins     []string `yaml:"rp_origins"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type AuthConfig struct {
	SessionSecret   string   `yaml:"session_secret"`
	RequireApproval bool     `yaml:"require_approval"`
	AllowedEmails   []string `yaml:"allowed_emails"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Validate config path to prevent path traversal attacks
	if err := validateConfigPath(configPath); err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	// Set defaults
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		WebAuthn: WebAuthnConfig{
			RPDisplayName: "Passkey Auth",
			RPID:          "localhost",
			RPOrigins:     []string{"http://localhost:8080"},
		},
		Database: DatabaseConfig{
			Path: "passkey-auth.db",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
		},
		Auth: AuthConfig{
			SessionSecret:   "change-me-in-production",
			RequireApproval: true,
			AllowedEmails:   []string{}, // Empty means no email restrictions
		},
	}

	// Load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		// #nosec G304 - configPath is validated above to prevent path traversal
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}
	if rpid := os.Getenv("WEBAUTHN_RP_ID"); rpid != "" {
		config.WebAuthn.RPID = rpid
	}
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}
	if secret := os.Getenv("SESSION_SECRET"); secret != "" {
		config.Auth.SessionSecret = secret
	}
	if allowedEmails := os.Getenv("ALLOWED_EMAILS"); allowedEmails != "" {
		config.Auth.AllowedEmails = strings.Split(allowedEmails, ",")
		// Trim whitespace from emails
		for i, email := range config.Auth.AllowedEmails {
			config.Auth.AllowedEmails[i] = strings.TrimSpace(email)
		}
	}

	return config, nil
}

// IsEmailAllowed checks if an email address is in the allowed list
// Returns true if the allowlist is empty (no restrictions) or if the email is in the list
func (c *Config) IsEmailAllowed(email string) bool {
	// If no allowed emails specified, allow all
	if len(c.Auth.AllowedEmails) == 0 {
		return true
	}

	// Check if email is in the allowed list
	for _, allowedEmail := range c.Auth.AllowedEmails {
		if email == allowedEmail {
			return true
		}
	}

	return false
}

// validateConfigPath ensures the config path is safe and doesn't allow path traversal
func validateConfigPath(path string) error {
	// Clean the path and check for path traversal attempts
	cleanPath := filepath.Clean(path)

	// Don't allow paths that try to go up directories
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Only allow certain file extensions
	ext := filepath.Ext(cleanPath)
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("only .yaml and .yml files are allowed")
	}

	// Convert to absolute path to check if it's within allowed directories
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Only allow config files in current directory or its subdirectories
	if !strings.HasPrefix(absPath, wd) {
		return fmt.Errorf("config file must be in current directory or subdirectories")
	}

	return nil
}
