package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v7"
	gutils "github.com/Laisky/go-utils/v6"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/decardlabs/uniapi/common"
	"github.com/decardlabs/uniapi/common/config"
	"github.com/decardlabs/uniapi/common/ctxkey"
	"github.com/decardlabs/uniapi/common/helper"
	"github.com/decardlabs/uniapi/model"
	"github.com/decardlabs/uniapi/relay/billing/ratio"
	"github.com/decardlabs/uniapi/relay/channeltype"
	"github.com/decardlabs/uniapi/relay/relaymode"
)

type ModelRequest struct {
	Model string `json:"model" form:"model"`
}

type responsePreviousIDRequest struct {
	PreviousResponseID *string `json:"previous_response_id"`
}

// isResponseAPIWebSocketHandshake reports whether current request is the websocket
// handshake for `/v1/responses`.
//
// Parameters:
//   - c: current gin context.
//   - relayMode: resolved relay mode for this request path.
//
// Returns:
//   - bool: true when request is a Response API websocket handshake.
func isResponseAPIWebSocketHandshake(c *gin.Context, relayMode int) bool {
	if c == nil || c.Request == nil {
		return false
	}

	return relayMode == relaymode.ResponseAPI &&
		c.Request.Method == http.MethodGet &&
		websocket.IsWebSocketUpgrade(c.Request)
}

// channelSupportsGroup reports whether channel serves the given group.
//
// Parameters:
//   - channel: channel candidate to validate.
//   - userGroup: caller's resolved group.
//
// Returns:
//   - bool: true when channel declares the user group.
func channelSupportsGroup(channel *model.Channel, userGroup string) bool {
	if channel == nil {
		return false
	}

	targetGroup := strings.TrimSpace(userGroup)
	if targetGroup == "" {
		return false
	}

	for grp := range strings.SplitSeq(channel.Group, ",") {
		if strings.EqualFold(strings.TrimSpace(grp), targetGroup) {
			return true
		}
	}

	return false
}

// selectResponseWebSocketChannelWithoutModel selects a channel for Response API
// websocket handshakes where model is unavailable during HTTP upgrade.
//
// Parameters:
//   - userGroup: caller's resolved group.
//   - relayMode: resolved relay mode.
//   - ignoreFirstPriority: whether to skip highest-priority tier.
//
// Returns:
//   - *model.Channel: selected channel.
//   - error: selection error when no channel matches.
func selectResponseWebSocketChannelWithoutModel(userGroup string, relayMode int, ignoreFirstPriority bool) (*model.Channel, error) {
	channels, err := model.GetAllEnabledChannels()
	if err != nil {
		return nil, errors.Wrap(err, "list enabled channels")
	}

	if len(channels) == 0 {
		return nil, errors.New("no enabled channels")
	}

	var candidates []*model.Channel
	for _, candidate := range channels {
		if candidate == nil {
			continue
		}
		if !channelSupportsGroup(candidate, userGroup) {
			continue
		}
		if !channelSupportsEndpoint(candidate, relayMode) {
			continue
		}
		if !channelSupportsResponseWebSocket(candidate, relayMode, true) {
			continue
		}

		candidates = append(candidates, candidate)
	}

	if len(candidates) == 0 {
		return nil, errors.New("no endpoint-compatible channel found for websocket request")
	}

	slices.SortStableFunc(candidates, func(a, b *model.Channel) int {
		switch {
		case a.GetPriority() > b.GetPriority():
			return -1
		case a.GetPriority() < b.GetPriority():
			return 1
		default:
			if a.Id < b.Id {
				return -1
			}
			if a.Id > b.Id {
				return 1
			}
			return 0
		}
	})

	highestPriority := candidates[0].GetPriority()
	if ignoreFirstPriority {
		for _, candidate := range candidates {
			if candidate.GetPriority() < highestPriority {
				return candidate, nil
			}
		}

		return nil, errors.New("no lower-priority websocket channel available")
	}

	return candidates[0], nil
}

// channelSupportsEndpoint checks if a channel supports the given relay mode (endpoint).
// It first checks the channel's custom supported_endpoints configuration,
// then falls back to the default endpoints for the channel type.
func channelSupportsEndpoint(channel *model.Channel, relayMode int) bool {
	// Get endpoint name for the relay mode
	endpointName := channeltype.RelayModeToEndpointName(relayMode)
	if endpointName == "" {
		// Unknown relay mode, allow it (backward compatibility)
		return true
	}

	// Check channel's custom supported endpoints
	customEndpoints := channel.GetSupportedEndpoints()
	if len(customEndpoints) > 0 {
		return channeltype.IsEndpointSupportedByName(endpointName, customEndpoints)
	}

	// Fall back to default endpoints for channel type
	defaultEndpoints := channeltype.DefaultEndpointNamesForChannelType(channel.Type)
	return channeltype.IsEndpointSupportedByName(endpointName, defaultEndpoints)
}

// channelSupportsResponseWebSocket reports whether channel can serve Response API
// websocket transport.
//
// Parameters:
//   - channel: candidate channel.
//   - relayMode: resolved relay mode.
//   - isResponseWSHandshake: whether current request is a response websocket handshake.
//
// Returns:
//   - bool: true when channel supports current transport constraints.
func channelSupportsResponseWebSocket(channel *model.Channel, relayMode int, isResponseWSHandshake bool) bool {
	if !isResponseWSHandshake {
		return true
	}

	if relayMode != relaymode.ResponseAPI {
		return true
	}

	return channel != nil && channel.Type == channeltype.OpenAI
}

// channelRateLimitValue returns the configured per-channel rate limit.
func channelRateLimitValue(channel *model.Channel) int {
	if channel == nil || channel.RateLimit == nil {
		return 0
	}

	return *channel.RateLimit
}

// extractPreviousResponseID reads previous_response_id from /v1/responses POST payload.
func extractPreviousResponseID(c *gin.Context) string {
	if c == nil || c.Request == nil || c.Request.Method != http.MethodPost {
		return ""
	}

	if !strings.HasSuffix(c.Request.URL.Path, "/v1/responses") {
		return ""
	}

	body, err := common.GetRequestBody(c)
	if err != nil || len(body) == 0 {
		return ""
	}

	var req responsePreviousIDRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		return ""
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	if req.PreviousResponseID == nil {
		return ""
	}

	return strings.TrimSpace(*req.PreviousResponseID)
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		lg := gmw.GetLogger(c)
		userId := c.GetInt(ctxkey.Id)
		ctx := gmw.Ctx(c)
		var userGroup string
		if userObj, exists := c.Get(ctxkey.UserObj); exists {
			if u, ok := userObj.(*model.User); ok {
				userGroup = u.Group
			}
		}
		if userGroup == "" {
			userGroup, _ = model.CacheGetUserGroup(ctx, userId)
		}
		c.Set(ctxkey.Group, userGroup)

		// Get relay mode for endpoint validation
		relayMode := relaymode.GetByPath(c.Request.URL.Path)

		var requestModel string
		var channel *model.Channel
		channelBoundFromSession := false
		selectedBy := "auto_selection"
		stickyEnabled := config.StickySessionEnabled
		channelId := c.GetInt(ctxkey.SpecificChannelId)
		if channelId == 0 {
			if responseID := strings.TrimSpace(c.Param("response_id")); responseID != "" {
				boundChannelID, found, bindErr := model.GetResponseBoundChannel(ctx, userId, responseID)
				if bindErr != nil {
					lg.Warn("failed to resolve response channel binding",
						zap.Error(bindErr),
						zap.Int("user_id", userId),
						zap.String("response_id", responseID),
					)
				} else if found {
					channelId = boundChannelID
					c.Set(ctxkey.SpecificChannelId, boundChannelID)
					channelBoundFromSession = true
					lg.Debug("response action routed by bound channel",
						zap.Int("user_id", userId),
						zap.String("response_id", responseID),
						zap.Int("channel_id", boundChannelID),
					)
				}
			}

			if channelId == 0 {
				if previousResponseID := extractPreviousResponseID(c); previousResponseID != "" {
					boundChannelID, found, bindErr := model.GetResponseBoundChannel(ctx, userId, previousResponseID)
					if bindErr != nil {
						lg.Warn("failed to resolve previous response channel binding",
							zap.Error(bindErr),
							zap.Int("user_id", userId),
							zap.String("previous_response_id", previousResponseID),
						)
					} else if found {
						channelId = boundChannelID
						c.Set(ctxkey.SpecificChannelId, boundChannelID)
						channelBoundFromSession = true
						lg.Debug("response request routed by previous_response_id binding",
							zap.Int("user_id", userId),
							zap.String("previous_response_id", previousResponseID),
							zap.Int("channel_id", boundChannelID),
						)
					}
				}
			}
		}
		if channelId != 0 {
			if channelBoundFromSession {
				selectedBy = "response_binding"
			} else {
				selectedBy = "specific_channel"
			}
			var err error
			channel, err = model.GetChannelById(channelId, true)
			if err != nil {
				AbortWithError(c, http.StatusBadRequest, errors.New("Invalid Channel Id"))
				return
			}
			if channel.Status != model.ChannelStatusEnabled {
				AbortWithError(c, http.StatusForbidden, errors.New("The channel has been disabled"))
				return
			}
			requestModel = c.GetString(ctxkey.RequestModel)
			if requestModel != "" && !channel.SupportsModel(requestModel) {
				AbortWithError(c, http.StatusBadRequest,
					errors.Errorf("Channel #%d does not support the requested model: %s", channelId, requestModel))
				return
			}
			isResponseWSHandshake := isResponseAPIWebSocketHandshake(c, relayMode)
			if !channelSupportsResponseWebSocket(channel, relayMode, isResponseWSHandshake) {
				AbortWithError(c, http.StatusBadRequest,
					errors.Errorf("Channel #%d does not support Response API websocket transport", channelId))
				return
			}

			// Check endpoint support for specific channel
			if !channelSupportsEndpoint(channel, relayMode) {
				endpointName := channeltype.RelayModeToEndpointName(relayMode)
				AbortWithError(c, http.StatusBadRequest,
					errors.Errorf("Channel #%d does not support the requested endpoint: %s", channelId, endpointName))
				return
			}

			if channelBoundFromSession {
				limitBlocked, limitErr := WouldChannelRateLimitBlock(c, channel.Id, channelRateLimitValue(channel))
				if limitErr != nil {
					lg.Warn("session-bound channel rate-limit precheck failed",
						zap.Error(limitErr),
						zap.Int("channel_id", channel.Id),
					)
				} else if limitBlocked {
					lg.Info("session-bound channel is rate limited, falling back to alternative channel",
						zap.Int("user_id", userId),
						zap.Int("channel_id", channel.Id),
					)
					channel = nil
					channelId = 0
					c.Set(ctxkey.SpecificChannelId, 0)
				}
			}
		} else {
			// keep auto-selection branch in sync with explicit-channel fallback behavior
		}

		if channel == nil {
			requestModel = c.GetString(ctxkey.RequestModel)
			if stickyEnabled && requestModel != "" {
				stickyChannelID, found, stickyErr := model.GetStickySessionChannel(ctx, userId, requestModel)
				if stickyErr != nil {
					lg.Warn("failed to load sticky session channel",
						zap.Error(stickyErr),
						zap.Int("user_id", userId),
						zap.String("model", requestModel),
					)
				} else if found {
					stickyChannel, loadErr := model.GetChannelById(stickyChannelID, true)
					if loadErr != nil {
						_ = model.DeleteStickySessionChannel(ctx, userId, requestModel)
					} else if stickyChannel.Status != model.ChannelStatusEnabled || !stickyChannel.SupportsModel(requestModel) {
						_ = model.DeleteStickySessionChannel(ctx, userId, requestModel)
					} else {
						limitBlocked, limitErr := WouldChannelRateLimitBlock(c, stickyChannel.Id, channelRateLimitValue(stickyChannel))
						if limitErr != nil {
							lg.Warn("sticky channel rate-limit precheck failed",
								zap.Error(limitErr),
								zap.Int("channel_id", stickyChannel.Id),
							)
						} else if !limitBlocked {
							channel = stickyChannel
							selectedBy = "sticky_session"
							lg.Debug("sticky session channel selected",
								zap.Int("user_id", userId),
								zap.String("model", requestModel),
								zap.Int("channel_id", stickyChannel.Id),
								zap.String("request_id", c.GetString(helper.RequestIdKey)),
							)
						} else {
							lg.Info("sticky session channel is rate limited, selecting alternative channel",
								zap.Int("user_id", userId),
								zap.String("model", requestModel),
								zap.Int("channel_id", stickyChannel.Id),
							)
						}
					}
				}
			} else if !stickyEnabled && requestModel != "" {
				lg.Debug("sticky session disabled, skip sticky lookup",
					zap.Int("user_id", userId),
					zap.String("model", requestModel),
				)
			}

			isResponseWSHandshake := isResponseAPIWebSocketHandshake(c, relayMode)
			if requestModel == "" && isResponseWSHandshake {
				if hintedModel := strings.TrimSpace(c.Query("model")); hintedModel != "" {
					requestModel = hintedModel
					c.Set(ctxkey.RequestModel, hintedModel)
					lg.Debug("response websocket handshake uses query model hint",
						zap.String("model", hintedModel),
					)
				} else {
					lg.Debug("response websocket handshake has no model in pre-upgrade request; selecting channel by group+endpoint")
				}
			}

			selectChannel := func(ignoreFirstPriority bool, exclude map[int]bool) (*model.Channel, error) {
				if requestModel == "" && isResponseWSHandshake {
					return selectResponseWebSocketChannelWithoutModel(userGroup, relayMode, ignoreFirstPriority)
				}

				for {
					candidate, err := model.CacheGetRandomSatisfiedChannelExcluding(userGroup, requestModel, ignoreFirstPriority, exclude, false)
					if err != nil {
						return nil, errors.Wrap(err, "select channel from cache")
					}

					limitBlocked, limitErr := WouldChannelRateLimitBlock(c, candidate.Id, channelRateLimitValue(candidate))
					if limitErr != nil {
						lg.Warn("candidate channel rate-limit precheck failed",
							zap.Error(limitErr),
							zap.Int("channel_id", candidate.Id),
						)
					} else if limitBlocked {
						exclude[candidate.Id] = true
						lg.Debug("channel skipped - rate limit reached",
							zap.Int("channel_id", candidate.Id),
							zap.String("channel_name", candidate.Name),
						)
						continue
					}

					// Check endpoint support
					if !channelSupportsEndpoint(candidate, relayMode) {
						exclude[candidate.Id] = true
						lg.Debug("channel skipped - does not support requested endpoint",
							zap.Int("channel_id", candidate.Id),
							zap.String("channel_name", candidate.Name),
							zap.String("endpoint", channeltype.RelayModeToEndpointName(relayMode)))
						continue
					}
					if !channelSupportsResponseWebSocket(candidate, relayMode, isResponseWSHandshake) {
						exclude[candidate.Id] = true
						lg.Debug("channel skipped - does not support response websocket transport",
							zap.Int("channel_id", candidate.Id),
							zap.String("channel_name", candidate.Name),
							zap.Int("channel_type", candidate.Type))
						continue
					}
					return candidate, nil
				}
			}

			if channel == nil {
				exclude := make(map[int]bool)
				var err error
				channel, err = selectChannel(false, exclude)
				if err != nil {
					lg.Info(fmt.Sprintf("No highest priority channels available for model %s in group %s, trying lower priority channels", requestModel, userGroup))
					channel, err = selectChannel(true, exclude)
					if err != nil {
						// Before returning 503, check whether all abilities are merely suspended
						// and will recover within ChannelPoolRecoveryWaitMax. If so, wait and retry
						// once to smooth over short rate-limit bursts without failing the request.
						if requestModel != "" && config.ChannelPoolRecoveryWaitMax > 0 {
							soonest, soonestErr := model.GetSoonestSuspendUntil(userGroup, requestModel)
							if soonestErr != nil {
								lg.Warn("failed to query soonest suspend_until for pool recovery wait",
									zap.Error(soonestErr),
									zap.String("model", requestModel),
								)
							} else if soonest != nil {
								waitDur := time.Until(*soonest)
								if waitDur > 0 && waitDur <= config.ChannelPoolRecoveryWaitMax {
									lg.Info("all channels suspended; waiting for earliest recovery before retrying",
										zap.Duration("wait_duration", waitDur),
										zap.Time("soonest_recovery", *soonest),
										zap.String("model", requestModel),
										zap.String("group", userGroup),
									)
									select {
									case <-ctx.Done():
										// Client disconnected; fall through to 503
									case <-time.After(waitDur):
										exclude2 := make(map[int]bool)
										channel, err = selectChannel(false, exclude2)
										if err != nil {
											channel, err = selectChannel(true, exclude2)
										}
									}
								}
							}
						}
						if channel == nil {
							message := fmt.Sprintf("No available channels for Model %s under Group %s", requestModel, userGroup)
							AbortWithError(c, http.StatusServiceUnavailable, errors.New(message))
							return
						}
					}
				}
				selectedBy = "auto_selection"
			}
		}

		if channel != nil {
			limitBlocked, limitErr := WouldChannelRateLimitBlock(c, channel.Id, channelRateLimitValue(channel))
			if limitErr != nil {
				lg.Warn("selected channel rate-limit precheck failed",
					zap.Error(limitErr),
					zap.Int("channel_id", channel.Id),
				)
			} else if limitBlocked {
				AbortWithError(c, http.StatusTooManyRequests, errors.New("selected channel is rate limited"))
				return
			}
		}
		lg.Debug("channel selected for request",
			zap.Int("user_id", userId),
			zap.String("user_group", userGroup),
			zap.String("request_model", requestModel),
			zap.Int("selected_channel_id", channel.Id),
			zap.String("selected_channel_name", channel.Name),
			zap.String("selected_by", selectedBy),
			zap.Bool("sticky_enabled", stickyEnabled),
		)
		SetupContextForSelectedChannel(c, channel, requestModel)
		if stickyEnabled && c.GetInt(ctxkey.SpecificChannelId) == 0 && requestModel != "" {
			if err := model.SetStickySessionChannel(ctx, userId, requestModel, channel.Id); err != nil {
				lg.Warn("failed to persist sticky session channel",
					zap.Error(err),
					zap.Int("user_id", userId),
					zap.String("model", requestModel),
					zap.Int("channel_id", channel.Id),
				)
			}
		} else if !stickyEnabled && requestModel != "" {
			lg.Debug("sticky session disabled, skip sticky persistence",
				zap.Int("user_id", userId),
				zap.String("model", requestModel),
				zap.Int("channel_id", channel.Id),
			)
		}
		c.Next()
	}
}

func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	lg := gmw.GetLogger(c)
	// one channel could relates to multiple groups,
	// and each groud has individual ratio,
	// set minimal group ratio as channel_ratio
	var minimalRatio float64 = -1
	for grp := range strings.SplitSeq(channel.Group, ",") {
		v := ratio.GetGroupRatio(grp)
		if minimalRatio < 0 || v < minimalRatio {
			minimalRatio = v
		}
	}
	lg.Info(fmt.Sprintf("set channel %s ratio to %f", channel.Name, minimalRatio))
	c.Set(ctxkey.ChannelRatio, minimalRatio)
	c.Set(ctxkey.ChannelModel, channel)

	// generate an unique cost id for each request
	if _, ok := c.Get(ctxkey.RequestId); !ok {
		c.Set(ctxkey.RequestId, gutils.UUID7())
	}

	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.ChannelId, channel.Id)
	c.Set(ctxkey.ChannelName, channel.Name)
	c.Set(ctxkey.ContentType, c.Request.Header.Get("Content-Type"))
	if channel.SystemPrompt != nil && *channel.SystemPrompt != "" {
		c.Set(ctxkey.SystemPrompt, *channel.SystemPrompt)
	}
	c.Set(ctxkey.ModelMapping, channel.GetModelMapping())
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	if channel.RateLimit != nil {
		c.Set(ctxkey.RateLimit, *channel.RateLimit)
	} else {
		c.Set(ctxkey.RateLimit, 0)
	}

	cfg, _ := channel.LoadConfig()
	// this is for backward compatibility
	if channel.Other != nil {
		switch channel.Type {
		case channeltype.Azure:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Xunfei:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.Gemini:
			if cfg.APIVersion == "" {
				cfg.APIVersion = *channel.Other
			}
		case channeltype.AIProxyLibrary:
			if cfg.LibraryID == "" {
				cfg.LibraryID = *channel.Other
			}
		case channeltype.Ali:
			if cfg.Plugin == "" {
				cfg.Plugin = *channel.Other
			}
		}
	}
	c.Set(ctxkey.Config, cfg)
}
