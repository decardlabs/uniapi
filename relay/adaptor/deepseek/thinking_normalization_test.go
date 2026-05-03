package deepseek

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/songquanpeng/one-api/common/ctxkey"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// TestConvertRequest_NormalizesAdaptiveThinkingType verifies DeepSeek adaptor converts
// unsupported thinking.type values into DeepSeek-compatible enums.
func TestConvertRequest_NormalizesAdaptiveThinkingType(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	request := &relaymodel.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []relaymodel.Message{
			{Role: "user", Content: "hello"},
		},
		Thinking: &relaymodel.Thinking{Type: "adaptive", BudgetTokens: relaymodel.IntPtr(2048)},
	}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = &http.Request{}
	context.Set(ctxkey.OriginalClaudeRequest, &relaymodel.ClaudeRequest{Thinking: &relaymodel.Thinking{Type: "adaptive", BudgetTokens: relaymodel.IntPtr(2048)}})

	adaptor := &Adaptor{}
	convertedAny, err := adaptor.ConvertRequest(context, relaymode.ChatCompletions, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*relaymodel.GeneralOpenAIRequest)
	require.True(t, ok)
	require.NotNil(t, converted.Thinking)
	require.Equal(t, "enabled", converted.Thinking.Type)
	require.Equal(t, 2048, *converted.Thinking.BudgetTokens)
}

// TestConvertRequest_PreservesSupportedThinkingType verifies already-supported values are unchanged.
func TestConvertRequest_PreservesSupportedThinkingType(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	request := &relaymodel.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []relaymodel.Message{
			{Role: "user", Content: "hello"},
		},
		Thinking: &relaymodel.Thinking{Type: "enabled", BudgetTokens: relaymodel.IntPtr(1024)},
	}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = &http.Request{}
	context.Set(ctxkey.OriginalClaudeRequest, &relaymodel.ClaudeRequest{Thinking: &relaymodel.Thinking{Type: "enabled", BudgetTokens: relaymodel.IntPtr(1024)}})

	adaptor := &Adaptor{}
	convertedAny, err := adaptor.ConvertRequest(context, relaymode.ChatCompletions, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*relaymodel.GeneralOpenAIRequest)
	require.True(t, ok)
	// Already-supported type does not require normalization, so Thinking remains unset.
	require.Nil(t, converted.Thinking)
}
