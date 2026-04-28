package deepseek

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/common/deepseekcompat"
	"github.com/songquanpeng/one-api/relay/adaptor/common/structuredjson"
	"github.com/songquanpeng/one-api/relay/adaptor/openai_compatible"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
	adaptor.DefaultPricingMethods
}

func (a *Adaptor) GetChannelName() string {
	return "deepseek"
}

func (a *Adaptor) GetModelList() []string {
	return adaptor.GetModelListFromPricing(ModelRatios)
}

// GetDefaultModelPricing returns the pricing information for DeepSeek models
// Based on official DeepSeek pricing: https://platform.deepseek.com/api-docs/pricing/
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig {
	return ModelRatios
}

// DefaultToolingConfig returns DeepSeek's provider-level tooling defaults (none published as of 2025-11-12).
func (a *Adaptor) DefaultToolingConfig() adaptor.ChannelToolConfig {
	return DeepseekToolingDefaults
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetModelRatio(modelName)
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetCompletionRatio(modelName)
}

// Implement required adaptor interface methods (DeepSeek uses OpenAI-compatible API)
func (a *Adaptor) Init(meta *meta.Meta) {}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// Handle Claude Messages requests - convert to OpenAI Chat Completions endpoint
	requestPath := meta.RequestURLPath
	if idx := strings.Index(requestPath, "?"); idx >= 0 {
		requestPath = requestPath[:idx]
	}
	if requestPath == "/v1/messages" {
		// Claude Messages requests should use OpenAI's chat completions endpoint
		chatCompletionsPath := "/v1/chat/completions"
		return openai_compatible.GetFullRequestURL(meta.BaseURL, chatCompletionsPath, meta.ChannelType), nil
	}

	// DeepSeek uses OpenAI-compatible API endpoints
	return openai_compatible.GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	// DeepSeek is OpenAI-compatible, so we can pass the request through with minimal changes
	// Remove reasoning_effort as DeepSeek doesn't support it
	if request.ReasoningEffort != nil {
		request.ReasoningEffort = nil
	}

	// Remove top-level Thinking field (not supported by DeepSeek's OpenAI-compatible API)
	// after extracting whether thinking mode is enabled.
	thinkingEnabled := request.Thinking != nil
	request.Thinking = nil

	normalizeDeepSeekThinkingConfigFromOriginal(c, request)

	normalizeDeepSeekToolMessageContent(c, request)

	// DeepSeek requires reasoning_content on all assistant messages when thinking mode is active.
	// Claude Code does not replay reasoning_content from previous turns, so we inject empty
	// values to avoid "The reasoning_content in the thinking mode must be passed back to the API."
	if thinkingEnabled {
		injectMissingReasoningContent(c, request)
	}

	if request.ResponseFormat != nil {
		if request.ResponseFormat.JsonSchema != nil {
			structuredjson.EnsureInstruction(request)
		}
		request.ResponseFormat = nil
	}

	return request, nil
}

// normalizeDeepSeekThinkingConfigFromOriginal reads the original ClaudeRequest from context
// to normalize thinking.type for DeepSeek. The ClaudeRequest.Thinking field is not
// propagated to GeneralOpenAIRequest (to avoid sending it to providers that don't support it),
// so we read it from the original request stored in context.
func normalizeDeepSeekThinkingConfigFromOriginal(c *gin.Context, request *model.GeneralOpenAIRequest) {
	if request == nil {
		return
	}

	// Try to get the original ClaudeRequest from context
	var claudeThinking *model.Thinking
	if raw, exists := c.Get(ctxkey.OriginalClaudeRequest); exists {
		if originalReq, ok := raw.(*model.ClaudeRequest); ok && originalReq.Thinking != nil {
			claudeThinking = originalReq.Thinking
		}
	}

	if claudeThinking == nil {
		return
	}

	normalizedType, changed := deepseekcompat.NormalizeThinkingType(claudeThinking.Type, claudeThinking.BudgetTokens)
	if !changed {
		return
	}

	// Store the normalized thinking config back in the OpenAI request so DeepSeek
	// can receive it as a top-level parameter it understands.
	request.Thinking = &model.Thinking{
		Type:         normalizedType,
		BudgetTokens: claudeThinking.BudgetTokens,
	}
	gmw.GetLogger(c).Debug("normalized deepseek thinking type from original ClaudeRequest",
		zap.String("model", request.Model),
		zap.String("original_type", claudeThinking.Type),
		zap.String("normalized_type", normalizedType),
		zap.Intp("budget_tokens", claudeThinking.BudgetTokens),
	)
}

// injectMissingReasoningContent ensures all assistant messages have reasoning_content when
// thinking mode is active. DeepSeek rejects requests where any assistant message lacks
// reasoning_content with: "The reasoning_content in the thinking mode must be passed back to the API."
//
// This handles cases where external clients (e.g. Claude Code) do not replay reasoning_content
// from previous turns. It also converts reasoning (OpenRouter format) and thinking (Anthropic
// format from Claude message blocks) to reasoning_content.
func injectMissingReasoningContent(c *gin.Context, request *model.GeneralOpenAIRequest) {
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

// normalizeDeepSeekToolMessageContent converts non-string tool message content into strings for DeepSeek compatibility.
// DeepSeek requires `messages[].content` for role=tool to be a string and rejects arrays/maps.
func normalizeDeepSeekToolMessageContent(c *gin.Context, request *model.GeneralOpenAIRequest) {
	lg := gmw.GetLogger(c)
	normalizedCount := 0

	for i := range request.Messages {
		message := &request.Messages[i]
		if message.Role != "tool" {
			continue
		}

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
					lg.Debug("deepseek tool message fallback marshal failed",
						zap.Int("message_index", i),
						zap.String("original_content_type", fmt.Sprintf("%T", message.Content)),
						zap.Error(err),
					)
					normalized = fmt.Sprintf("%v", message.Content)
				} else {
					normalized = string(encoded)
				}
			}
		}

		message.Content = normalized
		normalizedCount++
		lg.Debug("normalized deepseek tool message content",
			zap.Int("message_index", i),
			zap.String("normalized_content_type", "string"),
			zap.Int("normalized_content_length", len(normalized)),
		)
	}

	if normalizedCount > 0 {
		lg.Debug("normalized deepseek tool messages for provider compatibility",
			zap.Int("normalized_count", normalizedCount),
			zap.Int("message_count", len(request.Messages)),
		)
	}
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, errors.New("deepseek does not support image generation")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	// 1. Shared Claude Messages → OpenAI conversion
	converted, err := openai_compatible.ConvertClaudeRequest(c, request)
	if err != nil {
		return nil, errors.Wrap(err, "convert claude request")
	}

	openaiReq, ok := converted.(*model.GeneralOpenAIRequest)
	if !ok {
		return converted, nil
	}

	// 2. DeepSeek-specific: normalize thinking config from original Claude request
	normalizeDeepSeekThinkingConfigFromOriginal(c, openaiReq)

	// 3. DeepSeek-specific: normalize tool message content to strings
	normalizeDeepSeekToolMessageContent(c, openaiReq)

	// 4. DeepSeek V4 enables thinking mode by default even when the request
	// omits the "thinking" field entirely. We only skip injection when thinking
	// is explicitly set to "disabled".
	thinkingDisabled := openaiReq.Thinking != nil && openaiReq.Thinking.Type == "disabled"
	if !thinkingDisabled {
		injectMissingReasoningContent(c, openaiReq)
	}

	// 5. Remove top-level Thinking field (not a valid OpenAI Chat Completions param)
	openaiReq.Thinking = nil

	return openaiReq, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	return openai_compatible.HandleClaudeMessagesResponse(c, resp, meta, func(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
		if meta.IsStream {
			return openai_compatible.StreamHandler(c, resp, promptTokens, modelName)
		}
		return openai_compatible.Handler(c, resp, promptTokens, modelName)
	})
}
