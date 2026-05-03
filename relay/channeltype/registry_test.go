package channeltype

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelTypeRegistry(t *testing.T) {
	ClearChannelTypes()
	info := ChannelTypeInfoV2{
		ID:    100,
		Name:  "test",
		Label: "测试类型",
		Category: "test_group",
		Description: "测试用类型",
		Template: ChannelTypeTemplate{
			{Name: "api_base", Type: "string", Required: true, Default: "http://test", Description: "base url"},
		},
	}
	RegisterChannelType(info)
	got, ok := GetChannelType(100)
	require.True(t, ok)
	require.Equal(t, "test", got.Name)
	all := AllChannelTypes()
	found := false
	for _, v := range all {
		if v.ID == 100 {
			found = true
		}
	}
	require.True(t, found)
	ClearChannelTypes()
	_, ok = GetChannelType(100)
	require.False(t, ok)
}
