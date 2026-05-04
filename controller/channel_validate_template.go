package controller

import (
	"fmt"

	"github.com/decardlabs/uniapi/model"
	"github.com/decardlabs/uniapi/relay/channeltype"
)

// ValidateChannelParamsByTemplate 校验渠道参数是否符合类型模板
func ValidateChannelParamsByTemplate(channel *model.Channel) error {
	info, ok := channeltype.GetChannelType(channel.Type)
	if !ok {
		return fmt.Errorf("unknown channel type: %d", channel.Type)
	}
	cfg, err := channel.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}
	for _, field := range info.Template {
		val, exists := getFieldValue(channel, cfg, field.Name)
		if field.Required && (!exists || isEmpty(val)) {
			return fmt.Errorf("参数 %s 必填", field.Name)
		}
		if field.Pattern != "" && exists && !matchPattern(val, field.Pattern) {
			return fmt.Errorf("参数 %s 格式不合法", field.Name)
		}
	}
	return nil
}

// getFieldValue 获取 config 中字段值（支持一级字段）
func getFieldValue(channel *model.Channel, cfg model.ChannelConfig, name string) (interface{}, bool) {
	switch name {
	case "api_base":
		if channel.BaseURL == nil {
			return "", false
		}
		return *channel.BaseURL, *channel.BaseURL != ""
	case "key":
		return channel.Key, channel.Key != ""
	case "region":
		return cfg.Region, cfg.Region != ""
	case "sk":
		return cfg.SK, cfg.SK != ""
	case "ak":
		return cfg.AK, cfg.AK != ""
	case "user_id":
		return cfg.UserID, cfg.UserID != ""
	case "api_version":
		return cfg.APIVersion, cfg.APIVersion != ""
	case "library_id":
		return cfg.LibraryID, cfg.LibraryID != ""
	case "plugin":
		return cfg.Plugin, cfg.Plugin != ""
	case "vertex_ai_project_id":
		return cfg.VertexAIProjectID, cfg.VertexAIProjectID != ""
	case "vertex_ai_adc":
		return cfg.VertexAIADC, cfg.VertexAIADC != ""
	case "auth_type":
		return cfg.AuthType, cfg.AuthType != ""
	case "api_format":
		return cfg.APIFormat, cfg.APIFormat != ""
	default:
		return nil, false
	}
}

func isEmpty(val interface{}) bool {
	switch v := val.(type) {
	case string:
		return v == ""
	case nil:
		return true
	}
	return false
}

func matchPattern(val interface{}, pattern string) bool {
	// TODO: 支持正则校验
	return true
}
