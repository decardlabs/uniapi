package openai

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/common/ctxkey"
	"github.com/decardlabs/uniapi/relay/channeltype"
	"github.com/decardlabs/uniapi/relay/meta"
	"github.com/decardlabs/uniapi/relay/relaymode"
)

func TestRecordWebSearchPreviewInvocationNonReasoning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	metaInfo := &meta.Meta{
		ChannelType:     channeltype.OpenAI,
		Mode:            relaymode.ChatCompletions,
		ActualModelName: "gpt-4o-mini-search-preview",
	}

	recordWebSearchPreviewInvocation(c, metaInfo)

	rawCounts, exists := c.Get(ctxkey.ToolInvocationCounts)
	require.True(t, exists)
	counts := rawCounts.(map[string]int)
	require.Equal(t, 1, counts["web_search_preview_non_reasoning"])
}

func TestRecordWebSearchPreviewInvocationSkipsWhenCountsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	metaInfo := &meta.Meta{
		ChannelType:     channeltype.OpenAI,
		Mode:            relaymode.ChatCompletions,
		ActualModelName: "gpt-4o-mini-search-preview",
	}

	c.Set(ctxkey.WebSearchCallCount, 2)

	recordWebSearchPreviewInvocation(c, metaInfo)

	_, exists := c.Get(ctxkey.ToolInvocationCounts)
	require.False(t, exists, "should not record implicit count when upstream already provided one")
}

func TestRecordWebSearchPreviewInvocationReasoningTier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	metaInfo := &meta.Meta{
		ChannelType:     channeltype.OpenAI,
		Mode:            relaymode.ChatCompletions,
		ActualModelName: "gpt-5-search-preview",
	}

	recordWebSearchPreviewInvocation(c, metaInfo)

	rawCounts, exists := c.Get(ctxkey.ToolInvocationCounts)
	require.True(t, exists)
	counts := rawCounts.(map[string]int)
	require.Equal(t, 1, counts["web_search_preview_reasoning"])
}
