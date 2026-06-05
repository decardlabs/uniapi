package channeltype

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/relay/apitype"
)

func TestToAPITypeGLMUsesZhipuAdaptor(t *testing.T) {
	t.Parallel()

	require.Equal(t, apitype.Zhipu, ToAPIType(GLM))
}

func TestIdToNameGLM(t *testing.T) {
	t.Parallel()

	require.Equal(t, "glm", IdToName(GLM))
}
