# UniAPI 国产 LLM 四家兼容性 Review

> 审查日期: 2025-05-25 | 对照版本: DeepSeek V4、Kimi K2.6/K2.5、智谱 GLM-5.1/4-Plus、MiniMax M2.7
> 审查范围: ChatCompletions API、流式、Function Calling、Thinking/推理、Vision、Claude Messages 转发、Response API、Structured Output

---

## 1. DeepSeek

### 官方 API 特性概要

| 特性 | 官方支持 | UniAPI 处理 | 兼容状态 |
|------|---------|------------|---------|
| ChatCompletions | ✅ OpenAI 兼容 | ✅ 直通 | ✅ 完全兼容 |
| Streaming SSE | ✅ 含 reasoning_content | ✅ unified streaming 处理 | ✅ 完全兼容 |
| Function Calling | ✅ tools/tool_calls，支持 strict 模式 | ✅ 直通 + tool message name 回填 | ✅ 完全兼容 |
| Thinking/推理 | ✅ thinking.type + reasoning_content | ✅ 深度适配：thinking normalization + reasoning_content 注入 | ✅ 最佳适配 |
| Vision/多模态 | ✅ image_url（文档未详述但支持） | ⚠️ 无特殊处理 | ⚠️ 透传，未验证 |
| Claude Messages | ✅ 代理支持 | ✅ ConvertClaudeRequest → OpenAI 格式 | ✅ 完全兼容 |
| Response API | ✅ 代理支持 | ✅ 路由到 /v1/chat/completions | ✅ 兼容 |
| Structured Output | ✅ json_object + json_schema | ⚠️ json_schema 被降级为 system instruction | ⚠️ 部分兼容 |
| logprobs | ✅ 支持 | ✅ 直通（OpenAI 兼容层处理） | ✅ 兼容 |
| frequency/presence_penalty | ❌ 已弃用 | ✅ 无需剥离（上游忽略） | ✅ 无影响 |
| top_k | ❌ 不支持 | ✅ 主动剥离 | ✅ 正确处理 |

### 代码亮点
- **deepseekcompat 模块**: 共享的 thinking normalization（adaptive → enabled/disabled）和 tool message content 标准化
- **V4 默认 thinking**: 代码正确处理了 DeepSeek V4 默认启用 thinking 的行为
- **ClaudeRequest 上下文恢复**: 从 ClaudeRequest 中读取原始 thinking 配置，避免 Claude→OpenAI 转换中丢失 thinking 参数

### 已知问题
1. **旧模型缺失**: constants.go 只有 v4-pro 和 v4-flash，缺少 deepseek-chat、deepseek-reasoner 等旧模型定价
2. **json_schema 被降级**: response_format.json_schema 被移除，改为在 system message 中注入指令。DeepSeek 官方文档未明确支持 json_schema type，此处理合理但意味着严格模式不可用
3. **Vision 未验证**: 代码中无 vision 相关特殊处理，依赖 OpenAI 兼容层透传

---

## 2. Kimi (Moonshot)

### 官方 API 特性概要

| 特性 | 官方支持 | UniAPI 处理 | 兼容状态 |
|------|---------|------------|---------|
| ChatCompletions | ✅ OpenAI 兼容 | ✅ 直通 | ✅ 完全兼容 |
| Streaming SSE | ✅ 含 reasoning_content | ✅ unified streaming 处理 | ✅ 完全兼容 |
| Function Calling | ✅ tools/tool_calls，MFJS 规范 | ✅ 直通 + tool message name 回填 | ✅ 兼容 |
| Thinking/推理 | ✅ thinking.type（K2.6/K2.5），K2 Thinking 内置 | ⚠️ **无 thinking normalization** | ⚠️ 部分兼容 |
| Vision/多模态 | ✅ image_url + video_url | ⚠️ 无特殊处理 | ⚠️ 透传，未验证 |
| Claude Messages | ✅ 代理支持 | ✅ ConvertClaudeRequest | ✅ 兼容 |
| Response API | ✅ 代理支持 | ✅ 路由到 /v1/chat/completions | ✅ 兼容 |
| Structured Output | ✅ json_object + json_schema（MFJS） | ⚠️ json_schema 被降级为 system instruction | ⚠️ 部分兼容 |
| top_k | ❌ 不支持 | ✅ 主动剥离 | ✅ 正确处理 |
| prompt_cache_key | ✅ Kimi 特有缓存优化 | ❌ 未传递 | ❌ 缺失 |
| safety_identifier | ✅ 用户安全标识 | ❌ 未传递 | ❌ 缺失 |

### 代码亮点
- 结构清晰，直接使用 OpenAI 兼容层，最小化自定义逻辑
- tool message name 回填处理正确

### 已知问题
1. **模型列表严重不足**: constants.go 只有 kimi-k2.6 一个模型，缺少 K2.5、K2、K2 Thinking、Moonshot V1 全系列
2. **Thinking 模式未适配**: K2.6 默认 `thinking.type=enabled`，K2 Thinking 内置推理。当前代码不处理 thinking normalization，从 Claude Messages 转发时 thinking 参数会丢失
3. **reasoning_content 未处理**: Kimi K2.6 的 thinking 模式返回 reasoning_content，当前代码依赖 OpenAI 兼容层透传，但未做 Thinking→reasoning_content 的映射
4. **Vision 透传未验证**: Moonshot V1 vision 和 K2.5 视觉能力的 image_url/video_url 格式未做任何验证
5. **缺少 prompt_cache_key 传递**: Kimi 官方推荐使用 prompt_cache_key 提升缓存命中率
6. **缺少 safety_identifier 传递**: 官方推荐的安全标识未传递
7. **API 端点**: BaseURL 配置为 `api.moonshot.cn/v1`，但 K2 系列实际使用 `api.kimi.com`（moonshot.cn 主要服务 Moonshot V1）

---

## 3. 智谱 GLM (Zhipu)

### 官方 API 特性概要

| 特性 | 官方支持 | UniAPI 处理 | 兼容状态 |
|------|---------|------------|---------|
| ChatCompletions | ✅ v4 OpenAI 兼容 / v3 私有格式 | ✅ 双版本自动路由（glm-*→v4，其他→v3） | ✅ 完全兼容 |
| Streaming SSE | ✅ 含 reasoning_content | ✅ v4 用 OpenAI handler / v3 用自有 handler | ✅ 完全兼容 |
| Function Calling | ✅ tools/tool_calls（仅 auto） | ✅ 直通 | ✅ 兼容 |
| Thinking/推理 | ✅ thinking.type + reasoning_content | ⚠️ **无 thinking normalization** | ⚠️ 部分兼容 |
| Vision/多模态 | ✅ image_url + video_url + file_url | ⚠️ 无特殊处理 | ⚠️ 透传 |
| Claude Messages | ✅ 代理支持 | ✅ ConvertClaudeRequest（含 v3/v4 分支） | ✅ 完全兼容 |
| Response API | ✅ 代理支持 | ⚠️ 路由到 /api/paas/v4/chat/completions | ✅ 兼容 |
| Structured Output | ✅ json_object | ⚠️ json_schema 被降级为 system instruction | ⚠️ 部分兼容 |
| OCR | ✅ layout_parsing 端点 | ✅ 独立 OCR adaptor + 路由 | ✅ 独有特性 |
| Embeddings | ✅ v4 embeddings 端点 | ✅ 独立 embedding handler | ✅ 独有特性 |
| 图片生成 | ✅ images/generations 端点 | ✅ ConvertImageRequest | ✅ 独有特性 |
| tool_choice | ⚠️ 仅支持 auto | ✅ 无需处理（OpenAI 兼容层自动处理） | ✅ 正确 |
| Temperature/TopP | ⚠️ 范围 [0,1] | ✅ 主动 clamp | ✅ 正确处理 |
| 搜索工具定价 | ✅ search_std/pro/pro_sogou/pro_quark | ✅ ZhipuToolingDefaults 已配置 | ✅ 完善 |

### 代码亮点
- **v3/v4 双版本自动路由**: 根据 model name 前缀（glm-* → v4）自动选择 API 格式，向后兼容
- **独有功能完整**: OCR、Embeddings、Image Generation 三个端点都有独立适配
- **搜索工具定价**: 已配置 web_search 工具的分级定价
- **Temperature/TopP clamp**: 正确限制在 [0,1] 范围

### 已知问题
1. **模型列表过少**: constants.go 只有 glm-5.1 一个模型，缺少 glm-4-plus、glm-4-air、glm-4-flash、glm-4-flashX、glm-4.6v 等全系列
2. **Thinking 未适配**: GLM-4.6V 支持 `thinking.type=enabled` + `reasoning_content`，但 adaptor 不做 thinking normalization
3. **Vision/文件理解未验证**: GLM-4.6V 的 image_url、video_url、file_url 多模态格式未做验证
4. **web_search 工具未透传**: 官方有 search_std/pro 等内置搜索工具，但 adaptor 不做特殊处理（依赖用户在 tools 中手动定义）

---

## 4. MiniMax

### 官方 API 特性概要

| 特性 | 官方支持 | UniAPI 处理 | 兼容状态 |
|------|---------|------------|---------|
| ChatCompletions | ✅ OpenAI 兼容 | ✅ 直通 | ✅ 完全兼容 |
| Streaming SSE | ✅ 标准格式 | ✅ unified streaming | ✅ 兼容 |
| Function Calling | ✅ tools/tool_calls | ✅ 直通 + tool message name 回填 | ✅ 兼容 |
| Thinking/推理 | ✅ MiniMax-M1 推理模型 | ⚠️ **无 thinking/reasoning 处理** | ⚠️ 未适配 |
| Vision/多模态 | ✅（文档获取失败，需确认） | ⚠️ 无处理 | ⚠️ 未验证 |
| Claude Messages | ✅ 代理支持 | ✅ ConvertClaudeRequest | ✅ 兼容 |
| Response API | ✅ 代理支持 | ✅ 路由到 /v1/chat/completions | ✅ 兼容 |
| Structured Output | ⚠️ 未确认 | ⚠️ json_schema 被降级为 system instruction | ⚠️ 部分兼容 |
| top_k | ❌ 不支持 | ✅ 主动剥离 | ✅ 正确处理 |
| API 端点 | 国内 api.minimaxi.com / 海外 api.minimax.io | ⚠️ BaseURL 默认 api.minimaxi.com，可编辑 | ✅ 兼容 |

### 代码亮点
- 最简洁的 adaptor 实现之一
- tool message name 回填处理正确
- BaseURL 配置为可编辑（`Editable: true`），方便用户切换国内外端点

### 已知问题
1. **模型列表不完整**: 缺少 MiniMax-M2、MiniMax-M2.1、MiniMax-M2.1-highspeed、MiniMax-M2.5-highspeed
2. **M1 推理模型未适配**: MiniMax-M1 是推理模型，可能有 thinking/reasoning_content，当前代码完全不处理
3. **Vision 完全未验证**: MiniMax 文档无法完整获取，vision 支持情况不明
4. **官方文档获取困难**: MiniMax 的 API 文档为 SPA/动态渲染，WebFetch 无法获取完整内容，增加了审查难度

---

## 5. 横向对比总结

### 兼容性评分（满分 10 分）

| 维度 | DeepSeek | Kimi | 智谱 GLM | MiniMax |
|------|---------|------|---------|---------|
| ChatCompletions 基础 | 10 | 9 | 9 | 9 |
| Streaming SSE | 10 | 9 | 9 | 9 |
| Function Calling | 9 | 8 | 8 | 8 |
| Thinking/推理 | **10** | 5 | 5 | 2 |
| Vision/多模态 | 5（未验证） | 5（未验证） | 5（未验证） | 3（未知） |
| Claude Messages 转发 | 9 | 8 | 9 | 8 |
| Response API 转发 | 9 | 8 | 8 | 8 |
| Structured Output | 5 | 5 | 5 | 4 |
| 模型覆盖度 | 4 | 2 | 2 | 6 |
| 独有功能适配 | 6（thinking 深度适配） | 3 | **9**（OCR+Embedding+Image） | 4 |
| **综合评分** | **77/100** | **62/100** | **69/100** | **61/100** |

### 优先修复建议

| 优先级 | 问题 | 涉及厂商 | 影响 |
|--------|------|---------|------|
| **P0** | 模型列表补全 | Kimi、GLM、MiniMax | 用户无法看到/选择大部分模型 |
| **P0** | Thinking 模式适配 | Kimi、GLM | K2.6/GLM-4.6V 默认 thinking，无适配会导致推理异常 |
| **P1** | Vision 格式验证 | 全部 | 多模态请求可能静默失败 |
| **P1** | json_schema response_format 降级 | 全部 | 结构化输出能力受限 |
| **P2** | Kimi prompt_cache_key | Kimi | 缓存命中率低，成本增加 |
| **P2** | MiniMax M1 推理模型 | MiniMax | 推理能力无法利用 |
| **P3** | Kimi API 端点区分 | Kimi | K2 系列应使用 api.kimi.com 而非 api.moonshot.cn |

---

## 6. 架构层面的观察

### 做得好的地方
- **共享 deepseekcompat 模块**: thinking normalization 和 tool message 标准化被抽象为共享库，DeepSeek/VertexAI-DeepSeek/OpenAI-Compatible 都复用
- **structuredjson 模块**: json_schema 降级为 system instruction 的处理是合理的降级策略
- **Claude Messages 三路转发**: ChatCompletions、Claude Messages、Response API 三种入口格式都能正确路由到上游 OpenAI 兼容 API
- **OpenAI 兼容层**: unified streaming 的 thinking processor 能正确处理 reasoning_content/reasoning/thinking 三种格式

### 需要改进的方向
- **Kimi/GLM 缺少 thinking 适配**: 可以复用 deepseekcompat 模块的 thinking normalization 逻辑
- **Vision 兼容性需要端到端测试**: 四家的 image_url 格式可能有细微差异（如 base64 前缀、URL vs data URI）
- **模型列表维护机制**: 当前模型列表硬编码在 constants.go 中，应该考虑定期同步或提供更新机制
