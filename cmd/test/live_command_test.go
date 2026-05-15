package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestParseLiveArgsDefaults verifies default values and model fallback behavior.
func TestParseLiveArgsDefaults(t *testing.T) {
	t.Parallel()

	cfg := config{Models: []string{"gpt-4o-mini"}}
	opts, err := parseLiveArgs(nil, cfg)
	require.NoError(t, err)
	require.Equal(t, "gpt-4o-mini", opts.model)
	require.Equal(t, 3, opts.rounds)
	require.Equal(t, 1, opts.concurrency)
	require.Equal(t, 90*time.Second, opts.timeout)
}

// TestParseLiveArgsValidate verifies validation for invalid live probe options.
func TestParseLiveArgsValidate(t *testing.T) {
	t.Parallel()

	_, err := parseLiveArgs([]string{"--model", "", "--rounds", "0"}, config{})
	require.Error(t, err)

	_, err = parseLiveArgs([]string{"--model", "gpt-4o-mini", "--concurrency", "0"}, config{})
	require.Error(t, err)

	_, err = parseLiveArgs([]string{"--model", "gpt-4o-mini", "--timeout", "0s"}, config{})
	require.Error(t, err)
}

// TestExtractResponseID verifies id extraction from Response API payloads.
func TestExtractResponseID(t *testing.T) {
	t.Parallel()

	id := extractResponseID(`{"id":"resp_123","object":"response"}`)
	require.Equal(t, "resp_123", id)

	id = extractResponseID(`{"object":"response"}`)
	require.Empty(t, id)
}

// TestLiveSupportsToolInvocation verifies that DeepSeek models skip the forced tool-calling probe.
func TestLiveSupportsToolInvocation(t *testing.T) {
	t.Parallel()

	require.False(t, liveSupportsToolInvocation("deepseek-v4-pro"))
	require.False(t, liveSupportsToolInvocation("DeepSeek-Reasoner"))
	require.True(t, liveSupportsToolInvocation("gpt-4o-mini"))
}
