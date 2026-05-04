package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/relay/channeltype"
)

func TestGetChannelTypes(t *testing.T) {
	channeltype.ClearChannelTypes()
	channeltype.RegisterChannelType(channeltype.ChannelTypeInfoV2{
		ID:          9999,
		Name:        "mock",
		Label:       "Mock Channel",
		Category:    "test",
		Description: "mock type for testing",
		Template: channeltype.ChannelTypeTemplate{
			{Name: "api_base", Type: "string", Required: true, Description: "API Base URL"},
			{Name: "key", Type: "string", Required: true, Description: "API key"},
			{Name: "auth_mode", Type: "select", Required: false, Options: []interface{}{"token", "jwt"}},
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	GetChannelTypes(c)
	require.Equal(t, http.StatusOK, w.Code)

	var out struct {
		Success bool                  `json:"success"`
		Data    []ChannelTypeResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &out)
	require.NoError(t, err)
	require.True(t, out.Success)
	require.NotEmpty(t, out.Data)

	var found *ChannelTypeResponse
	for i := range out.Data {
		if out.Data[i].Value == 9999 {
			found = &out.Data[i]
			break
		}
	}
	require.NotNil(t, found)
	require.Equal(t, 9999, found.Key)
	require.Equal(t, "Mock Channel", found.Text)
	require.Equal(t, "test", found.Category)

	// api_base/key are core controls and must not be duplicated in dynamic template.
	require.Len(t, found.Template.Fields, 1)
	require.Equal(t, "auth_mode", found.Template.Fields[0].Key)
	require.Equal(t, "Auth Mode", found.Template.Fields[0].Label)
	require.Contains(t, found.Template.Fields[0].Help, "Configure")
	require.Len(t, found.Template.Fields[0].Options, 2)
	require.Equal(t, "token", found.Template.Fields[0].Options[0].Label)
}
