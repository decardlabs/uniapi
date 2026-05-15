package model

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/common"
)

// TestStickySessionChannelMemoryFallback verifies sticky session channel bindings work when Redis is disabled.
func TestStickySessionChannelMemoryFallback(t *testing.T) {
	ctx := context.Background()
	originalRedisEnabled := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	defer common.SetRedisEnabled(originalRedisEnabled)

	userID := 101
	modelName := "gpt-4o-mini"
	channelID := 88

	require.NoError(t, SetStickySessionChannel(ctx, userID, modelName, channelID))

	actualChannelID, found, err := GetStickySessionChannel(ctx, userID, modelName)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, channelID, actualChannelID)

	require.NoError(t, DeleteStickySessionChannel(ctx, userID, modelName))
	_, foundAfterDelete, err := GetStickySessionChannel(ctx, userID, modelName)
	require.NoError(t, err)
	require.False(t, foundAfterDelete)
}

// TestResponseBoundChannelMemoryFallback verifies response_id channel bindings work when Redis is disabled.
func TestResponseBoundChannelMemoryFallback(t *testing.T) {
	ctx := context.Background()
	originalRedisEnabled := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	defer common.SetRedisEnabled(originalRedisEnabled)

	userID := 202
	responseID := "resp_abc123"
	channelID := 99

	require.NoError(t, SetResponseBoundChannel(ctx, userID, responseID, channelID))

	actualChannelID, found, err := GetResponseBoundChannel(ctx, userID, responseID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, channelID, actualChannelID)
}

// TestStickySessionChannelIsolation verifies sticky bindings are isolated by user+model.
func TestStickySessionChannelIsolation(t *testing.T) {
	ctx := context.Background()
	originalRedisEnabled := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	defer common.SetRedisEnabled(originalRedisEnabled)

	// Same model, different users should not overwrite each other.
	require.NoError(t, SetStickySessionChannel(ctx, 1001, "MiniMax-M2.7", 1))
	require.NoError(t, SetStickySessionChannel(ctx, 1002, "MiniMax-M2.7", 2))

	ch1, found1, err := GetStickySessionChannel(ctx, 1001, "MiniMax-M2.7")
	require.NoError(t, err)
	require.True(t, found1)
	require.Equal(t, 1, ch1)

	ch2, found2, err := GetStickySessionChannel(ctx, 1002, "MiniMax-M2.7")
	require.NoError(t, err)
	require.True(t, found2)
	require.Equal(t, 2, ch2)

	// Same user, different models should not overwrite each other.
	require.NoError(t, SetStickySessionChannel(ctx, 1001, "gpt-4o", 3))
	require.NoError(t, SetStickySessionChannel(ctx, 1001, "claude-3-5", 4))

	chModelA, foundModelA, err := GetStickySessionChannel(ctx, 1001, "gpt-4o")
	require.NoError(t, err)
	require.True(t, foundModelA)
	require.Equal(t, 3, chModelA)

	chModelB, foundModelB, err := GetStickySessionChannel(ctx, 1001, "claude-3-5")
	require.NoError(t, err)
	require.True(t, foundModelB)
	require.Equal(t, 4, chModelB)
}

// TestStickySessionChannelModelNormalization verifies model key normalization is case-insensitive.
func TestStickySessionChannelModelNormalization(t *testing.T) {
	ctx := context.Background()
	originalRedisEnabled := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	defer common.SetRedisEnabled(originalRedisEnabled)

	require.NoError(t, SetStickySessionChannel(ctx, 2001, "  MiniMax-M2.7  ", 7))

	ch, found, err := GetStickySessionChannel(ctx, 2001, "minimax-m2.7")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 7, ch)
}
