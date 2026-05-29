package client

import (
	"net/http"
	"testing"
	"time"

	"github.com/decardlabs/uniapi/common/config"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	// Test that Init() creates properly configured HTTP clients
	Init()

	// Verify UserContentRequestHTTPClient is created
	require.NotNil(t, UserContentRequestHTTPClient)
	require.NotNil(t, UserContentRequestHTTPClient.Transport)

	// Verify it has a timeout set
	require.Greater(t, UserContentRequestHTTPClient.Timeout.Seconds(), 0.0)

	// Verify HTTP/2 is disabled (TLSNextProto should be empty map)
	if transport, ok := UserContentRequestHTTPClient.Transport.(*http.Transport); ok {
		require.NotNil(t, transport.TLSNextProto)
		require.Empty(t, transport.TLSNextProto)
	}

	// Verify other clients are created
	require.NotNil(t, HTTPClient)
	require.NotNil(t, ImpatientHTTPClient)
}

func TestInit_UserContentTimeoutNoProxyUsesConfig(t *testing.T) {
	t.Helper()

	originalProxy := config.UserContentRequestProxy
	originalTimeout := config.UserContentRequestTimeout
	t.Cleanup(func() {
		config.UserContentRequestProxy = originalProxy
		config.UserContentRequestTimeout = originalTimeout
	})

	config.UserContentRequestProxy = ""
	config.UserContentRequestTimeout = 7

	Init()

	require.NotNil(t, UserContentRequestHTTPClient)
	require.Equal(t, 7*time.Second, UserContentRequestHTTPClient.Timeout)
}
