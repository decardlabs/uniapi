package model

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/common"
)

func TestPackageSQLitePathIsIsolated(t *testing.T) {
	t.Helper()
	require.NotEqual(t, "uniapi.db", filepath.Clean(common.SQLitePath))
}