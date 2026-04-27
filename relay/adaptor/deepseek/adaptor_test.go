package deepseek

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestConvertRequest_NormalizesToolArrayContentToString(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "user", Content: "hello"},
			{
				Role:       "tool",
				ToolCallId: "call_1",
				Content: []any{
					map[string]any{"type": "text", "text": "README.md\n"},
				},
			},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Equal(t, "README.md\n", converted.Messages[1].Content)
}

func TestConvertRequest_NormalizesToolMapContentByJSONFallback(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{
				Role:       "tool",
				ToolCallId: "call_2",
				Content: map[string]any{
					"stdout":    "ok",
					"exit_code": 0,
				},
			},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)

	contentStr, ok := converted.Messages[0].Content.(string)
	require.True(t, ok)
	require.Contains(t, contentStr, `"stdout":"ok"`)
	require.Contains(t, contentStr, `"exit_code":0`)
}

func TestConvertRequest_NormalizesNilToolContentToEmptyString(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "tool", ToolCallId: "call_3", Content: nil},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Equal(t, "", converted.Messages[0].Content)
}

func TestConvertRequest_DoesNotChangeNonToolArrayContent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	originalContent := []any{map[string]any{"type": "text", "text": "hello"}}
	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "user", Content: originalContent},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Equal(t, originalContent, converted.Messages[0].Content)
}

func TestConvertRequest_ConvertsReasoningToReasoningContent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	reasoningText := "Let me think about this step by step..."
	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "user", Content: "hello"},
			{
				Role:      "assistant",
				Content:   "The answer is 42.",
				Reasoning: &reasoningText,
			},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	// reasoning should be converted to reasoning_content
	require.Nil(t, converted.Messages[1].Reasoning)
	require.NotNil(t, converted.Messages[1].ReasoningContent)
	require.Equal(t, reasoningText, *converted.Messages[1].ReasoningContent)
}

func TestConvertRequest_ConvertsThinkingToReasoningContent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	thinkingText := "I need to analyze this carefully..."
	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "user", Content: "hello"},
			{
				Role:     "assistant",
				Content:  "Here is my analysis.",
				Thinking: &thinkingText,
			},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	// thinking should be converted to reasoning_content
	require.Nil(t, converted.Messages[1].Thinking)
	require.NotNil(t, converted.Messages[1].ReasoningContent)
	require.Equal(t, thinkingText, *converted.Messages[1].ReasoningContent)
}

func TestConvertRequest_PreservesExistingReasoningContent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	reasoningContent := "DeepSeek reasoning..."
	reasoningText := "OpenRouter reasoning..."
	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{
				Role:             "assistant",
				Content:          "Answer",
				ReasoningContent: &reasoningContent,
				Reasoning:        &reasoningText,
			},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	// reasoning_content already present, should NOT be overwritten
	require.Equal(t, reasoningContent, *converted.Messages[0].ReasoningContent)
}

func TestConvertRequest_DoesNotConvertNonAssistantReasoning(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	reasoningText := "Should not be converted"
	adaptor := &Adaptor{}
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "user", Content: "hello", Reasoning: &reasoningText},
		},
	}

	convertedAny, err := adaptor.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	// non-assistant messages should not be touched
	require.NotNil(t, converted.Messages[0].Reasoning)
	require.Nil(t, converted.Messages[0].ReasoningContent)
}
