package deepseekcompat

import (
	"encoding/json"
	"fmt"

	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/model"
)

// NormalizeLogger defines the logger interface used by DeepSeek normalization functions.
type NormalizeLogger interface {
	Debug(msg string, fields ...zap.Field)
}

// NormalizeToolMessageContent converts non-string tool message content into strings
// for DeepSeek compatibility. DeepSeek requires messages[].content for role=tool
// to be a string and rejects arrays/maps.
func NormalizeToolMessageContent(lg NormalizeLogger, request *model.GeneralOpenAIRequest) {
	if request == nil {
		return
	}

	toolMessageCount := 0
	normalizedCount := 0

	for idx := range request.Messages {
		message := &request.Messages[idx]
		if message.Role != "tool" {
			continue
		}

		toolMessageCount++
		if _, ok := message.Content.(string); ok {
			continue
		}

		normalized := message.StringContent()
		if normalized == "" {
			if message.Content == nil {
				normalized = ""
			} else {
				encoded, err := json.Marshal(message.Content)
				if err != nil {
					if lg != nil {
						lg.Debug("deepseek tool message fallback marshal failed",
							zap.Int("message_index", idx),
							zap.String("original_content_type", fmt.Sprintf("%T", message.Content)),
							zap.Error(err),
						)
					}
					normalized = fmt.Sprintf("%v", message.Content)
				} else {
					normalized = string(encoded)
				}
			}
		}

		message.Content = normalized
		normalizedCount++
		if lg != nil {
			lg.Debug("normalized deepseek tool message content",
				zap.Int("message_index", idx),
				zap.Int("normalized_content_length", len(normalized)),
			)
		}
	}

	if lg != nil && toolMessageCount > 0 {
		lg.Debug("deepseek tool message normalization summary",
			zap.Int("tool_message_count", toolMessageCount),
			zap.Int("normalized_count", normalizedCount),
		)
	}
}

// InjectMissingReasoningContent ensures all assistant messages have reasoning_content
// when thinking mode is active. DeepSeek rejects requests where any assistant message
// lacks reasoning_content with: "The reasoning_content in the thinking mode must be
// passed back to the API."
//
// This handles cases where external clients (e.g. Claude Code) do not replay
// reasoning_content from previous turns. It also converts reasoning (OpenRouter format)
// and thinking (Anthropic format from Claude message blocks) to reasoning_content.
func InjectMissingReasoningContent(c *gin.Context, request *model.GeneralOpenAIRequest) {
	lg := gmw.GetLogger(c)
	injectedCount := 0

	for i := range request.Messages {
		msg := &request.Messages[i]
		if msg.Role != "assistant" {
			continue
		}

		// Already has reasoning_content — nothing to do
		if msg.ReasoningContent != nil {
			continue
		}

		// reasoning (OpenRouter format) → reasoning_content
		if msg.Reasoning != nil {
			msg.ReasoningContent = msg.Reasoning
			msg.Reasoning = nil
			msg.Thinking = nil
			injectedCount++
			lg.Debug("converted reasoning → reasoning_content for deepseek",
				zap.Int("message_index", i),
			)
			continue
		}

		// thinking (Anthropic format, from Claude message blocks) → reasoning_content
		if msg.Thinking != nil {
			msg.ReasoningContent = msg.Thinking
			msg.Thinking = nil
			msg.Reasoning = nil
			injectedCount++
			lg.Debug("converted thinking → reasoning_content for deepseek",
				zap.Int("message_index", i),
			)
			continue
		}

		// Thinking mode active but no reasoning content at all — inject empty string
		empty := ""
		msg.ReasoningContent = &empty
		injectedCount++
		lg.Debug("injected empty reasoning_content for deepseek assistant message (thinking mode active)",
			zap.Int("message_index", i),
		)
	}

	if injectedCount > 0 {
		lg.Debug("normalized reasoning fields for deepseek compatibility",
			zap.Int("injected_count", injectedCount),
		)
	}
}
