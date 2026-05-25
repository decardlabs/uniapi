package minimax

import (
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/relay/adaptor"
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
	backfillToolMessageNamesFromToolCalls(request)
	return request, nil
}

// backfillToolMessageNamesFromToolCalls backfills role=tool message names from prior assistant
// tool_calls when tool_call_id can be mapped to a function name.
func backfillToolMessageNamesFromToolCalls(request *model.GeneralOpenAIRequest) {
	if request == nil || len(request.Messages) == 0 {
		return
	}

	toolCallNames := make(map[string]string)
	for i := range request.Messages {
		message := &request.Messages[i]
		for _, toolCall := range message.ToolCalls {
			if toolCall.Id == "" || toolCall.Function == nil || toolCall.Function.Name == "" {
				continue
			}
			for _, key := range toolCallIDVariants(toolCall.Id) {
				if key == "" {
					continue
				}
				toolCallNames[key] = toolCall.Function.Name
			}
		}
	}

	for i := range request.Messages {
		message := &request.Messages[i]
		if message.Role != "tool" || message.Name != nil || message.ToolCallId == "" {
			continue
		}

		if name, ok := toolCallNames[message.ToolCallId]; ok && name != "" {
			nameCopy := name
			message.Name = &nameCopy
		}
	}
}

// toolCallIDVariants returns normalized key variants for matching call IDs from
// different protocol representations (call_*, fc_*, or bare suffix).
func toolCallIDVariants(id string) []string {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil
	}

	variants := []string{trimmed}
	if strings.HasPrefix(trimmed, "call_") {
		suffix := strings.TrimPrefix(trimmed, "call_")
		variants = append(variants, "fc_"+suffix, suffix)
	} else if strings.HasPrefix(trimmed, "fc_") {
		suffix := strings.TrimPrefix(trimmed, "fc_")
		variants = append(variants, "call_"+suffix, suffix)
	} else {
		variants = append(variants, "call_"+trimmed, "fc_"+trimmed)
	}

	return variants
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, errors.New("minimax does not support image generation")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	return openai_compatible.ConvertClaudeRequest(c, request)
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
