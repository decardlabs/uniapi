package moonshot

import (
	"io"
	"net/http"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/relay/adaptor"
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
	return "moonshot"
}

func (a *Adaptor) GetModelList() []string {
	return adaptor.GetModelListFromPricing(ModelRatios)
}

// GetDefaultModelPricing returns the pricing information for Moonshot models
// Based on official Moonshot pricing: https://platform.moonshot.cn/pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig {
	return ModelRatios
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

// DefaultToolingConfig returns Moonshot tooling defaults (none published as of 2025-11-12).
func (a *Adaptor) DefaultToolingConfig() adaptor.ChannelToolConfig {
	return MoonshotToolingDefaults
}

// Implement required adaptor interface methods (Moonshot uses OpenAI-compatible API)
func (a *Adaptor) Init(meta *meta.Meta) {}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// Route all chat-based modes to the chat completions endpoint
	switch meta.Mode {
	case relaymode.ChatCompletions, relaymode.ClaudeMessages, relaymode.ResponseAPI:
		return openai_compatible.GetFullRequestURL(meta.BaseURL, "/v1/chat/completions", meta.ChannelType), nil
	}
	// Moonshot uses OpenAI-compatible API endpoints
	return openai_compatible.GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	// Moonshot requires name on role=tool messages; backfill from tool_calls
	openai_compatible.BackfillToolMessageNamesFromToolCalls(request)

	// Moonshot is OpenAI-compatible, so we can pass the request through with minimal changes
	// Remove reasoning_effort as Moonshot doesn't support it
	if request.ReasoningEffort != nil {
		request.ReasoningEffort = nil
	}
	// Moonshot does not support top_k
	request.TopK = nil
	// Moonshot does not support json_schema response_format; preserve the schema as a system instruction
	if request.ResponseFormat != nil && request.ResponseFormat.JsonSchema != nil {
		structuredjson.EnsureInstruction(request)
		request.ResponseFormat = nil
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, errors.New("moonshot does not support image generation")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	converted, err := openai_compatible.ConvertClaudeRequest(c, request)
	if err != nil {
		return nil, errors.Wrap(err, "convert claude request")
	}
	openaiReq, ok := converted.(*model.GeneralOpenAIRequest)
	if !ok {
		return converted, nil
	}
	// Apply same post-processing as ConvertRequest
	openaiReq.ReasoningEffort = nil
	openaiReq.TopK = nil
	if openaiReq.ResponseFormat != nil && openaiReq.ResponseFormat.JsonSchema != nil {
		structuredjson.EnsureInstruction(openaiReq)
		openaiReq.ResponseFormat = nil
	}
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
