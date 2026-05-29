package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseChannelIDs verifies CSV parsing for conservative vision backfill flags.
func TestParseChannelIDs(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		ids, err := parseChannelIDs("")
		require.NoError(t, err)
		require.Nil(t, ids)
	})

	t.Run("valid ids", func(t *testing.T) {
		ids, err := parseChannelIDs("101, 205,205")
		require.NoError(t, err)
		require.Len(t, ids, 2)
		_, ok101 := ids[101]
		_, ok205 := ids[205]
		require.True(t, ok101)
		require.True(t, ok205)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, err := parseChannelIDs("abc")
		require.Error(t, err)
	})

	t.Run("non-positive id", func(t *testing.T) {
		_, err := parseChannelIDs("0")
		require.Error(t, err)
	})

	t.Run("empty item", func(t *testing.T) {
		_, err := parseChannelIDs("101,,205")
		require.Error(t, err)
	})
}
