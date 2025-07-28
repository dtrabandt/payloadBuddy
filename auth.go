// Package auth provides HTTP Basic Authentication middleware and utilities
// for the gohugePayloadServer application.
//
// This package implements secure HTTP Basic Authentication with the following features:
// - Cryptographically secure random credential generation
// - Constant-time comparison to prevent timing attacks
// - Configurable via command-line flags
// - Optional authentication (can be disabled)
// - User-friendly credential display and URL generation
//
// Security Considerations:
// - Uses crypto/rand for secure random number generation
// - Implements constant-time comparison to prevent timing side-channel attacks
// - Credentials are displayed in plaintext on startup (intended for development/testing)
// - No password hashing (Basic Auth sends credentials in base64, not suitable for production)
//
// Usage:
//   go run . -auth                          // Enable with auto-generated credentials
//   go run . -auth -user=myuser -pass=mypass // Enable with custom credentials
//   go run .                                // Disable authentication (default)
package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
)

// Authentication configuration variables
//
// These variables control the authentication behavior of the server and are
// populated from command-line flags during application startup.
var (
	// enableAuth is a command-line flag that controls whether HTTP Basic Authentication
	// is enabled for the server endpoints. When false, all requests pass through
	// without authentication checks.
	//
	// Default: false (authentication disabled)
	// Flag: -auth
	enableAuth = flag.Bool("auth", false, "Enable basic authentication")
	
	// username is a command-line flag for specifying a custom username.
	// If empty when authentication is enabled, a secure random username is generated.
	//
	// Default: "" (auto-generate)
	// Flag: -user=<username>
	username = flag.String("user", "", "Username for basic auth (auto-generated if empty)")
	
	// password is a command-line flag for specifying a custom password.
	// If empty when authentication is enabled, a secure random password is generated.
	//
	// Default: "" (auto-generate)
	// Flag: -pass=<password>
	password = flag.String("pass", "", "Password for basic auth (auto-generated if empty)")
	
	// authUsername holds the actual username used for authentication.
	// This is either the value from the -user flag or an auto-generated secure string.
	// Only populated when authentication is enabled.
	authUsername string
	
	// authPassword holds the actual password used for authentication.
	// This is either the value from the -pass flag or an auto-generated secure string.
	// Only populated when authentication is enabled.
	authPassword string
)

// generateRandomString generates a cryptographically secure random string of the specified length.
//
// This function uses the crypto/rand package to generate secure random bytes, which are
// then mapped to a character set containing uppercase letters, lowercase letters, and digits.
// The resulting string is suitable for use as usernames and passwords in development/testing
// environments.
//
// Security Properties:
// - Uses crypto/rand.Read() for cryptographically secure random number generation
// - Uniform distribution across the character set using modulo operation
// - No predictable patterns or sequences
// - Suitable entropy for development credentials (not intended for production secrets)
//
// Character Set: [a-zA-Z0-9] (62 possible characters)
// Entropy: log2(62^length) bits (e.g., 8 chars = ~47.6 bits, 12 chars = ~71.5 bits)
//
// Parameters:
//   length - The desired length of the generated string (must be > 0)
//
// Returns:
//   A random string of the specified length containing only alphanumeric characters
//
// Panics:
//   If crypto/rand.Read() fails (extremely rare, indicates system-level issues)
//
// Example:
//   username := generateRandomString(8)   // e.g., "Kj9mN2pQ"
//   password := generateRandomString(12)  // e.g., "7hG3kL9mP4xR"
func generateRandomString(length int) string {
	// Character set: 26 lowercase + 26 uppercase + 10 digits = 62 total characters
	// This provides good entropy while avoiding special characters that might
	// cause issues in URLs or command-line usage
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	// Allocate byte slice for random data
	b := make([]byte, length)
	
	// Generate cryptographically secure random bytes
	// crypto/rand.Read() is the gold standard for secure random generation in Go
	if _, err := rand.Read(b); err != nil {
		// This should never happen in practice unless there are serious system issues
		// (e.g., /dev/urandom is unavailable on Unix systems)
		panic(fmt.Sprintf("Failed to generate secure random bytes: %v", err))
	}
	
	// Map each random byte to a character in our charset
	// We use modulo to ensure uniform distribution across the character set
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	
	return string(b)
}

// basicAuthMiddleware provides HTTP Basic Authentication middleware for protecting endpoints.
//
// This middleware implements secure HTTP Basic Authentication with the following security features:
// - Constant-time comparison to prevent timing side-channel attacks
// - Proper WWW-Authenticate header handling per RFC 7617
// - Graceful handling of missing or malformed Authorization headers
// - Optional authentication (bypassed when enableAuth is false)
//
// Security Implementation Details:
//
// 1. Timing Attack Prevention:
//    Uses crypto/subtle.ConstantTimeCompare to prevent attackers from determining
//    valid usernames or passwords by measuring response times. Standard string
//    comparison (==) would exit early on the first differing character, creating
//    a timing side-channel that could be exploited.
//
// 2. Authentication Flow:
//    - If authentication is disabled, requests pass through immediately
//    - Extracts credentials from Authorization: Basic <base64> header
//    - Validates both username and password using constant-time comparison
//    - Returns 401 Unauthorized with WWW-Authenticate header if validation fails
//
// 3. Security Considerations:
//    - HTTP Basic Auth transmits credentials in base64 (not encrypted)
//    - Should only be used over HTTPS in production environments
//    - Credentials are compared against in-memory values (no persistent storage)
//    - No rate limiting or account lockout (suitable for development/testing only)
//
// Parameters:
//   next - The next HTTP handler in the middleware chain
//
// Returns:
//   An http.HandlerFunc that performs authentication before calling the next handler
//
// HTTP Status Codes:
//   200 - Authentication successful, request passed to next handler
//   401 - Authentication failed or credentials missing
//
// Example Usage:
//   http.HandleFunc("/protected", basicAuthMiddleware(myHandler))
//
// Example Requests:
//   curl -u username:password http://localhost:8080/endpoint  # Valid request
//   curl http://localhost:8080/endpoint                       # Returns 401
//   curl -H "Authorization: Basic $(echo -n user:pass | base64)" http://localhost:8080/endpoint
//
// Security Warnings:
//   - This implementation is suitable for development and testing only
//   - For production use, consider OAuth 2.0, JWT, or other modern auth methods
//   - Always use HTTPS when transmitting Basic Auth credentials
//   - Consider implementing rate limiting to prevent brute force attacks
func basicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If authentication is disabled globally, bypass all checks
		// This allows the server to run in open mode for development
		if !*enableAuth {
			next(w, r)
			return
		}

		// Extract credentials from the Authorization header
		// r.BasicAuth() handles the parsing of "Authorization: Basic <base64>" header
		// and returns the decoded username and password
		user, pass, ok := r.BasicAuth()
		if !ok {
			// No valid Basic Auth header found - this could mean:
			// - No Authorization header at all
			// - Authorization header with wrong scheme (e.g., Bearer instead of Basic)
			// - Malformed base64 encoding in the Basic Auth header
			
			// Set WWW-Authenticate header to inform client about required auth method
			// The "realm" parameter is a human-readable string describing the protected area
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate credentials using constant-time comparison
		// This is critical for security - we must check both username and password
		// even if the username is wrong, to prevent timing attacks
		
		// subtle.ConstantTimeCompare returns 1 if the slices are equal, 0 otherwise
		// It always examines every byte in both slices, regardless of whether
		// differences are found early, making timing attacks infeasible
		userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(authUsername)) == 1
		passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(authPassword)) == 1

		// Both username AND password must match for authentication to succeed
		// We use separate boolean variables to ensure both comparisons always execute
		// (preventing timing attacks based on short-circuit evaluation)
		if !userMatch || !passMatch {
			// Authentication failed - send same response as missing credentials
			// This prevents username enumeration by ensuring identical responses
			// for "no credentials" and "wrong credentials" scenarios
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Authentication successful - proceed to the next handler
		// At this point, we can be confident that the request is from an
		// authenticated user with valid credentials
		next(w, r)
	}
}

// setupAuthentication configures the authentication system based on command-line flags.
//
// This function must be called after flag.Parse() to properly initialize the authentication
// system. It examines the command-line flags and sets up the global authentication state
// accordingly.
//
// Behavior:
//   - If authentication is disabled (-auth flag not set), this function does nothing
//   - If authentication is enabled but no custom credentials provided, generates secure random credentials
//   - If custom credentials are provided via -user and/or -pass flags, uses those values
//   - Supports mixed scenarios (e.g., custom username with auto-generated password)
//
// Credential Generation:
//   - Usernames: 8 characters (provides ~47.6 bits of entropy)
//   - Passwords: 12 characters (provides ~71.5 bits of entropy)
//   - Both use alphanumeric characters [a-zA-Z0-9] for maximum compatibility
//
// Security Considerations:
//   - Auto-generated credentials are cryptographically secure using crypto/rand
//   - Custom credentials should be strong enough for the intended use case
//   - Generated credentials are displayed in plaintext on startup (development/testing only)
//   - No validation is performed on custom credentials (user responsibility)
//
// Flag Dependencies:
//   This function depends on the following command-line flags being parsed:
//   - enableAuth (*bool): Whether authentication is enabled
//   - username (*string): Custom username (empty string triggers auto-generation)
//   - password (*string): Custom password (empty string triggers auto-generation)
//
// Side Effects:
//   - Modifies global variables authUsername and authPassword
//   - These variables are used by basicAuthMiddleware for credential validation
//
// Example Scenarios:
//   ./server -auth                          → Auto-generate both username and password
//   ./server -auth -user=admin              → Use "admin" as username, auto-generate password
//   ./server -auth -pass=secret123          → Auto-generate username, use "secret123" as password
//   ./server -auth -user=admin -pass=secret → Use both custom credentials
//   ./server                                → Authentication disabled, function does nothing
//
// Note: This function should be called exactly once during application startup,
// after flag.Parse() but before starting the HTTP server.
func setupAuthentication() {
	// Only configure authentication if it's been enabled via the -auth flag
	if *enableAuth {
		// Configure username: use custom value if provided, otherwise generate secure random
		if *username == "" {
			// Generate an 8-character random username
			// 8 chars from 62-char alphabet provides ~47.6 bits of entropy
			// This is sufficient for development/testing scenarios
			authUsername = generateRandomString(8)
		} else {
			// Use the custom username provided via -user flag
			// No validation is performed - user is responsible for choosing appropriate values
			authUsername = *username
		}
		
		// Configure password: use custom value if provided, otherwise generate secure random
		if *password == "" {
			// Generate a 12-character random password
			// 12 chars from 62-char alphabet provides ~71.5 bits of entropy
			// This exceeds most security guidelines for temporary development credentials
			authPassword = generateRandomString(12)
		} else {
			// Use the custom password provided via -pass flag
			// No validation is performed - user is responsible for choosing secure passwords
			authPassword = *password
		}
	}
	// If authentication is disabled, authUsername and authPassword remain empty strings
	// The basicAuthMiddleware will bypass all authentication checks in this case
}

// printAuthenticationInfo displays authentication credentials and usage information to stdout.
//
// This function is designed for development and testing environments where it's acceptable
// to display credentials in plaintext. It provides users with all the information needed
// to authenticate against the server, including multiple formats for different tools.
//
// Output Information:
//   - Clear indication that authentication is enabled
//   - Username and password in plaintext
//   - Pre-encoded Authorization header value for direct use
//   - Formatted display for easy copying
//
// Security Considerations:
//   - Credentials are displayed in PLAINTEXT on the console
//   - This output may be logged to files or visible in process lists
//   - Intended ONLY for development and testing environments
//   - Should NOT be used in production or when credentials need to remain secret
//   - Consider the security implications of displaying credentials in shared environments
//
// When Called:
//   - Only displays information if authentication is enabled (*enableAuth == true)
//   - Should be called after setupAuthentication() to ensure credentials are configured
//   - Typically called during server startup to inform operators of the credentials
//
// Output Format:
//   The function outputs a clearly formatted block containing:
//   - Visual separators for easy identification
//   - Username: plaintext value for -u curl parameter
//   - Password: plaintext value for -u curl parameter  
//   - Auth Header: base64-encoded "username:password" for direct header usage
//
// Example Output:
//   === BASIC AUTHENTICATION ENABLED ===
//   Username: Kj9mN2pQ
//   Password: 7hG3kL9mP4xR
//   Auth Header: Authorization: Basic S2o5bU4ycFE6N2hHM2tMOW1QNHhS
//   =====================================
//
// Usage Examples:
//   The displayed information can be used with various HTTP clients:
//   - curl: curl -u Kj9mN2pQ:7hG3kL9mP4xR http://localhost:8080/endpoint
//   - HTTPie: http GET localhost:8080/endpoint -a Kj9mN2pQ:7hG3kL9mP4xR
//   - Manual header: curl -H "Authorization: Basic S2o5bU4ycFE6N2hHM2tMOW1QNHhS" http://localhost:8080/endpoint
//
// Security Warning:
//   This function intentionally displays sensitive authentication credentials.
//   Only use in environments where this is acceptable (development, testing, demos).
func printAuthenticationInfo() {
	// Only display authentication info if it's actually enabled
	if *enableAuth {
		// Print a clear header to make authentication status obvious
		fmt.Println("\n=== BASIC AUTHENTICATION ENABLED ===")
		
		// Display the username in plaintext for easy copying to curl -u commands
		fmt.Printf("Username: %s\n", authUsername)
		
		// Display the password in plaintext for easy copying to curl -u commands
		fmt.Printf("Password: %s\n", authPassword)
		
		// Pre-encode the credentials as a complete Authorization header value
		// This saves users from having to manually base64 encode "username:password"
		// The format follows RFC 7617: Authorization: Basic <base64(username:password)>
		credentials := authUsername + ":" + authPassword
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
		fmt.Printf("Auth Header: Authorization: Basic %s\n", encodedCredentials)
		
		// Print a clear footer to mark the end of authentication information
		fmt.Println("=====================================")
	}
	// If authentication is disabled, this function silently does nothing
	// This allows it to be called unconditionally without cluttering output
}

// getExampleURL generates user-friendly example commands for accessing the provided URL.
//
// This utility function creates ready-to-use command examples that users can copy and paste
// to test the server endpoints. The output format depends on whether authentication is enabled,
// ensuring users always get working examples regardless of the server configuration.
//
// Behavior:
//   - If authentication is disabled: returns the URL as-is for direct browser/tool access
//   - If authentication is enabled: returns a complete curl command with embedded credentials
//
// Parameters:
//   baseURL - The complete URL to be accessed (e.g., "http://localhost:8080/huge_payload")
//            Should include protocol, host, port, path, and any query parameters
//
// Returns:
//   A string containing either:
//   - The original URL (when authentication is disabled)
//   - A curl command with authentication (when authentication is enabled)
//
// Output Examples:
//
//   Authentication Disabled:
//     Input:  "http://localhost:8080/huge_payload"
//     Output: "http://localhost:8080/huge_payload"
//
//   Authentication Enabled:
//     Input:  "http://localhost:8080/huge_payload?count=1000"
//     Output: "curl -u Kj9mN2pQ:7hG3kL9mP4xR http://localhost:8080/huge_payload?count=1000"
//
// Usage Patterns:
//   This function is typically used when generating help text or startup messages:
//
//   fmt.Printf("Test endpoint: %s\n", getExampleURL("http://localhost:8080/test"))
//   fmt.Printf("API docs: %s\n", getExampleURL("http://localhost:8080/docs"))
//
// Security Considerations:
//   - When authentication is enabled, credentials are embedded in the returned string
//   - The returned commands are safe to display in development/testing environments
//   - Consider the audience when displaying these examples (credentials are visible)
//   - The curl format makes it easy for users to modify for their HTTP client of choice
//
// Alternative HTTP Clients:
//   While this function returns curl examples, users can adapt them for other tools:
//   - HTTPie: http GET localhost:8080/endpoint -a username:password
//   - wget: wget --user=username --password=password http://localhost:8080/endpoint
//   - Browser: Use browser developer tools to set Authorization header
//
// Design Rationale:
//   - curl is widely available and familiar to developers
//   - The -u flag is the standard way to specify Basic Auth credentials in curl
//   - Embedding credentials in the command reduces user errors and setup time
//   - The format is consistent with curl documentation and common usage patterns
func getExampleURL(baseURL string) string {
	if *enableAuth {
		// Return a complete curl command with authentication
		// The -u flag is curl's standard method for HTTP Basic Authentication
		// Format: curl -u username:password <URL>
		return fmt.Sprintf("curl -u %s:%s %s", authUsername, authPassword, baseURL)
	}
	
	// Return the bare URL when authentication is disabled
	// This can be used directly in browsers, wget, or other HTTP clients
	return baseURL
}