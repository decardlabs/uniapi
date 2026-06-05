package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/common/config"
	"github.com/decardlabs/uniapi/common/ctxkey"
	"github.com/decardlabs/uniapi/common/logger"
	"github.com/decardlabs/uniapi/model"
	"github.com/decardlabs/uniapi/relay/channeltype"
)

// TestDistributeClaudeImageRejectionThenTextContinues verifies that Claude Messages
// requests can continue with text after an image rejection on text-only models.
func TestDistributeClaudeImageRejectionThenTextContinues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := setupDistributorTestDB(t)
	defer cleanup()

	originalRouteMode := config.MultimodalRouteMode
	config.MultimodalRouteMode = "capability_based"
	defer func() { config.MultimodalRouteMode = originalRouteMode }()

	user := &model.User{Id: 893, Username: "claude-continue", Password: "hashed", Group: "default", Status: model.UserStatusEnabled}
	require.NoError(t, db.Create(user).Error)

	priority := int64(200)
	textChannel := &model.Channel{
		Id:       8930,
		Name:     "text-only-channel",
		Type:     channeltype.OpenAI,
		Models:   "openai/gpt-oss-120b",
		Group:    "default",
		Status:   model.ChannelStatusEnabled,
		Priority: &priority,
	}
	require.NoError(t, db.Create(textChannel).Error)
	require.NoError(t, textChannel.AddAbilities())

	imageReqBody := `{"model":"openai/gpt-oss-120b","messages":[{"role":"user","content":[{"type":"text","text":"analyze screenshot"},{"type":"image","source":{"type":"url","url":"https://example.com/screenshot.png","media_type":"image/png"}}]}]}`
	imageReq := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(imageReqBody))
	imageReq.Header.Set("Content-Type", "application/json")
	imageRec := httptest.NewRecorder()
	imageCtx, _ := gin.CreateTestContext(imageRec)
	imageCtx.Request = imageReq
	imageCtx.Set(ctxkey.Id, user.Id)
	imageCtx.Set(ctxkey.UserObj, user)
	imageCtx.Set(ctxkey.RequestModel, "openai/gpt-oss-120b")
	imageCtx.Set(ctxkey.TokenId, 89301)
	gmw.SetLogger(imageCtx, logger.Logger)

	Distribute()(imageCtx)

	require.True(t, imageCtx.IsAborted())
	require.Equal(t, http.StatusBadRequest, imageRec.Code)
	require.Contains(t, imageRec.Body.String(), "only supports text content")
	require.Contains(t, imageRec.Body.String(), "content types: image")

	textReqBody := `{"model":"openai/gpt-oss-120b","messages":[{"role":"user","content":[{"type":"text","text":"continue with text only"}]}]}`
	textReq := httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(textReqBody))
	textReq.Header.Set("Content-Type", "application/json")
	textRec := httptest.NewRecorder()
	textCtx, _ := gin.CreateTestContext(textRec)
	textCtx.Request = textReq
	textCtx.Set(ctxkey.Id, user.Id)
	textCtx.Set(ctxkey.UserObj, user)
	textCtx.Set(ctxkey.RequestModel, "openai/gpt-oss-120b")
	textCtx.Set(ctxkey.TokenId, 89302)
	gmw.SetLogger(textCtx, logger.Logger)

	Distribute()(textCtx)

	require.False(t, textCtx.IsAborted(), "Claude text-only request should proceed after previous image rejection")
	require.Equal(t, textChannel.Id, textCtx.GetInt(ctxkey.ChannelId))
	require.Empty(t, textCtx.GetString(ctxkey.AutoRoutedModel))
}

// TestDistributeResponseImageRejectionThenTextContinues verifies that Response API
// requests can continue with text after an image rejection on text-only models.
func TestDistributeResponseImageRejectionThenTextContinues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, cleanup := setupDistributorTestDB(t)
	defer cleanup()

	originalRouteMode := config.MultimodalRouteMode
	config.MultimodalRouteMode = "capability_based"
	defer func() { config.MultimodalRouteMode = originalRouteMode }()

	user := &model.User{Id: 894, Username: "response-continue", Password: "hashed", Group: "default", Status: model.UserStatusEnabled}
	require.NoError(t, db.Create(user).Error)

	priority := int64(200)
	textChannel := &model.Channel{
		Id:       8940,
		Name:     "text-only-response-channel",
		Type:     channeltype.OpenAI,
		Models:   "openai/gpt-oss-120b",
		Group:    "default",
		Status:   model.ChannelStatusEnabled,
		Priority: &priority,
	}
	require.NoError(t, db.Create(textChannel).Error)
	require.NoError(t, textChannel.AddAbilities())

	imageReqBody := `{"model":"openai/gpt-oss-120b","input":[{"role":"user","content":[{"type":"input_text","text":"analyze screenshot"},{"type":"input_image","image_url":"https://example.com/screenshot.png"}]}]}`
	imageReq := httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(imageReqBody))
	imageReq.Header.Set("Content-Type", "application/json")
	imageRec := httptest.NewRecorder()
	imageCtx, _ := gin.CreateTestContext(imageRec)
	imageCtx.Request = imageReq
	imageCtx.Set(ctxkey.Id, user.Id)
	imageCtx.Set(ctxkey.UserObj, user)
	imageCtx.Set(ctxkey.RequestModel, "openai/gpt-oss-120b")
	imageCtx.Set(ctxkey.TokenId, 89401)
	gmw.SetLogger(imageCtx, logger.Logger)

	Distribute()(imageCtx)

	require.True(t, imageCtx.IsAborted())
	require.Equal(t, http.StatusBadRequest, imageRec.Code)
	require.Contains(t, imageRec.Body.String(), "only supports text content")
	require.Contains(t, imageRec.Body.String(), "content types: image_url,input_image")

	textReqBody := `{"model":"openai/gpt-oss-120b","input":[{"role":"user","content":[{"type":"input_text","text":"continue with text only"}]}]}`
	textReq := httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(textReqBody))
	textReq.Header.Set("Content-Type", "application/json")
	textRec := httptest.NewRecorder()
	textCtx, _ := gin.CreateTestContext(textRec)
	textCtx.Request = textReq
	textCtx.Set(ctxkey.Id, user.Id)
	textCtx.Set(ctxkey.UserObj, user)
	textCtx.Set(ctxkey.RequestModel, "openai/gpt-oss-120b")
	textCtx.Set(ctxkey.TokenId, 89402)
	gmw.SetLogger(textCtx, logger.Logger)

	Distribute()(textCtx)

	require.False(t, textCtx.IsAborted(), "Response API text-only request should proceed after previous image rejection")
	require.Equal(t, textChannel.Id, textCtx.GetInt(ctxkey.ChannelId))
	require.Empty(t, textCtx.GetString(ctxkey.AutoRoutedModel))
}
