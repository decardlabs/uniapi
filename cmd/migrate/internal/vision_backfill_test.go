package internal

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/decardlabs/uniapi/model"
)

// TestDetectVisionModels verifies detection keeps known vision models and removes duplicates.
func TestDetectVisionModels(t *testing.T) {
	models := []string{
		"openai/gpt-oss-120b",
		"gpt-4o",
		"Gemini-2.5-Pro",
		"gpt-4o",
		"",
	}

	visionModels := detectVisionModels(models)
	require.Equal(t, []string{"gpt-4o", "Gemini-2.5-Pro"}, visionModels)
}

// TestBackfillChannelVisionConfig verifies missing capability fields are populated from model list.
func TestBackfillChannelVisionConfig(t *testing.T) {
	channel := &model.Channel{Models: "openai/gpt-oss-120b,gpt-4o"}

	cfg, changed, err := backfillChannelVisionConfig(channel)
	require.NoError(t, err)
	require.True(t, changed)
	require.True(t, cfg.SupportsVision)
	require.Equal(t, []string{"gpt-4o"}, cfg.VisionModels)
}

// TestBackfillChannelVisionConfigPreservesExisting verifies explicit capability config is not overwritten.
func TestBackfillChannelVisionConfigPreservesExisting(t *testing.T) {
	channel := &model.Channel{
		Models:  "gpt-4o",
		Config:  `{"supports_vision":true,"vision_models":["custom-vision-model"]}`,
		Status:  model.ChannelStatusEnabled,
		Name:    "test-channel",
		Weight:  nil,
		BaseURL: nil,
	}

	cfg, changed, err := backfillChannelVisionConfig(channel)
	require.NoError(t, err)
	require.False(t, changed)
	require.True(t, cfg.SupportsVision)
	require.Equal(t, []string{"custom-vision-model"}, cfg.VisionModels)
}

// TestShouldBackfillChannel verifies enabled-only and channel-id filtering behavior.
func TestShouldBackfillChannel(t *testing.T) {
	channelEnabled := &model.Channel{Id: 10, Status: model.ChannelStatusEnabled}
	channelDisabled := &model.Channel{Id: 20, Status: model.ChannelStatusManuallyDisabled}

	require.True(t, shouldBackfillChannel(channelEnabled, VisionBackfillOptions{}))
	require.True(t, shouldBackfillChannel(channelDisabled, VisionBackfillOptions{}))

	require.True(t, shouldBackfillChannel(channelEnabled, VisionBackfillOptions{OnlyEnabled: true}))
	require.False(t, shouldBackfillChannel(channelDisabled, VisionBackfillOptions{OnlyEnabled: true}))

	whitelist := map[int]struct{}{20: {}}
	require.False(t, shouldBackfillChannel(channelEnabled, VisionBackfillOptions{ChannelIDs: whitelist}))
	require.True(t, shouldBackfillChannel(channelDisabled, VisionBackfillOptions{ChannelIDs: whitelist}))

	require.False(t, shouldBackfillChannel(channelDisabled, VisionBackfillOptions{OnlyEnabled: true, ChannelIDs: whitelist}))
}
