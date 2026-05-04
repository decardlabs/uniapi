package alibailian

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestModelListMatchesCompatibleDocs verifies that the exported public model list
// only contains models explicitly listed in Alibaba's OpenAI-compatible docs.
func TestModelListMatchesCompatibleDocs(t *testing.T) {
	t.Parallel()

	require.Contains(t, ModelList, "qwen3-max")
	require.Contains(t, ModelList, "qwen3.6-plus")
	require.Contains(t, ModelList, "qwen3.5-flash-2026-02-23")
	require.Contains(t, ModelList, "qwen3-next-80b-a3b-thinking")

	require.NotContains(t, ModelList, "qwen-turbo")
	require.NotContains(t, ModelList, "qwen-long")
	require.NotContains(t, ModelList, "qwen-coder-plus")
	require.NotContains(t, ModelList, "qwq-32b-preview")
}
