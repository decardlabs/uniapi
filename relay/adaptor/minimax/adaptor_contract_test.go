package minimax

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/relay/model"
)

// TestAdaptorConvertRequest_BackfillsToolMessageNameFromToolCalls verifies that ConvertRequest
// backfills role=tool message names from previous assistant tool_calls for MiniMax compatibility.
func TestAdaptorConvertRequest_BackfillsToolMessageNameFromToolCalls(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(nil)

	req := &model.GeneralOpenAIRequest{
		Messages: []model.Message{
			{
				Role: "assistant",
				ToolCalls: []model.Tool{{
					Id:   "call_1",
					Type: "function",
					Function: &model.Function{
						Name: "get_weather",
					},
				}},
			},
			{
				Role:       "tool",
				ToolCallId: "call_1",
				Content:    "{\"temperature\":26}",
			},
		},
	}

	adaptor := &Adaptor{}
	convertedAny, err := adaptor.ConvertRequest(ctx, 0, req)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Len(t, converted.Messages, 2)
	require.NotNil(t, converted.Messages[1].Name)
	require.Equal(t, "get_weather", *converted.Messages[1].Name)
}

// TestAdaptorConvertRequest_PreservesExistingToolMessageName verifies that ConvertRequest
// keeps an existing role=tool message name instead of overwriting it.
func TestAdaptorConvertRequest_PreservesExistingToolMessageName(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(nil)

	existingName := "custom_name"
	req := &model.GeneralOpenAIRequest{
		Messages: []model.Message{
			{
				Role: "assistant",
				ToolCalls: []model.Tool{{
					Id:   "call_2",
					Type: "function",
					Function: &model.Function{
						Name: "get_weather",
					},
				}},
			},
			{
				Role:       "tool",
				ToolCallId: "call_2",
				Name:       &existingName,
				Content:    "ok",
			},
		},
	}

	adaptor := &Adaptor{}
	convertedAny, err := adaptor.ConvertRequest(ctx, 0, req)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Len(t, converted.Messages, 2)
	require.NotNil(t, converted.Messages[1].Name)
	require.Equal(t, "custom_name", *converted.Messages[1].Name)
}

// TestAdaptorConvertRequest_LeavesToolMessageNameNilWithoutMatch verifies that ConvertRequest
// leaves role=tool message names empty when there is no matching tool_call_id mapping.
func TestAdaptorConvertRequest_LeavesToolMessageNameNilWithoutMatch(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(nil)

	req := &model.GeneralOpenAIRequest{
		Messages: []model.Message{
			{
				Role: "assistant",
				ToolCalls: []model.Tool{{
					Id:   "call_3",
					Type: "function",
					Function: &model.Function{
						Name: "get_weather",
					},
				}},
			},
			{
				Role:       "tool",
				ToolCallId: "call_not_found",
				Content:    "ok",
			},
		},
	}

	adaptor := &Adaptor{}
	convertedAny, err := adaptor.ConvertRequest(ctx, 0, req)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Len(t, converted.Messages, 2)
	require.Nil(t, converted.Messages[1].Name)
}

// TestAdaptorConvertRequest_BackfillsToolMessageNameWithFCID verifies that ConvertRequest
// maps tool_call_id values between call_* and fc_* forms for MiniMax compatibility.
func TestAdaptorConvertRequest_BackfillsToolMessageNameWithFCID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(nil)

	req := &model.GeneralOpenAIRequest{
		Messages: []model.Message{
			{
				Role: "assistant",
				ToolCalls: []model.Tool{{
					Id:   "call_123",
					Type: "function",
					Function: &model.Function{
						Name: "get_weather",
					},
				}},
			},
			{
				Role:       "tool",
				ToolCallId: "fc_123",
				Content:    "ok",
			},
		},
	}

	adaptor := &Adaptor{}
	convertedAny, err := adaptor.ConvertRequest(ctx, 0, req)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Len(t, converted.Messages, 2)
	require.NotNil(t, converted.Messages[1].Name)
	require.Equal(t, "get_weather", *converted.Messages[1].Name)
}

// TestAdaptorConvertRequest_BackfillsToolMessageNameWithBareID verifies that ConvertRequest
// can match bare tool_call_id values against call_* assistant tool call IDs.
func TestAdaptorConvertRequest_BackfillsToolMessageNameWithBareID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(nil)

	req := &model.GeneralOpenAIRequest{
		Messages: []model.Message{
			{
				Role: "assistant",
				ToolCalls: []model.Tool{{
					Id:   "call_456",
					Type: "function",
					Function: &model.Function{
						Name: "get_time",
					},
				}},
			},
			{
				Role:       "tool",
				ToolCallId: "456",
				Content:    "ok",
			},
		},
	}

	adaptor := &Adaptor{}
	convertedAny, err := adaptor.ConvertRequest(ctx, 0, req)
	require.NoError(t, err)

	converted, ok := convertedAny.(*model.GeneralOpenAIRequest)
	require.True(t, ok)
	require.Len(t, converted.Messages, 2)
	require.NotNil(t, converted.Messages[1].Name)
	require.Equal(t, "get_time", *converted.Messages[1].Name)
}
