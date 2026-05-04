package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/common"
	"github.com/decardlabs/uniapi/common/ctxkey"
	"github.com/decardlabs/uniapi/model"
)

func setupUserControllerTest(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	originalRedisEnabled := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	t.Cleanup(func() {
		common.SetRedisEnabled(originalRedisEnabled)
	})

	tempDir := t.TempDir()
	originalSQLitePath := common.SQLitePath
	common.SQLitePath = filepath.Join(tempDir, "user-controller.db")
	t.Cleanup(func() {
		common.SQLitePath = originalSQLitePath
	})

	model.InitDB()
	model.InitLogDB()

	t.Cleanup(func() {
		if model.DB != nil {
			require.NoError(t, model.CloseDB())
			model.DB = nil
			model.LOG_DB = nil
		}
	})
}

func TestUpdateUserQuotaToZero(t *testing.T) {
	setupUserControllerTest(t)

	user := &model.User{
		Username:    "quota-user",
		Password:    "hashed-password",
		DisplayName: "Original",
		Quota:       100,
		Group:       "default",
		Status:      model.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)

	router := gin.New()
	router.PUT("/api/user/", func(c *gin.Context) {
		c.Set(ctxkey.Role, model.RoleRootUser)
		UpdateUser(c)
	})

	payload := map[string]any{
		"id":    user.Id,
		"quota": 0,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/user/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.True(t, resp.Success, resp.Message)

	updated, err := model.GetUserById(user.Id, true)
	require.NoError(t, err)
	require.Equal(t, int64(0), updated.Quota)
	require.Equal(t, "Original", updated.DisplayName)
}

func TestUpdateUserClearEmail(t *testing.T) {
	setupUserControllerTest(t)

	user := &model.User{
		Username:    "email-user",
		Password:    "hashed-password",
		DisplayName: "Email User",
		Email:       "user@example.com",
		Quota:       50,
		Group:       "default",
		Status:      model.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)

	router := gin.New()
	router.PUT("/api/user/", func(c *gin.Context) {
		c.Set(ctxkey.Role, model.RoleRootUser)
		UpdateUser(c)
	})

	payload := map[string]any{
		"id":    user.Id,
		"email": "",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/user/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.True(t, resp.Success, resp.Message)

	updated, err := model.GetUserById(user.Id, true)
	require.NoError(t, err)
	require.Equal(t, "", updated.Email)
	require.Equal(t, int64(50), updated.Quota)
}

func TestUpdateUserEmailNullSkipsChange(t *testing.T) {
	setupUserControllerTest(t)

	user := &model.User{
		Username:    "null-email-user",
		Password:    "hashed-password",
		DisplayName: "Null Email User",
		Email:       "existing@example.com",
		Quota:       42,
		Group:       "default",
		Status:      model.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)

	router := gin.New()
	router.PUT("/api/user/", func(c *gin.Context) {
		c.Set(ctxkey.Role, model.RoleRootUser)
		UpdateUser(c)
	})

	payload := map[string]any{
		"id":    user.Id,
		"email": nil,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/user/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.True(t, resp.Success, resp.Message)

	updated, err := model.GetUserById(user.Id, true)
	require.NoError(t, err)
	require.Equal(t, "existing@example.com", updated.Email)
	require.Equal(t, int64(42), updated.Quota)
}
