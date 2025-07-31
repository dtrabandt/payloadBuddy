package main

import (
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short string", 8},
		{"medium string", 12},
		{"long string", 32},
		{"single character", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRandomString(tt.length)

			// Check length
			if len(result) != tt.length {
				t.Errorf("Expected length %d, got %d", tt.length, len(result))
			}

			// Check character set (alphanumeric only)
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, char := range result {
				if !strings.ContainsRune(charset, char) {
					t.Errorf("Invalid character '%c' in generated string", char)
				}
			}

			// Check uniqueness (run multiple times)
			// Note: For very short strings (1 char), duplicates are likely due to limited charset
			if tt.length > 1 {
				results := make(map[string]bool)
				for i := 0; i < 10; i++ {
					s := generateRandomString(tt.length)
					if results[s] {
						t.Errorf("Generated duplicate string: %s", s)
					}
					results[s] = true
				}
			}
		})
	}
}

func TestGenerateRandomString_Uniqueness(t *testing.T) {
	// Test that multiple calls generate different strings
	length := 16
	iterations := 100
	results := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		result := generateRandomString(length)
		if results[result] {
			t.Errorf("Generated duplicate string after %d iterations: %s", i+1, result)
		}
		results[result] = true
	}
}

func TestSetupAuthentication(t *testing.T) {
	// Save original values
	originalEnableAuth := *enableAuth
	originalUsername := *username
	originalPassword := *password
	originalAuthUsername := authUsername
	originalAuthPassword := authPassword

	defer func() {
		// Restore original values
		*enableAuth = originalEnableAuth
		*username = originalUsername
		*password = originalPassword
		authUsername = originalAuthUsername
		authPassword = originalAuthPassword
	}()

	tests := []struct {
		name           string
		enableAuth     bool
		inputUsername  string
		inputPassword  string
		expectUsername bool
		expectPassword bool
	}{
		{
			name:           "auth disabled",
			enableAuth:     false,
			inputUsername:  "",
			inputPassword:  "",
			expectUsername: false,
			expectPassword: false,
		},
		{
			name:           "auth enabled with auto-generated credentials",
			enableAuth:     true,
			inputUsername:  "",
			inputPassword:  "",
			expectUsername: true,
			expectPassword: true,
		},
		{
			name:           "auth enabled with custom username",
			enableAuth:     true,
			inputUsername:  "testuser",
			inputPassword:  "",
			expectUsername: true,
			expectPassword: true,
		},
		{
			name:           "auth enabled with custom password",
			enableAuth:     true,
			inputUsername:  "",
			inputPassword:  "testpass",
			expectUsername: true,
			expectPassword: true,
		},
		{
			name:           "auth enabled with both custom credentials",
			enableAuth:     true,
			inputUsername:  "admin",
			inputPassword:  "secret123",
			expectUsername: true,
			expectPassword: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			*enableAuth = tt.enableAuth
			*username = tt.inputUsername
			*password = tt.inputPassword
			authUsername = ""
			authPassword = ""

			// Call function
			setupAuthentication()

			// Check results
			if tt.expectUsername {
				if authUsername == "" {
					t.Error("Expected authUsername to be set, but it's empty")
				}
				if tt.inputUsername != "" && authUsername != tt.inputUsername {
					t.Errorf("Expected custom username %q, got %q", tt.inputUsername, authUsername)
				}
				if tt.inputUsername == "" && len(authUsername) != 8 {
					t.Errorf("Expected auto-generated username length 8, got %d", len(authUsername))
				}
			} else {
				if authUsername != "" {
					t.Errorf("Expected authUsername to be empty, got %q", authUsername)
				}
			}

			if tt.expectPassword {
				if authPassword == "" {
					t.Error("Expected authPassword to be set, but it's empty")
				}
				if tt.inputPassword != "" && authPassword != tt.inputPassword {
					t.Errorf("Expected custom password %q, got %q", tt.inputPassword, authPassword)
				}
				if tt.inputPassword == "" && len(authPassword) != 12 {
					t.Errorf("Expected auto-generated password length 12, got %d", len(authPassword))
				}
			} else {
				if authPassword != "" {
					t.Errorf("Expected authPassword to be empty, got %q", authPassword)
				}
			}
		})
	}
}

func TestPrintAuthenticationInfo(t *testing.T) {
	// Save original values
	originalEnableAuth := *enableAuth
	originalAuthUsername := authUsername
	originalAuthPassword := authPassword

	defer func() {
		// Restore original values
		*enableAuth = originalEnableAuth
		authUsername = originalAuthUsername
		authPassword = originalAuthPassword
	}()

	tests := []struct {
		name         string
		enableAuth   bool
		username     string
		password     string
		shouldOutput bool
	}{
		{
			name:         "auth disabled",
			enableAuth:   false,
			username:     "",
			password:     "",
			shouldOutput: false,
		},
		{
			name:         "auth enabled",
			enableAuth:   true,
			username:     "testuser",
			password:     "testpass",
			shouldOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*enableAuth = tt.enableAuth
			authUsername = tt.username
			authPassword = tt.password

			// This function prints to stdout, so we can't easily capture output
			// But we can call it to ensure it doesn't panic
			printAuthenticationInfo()

			// The function should complete without error
			// In a more sophisticated test, we could capture stdout
		})
	}
}

func TestGetExampleURL(t *testing.T) {
	// Save original values
	originalEnableAuth := *enableAuth
	originalAuthUsername := authUsername
	originalAuthPassword := authPassword

	defer func() {
		// Restore original values
		*enableAuth = originalEnableAuth
		authUsername = originalAuthUsername
		authPassword = originalAuthPassword
	}()

	tests := []struct {
		name       string
		enableAuth bool
		username   string
		password   string
		baseURL    string
		expected   string
	}{
		{
			name:       "auth disabled",
			enableAuth: false,
			username:   "",
			password:   "",
			baseURL:    "http://localhost:8080/test",
			expected:   "http://localhost:8080/test",
		},
		{
			name:       "auth enabled",
			enableAuth: true,
			username:   "user123",
			password:   "pass456",
			baseURL:    "http://localhost:8080/api",
			expected:   "curl -u user123:pass456 http://localhost:8080/api",
		},
		{
			name:       "auth enabled with query parameters",
			enableAuth: true,
			username:   "admin",
			password:   "secret",
			baseURL:    "http://localhost:8080/stream?count=100",
			expected:   "curl -u admin:secret http://localhost:8080/stream?count=100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*enableAuth = tt.enableAuth
			authUsername = tt.username
			authPassword = tt.password

			result := getExampleURL(tt.baseURL)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
