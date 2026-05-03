package openai

import (
	"strings"

	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor/common/deepseekcompat"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// shouldNormalizeToolMessageContentForDeepSeek reports whether tool message content should
// be normalized to string for DeepSeek-compatible upstreams.
func shouldNormalizeToolMessageContentForDeepSeek(metaInfo *meta.Meta, request *model.GeneralOpenAIRequest) bool {
	if metaInfo != nil {
		if metaInfo.ChannelType == channeltype.DeepSeek {
			return true
		}
		if strings.Contains(strings.ToLower(strings.TrimSpace(metaInfo.BaseURL)), "deepseek") {
			return true
		}
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(metaInfo.ActualModelName)), "deepseek-") {
			return true
		}
	}

	if request != nil {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(request.Model)), "deepseek-") {
			return true
		}
	}

	return false
}

// normalizeClaudeThinkingForDeepSeek coerces Claude thinking payloads into DeepSeek-compatible values.
// DeepSeek currently accepts only enabled or disabled for thinking.type.
func normalizeClaudeThinkingForDeepSeek(c *gin.Context, request *model.GeneralOpenAIRequest) {
	if request == nil || request.Thinking == nil {
		return
	}

	originalType := request.Thinking.Type
	normalizedType, changed := deepseekcompat.NormalizeThinkingType(originalType, request.Thinking.BudgetTokens)
	if !changed {
		return
	}

	request.Thinking.Type = normalizedType
}

// isDeepSeekClaudeThinkingEnabled checks whether thinking mode should be treated as active
// for DeepSeek in the Claude Messages conversion path.
// DeepSeek V4 enables thinking mode by default; we only skip injection when thinking
// is explicitly disabled.
func isDeepSeekClaudeThinkingEnabled(c *gin.Context, request *model.GeneralOpenAIRequest) bool {
	// If the original ClaudeRequest has thinking set to "disabled", skip injection.
	if raw, exists := c.Get(ctxkey.OriginalClaudeRequest); exists {
		if originalReq, ok := raw.(*model.ClaudeRequest); ok && originalReq.Thinking != nil {
			if originalReq.Thinking.Type == "disabled" {
				return false
			}
		}
	}

	// If the OpenAI request has thinking set to "disabled", skip injection.
	if request.Thinking != nil && request.Thinking.Type == "disabled" {
		return false
	}

	return true
}

// deepSeekThinkingNormalizeLogger defines the logger interface used by DeepSeek thinking normalization.
type deepSeekThinkingNormalizeLogger interface {
	Debug(msg string, fields ...zap.Field)
}

// deepSeekToolNormalizeLogger defines the logger interface used by DeepSeek tool normalization.
type deepSeekToolNormalizeLogger interface {
	Debug(msg string, fields ...zap.Field)
}
