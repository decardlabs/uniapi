package controller

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/common"
)

// TestMain redirects package-level SQLite usage to a temp directory so controller
// tests never create repository-local uniapi.db files.
func TestMain(m *testing.M) {
	originalSQLitePath := common.SQLitePath
	tempDir, err := os.MkdirTemp("", "uniapi-controller-test-*")
	if err != nil {
		panic(err)
	}

	common.SQLitePath = filepath.Join(tempDir, "controller-test.db")
	code := m.Run()
	common.SQLitePath = originalSQLitePath
	_ = os.RemoveAll(tempDir)
	os.Exit(code)
}

// TestPackageSQLitePathIsIsolated verifies controller tests do not fall back to
// the default relative SQLite path.
func TestPackageSQLitePathIsIsolated(t *testing.T) {
	t.Helper()
	require.NotEqual(t, "uniapi.db", filepath.Clean(common.SQLitePath))
}