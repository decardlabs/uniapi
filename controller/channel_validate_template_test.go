package controller

import (
	"testing"

	"github.com/decardlabs/uniapi/model"
	"github.com/decardlabs/uniapi/relay/channeltype"
	"github.com/stretchr/testify/require"
)

func TestValidateChannelParamsByTemplate(t *testing.T) {
	channeltype.ClearChannelTypes()
	channeltype.RegisterChannelType(channeltype.ChannelTypeInfoV2{
		ID:       1001,
		Name:     "mock",
		Label:    "Mock类型",
		Category: "test",
		Template: channeltype.ChannelTypeTemplate{
			{Name: "region", Type: "string", Required: true},
			{Name: "sk", Type: "string", Required: false},
		},
	})
	ch := &model.Channel{Type: 1001, Config: `{"region":"cn-beijing"}`}
	err := ValidateChannelParamsByTemplate(ch)
	require.NoError(t, err)

	ch2 := &model.Channel{Type: 1001, Config: `{"region":""}`}
	err = ValidateChannelParamsByTemplate(ch2)
	require.Error(t, err)
	ch3 := &model.Channel{Type: 1001, Config: `{}`}
	err = ValidateChannelParamsByTemplate(ch3)
	require.Error(t, err)
}
