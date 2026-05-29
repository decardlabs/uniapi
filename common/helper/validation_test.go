package helper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBuildTextOnlyModelImageInputValidationMessageIncludesTypes verifies the
// error message includes deduplicated image content types and routing guidance.
func TestBuildTextOnlyModelImageInputValidationMessageIncludesTypes(t *testing.T) {
	t.Parallel()

	msg := BuildTextOnlyModelImageInputValidationMessage(
		"deepseek-v4-pro",
		[]string{"input_image", "image_url", "image_url"},
		"no vision-capable model is currently available",
	)

	require.Contains(t, msg, "deepseek-v4-pro")
	require.Contains(t, msg, "content types: image_url,input_image")
	require.Contains(t, msg, "no vision-capable model is currently available")
	require.Contains(t, msg, "Remove image input and retry with text")
}

// TestBuildTextOnlyModelImageInputValidationMessageWithoutTypes verifies
// the fallback summary stays readable even when content types are unavailable.
func TestBuildTextOnlyModelImageInputValidationMessageWithoutTypes(t *testing.T) {
	t.Parallel()

	msg := BuildTextOnlyModelImageInputValidationMessage("deepseek-v4-pro", nil, "")

	require.Contains(t, msg, "content types: image")
	require.Contains(t, msg, "switch to a vision-capable model")
}