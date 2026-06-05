package zhipu

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/common/ctxkey"
	"github.com/decardlabs/uniapi/common/helper"
	"github.com/decardlabs/uniapi/relay/adaptor"
	"github.com/decardlabs/uniapi/relay/adaptor/common/structuredjson"
	"github.com/decardlabs/uniapi/relay/adaptor/openai"
	"github.com/decardlabs/uniapi/relay/adaptor/openai_compatible"
	"github.com/decardlabs/uniapi/relay/meta"
	"github.com/decardlabs/uniapi/relay/model"
	"github.com/decardlabs/uniapi/relay/relaymode"
)

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {

}

// getAPIVersion determines the API version from the model name.
// Models with "glm-" prefix use v4 (OpenAI-compatible), others use v3 (proprietary format).
func getAPIVersion(modelName string) string {
	if strings.HasPrefix(modelName, "glm-") {
		return "v4"
	}
	return "v3"
}

// normalizeZhipuBaseURL returns the Zhipu service host base URL.
// It accepts both legacy forms with /api/paas/v4 suffix and plain host forms.
func normalizeZhipuBaseURL(rawBaseURL string) string {
	baseURL := strings.TrimSpace(rawBaseURL)
	baseURL = strings.TrimRight(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api/paas/v4")
	return strings.TrimRight(baseURL, "/")
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	baseURL := normalizeZhipuBaseURL(meta.BaseURL)
	switch meta.Mode {
	case relaymode.ImagesGenerations:
		return fmt.Sprintf("%s/api/paas/v4/images/generations", baseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/api/paas/v4/embeddings", baseURL), nil
	case relaymode.OCR:
		return fmt.Sprintf("%s/api/paas/v4/layout_parsing", baseURL), nil
	}
	// OCR model detection by model name takes priority for backward compatibility
	if isOCRModel(meta.ActualModelName) {
		return fmt.Sprintf("%s/api/paas/v4/layout_parsing", baseURL), nil
	}
	// All other modes (ChatCompletions, ClaudeMessages, ResponseAPI, etc.)
	// route to the chat completions endpoint
	if getAPIVersion(meta.ActualModelName) == "v4" {
		return fmt.Sprintf("%s/api/paas/v4/chat/completions", baseURL), nil
	}
	method := "invoke"
	if meta.IsStream {
		method = "sse-invoke"
	}
	return fmt.Sprintf("%s/api/paas/v3/model-api/%s/%s", baseURL, meta.ActualModelName, method), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if getAPIVersion(meta.ActualModelName) == "v4" {
		apiKey := strings.TrimSpace(meta.APIKey)
		if apiKey == "" {
			return errors.New("zhipu api key is required")
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		return nil
	}

	token, err := GetToken(meta.APIKey)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case relaymode.Embeddings:
		baiduEmbeddingRequest, err := ConvertEmbeddingRequest(*request)
		if err != nil {
			return nil, errors.Wrap(err, "convert zhipu embedding request")
		}
		return baiduEmbeddingRequest, nil
	default:
		if isOCRModel(request.Model) {
			return ConvertOCRRequest(*request)
		}

		// TopP [0.0, 1.0]
		request.TopP = helper.Float64PtrMax(request.TopP, 1)
		request.TopP = helper.Float64PtrMin(request.TopP, 0)

		// Temperature [0.0, 1.0]
		request.Temperature = helper.Float64PtrMax(request.Temperature, 1)
		request.Temperature = helper.Float64PtrMin(request.Temperature, 0)

		// Zhipu does not support reasoning_effort
		request.ReasoningEffort = nil

		// Zhipu does not support top_k
		request.TopK = nil

		// Zhipu does not support json_schema response_format; preserve the schema as a system instruction
		if request.ResponseFormat != nil && request.ResponseFormat.JsonSchema != nil {
			structuredjson.EnsureInstruction(request)
			request.ResponseFormat = nil
		}

		if getAPIVersion(request.Model) == "v4" {
			return request, nil
		}
		v3Req, err := ConvertRequest(*request)
		if err != nil {
			return nil, err
		}
		return v3Req, nil
	}
}

func (a *Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	newRequest := ImageRequest{
		Model:  request.Model,
		Prompt: request.Prompt,
		UserId: request.User,
	}
	return newRequest, nil
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	// 1. Use the shared Claude-to-OpenAI conversion (handles structured output
	//    promotion, tool normalization, tool_choice normalization, and all
	//    message/block conversion)
	converted, err := openai_compatible.ConvertClaudeRequest(c, request)
	if err != nil {
		return nil, errors.Wrap(err, "convert claude request")
	}

	openaiReq, ok := converted.(*model.GeneralOpenAIRequest)
	if !ok {
		return converted, nil
	}

	// 2. Zhipu-specific: clamp Temperature and TopP to [0, 1]
	openaiReq.Temperature = helper.Float64PtrMax(openaiReq.Temperature, 1)
	openaiReq.Temperature = helper.Float64PtrMin(openaiReq.Temperature, 0)
	openaiReq.TopP = helper.Float64PtrMax(openaiReq.TopP, 1)
	openaiReq.TopP = helper.Float64PtrMin(openaiReq.TopP, 0)

	// 3. Zhipu-specific: strip unsupported fields
	openaiReq.ReasoningEffort = nil
	openaiReq.TopK = nil
	if openaiReq.ResponseFormat != nil && openaiReq.ResponseFormat.JsonSchema != nil {
		structuredjson.EnsureInstruction(openaiReq)
		openaiReq.ResponseFormat = nil
	}

	// 4. Zhipu-specific: v3 format conversion vs v4 passthrough
	if getAPIVersion(openaiReq.Model) == "v4" {
		return openaiReq, nil
	}
	v3Req, err := ConvertRequest(*openaiReq)
	if err != nil {
		return nil, err
	}
	return v3Req, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponseV4(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	// Handle Claude Messages conversion when needed
	if isClaudeConversion, exists := c.Get(ctxkey.ClaudeMessagesConversion); exists && isClaudeConversion.(bool) {
		return openai_compatible.HandleClaudeMessagesResponse(c, resp, meta, func(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
			if meta.IsStream {
				err, _, usage := openai.StreamHandler(c, resp, meta.Mode)
				return err, usage
			}
			err, usage := openai.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
			return err, usage
		})
	}

	if meta.IsStream {
		err, _, usage = openai.StreamHandler(c, resp, meta.Mode)
	} else {
		err, usage = openai.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	switch meta.Mode {
	case relaymode.Embeddings:
		err, usage = EmbeddingsHandler(c, resp)
		return
	case relaymode.ImagesGenerations:
		err, usage = openai.ImageHandler(c, resp)
		return
	}
	if isOCRModel(meta.ActualModelName) {
		err, usage = OCRHandler(c, resp, meta.ActualModelName)
		return
	}
	if getAPIVersion(meta.ActualModelName) == "v4" {
		return a.DoResponseV4(c, resp, meta)
	}
	if meta.IsStream {
		err, usage = StreamHandler(c, resp, meta.ActualModelName)
	} else {
		if meta.Mode == relaymode.Embeddings {
			err, usage = EmbeddingsHandler(c, resp)
		} else {
			err, usage = Handler(c, resp, meta.ActualModelName)
		}
	}
	return
}

func ConvertEmbeddingRequest(request model.GeneralOpenAIRequest) (*EmbeddingRequest, error) {
	inputs := request.ParseInput()
	if len(inputs) != 1 {
		return nil, errors.New("invalid input length, zhipu only support one input")
	}
	return &EmbeddingRequest{
		Model: request.Model,
		Input: inputs[0],
	}, nil
}

func (a *Adaptor) GetModelList() []string {
	return adaptor.GetModelListFromPricing(ModelRatios)
}

func (a *Adaptor) GetChannelName() string {
	return "zhipu"
}

// Pricing methods - Zhipu adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig {
	const MilliRmb = 0.0001

	// Direct map definition - much easier to maintain and edit
	return ModelRatios
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default Zhipu pricing
	return 0.001 * 0.0001 // Default RMB pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	return 1.0 // Default completion ratio for Zhipu
}

// DefaultToolingConfig returns Zhipu tooling defaults (search tool tiers and rates).
func (a *Adaptor) DefaultToolingConfig() adaptor.ChannelToolConfig {
	return ZhipuToolingDefaults
}

// ConvertOCRRequest converts the user-facing OCR request to the Zhipu-native format.
func (a *Adaptor) ConvertOCRRequest(_ *gin.Context, request *model.OCRRequest) (any, error) {
	if request == nil {
		return nil, errors.New("OCR request is nil")
	}
	return &OCRRequest{
		Model:                   request.Model,
		File:                    request.File,
		RequestID:               request.RequestID,
		UserID:                  request.UserID,
		ReturnCropImages:        request.ReturnCropImages,
		NeedLayoutVisualization: request.NeedLayoutVisualization,
		StartPageID:             request.StartPageID,
		EndPageID:               request.EndPageID,
	}, nil
}

// DoOCRResponse handles the upstream OCR response and writes the result to the client.
func (a *Adaptor) DoOCRResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	err, usage = OCRHandler(c, resp, meta.ActualModelName)
	return
}
