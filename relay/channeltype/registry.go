package channeltype

import "sync"

// ChannelTypeRegistry manages all channel types and their templates.
var (
	channelTypeRegistryMu sync.RWMutex
	channelTypeRegistry   = make(map[int]ChannelTypeInfoV2)
)

// RegisterChannelType registers or updates a channel type.
func RegisterChannelType(info ChannelTypeInfoV2) {
	channelTypeRegistryMu.Lock()
	defer channelTypeRegistryMu.Unlock()
	channelTypeRegistry[info.ID] = info
}

// GetChannelType returns a channel type by id.
func GetChannelType(id int) (ChannelTypeInfoV2, bool) {
	channelTypeRegistryMu.RLock()
	defer channelTypeRegistryMu.RUnlock()
	info, ok := channelTypeRegistry[id]
	return info, ok
}

// AllChannelTypes returns all registered channel types.
func AllChannelTypes() []ChannelTypeInfoV2 {
	channelTypeRegistryMu.RLock()
	defer channelTypeRegistryMu.RUnlock()
	out := make([]ChannelTypeInfoV2, 0, len(channelTypeRegistry))
	for _, info := range channelTypeRegistry {
		out = append(out, info)
	}
	return out
}

// ClearChannelTypes clears all registered channel types (for testing).
func ClearChannelTypes() {
	channelTypeRegistryMu.Lock()
	defer channelTypeRegistryMu.Unlock()
	channelTypeRegistry = make(map[int]ChannelTypeInfoV2)
}
