package deepseek

import (
	"io"
	"net/http"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/common/ctxkey"
	"github.com/decardlabs/uniapi/relay/adaptor"
	"github.com/decardlabs/uniapi/relay/adaptor/common/deepseekcompat"
	"github.com/decardlabs/uniapi/relay/adaptor/common/structuredjson"
	"github.com/decardlabs/uniapi/relay/adaptor/openai_compatible"
	"github.com/decardlabs/uniapi/relay/meta"
	"github.com/decardlabs/uniapi/relay/model"
	"github.com/decardlabs/uniapi/relay/relaymode"
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
	// Route all chat-based modes to the chat completions endpoint
	switch meta.Mode {
	case relaymode.ChatCompletions, relaymode.ClaudeMessages, relaymode.ResponseAPI:
		return openai_compatible.GetFullRequestURL(meta.BaseURL, "/v1/chat/completions", meta.ChannelType), nil
	}
	return openai_compatible.GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	// DeepSeek requires name on role=tool messages; backfill from tool_calls
	openai_compatible.BackfillToolMessageNamesFromToolCalls(request)

	// DeepSeek is OpenAI-compatible, so we can pass the request through with minimal changes
	// Remove reasoning_effort as DeepSeek doesn't support it
	if request.ReasoningEffort != nil {
		request.ReasoningEffort = nil
	}

	// DeepSeek does not support top_k
	request.TopK = nil

	// Remove top-level Thinking field first, then re-apply from Claude context if present.
	// Check thinking status AFTER normalization to correctly handle the case where
	// normalizeDeepSeekThinkingConfigFromOriginal recovers Thinking from a ClaudeRequest.
	request.Thinking = nil
	normalizeDeepSeekThinkingConfigFromOriginal(c, request)

	normalizeDeepSeekToolMessageContent(c, request)

	// Inject reasoning_content when thinking mode is active (i.e. Thinking was
	// explicitly set on the original request or recovered from Claude context).
	if request.Thinking != nil {
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

// injectMissingReasoningContent delegates to the shared DeepSeek compatibility
// layer to ensure all assistant messages carry reasoning_content when thinking mode is active.
func injectMissingReasoningContent(c *gin.Context, request *model.GeneralOpenAIRequest) {
	deepseekcompat.InjectMissingReasoningContent(c, request)
}

// normalizeDeepSeekToolMessageContent delegates to the shared DeepSeek compatibility
// layer to convert non-string tool message content into strings.
func normalizeDeepSeekToolMessageContent(c *gin.Context, request *model.GeneralOpenAIRequest) {
	deepseekcompat.NormalizeToolMessageContent(gmw.GetLogger(c), request)
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

	// 5. DeepSeek-specific: strip unsupported fields
	openaiReq.ReasoningEffort = nil
	openaiReq.TopK = nil
	if openaiReq.ResponseFormat != nil && openaiReq.ResponseFormat.JsonSchema != nil {
		structuredjson.EnsureInstruction(openaiReq)
		openaiReq.ResponseFormat = nil
	}
	// openaiReq.Thinking = nil // REMOVED: DeepSeek API supports thinking parameter

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
