package doubao

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestModelListUsesOfficialArkIDs verifies that the exported public model list uses
// the official Ark model identifiers rather than legacy pricing aliases.
func TestModelListUsesOfficialArkIDs(t *testing.T) {
	t.Parallel()

	require.Contains(t, ModelList, "doubao-seed-2-0-pro-260215")
	require.Contains(t, ModelList, "doubao-seed-1-6-251015")
	require.Contains(t, ModelList, "doubao-embedding-vision-251215")

	require.NotContains(t, ModelList, "doubao-seed-2.0-pro")
	require.NotContains(t, ModelList, "doubao-seed-1.6")
	require.NotContains(t, ModelList, "Doubao-embedding")

	adaptor := &Adaptor{}
	require.Equal(t, ModelList, adaptor.GetModelList())
}
