package internal

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"

	"github.com/decardlabs/uniapi/common/logger"
	"github.com/decardlabs/uniapi/model"
)

// VisionBackfillSummary describes the result of channel vision capability backfill.
type VisionBackfillSummary struct {
	ScannedChannels int
	UpdatedChannels int
	DryRun          bool
}

// VisionBackfillOptions controls channel selection for capability backfill.
type VisionBackfillOptions struct {
	OnlyEnabled bool
	ChannelIDs  map[int]struct{}
}

// BackfillVisionCapabilities updates channel config with vision capability fields based on supported models.
func BackfillVisionCapabilities(ctx context.Context, sourceDSN string, dryRun bool, opts VisionBackfillOptions) (*VisionBackfillSummary, error) {
	if strings.TrimSpace(sourceDSN) == "" {
		return nil, errors.New("source-dsn is required")
	}

	conn, err := ConnectDatabaseFromDSN(sourceDSN)
	if err != nil {
		return nil, errors.Wrap(err, "connect source database")
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logger.Logger.Warn("close source database connection failed", zap.Error(closeErr))
		}
	}()

	var channels []model.Channel
	if err := conn.DB.WithContext(ctx).Find(&channels).Error; err != nil {
		return nil, errors.Wrap(err, "load channels")
	}

	summary := &VisionBackfillSummary{ScannedChannels: len(channels), DryRun: dryRun}
	for idx := range channels {
		channel := &channels[idx]
		if !shouldBackfillChannel(channel, opts) {
			continue
		}

		cfg, changed, err := backfillChannelVisionConfig(channel)
		if err != nil {
			return nil, errors.Wrapf(err, "backfill channel %d", channel.Id)
		}
		if !changed {
			continue
		}

		encodedConfig, err := json.Marshal(cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal channel %d config", channel.Id)
		}

		summary.UpdatedChannels++
		if dryRun {
			logger.Logger.Info("vision capability backfill preview",
				zap.Int("channel_id", channel.Id),
				zap.String("channel_name", channel.Name),
				zap.String("config", string(encodedConfig)),
			)
			continue
		}

		if err := conn.DB.WithContext(ctx).Model(&model.Channel{}).Where("id = ?", channel.Id).Update("config", string(encodedConfig)).Error; err != nil {
			return nil, errors.Wrapf(err, "persist channel %d config", channel.Id)
		}
	}

	return summary, nil
}

// shouldBackfillChannel checks whether channel is selected by backfill options.
func shouldBackfillChannel(channel *model.Channel, opts VisionBackfillOptions) bool {
	if channel == nil {
		return false
	}

	if opts.OnlyEnabled && channel.Status != model.ChannelStatusEnabled {
		return false
	}

	if len(opts.ChannelIDs) == 0 {
		return true
	}

	_, ok := opts.ChannelIDs[channel.Id]
	return ok
}

// backfillChannelVisionConfig calculates channel vision capability fields from supported models.
func backfillChannelVisionConfig(channel *model.Channel) (model.ChannelConfig, bool, error) {
	if channel == nil {
		return model.ChannelConfig{}, false, errors.New("channel is nil")
	}

	cfg, err := channel.LoadConfig()
	if err != nil {
		return model.ChannelConfig{}, false, errors.Wrap(err, "load channel config")
	}

	if cfg.SupportsVision && len(cfg.VisionModels) > 0 {
		return cfg, false, nil
	}

	visionModels := detectVisionModels(channel.GetSupportedModelNames())
	if len(visionModels) == 0 {
		return cfg, false, nil
	}

	if !cfg.SupportsVision {
		cfg.SupportsVision = true
	}
	if len(cfg.VisionModels) == 0 {
		cfg.VisionModels = visionModels
	}

	return cfg, true, nil
}

// detectVisionModels filters model names that likely support image understanding.
func detectVisionModels(models []string) []string {
	if len(models) == 0 {
		return nil
	}

	var visionModels []string
	seen := make(map[string]struct{}, len(models))
	for idx := range models {
		candidate := strings.TrimSpace(models[idx])
		if candidate == "" || !isVisionModelName(candidate) {
			continue
		}
		key := strings.ToLower(candidate)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		visionModels = append(visionModels, candidate)
	}

	if len(visionModels) == 0 {
		return nil
	}

	return visionModels
}

// isVisionModelName returns true when model name belongs to known vision-capable families.
func isVisionModelName(modelName string) bool {
	normalized := strings.ToLower(strings.TrimSpace(modelName))
	if normalized == "" {
		return false
	}

	visionMarkers := []string{
		"gpt-4o",
		"gpt-4.1",
		"gpt-image",
		"claude-3",
		"gemini",
		"glm-4v",
		"qwen-vl",
		"vision",
		"multimodal",
		"llava",
		"pixtral",
		"kimi-v",
	}
	for idx := range visionMarkers {
		if strings.Contains(normalized, visionMarkers[idx]) {
			return true
		}
	}

	return false
}
