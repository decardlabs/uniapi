package minimax

import (
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/relay/adaptor"
	"github.com/decardlabs/uniapi/relay/adaptor/common/structuredjson"
	"github.com/decardlabs/uniapi/relay/adaptor/openai_compatible"
	"github.com/decardlabs/uniapi/relay/meta"
	"github.com/decardlabs/uniapi/relay/model"
)

type Adaptor struct {
	adaptor.DefaultPricingMethods
}

func (a *Adaptor) Init(meta *meta.Meta) {}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return GetRequestURL(meta)
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	openai_compatible.BackfillToolMessageNamesFromToolCalls(request)
	// MiniMax does not support reasoning_effort
	request.ReasoningEffort = nil
	// M2 series support top_k; strip only for legacy models
	if !isMiniMaxM2Model(request.Model) {
		request.TopK = nil
	}
	// MiniMax does not support json_schema response_format; preserve the schema as a system instruction
	if request.ResponseFormat != nil && request.ResponseFormat.JsonSchema != nil {
		structuredjson.EnsureInstruction(request)
		request.ResponseFormat = nil
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, errors.New("minimax does not support image generation")
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
	if !isMiniMaxM2Model(openaiReq.Model) {
		openaiReq.TopK = nil
	}
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
			return openai_compatible.StreamHandlerWithThinking(c, resp, promptTokens, modelName)
		}
		return openai_compatible.HandlerWithThinking(c, resp, promptTokens, modelName)
	})
}

func (a *Adaptor) GetModelList() []string {
	return adaptor.GetModelListFromPricing(ModelRatios)
}

func (a *Adaptor) GetChannelName() string {
	return "minimax"
}

// GetDefaultModelPricing returns the pricing information for Minimax models
// Based on Minimax pricing: https://api.minimax.chat/document/price
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

// DefaultToolingConfig returns MiniMax tooling defaults (no published per-call pricing as of 2025-11-12).
func (a *Adaptor) DefaultToolingConfig() adaptor.ChannelToolConfig {
	return MinimaxToolingDefaults
}

// isMiniMaxM2Model reports whether the model name indicates a MiniMax M2 series model.
func isMiniMaxM2Model(modelName string) bool {
	return strings.HasPrefix(modelName, "MiniMax-M2")
}
