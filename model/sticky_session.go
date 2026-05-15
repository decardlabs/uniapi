package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"

	"github.com/decardlabs/uniapi/common"
	"github.com/decardlabs/uniapi/common/config"
	"github.com/decardlabs/uniapi/common/logger"
)

type stickyCacheEntry struct {
	channelID int
	expiresAt time.Time
}

type functionContextEntry struct {
	RequestID          string `json:"request_id"`
	UserID             int    `json:"user_id"`
	Model              string `json:"model"`
	ChannelID          int    `json:"channel_id"`
	PreviousResponseID string `json:"previous_response_id,omitempty"`
	ToolCount          int    `json:"tool_count"`
	CreatedAtUTC       int64  `json:"created_at_utc"`
}

var (
	stickySessionMemoryLock sync.RWMutex
	stickySessionMemory     = map[string]stickyCacheEntry{}
)

// stickySessionTTL returns the configured sticky-session timeout as a duration.
func stickySessionTTL() time.Duration {
	ttlSeconds := config.StickySessionTimeoutSeconds
	if ttlSeconds <= 0 {
		ttlSeconds = 30 * 60
	}

	return time.Duration(ttlSeconds) * time.Second
}

// normalizedModelKey normalizes model names for sticky/session storage keys.
func normalizedModelKey(model string) string {
	trimmed := strings.TrimSpace(strings.ToLower(model))
	if trimmed == "" {
		return "unknown"
	}

	return trimmed
}

// stickySessionKey builds the storage key for user-to-channel sticky binding.
func stickySessionKey(userID int, model string) string {
	return fmt.Sprintf("sticky_session:user:%d:model:%s", userID, normalizedModelKey(model))
}

// responseChannelKey builds the storage key for response_id-to-channel binding.
func responseChannelKey(userID int, responseID string) string {
	return fmt.Sprintf("response_channel:user:%d:response:%s", userID, strings.TrimSpace(responseID))
}

// functionContextKey builds the storage key for centralized function-call context history.
func functionContextKey(userID int, model string) string {
	return fmt.Sprintf("function_context:user:%d:model:%s", userID, normalizedModelKey(model))
}

// SetStickySessionChannel stores or refreshes a sticky channel binding for a user/model.
func SetStickySessionChannel(ctx context.Context, userID int, model string, channelID int) error {
	if userID <= 0 || channelID <= 0 || strings.TrimSpace(model) == "" {
		return errors.New("invalid sticky-session binding arguments")
	}

	ttl := stickySessionTTL()
	key := stickySessionKey(userID, model)

	if common.IsRedisEnabled() {
		if err := common.RedisSet(ctx, key, fmt.Sprintf("%d", channelID), ttl); err != nil {
			return errors.Wrap(err, "set sticky session channel in redis")
		}
		return nil
	}

	stickySessionMemoryLock.Lock()
	stickySessionMemory[key] = stickyCacheEntry{channelID: channelID, expiresAt: time.Now().UTC().Add(ttl)}
	stickySessionMemoryLock.Unlock()
	return nil
}

// GetStickySessionChannel loads the sticky channel binding for a user/model if it exists and is not expired.
func GetStickySessionChannel(ctx context.Context, userID int, model string) (int, bool, error) {
	if userID <= 0 || strings.TrimSpace(model) == "" {
		return 0, false, nil
	}

	key := stickySessionKey(userID, model)

	if common.IsRedisEnabled() {
		value, err := common.RedisGet(ctx, key)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "redis: nil") {
				return 0, false, nil
			}
			return 0, false, errors.Wrap(err, "get sticky session channel from redis")
		}

		var channelID int
		if _, scanErr := fmt.Sscanf(value, "%d", &channelID); scanErr != nil {
			return 0, false, errors.Wrap(scanErr, "parse sticky session channel id")
		}
		if channelID <= 0 {
			return 0, false, nil
		}
		return channelID, true, nil
	}

	stickySessionMemoryLock.RLock()
	entry, ok := stickySessionMemory[key]
	stickySessionMemoryLock.RUnlock()
	if !ok {
		return 0, false, nil
	}

	if time.Now().UTC().After(entry.expiresAt) {
		stickySessionMemoryLock.Lock()
		delete(stickySessionMemory, key)
		stickySessionMemoryLock.Unlock()
		return 0, false, nil
	}

	return entry.channelID, true, nil
}

// DeleteStickySessionChannel removes sticky channel binding for a user/model.
func DeleteStickySessionChannel(ctx context.Context, userID int, model string) error {
	if userID <= 0 || strings.TrimSpace(model) == "" {
		return nil
	}

	key := stickySessionKey(userID, model)

	if common.IsRedisEnabled() {
		if err := common.RedisDel(ctx, key); err != nil {
			return errors.Wrap(err, "delete sticky session channel from redis")
		}
		return nil
	}

	stickySessionMemoryLock.Lock()
	delete(stickySessionMemory, key)
	stickySessionMemoryLock.Unlock()
	return nil
}

// SetResponseBoundChannel stores response_id to channel binding so follow-up response actions keep affinity.
func SetResponseBoundChannel(ctx context.Context, userID int, responseID string, channelID int) error {
	if userID <= 0 || channelID <= 0 || strings.TrimSpace(responseID) == "" {
		return errors.New("invalid response-channel binding arguments")
	}

	key := responseChannelKey(userID, responseID)
	if err := common.RedisSet(ctx, key, fmt.Sprintf("%d", channelID), stickySessionTTL()); err != nil {
		if !common.IsRedisEnabled() {
			stickySessionMemoryLock.Lock()
			stickySessionMemory[key] = stickyCacheEntry{channelID: channelID, expiresAt: time.Now().UTC().Add(stickySessionTTL())}
			stickySessionMemoryLock.Unlock()
			return nil
		}
		return errors.Wrap(err, "set response-channel binding in redis")
	}

	return nil
}

// GetResponseBoundChannel loads bound channel by response_id for the current user.
func GetResponseBoundChannel(ctx context.Context, userID int, responseID string) (int, bool, error) {
	if userID <= 0 || strings.TrimSpace(responseID) == "" {
		return 0, false, nil
	}

	key := responseChannelKey(userID, responseID)
	if common.IsRedisEnabled() {
		value, err := common.RedisGet(ctx, key)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "redis: nil") {
				return 0, false, nil
			}
			return 0, false, errors.Wrap(err, "get response-channel binding from redis")
		}

		var channelID int
		if _, scanErr := fmt.Sscanf(value, "%d", &channelID); scanErr != nil {
			return 0, false, errors.Wrap(scanErr, "parse response-channel binding channel id")
		}
		if channelID <= 0 {
			return 0, false, nil
		}
		return channelID, true, nil
	}

	stickySessionMemoryLock.RLock()
	entry, ok := stickySessionMemory[key]
	stickySessionMemoryLock.RUnlock()
	if !ok {
		return 0, false, nil
	}

	if time.Now().UTC().After(entry.expiresAt) {
		stickySessionMemoryLock.Lock()
		delete(stickySessionMemory, key)
		stickySessionMemoryLock.Unlock()
		return 0, false, nil
	}

	return entry.channelID, true, nil
}

// RecordFunctionContext stores compact function-calling context history centrally for debugging and session continuity.
func RecordFunctionContext(ctx context.Context, userID int, model string, channelID int, requestID string, previousResponseID string, toolCount int) {
	if userID <= 0 || channelID <= 0 || strings.TrimSpace(model) == "" {
		return
	}
	if toolCount <= 0 && strings.TrimSpace(previousResponseID) == "" {
		return
	}

	entry := functionContextEntry{
		RequestID:          strings.TrimSpace(requestID),
		UserID:             userID,
		Model:              normalizedModelKey(model),
		ChannelID:          channelID,
		PreviousResponseID: strings.TrimSpace(previousResponseID),
		ToolCount:          toolCount,
		CreatedAtUTC:       time.Now().UTC().Unix(),
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		logger.Logger.Warn("failed to marshal function context entry", zap.Error(err))
		return
	}

	key := functionContextKey(userID, model)
	ttl := stickySessionTTL()
	maxRecords := config.FunctionContextHistoryLimit
	if maxRecords <= 0 {
		maxRecords = 50
	}

	if common.IsRedisEnabled() {
		if err := common.RDB.LPush(ctx, key, string(payload)).Err(); err != nil {
			logger.Logger.Warn("failed to push function context to redis", zap.Error(err), zap.String("key", key))
			return
		}
		if err := common.RDB.LTrim(ctx, key, 0, int64(maxRecords-1)).Err(); err != nil {
			logger.Logger.Warn("failed to trim function context list in redis", zap.Error(err), zap.String("key", key))
		}
		if err := common.RDB.Expire(ctx, key, ttl).Err(); err != nil {
			logger.Logger.Warn("failed to refresh function context ttl in redis", zap.Error(err), zap.String("key", key))
		}
		return
	}

	memoryKey := key + ":history"
	stickySessionMemoryLock.Lock()
	stickySessionMemory[memoryKey] = stickyCacheEntry{channelID: channelID, expiresAt: time.Now().UTC().Add(ttl)}
	stickySessionMemoryLock.Unlock()
}
