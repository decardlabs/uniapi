package deepseek

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/relay/model"
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
	// DeepSeek V4 enables thinking by default, so reasoning is converted to reasoning_content.
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
	// DeepSeek V4 enables thinking by default, so thinking is converted to reasoning_content.
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

// TestConvertClaudeRequest_StripsImageURLParts 验证从 Claude Messages 转换后
// image_url 内容块被剥离，不发送给 DeepSeek（否则会收到 400 错误）。
// 场景：Claude Code 读取 PDF 后，历史消息中含有 PDF 页面的 base64 图片。
func TestConvertClaudeRequest_StripsImageURLParts(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	a := &Adaptor{}

	// 构造含有 image_url content part 的 GeneralOpenAIRequest（模拟转换后的状态）
	imageURL := "data:image/png;base64,iVBORw0KGgo="
	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-v4-flash",
		Messages: []model.Message{
			{Role: "user", Content: "请分析这个文档"},
			{
				Role: "user",
				Content: []model.MessageContent{
					{Type: model.ContentTypeText, Text: strPtr("工具返回结果：")},
					{Type: model.ContentTypeImageURL, ImageURL: &model.ImageURL{Url: imageURL}},
					{Type: model.ContentTypeImageURL, ImageURL: &model.ImageURL{Url: imageURL}},
				},
			},
		},
	}

	convertedAny, err := a.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)

	// 第一条消息（纯字符串）应不受影响
	require.Equal(t, "请分析这个文档", converted.Messages[0].Content)

	// 第二条消息的 image_url 应被剥离，只剩文本部分
	parts, ok := converted.Messages[1].Content.([]model.MessageContent)
	require.True(t, ok, "应保留 MessageContent 切片")
	require.Len(t, parts, 1, "image_url 部分应被移除，只保留文本")
	require.Equal(t, model.ContentTypeText, parts[0].Type)
	require.Equal(t, "工具返回结果：", *parts[0].Text)
}

// TestConvertClaudeRequest_AllImagePartsReplacedWithPlaceholder 验证当消息
// 全部内容都是图片时，被替换为占位文本而非空消息。
func TestConvertClaudeRequest_AllImagePartsReplacedWithPlaceholder(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	writer := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(writer)

	a := &Adaptor{}

	request := &model.GeneralOpenAIRequest{
		Model: "deepseek-v4-flash",
		Messages: []model.Message{
			{
				Role: "user",
				Content: []model.MessageContent{
					{Type: model.ContentTypeImageURL, ImageURL: &model.ImageURL{Url: "data:image/png;base64,abc"}},
				},
			},
		},
	}

	convertedAny, err := a.ConvertRequest(c, 0, request)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)

	// 消息不应变为空，应有占位文本
	contentStr, ok := converted.Messages[0].Content.(string)
	require.True(t, ok, "全图片消息应被替换为占位字符串")
	require.Contains(t, contentStr, "图片内容已省略")
}

func strPtr(s string) *string { return &s }
