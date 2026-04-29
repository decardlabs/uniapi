package secure

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstantTimeEqual_ReturnsTrueForMatch(t *testing.T) {
	a := "secret-token-123"
	b := "secret-token-123"
	result := ConstantTimeEqual(a, b)
	require.True(t, result, "Should return true for matching strings")
}

func TestConstantTimeEqual_ReturnsFalseForMismatch(t *testing.T) {
	a := "secret-token-123"
	b := "secret-token-456"
	result := ConstantTimeEqual(a, b)
	require.False(t, result, "Should return false for mismatched strings")
}

func TestConstantTimeEqual_ReturnsFalseForDifferentLength(t *testing.T) {
	a := "short"
	b := "longer-string"
	result := ConstantTimeEqual(a, b)
	require.False(t, result, "Should return false for strings of different lengths")
}

func TestConstantTimeEqual_EmptyStrings(t *testing.T) {
	require.True(t, ConstantTimeEqual("", ""), "Empty strings should match")
	require.False(t, ConstantTimeEqual("", "non-empty"), "Empty vs non-empty should not match")
	require.False(t, ConstantTimeEqual("non-empty", ""), "Non-empty vs empty should not match")
}

func TestConstantTimeEqual_RealWorldToken(t *testing.T) {
	// Simulate token comparison
	token1 := "sk-1234567890abcdef"
	token2 := "sk-1234567890abcdef"
	token3 := "sk-0987654321fedcba"

	require.True(t, ConstantTimeEqual(token1, token2), "Same tokens should match")
	require.False(t, ConstantTimeEqual(token1, token3), "Different tokens should not match")
}
