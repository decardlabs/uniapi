package common

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
	"github.com/stretchr/testify/require"
)

func TestPassword2Hash_UsesSufficientCost(t *testing.T) {
	password := "TestPassword123!"
	hash, err := Password2Hash(password)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// Verify the hash uses cost >= 12
	cost, err := bcrypt.Cost([]byte(hash))
	require.NoError(t, err)
	require.GreaterOrEqual(t, cost, 12, "Password hashing cost must be >= 12 per OWASP/AGENTS.md")
}

func TestPassword2Hash_VerifyWorks(t *testing.T) {
	password := "TestPassword123!"
	hash, err := Password2Hash(password)
	require.NoError(t, err)

	// Should validate correctly
	valid := ValidatePasswordAndHash(password, hash)
	require.True(t, valid)

	// Should reject wrong password
	valid = ValidatePasswordAndHash("WrongPassword", hash)
	require.False(t, valid)
}

func TestPassword2Hash_CostConsistency(t *testing.T) {
	// Multiple hashes should all use the same cost
	password := "TestPassword123!"
	hash1, err1 := Password2Hash(password)
	hash2, err2 := Password2Hash(password)

	require.NoError(t, err1)
	require.NoError(t, err2)

	cost1, _ := bcrypt.Cost([]byte(hash1))
	cost2, _ := bcrypt.Cost([]byte(hash2))

	require.Equal(t, cost1, cost2, "Cost should be consistent across hashes")
	require.GreaterOrEqual(t, cost1, 12)
}
