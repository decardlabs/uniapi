package middleware

import (
	"encoding/json"
	"slices"
	"sort"
	"strings"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/decardlabs/uniapi/common"
	"github.com/decardlabs/uniapi/common/config"
	"github.com/decardlabs/uniapi/common/helper"
	"github.com/decardlabs/uniapi/relay/model"
)

// AbortWithError aborts the request with an error message
func AbortWithError(c *gin.Context, statusCode int, err error) {
	logger := gmw.GetLogger(c)
	if shouldLogAsWarning(statusCode, err) {
		logger.Warn("server abort",
			zap.Int("status_code", statusCode),
			zap.Error(err))
	} else {
		logger.Error("server abort",
			zap.Int("status_code", statusCode),
			zap.Error(err))
	}

	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": helper.MessageWithRequestId(err.Error(), c.GetString(helper.RequestIdKey)),
			"type":    string(model.ErrorTypeOneAPI),
		},
	})
	c.Abort()
}

// TokenInfo holds information about an API token for logging purposes.
// All fields are masked or ID-based to avoid exposing sensitive data.
type TokenInfo struct {
	MaskedKey   string // Masked API key (prefix...suffix)
	TokenId     int    // Token ID
	TokenName   string // Token name
	UserId      int    // User ID who owns the token
	Username    string // Username (optional, may be empty)
	RequestedAt string // Request model (optional)
}

// AbortWithTokenError aborts the request with an error message and logs detailed token information.
// This function should be used when rejecting API key requests to provide more context in logs
// for debugging purposes. The token information is safely masked to avoid exposing sensitive data.
func AbortWithTokenError(c *gin.Context, statusCode int, err error, tokenInfo *TokenInfo) {
	logger := gmw.GetLogger(c)
	logFields := []zap.Field{
		zap.Int("status_code", statusCode),
		zap.Error(err),
	}

	// Add token info fields if available
	if tokenInfo != nil {
		logFields = append(logFields,
			zap.String("api_key", tokenInfo.MaskedKey),
			zap.Int("token_id", tokenInfo.TokenId),
			zap.String("token_name", tokenInfo.TokenName),
			zap.Int("user_id", tokenInfo.UserId),
		)
		if tokenInfo.Username != "" {
			logFields = append(logFields, zap.String("username", tokenInfo.Username))
		}
		if tokenInfo.RequestedAt != "" {
			logFields = append(logFields, zap.String("requested_model", tokenInfo.RequestedAt))
		}
	}

	if shouldLogAsWarning(statusCode, err) {
		logger.Warn("server abort", logFields...)
	} else {
		logger.Error("server abort", logFields...)
	}

	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": helper.MessageWithRequestId(err.Error(), c.GetString(helper.RequestIdKey)),
			"type":    string(model.ErrorTypeOneAPI),
		},
	})
	c.Abort()
}

// shouldLogAsWarning determines whether an abort should be logged as WARN.
//
// Parameters:
//   - statusCode: HTTP status code returned to client.
//   - err: the error that triggers abort.
//
// Returns:
//   - true if this is a client-caused or intentionally ignored case.
//   - false if this is a server-side failure that should be logged as ERROR.
func shouldLogAsWarning(statusCode int, err error) bool {
	if statusCode >= 400 && statusCode < 500 {
		return true
	}

	if err == nil {
		return false
	}

	switch {
	case strings.Contains(err.Error(), "token not found for key:"):
		return true
	default:
		return false
	}
}

func getRequestModel(c *gin.Context) (string, error) {
	// Realtime WS uses model in query string
	if strings.HasPrefix(c.Request.URL.Path, "/v1/realtime") {
		m := c.Query("model")
		if m == "" {
			return "", errors.New("missing required query parameter: model")
		}
		return m, nil
	}

	var modelRequest ModelRequest
	err := common.UnmarshalBodyReusable(c, &modelRequest)
	if err != nil {
		return "", errors.Wrap(err, "common.UnmarshalBodyReusable failed")
	}

	switch {
	case strings.HasPrefix(c.Request.URL.Path, "/v1/moderations"):
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	case strings.HasSuffix(c.Request.URL.Path, "embeddings"):
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	case strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations"),
		strings.HasPrefix(c.Request.URL.Path, "/v1/images/edits"):
		if modelRequest.Model == "" {
			modelRequest.Model = "dall-e-2"
		}
	case strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions"),
		strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations"):
		if modelRequest.Model == "" {
			modelRequest.Model = "whisper-1"
		}
	}

	return modelRequest.Model, nil
}

func isTextOnlyChatModelName(modelName string) bool {
	normalized := strings.ToLower(strings.TrimSpace(modelName))
	if normalized == "" {
		return false
	}

	if strings.Contains(normalized, "gpt-oss") {
		return true
	}

	if strings.Contains(normalized, "deepseek") {
		// All DeepSeek models are passed through to upstream without multimodal routing.
		// The upstream API decides whether the model supports image input.
		return false
	}

	return false
}

func requestContainsImageInput(c *gin.Context) bool {
	return len(detectImageInputContentTypes(c)) > 0
}

// detectImageInputContentTypes inspects request payload content items and returns
// detected image-related content types (for example image_url/input_image/image).
// Parameters: c is the current request context.
// Returns: sorted unique content type markers found in the request payload.
func detectImageInputContentTypes(c *gin.Context) []string {
	if c == nil || c.Request == nil {
		return nil
	}

	path := c.Request.URL.Path
	if !strings.HasPrefix(path, "/v1/chat/completions") &&
		!strings.HasPrefix(path, "/v1/messages") &&
		!strings.HasPrefix(path, "/v1/responses") {
		return nil
	}

	body, err := common.GetRequestBody(c)
	if err != nil || len(body) == 0 {
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}

	types := make(map[string]struct{})
	collectImageMarkers(payload["messages"], types)
	collectImageMarkers(payload["input"], types)

	if len(types) == 0 {
		return nil
	}

	collected := make([]string, 0, len(types))
	for typeName := range types {
		collected = append(collected, typeName)
	}
	sort.Strings(collected)

	return collected
}

func collectImageMarkers(value any, output map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		if marker, ok := typed["type"].(string); ok {
			normalized := strings.ToLower(strings.TrimSpace(marker))
			if normalized == "image_url" || normalized == "input_image" || normalized == "image" {
				output[normalized] = struct{}{}
			}
		}

		if _, exists := typed["image_url"]; exists {
			output["image_url"] = struct{}{}
		}

		for _, nested := range typed {
			collectImageMarkers(nested, output)
		}
	case []any:
		for idx := range typed {
			collectImageMarkers(typed[idx], output)
		}
	}
}

func isModelInList(modelName string, models string) bool {
	modelList := strings.Split(models, ",")
	return slices.Contains(modelList, modelName)
}

// GetTokenKeyParts extracts the token key parts from the Authorization header
//
// key like `sk-{token}[-{channelid}]`
//
// For WebSocket upgrade requests from browsers (which cannot set custom headers),
// the API key may be passed via the Sec-WebSocket-Protocol header using the
// subprotocol format: "openai-insecure-api-key.{KEY}"
func GetTokenKeyParts(c *gin.Context) []string {
	key := c.Request.Header.Get("Authorization")
	if key == "" {
		// compatible with Anthropic
		key = c.Request.Header.Get("X-Api-Key")
	}

	// For WebSocket upgrade requests, also check subprotocol-based auth.
	// Browsers cannot set custom headers on WebSocket connections, so the
	// OpenAI Realtime API allows passing the key as a subprotocol:
	//   Sec-WebSocket-Protocol: realtime, openai-insecure-api-key.{KEY}, openai-beta.realtime-v1
	if key == "" {
		if sp := c.Request.Header.Get("Sec-WebSocket-Protocol"); sp != "" {
			for _, proto := range strings.Split(sp, ",") {
				proto = strings.TrimSpace(proto)
				if strings.HasPrefix(proto, "openai-insecure-api-key.") {
					key = strings.TrimPrefix(proto, "openai-insecure-api-key.")
					break
				}
			}
		}
	}

	key = strings.TrimPrefix(key, "Bearer ")
	// Trim current configured prefix first
	if p := config.TokenKeyPrefix; p != "" {
		key = strings.TrimPrefix(key, p)
	}
	// Backward compatibility with historical prefixes
	key = strings.TrimPrefix(key, "sk-")
	key = strings.TrimPrefix(key, "laisky-")
	return strings.Split(key, "-")
}
