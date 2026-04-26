// Package middleware provides authentication middleware functions for the One API system.
//
// This file contains several authentication mechanisms:
//
// 1. Session-based Authentication (UserAuth, AdminAuth, RootAuth):
//   - Used for web dashboard access via browser sessions/cookies
//   - Falls back to Authorization header tokens if no session exists
//   - Different permission levels: User < Admin < Root
//
// 2. Token-based Authentication (TokenAuth):
//   - Used for programmatic API access with API keys
//   - Includes advanced features like IP restrictions, model permissions, quotas
//   - Supports channel-specific routing for admin users
//
// Key Differences:
// - Session auth: For human users accessing the web interface
// - Token auth: For applications/scripts making API calls
// - Token auth has more granular controls (IP, models, quotas)
// - Session auth has simpler role-based access (user/admin/root)
package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v7"
	"github.com/Laisky/zap"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/blacklist"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/network"
	"github.com/songquanpeng/one-api/model"
)

// authResult holds the resolved user identity from session or token authentication.
// All fields are pointers so callers can distinguish "not set" from zero values.
type authResult struct {
	username interface{}
	role      interface{}
	id        interface{}
	status    interface{}
	userObj   *model.User
}

// resolveIdentity attempts to authenticate a user from session cookies first,
// then falls back to the Authorization header (access token).
// Returns (*authResult, true) if authentication succeeded, or (nil, false) if not.
func resolveIdentity(c *gin.Context) (*authResult, bool) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username != nil {
		return &authResult{
			username: username,
			role:      session.Get("role"),
			id:        session.Get("id"),
			status:    session.Get("status"),
		}, true
	}

	// No session — try access token
	accessToken := c.Request.Header.Get("Authorization")
	if accessToken == "" {
		return nil, false
	}

	user := model.ValidateAccessToken(accessToken)
	if user == nil || user.Username == "" {
		return nil, false
	}

	return &authResult{
		userObj:   user,
		username:  user.Username,
		role:      user.Role,
		id:        user.Id,
		status:    user.Status,
	}, true
}

// checkUserStatus validates that the user is not disabled or banned.
// Returns an error message string if the check fails, empty string if OK.
func checkUserStatus(c *gin.Context, id int, status int) string {
	if status == model.UserStatusDisabled || blacklist.IsUserBanned(id) {
		// Clear session for banned users
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		return "User has been banned"
	}
	return ""
}

// setUserContext populates Gin context with the resolved user identity.
func setUserContext(c *gin.Context, result *authResult) {
	if result.userObj != nil {
		c.Set(ctxkey.UserObj, result.userObj)
	} else if result.id != nil {
		ctx := gmw.Ctx(c)
		uid := result.id.(int)
		userObj, err := model.CacheGetUserById(ctx, uid)
		if err != nil {
			gmw.GetLogger(c).Warn("failed to fetch user object for context", zap.Int("user_id", uid), zap.Error(err))
		}
		if userObj != nil {
			c.Set(ctxkey.UserObj, userObj)
		}
	}
	c.Set(ctxkey.Username, result.username)
	c.Set(ctxkey.Role, result.role)
	c.Set(ctxkey.Id, result.id)
}

// authHelper is the shared authentication logic for UserAuth/AdminAuth/RootAuth.
// It requires successful authentication — rejects with 401/403 on failure.
func authHelper(c *gin.Context, minRole int) {
	result, ok := resolveIdentity(c)
	if !ok {
		respondAuthError(c, http.StatusUnauthorized, "No permission to perform this operation, not logged in and no access token provided")
		return
	}

	// Validate user status (disabled / banned)
	if msg := checkUserStatus(c, result.id.(int), result.status.(int)); msg != "" {
		respondAuthError(c, http.StatusForbidden, msg)
		return
	}

	// Check role permissions
	if result.role.(int) < minRole {
		respondAuthError(c, http.StatusForbidden, "No permission to perform this operation, insufficient permissions")
		return
	}

	setUserContext(c, result)
	c.Next()
}

// UserAuth returns a middleware function that requires basic user authentication.
// This allows access to any logged-in user (common users, admins, and root users).
// Use this for endpoints that require authentication but don't need special privileges.
func UserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, model.RoleCommonUser)
	}
}

// OptionalUserAuth returns a middleware that tries to authenticate the user from
// session or access token but does NOT reject the request when no credentials are
// present.  If authentication succeeds the usual context keys (Id, Username, Role)
// are populated; otherwise the request continues anonymously (Id defaults to 0).
func OptionalUserAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		result, ok := resolveIdentity(c)
		if !ok {
			c.Next()
			return
		}

		// Validate user status
		if msg := checkUserStatus(c, result.id.(int), result.status.(int)); msg != "" {
			gmw.GetLogger(c).Info("optional auth: user is banned, skipping context set", zap.String("msg", msg))
			c.Next()
			return
		}

		setUserContext(c, result)
		c.Next()
	}
}

// AdminAuth returns a middleware function that requires administrator privileges.
// This restricts access to admin users and root users only.
// Use this for management endpoints that regular users shouldn't access.
func AdminAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, model.RoleAdminUser)
	}
}

// RootAuth returns a middleware function that requires root user privileges.
// This restricts access to root users only (highest privilege level).
// Use this for system-critical endpoints like user management, system configuration.
func RootAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c, model.RoleRootUser)
	}
}

// TokenAuth returns a middleware function for API token-based authentication.
// This is different from the session-based auth functions above - it's specifically
// designed for API access using tokens (like API keys for programmatic access).
// It performs additional validations like:
//   - Token validity and expiration
//   - IP subnet restrictions (if configured)
//   - Model access permissions
//   - Quota limits
//   - Channel-specific access (for admin users)
//
// Use this for API endpoints that will be accessed programmatically with API tokens.
func TokenAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := gmw.Ctx(c)
		// Parse the token key from the request (could include channel specification)
		// Parse the token key from the request (could include channel specification)
		parts := GetTokenKeyParts(c)
		key := parts[0]

		// Validate the API token against the database
		token, err := model.ValidateUserToken(ctx, key)
		if err != nil {
			AbortWithError(c, http.StatusUnauthorized, err)
			return
		}

		// Build token info for error logging (masked key for security)
		tokenInfo := &TokenInfo{
			MaskedKey: helper.MaskAPIKey(key),
			TokenId:   token.Id,
			TokenName: token.Name,
			UserId:    token.UserId,
		}

		// Check IP subnet restrictions (if configured for this token)
		if token.Subnet != nil && *token.Subnet != "" {
			if !network.IsIpInSubnets(ctx, c.ClientIP(), *token.Subnet) {
				AbortWithTokenError(c, http.StatusForbidden, errors.Errorf("This API key can only be used in the specified subnet: %s, current IP: %s", *token.Subnet, c.ClientIP()), tokenInfo)
				return
			}
		}

		// Fetch the full user object once; downstream handlers read from context
		// instead of making redundant DB/cache lookups.
		user, err := model.CacheGetUserById(ctx, token.UserId)
		if err != nil {
			AbortWithTokenError(c, http.StatusInternalServerError, errors.Wrap(err, "failed to get user"), tokenInfo)
			return
		}

		// Verify the token owner (user) is still enabled and not banned
		if user.Status == model.UserStatusDisabled || blacklist.IsUserBanned(user.Id) {
			AbortWithTokenError(c, http.StatusForbidden, errors.New("User has been banned"), tokenInfo)
			return
		}

		// Extract and validate the requested model (for AI/ML API endpoints)
		requestModel, err := getRequestModel(c)
		if err != nil && shouldCheckModel(c) {
			AbortWithTokenError(c, http.StatusBadRequest, err, tokenInfo)
			return
		}
		c.Set(ctxkey.RequestModel, requestModel)
		tokenInfo.RequestedAt = requestModel

		// Check if token has model restrictions and validate access
		if token.Models != nil && *token.Models != "" {
			c.Set(ctxkey.AvailableModels, *token.Models)
			if requestModel != "" && !isModelInList(requestModel, *token.Models) {
				AbortWithTokenError(c, http.StatusForbidden, errors.Errorf("This API key does not have permission to use the model: %s", requestModel), tokenInfo)
				return
			}
		}

		// Set user and token context for downstream handlers
		c.Set(ctxkey.UserObj, user)
		c.Set(ctxkey.Id, user.Id)
		c.Set(ctxkey.Username, user.Username)
		c.Set(ctxkey.TokenId, token.Id)
		c.Set(ctxkey.TokenName, token.Name)
		c.Set(ctxkey.TokenQuota, token.RemainQuota)
		c.Set(ctxkey.TokenQuotaUnlimited, token.UnlimitedQuota)

		// Handle channel-specific routing (admin feature)
		// Format: token_key:channel_id allows admins to specify which channel to use
		if len(parts) > 1 {
			if user.Role >= model.RoleAdminUser {
				cid, err := strconv.Atoi(parts[1])
				if err != nil {
					AbortWithTokenError(c, http.StatusBadRequest, errors.Errorf("Invalid Channel Id: %s", parts[1]), tokenInfo)
					return
				}

				c.Set(ctxkey.SpecificChannelId, cid)
			} else {
				AbortWithTokenError(c, http.StatusForbidden, errors.New("Ordinary users do not support specifying channels"), tokenInfo)
				return
			}
		}

		// Handle channel specification via URL parameter (for proxy relay)
		if channelId := c.Param("channelid"); channelId != "" {
			cid, err := strconv.Atoi(channelId)
			if err != nil {
				AbortWithTokenError(c, http.StatusBadRequest, errors.Errorf("Invalid Channel Id: %s", channelId), tokenInfo)
				return
			}

			c.Set(ctxkey.SpecificChannelId, cid)
		}

		c.Next()
	}
}

// shouldCheckModel determines whether the current endpoint requires model validation.
// This helper function checks if the request path corresponds to AI/ML API endpoints
// that need to validate which AI model the user is trying to access.
// Returns true for endpoints like completions, chat, images, and audio processing.
func shouldCheckModel(c *gin.Context) bool {
	if strings.HasPrefix(c.Request.URL.Path, "/v1/completions") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/chat/completions") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/realtime") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images") {
		return true
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		return true
	}
	return false
}

// respondAuthError centralizes error responses for auth failures (DRY, KISS)
func respondAuthError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"success": false, "message": message})
	c.Abort()
}
