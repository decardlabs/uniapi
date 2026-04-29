package secure

import "crypto/subtle"

// ConstantTimeEqual compares two strings using constant-time comparison to prevent timing attacks.
// Returns true if the strings are equal, false otherwise.
// Use this for comparing sensitive values like tokens, API keys, and signatures.
func ConstantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
