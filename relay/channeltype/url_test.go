package channeltype

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestChannelBaseURLs(t *testing.T) {
	Convey("channel base urls", t, func() {
		So(len(ChannelBaseURLs), ShouldEqual, Dummy)
	})
}

// TestChannelBaseURLConfigs_Release37Channels ensures 3.7 release channels keep the confirmed default base URLs.
func TestChannelBaseURLConfigs_Release37Channels(t *testing.T) {
	Convey("release 3.7 channel base url defaults", t, func() {
		So(GetChannelBaseURLConfig(OpenRouter).URL, ShouldEqual, "https://openrouter.ai/api/v1")
		So(GetChannelBaseURLConfig(Minimax).URL, ShouldEqual, "https://api.minimax.chat/v1")
		So(GetChannelBaseURLConfig(DeepSeek).URL, ShouldEqual, "https://api.deepseek.com/v1")
		So(GetChannelBaseURLConfig(GLM).URL, ShouldEqual, "https://open.bigmodel.cn/api/paas/v4")
		So(GetChannelBaseURLConfig(Moonshot).URL, ShouldEqual, "https://api.moonshot.cn/v1")
		So(GetChannelBaseURLConfig(Doubao).URL, ShouldEqual, "https://ark.cn-beijing.volces.com/api/v3")
		So(GetChannelBaseURLConfig(SiliconFlow).URL, ShouldEqual, "https://api.siliconflow.cn/v1")
		So(GetChannelBaseURLConfig(Ali).URL, ShouldEqual, "https://dashscope.aliyuncs.com/compatible-mode/v1")
		So(GetChannelBaseURLConfig(TencentTokenHub).URL, ShouldEqual, "https://api.lkeap.cloud.tencent.com/TokenHub/v3")
	})
}
