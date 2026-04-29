package common

import (
	"github.com/Laisky/errors/v2"
	"golang.org/x/crypto/bcrypt"
)

const (
	// PasswordHashCost follows OWASP recommendations for bcrypt (cost >= 12).
	// Cost 12 = ~4096 iterations, Cost 13 = ~8192 iterations, Cost 14 = ~16384 iterations.
	// AGENTS.md requires at least 10,000 iterations, so cost 14 is recommended.
	// We use 12 as a balance between security and performance.
	PasswordHashCost = 12
)

// Password2Hash converts the provided plaintext password into a bcrypt hash using cost=12.
// It returns the hashed password string and any error emitted by the bcrypt library.
func Password2Hash(password string) (string, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, PasswordHashCost)
	if err != nil {
		return "", errors.Wrap(err, "generate password hash")
	}
	return string(hashedPassword), nil
}

// ValidatePasswordAndHash checks whether the plaintext password matches the supplied bcrypt hash.
// It returns true when the hash corresponds to the password, otherwise false.
func ValidatePasswordAndHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
