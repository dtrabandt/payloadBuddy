package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
)

// Config holds the runtime authentication state for the server.
type Config struct {
	Enabled  bool
	Username string
	Password string
}

// Setup configures authentication from the provided flag values.
// When enabled is true and a credential is empty, a secure random value is generated.
func Setup(enabled bool, username, password string) *Config {
	cfg := &Config{Enabled: enabled}
	if !enabled {
		return cfg
	}
	if username == "" {
		cfg.Username = generateRandomString(8)
	} else {
		cfg.Username = username
	}
	if password == "" {
		cfg.Password = generateRandomString(12)
	} else {
		cfg.Password = password
	}
	return cfg
}

// Middleware returns an http.HandlerFunc that enforces HTTP Basic Auth when
// authentication is enabled, using constant-time comparison to prevent timing attacks.
func (c *Config) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !c.Enabled {
			next(w, r)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(c.Username)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(c.Password)) == 1
		if !userMatch || !passMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// ExampleURL returns a ready-to-use curl command (auth enabled) or the bare URL (auth disabled).
func (c *Config) ExampleURL(baseURL string) string {
	if c.Enabled {
		return fmt.Sprintf("curl -u %s:%s %s", c.Username, c.Password, baseURL)
	}
	return baseURL
}

// PrintInfo displays credentials to stdout when authentication is enabled.
func (c *Config) PrintInfo() {
	if !c.Enabled {
		return
	}
	fmt.Println("\n=== BASIC AUTHENTICATION ENABLED ===")
	fmt.Printf("Username: %s\n", c.Username)
	fmt.Printf("Password: %s\n", c.Password)
	credentials := c.Username + ":" + c.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	fmt.Printf("Auth Header: Authorization: Basic %s\n", encoded)
	fmt.Println("=====================================")
}

// generateRandomString returns a cryptographically secure alphanumeric string.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate secure random bytes: %v", err))
	}
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
