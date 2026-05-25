# UniAPI 模型接入渠道兼容性 Review 报告

**审查日期**: 2026-05-25  
**审查范围**: DeepSeek、MiniMax、Moonshot(Kimi)、Zhipu(GLM) 四种模型的接入渠道  
**审查目标**: 验证对 ChatCompletions API、Response API、Claude Messages API 三种格式的兼容性

---

## 1. 执行摘要

| 模型通道 | ChatCompletions | Response API | Claude Messages | 综合评分 |
|---------|:-:|:-:|:-:|:-:|
| **DeepSeek** | ✅ 完整支持 | ❌ 未注册 | ✅ 完整支持 | **B+** |
| **MiniMax** | ✅ 完整支持 | ✅ 已注册 | ✅ 完整支持 | **A-** |
| **Moonshot** | ✅ 完整支持 | ❌ 未注册 | ✅ 完整支持 | **B+** |
| **Zhipu GLM** | ✅ 完整支持 | ❌ 未注册 | ⚠️ 部分支持 | **B** |

---

## 2. 详细兼容性分析

### 2.1 DeepSeek 适配器

**文件位置**: `relay/adaptor/deepseek/`

#### 接口实现情况

| 接口方法 | 状态 | 说明 |
|---------|:-:|------|
| `Init` | ✅ | 空实现 |
| `GetRequestURL` | ✅ | 支持 `/v1/messages` → `/v1/chat/completions` 转换 |
| `SetupRequestHeader` | ✅ | Bearer Token 认证 |
| `ConvertRequest` | ✅ | 移除 `reasoning_effort`，规范化 thinking 配置 |
| `ConvertImageRequest` | ✅ | 明确返回不支持错误 |
| `ConvertClaudeRequest` | ✅ | 使用 `openai_compatible.ConvertClaudeRequest` + DeepSeek 特殊处理 |
| `DoRequest` | ✅ | 使用标准 `DoRequestHelper` |
| `DoResponse` | ✅ | 支持 Claude 转换响应处理 + 流式/非流式 |
| `GetModelList` | ✅ | 从定价配置动态生成 |
| `GetChannelName` | ✅ | 返回 "deepseek" |
| `GetDefaultModelPricing` | ✅ | 完整定价信息 |
| `DefaultToolingConfig` | ✅ | 返回 `DeepseekToolingDefaults` |

#### Response API 支持

- **注册状态**: ❌ **未在 `main.go` 中注册 ResponseAPI 支持**
- **`GetRequestURL`**: 未处理 `relaymode.ResponseAPI` 分支
- **影响**: 当客户端直接向 DeepSeek 通道发送 `/v1/responses` 请求时，会返回 "unsupported relay mode" 错误

#### Claude Messages API 兼容性

| 功能点 | 状态 | 详细说明 |
|--------|:-:|------|
| 请求转换 | ✅ | 使用 `openai_compatible.ConvertClaudeRequest` 核心转换逻辑 |
| Thinking 处理 | ✅ | 特殊的 thinking 配置规范化 (`normalizeDeepSeekThinkingConfigFromOriginal`) |
| Tool Use 转换 | ✅ | 正确处理 tool_use → function 转换 |
| Tool Result 处理 | ✅ | 使用 `buildToolUseNames` 映射 tool_use ID → name |
| 流式响应 | ✅ | 通过 `openai_compatible.StreamHandler` 处理 |
| 非流式响应 | ✅ | 通过 `openai_compatible.Handler` 处理 |
| 结构化输出提升 | ⚠️ | **已禁用** (`structuredPromotionDisabled` 对 DeepSeek 返回 true) |
| reasoning_content 注入 | ✅ | `injectMissingReasoningContent` 处理历史消息 |

#### 特殊功能支持

| 功能 | 支持情况 | 备注 |
|------|---------|------|
| Thinking 模式 | ✅ | 支持 `thinking.type` = "enabled"/"disabled"，处理 budget_tokens |
| 多轮对话 | ✅ | 正确处理历史消息中的 reasoning_content |
| 工具调用 | ✅ | 支持单工具结构化输出（但被主动禁用） |
| 流式输出 | ✅ | SSE 格式 |
| 缓存提示 | ⚠️ | 定价配置中定义了 CacheWrite5mRatio/CacheWrite1hRatio，但需确认上游是否支持 |

#### 已知问题

1. **Response API 未注册**: `main.go` 中未包含 `relaymode.ResponseAPI` 分支
2. **结构化输出被禁用**: `structuredPromotionDisabled` 对 DeepSeek 返回 true，可能错过优化机会
3. **thinking 参数清理**: `ConvertRequest` 中先提取 thinking 配置，然后设置 `request.Thinking = nil`，但 `ConvertClaudeRequest` 中又通过 `c.Get(ctxkey.OriginalClaudeRequest)` 读取原始请求

---

### 2.2 MiniMax 适配器

**文件位置**: `relay/adaptor/minimax/`

#### 接口实现情况

| 接口方法 | 状态 | 说明 |
|---------|:-:|------|
| `Init` | ✅ | 空实现 |
| `GetRequestURL` | ✅ | **同时支持** ChatCompletions、ClaudeMessages、**ResponseAPI** |
| `SetupRequestHeader` | ✅ | Bearer Token 认证 |
| `ConvertRequest` | ✅ | 移除 `reasoning_effort` |
| `ConvertImageRequest` | ✅ | 明确返回不支持错误 |
| `ConvertClaudeRequest` | ✅ | 直接委托 `openai_compatible.ConvertClaudeRequest` |
| `DoRequest` | ✅ | 使用标准 `DoRequestHelper` |
| `DoResponse` | ✅ | 支持 Claude 转换响应处理，使用 `HandleClaudeMessagesResponse` |
| `GetModelList` | ✅ | 从定价配置动态生成 |
| `GetChannelName` | ✅ | 返回 "minimax" |
| `GetDefaultModelPricing` | ✅ | 完整定价信息 |
| `DefaultToolingConfig` | ✅ | 返回 `MoonshotToolingDefaults`（命名有误，应为 `MinimaxToolingDefaults`） |

#### Response API 支持

- **注册状态**: ✅ **已在 `main.go` 中注册**
- **`GetRequestURL`**: 正确处理 `relaymode.ResponseAPI`，路由到 `/v1/chat/completions`
- **影响**: MiniMax 是目前四个适配器中**唯一**完整注册 Response API 的通道

#### Claude Messages API 兼容性

| 功能点 | 状态 | 详细说明 |
|--------|:-:|------|
| 请求转换 | ✅ | 使用 `openai_compatible.ConvertClaudeRequest` 核心转换逻辑 |
| Thinking 处理 | ❌ | 未提及 MiniMax 支持 thinking，可能不支持 |
| Tool Use 转换 | ✅ | 使用共享逻辑，包含 `buildToolUseNames` 映射 |
| Tool Result 处理 | ⚠️ | **缺少 `backfillToolMessageNamesFromToolCalls`** 特殊处理 |
| 流式响应 | ✅ | 通过 `openai_compatible.StreamHandler` 处理 |
| 非流式响应 | ✅ | 通过 `openai_compatible.Handler` 处理 |
| 结构化输出提升 | ✅ | 支持（未在 `structuredPromotionDisabled` 中禁用） |

#### 特殊功能支持

| 功能 | 支持情况 | 备注 |
|------|---------|------|
| Thinking 模式 | ❌ | MiniMax 官方 API 未声明支持 |
| 多轮对话 | ✅ | 通过标准 OpenAI 格式 |
| 工具调用 | ✅ | 支持 function calling |
| 流式输出 | ✅ | SSE 格式 |
| Tool Message 名称回填 | ❌ | 未实现 `backfillToolMessageNamesFromToolCalls` |

#### 已知问题

1. **Tool Message 名称回填缺失**: `openai_compatible.ConvertClaudeRequest` 在转换时会丢失 tool_result 的 name 字段，MiniMax 适配器未像某些实现那样补充 `backfillToolMessageNamesFromToolCalls` 逻辑
2. **命名不一致**: `DefaultToolingConfig()` 返回 `MoonshotToolingDefaults`（应为 `MinimaxToolingDefaults`）
3. **Response API 实际支持存疑**: 虽然注册了 ResponseAPI 模式，但 `DoResponse` 并未区分 ResponseAPI 和 ChatCompletions 的响应格式差异

---

### 2.3 Moonshot(Kimi) 适配器

**文件位置**: `relay/adaptor/moonshot/`

#### 接口实现情况

| 接口方法 | 状态 | 说明 |
|---------|:-:|------|
| `Init` | ✅ | 空实现 |
| `GetRequestURL` | ⚠️ | 仅处理 ClaudeMessages → ChatCompletions，未处理 ResponseAPI |
| `SetupRequestHeader` | ✅ | Bearer Token 认证 |
| `ConvertRequest` | ✅ | 移除 `reasoning_effort` |
| `ConvertImageRequest` | ✅ | 明确返回不支持错误 |
| `ConvertClaudeRequest` | ✅ | 直接委托 `openai_compatible.ConvertClaudeRequest` |
| `DoRequest` | ✅ | 使用标准 `DoRequestHelper` |
| `DoResponse` | ✅ | 支持 Claude 转换响应处理，使用 `HandleClaudeMessagesResponse` |
| `GetModelList` | ✅ | 从定价配置动态生成 |
| `GetChannelName` | ✅ | 返回 "moonshot" |
| `GetDefaultModelPricing` | ✅ | 完整定价信息 |
| `DefaultToolingConfig` | ✅ | 返回 `MoonshotToolingDefaults` |

#### Response API 支持

- **注册状态**: ❌ **未在 `main.go` 中注册 ResponseAPI 支持**
- **`GetRequestURL`**: 未处理 `relaymode.ResponseAPI` 分支
- **影响**: 与 DeepSeek 类似，不支持直接 Response API 请求

#### Claude Messages API 兼容性

| 功能点 | 状态 | 详细说明 |
|--------|:-:|------|
| 请求转换 | ✅ | 使用 `openai_compatible.ConvertClaudeRequest` 核心转换逻辑 |
| Thinking 处理 | ❌ | Moonshot 不支持 thinking/reasoning |
| Tool Use 转换 | ✅ | 使用共享逻辑 |
| Tool Result 处理 | ✅ | 使用 `buildToolUseNames` 映射 |
| 流式响应 | ✅ | 通过 `openai_compatible.StreamHandler` 处理 |
| 非流式响应 | ✅ | 通过 `openai_compatible.Handler` 处理 |
| 结构化输出提升 | ✅ | 支持（未在 `structuredPromotionDisabled` 中禁用） |

#### 特殊功能支持

| 功能 | 支持情况 | 备注 |
|------|---------|------|
| Thinking 模式 | ❌ | Moonshot API 不支持 |
| 多轮对话 | ✅ | 通过标准 OpenAI 格式 |
| 工具调用 | ✅ | 支持 function calling |
| 流式输出 | ✅ | SSE 格式 |
| Moonshot 特殊参数 | ⚠️ | 仅移除了 `reasoning_effort`，未检查其他 Moonshot 特定参数 |

#### 已知问题

1. **Response API 未注册**: 与 DeepSeek 相同的问题
2. **过度依赖 `openai_compatible`**: 没有针对 Moonshot 的特殊处理（如 API 特定参数、错误码处理等）
3. **`GetRequestURL` 可以改进**: 虽然处理了 `/v1/messages` 转换，但逻辑较简单，未考虑边缘情况

---

### 2.4 Zhipu GLM 适配器

**文件位置**: `relay/adaptor/zhipu/`

#### 接口实现情况

| 接口方法 | 状态 | 说明 |
|---------|:-:|------|
| `Init` | ✅ | 空实现 |
| `GetRequestURL` | ⚠️ | 复杂逻辑：支持 v3/v4 API，处理 Embeddings、OCR、Images Generations |
| `SetupRequestHeader` | ✅ | 使用 `GetToken(meta.APIKey)` 特殊认证方式 |
| `ConvertRequest` | ⚠️ | 部分委托，部分自实现；对 v3 API 使用 `ConvertRequest` 自实现 |
| `ConvertImageRequest` | ✅ | 支持，转换为 GLM 图像生成格式 |
| `ConvertClaudeRequest` | ⚠️ | **自实现**，未使用 `openai_compatible.ConvertClaudeRequest` |
| `DoRequest` | ✅ | 使用标准 `DoRequestHelper` |
| `DoResponse` | ⚠️ | 复杂逻辑：根据 API 版本和模式分发到不同处理器 |
| `ConvertOCRRequest` | ✅ | 支持 OCR 请求转换 (`OCRAdaptor` 接口) |
| `DoOCRResponse` | ✅ | 支持 OCR 响应处理 (`OCRAdaptor` 接口) |
| `GetModelList` | ✅ | 从定价配置动态生成 |
| `GetChannelName` | ✅ | 返回 "zhipu" |
| `GetDefaultModelPricing` | ✅ | 完整定价信息（包含多种模型） |
| `DefaultToolingConfig` | ✅ | 返回 `ZhipuToolingDefaults` |

#### Response API 支持

- **注册状态**: ❌ **未在代码中找到 ResponseAPI 支持**
- **`GetRequestURL`**: 未处理 `relaymode.ResponseAPI` 分支
- **影响**: Zhipu 适配器不支持 Response API

#### Claude Messages API 兼容性

| 功能点 | 状态 | 详细说明 |
|--------|:-:|------|
| 请求转换 | ⚠️ | **自实现** `ConvertClaudeRequest`，与 `openai_compatible` 逻辑可能不同步 |
| Thinking 处理 | ❌ | GLM API 未声明支持 thinking |
| Tool Use 转换 | ⚠️ | 自实现，转换逻辑与标准可能有差异 |
| Tool Result 处理 | ⚠️ | 自实现，未使用 `buildToolUseNames` 映射 |
| 流式响应 | ✅ | v4 使用 `openai.StreamHandler`，v3 使用自实现 |
| 非流式响应 | ✅ | 根据 API 版本分发处理 |
| 结构化输出提升 | ❓ | 未明确处理 |

#### Zhipu 自实现 Claude Messages 转换分析

**文件**: `relay/adaptor/zhipu/adaptor.go` (第 111-253 行)

**优点**:
1. 正确处理了 system prompt（支持 string 和结构化格式）
2. 处理了 image 类型内容块（转换为 OpenAI 的 `image_url` 格式）
3. 转换后调用 `a.ConvertRequest(c, relaymode.ChatCompletions, openaiRequest)` 应用 Zhipu 特定逻辑

**缺点/风险**:
1. **未使用 `buildToolUseNames`**: 原生实现中 tool_result 消息可能缺少 name 字段
2. **工具转换不完整**: 仅转换了 `Name` 和 `Description`，未处理 `InputSchema`
3. **未处理 `tool_choice`**: 转换逻辑中未包含 `tool_choice` 的转换
4. **与 `openai_compatible` 不同步**: 当核心转换逻辑更新时，Zhipu 的自实现可能不会同步更新
5. **未设置 `ctxkey.ClaudeMessagesConversion`**: 虽然设置了，但时机可能不对

#### 特殊功能支持

| 功能 | 支持情况 | 备注 |
|------|---------|------|
| Thinking 模式 | ❌ | GLM API 不支持 |
| 多轮对话 | ✅ | 通过转换后的 OpenAI 格式 |
| 工具调用 | ⚠️ | 自实现，可能与标准有差异 |
| 流式输出 | ✅ | v4 使用 OpenAI 兼容格式，v3 自实现 |
| 图像生成 | ✅ | 支持 `/v1/images/generations` → GLM API |
| Embeddings | ✅ | 支持 `/v1/embeddings` → GLM API |
| OCR/Layout Parsing | ✅ | 支持 `/v1/layout_parsing` → GLM API (v4) |
| 多版本 API | ✅ | 同时支持 v3 (`/api/paas/v3/`) 和 v4 (`/api/paas/v4/`) |

#### 已知问题

1. **Claude Messages 转换自实现**: 与 `openai_compatible` 包不同步，可能缺少最新功能
2. **Response API 未注册**: 四个适配器中唯一支持 OCR 的，但也不支持 Response API
3. **工具调用转换不完整**: 缺少 `tool_choice` 和完整的 `InputSchema` 处理
4. **代码重复**: 自实现 `ConvertClaudeRequest` 导致与 `openai_compatible` 包的代码重复
5. **`DoResponse` 逻辑复杂**: 大量条件分支，难以维护

---

## 3. 三种 API 格式转换分析

### 3.1 ChatCompletions API (OpenAI 标准)

所有四个适配器都**完整支持** ChatCompletions API。

| 适配器 | 请求处理 | 响应处理 | 流式支持 | 备注 |
|--------|---------|---------|---------|------|
| DeepSeek | ✅ | ✅ | ✅ | 特殊处理 thinking |
| MiniMax | ✅ | ✅ | ✅ | 标准实现 |
| Moonshot | ✅ | ✅ | ✅ | 标准实现 |
| Zhipu | ✅ | ✅ | ✅ | v3/v4 不同处理 |

### 3.2 Response API (OpenAI 新标准)

| 适配器 | 注册状态 | 请求转换 | 响应转换 | 备注 |
|--------|---------|---------|---------|------|
| DeepSeek | ❌ 未注册 | N/A | N/A | `main.go` 缺少 `relaymode.ResponseAPI` |
| MiniMax | ✅ 已注册 | ⚠️ 同 ChatCompletions | ⚠️ 未区分格式 | `main.go` 已注册，但处理逻辑未区分 |
| Moonshot | ❌ 未注册 | N/A | N/A | `main.go` 缺少 `relaymode.ResponseAPI` |
| Zhipu | ❌ 未注册 | N/A | N/A | 未找到相关代码 |

**`openai_compatible/claude_convert.go` 中的 Response API → Claude 转换**:
- 已实现 `ConvertOpenAIResponseToClaudeResponse` 函数
- 支持 Response API 格式 → Claude Messages 格式转换
- 支持 ChatCompletions 格式 → Claude Messages 格式转换
- 但四个适配器都未充分利用此能力

### 3.3 Claude Messages API (Anthropic 标准)

| 适配器 | 请求转换 | 响应转换 | Thinking 支持 | 工具调用 | 备注 |
|--------|---------|---------|-------------|---------|------|
| DeepSeek | ✅ `openai_compatible` | ✅ `HandleClaudeMessagesResponse` | ✅ 特殊处理 | ✅ 完整 | 最完整实现 |
| MiniMax | ✅ `openai_compatible` | ✅ `HandleClaudeMessagesResponse` | ❌ 不支持 | ⚠️ 缺少回填 | 标准实现 |
| Moonshot | ✅ `openai_compatible` | ✅ `HandleClaudeMessagesResponse` | ❌ 不支持 | ✅ 完整 | 标准实现 |
| Zhipu | ⚠️ 自实现 | ✅ 分发处理 | ❌ 不支持 | ⚠️ 不完整 | 需重写为使用 `openai_compatible` |

---

## 4. 核心转换逻辑文件分析

### 4.1 `openai_compatible/claude_messages.go`

**核心功能**: Claude Messages → OpenAI ChatCompletions 转换

| 函数 | 功能 | 调用者 |
|------|------|-------|
| `ConvertClaudeRequest` | 主转换函数 | DeepSeek, MiniMax, Moonshot |
| `buildToolUseNames` | 预扫描构建 tool_use ID → name 映射 | `ConvertClaudeRequest` |
| `convertClaudeMessageToOpenAI` | 转换单条消息 | `ConvertClaudeRequest` |
| `detectStructuredToolSchema` | 检测结构化输出并提升 | `ConvertClaudeRequest` |
| `normalizeClaudeToolChoice` | 规范化 tool_choice | `ConvertClaudeRequest` |

### 4.2 `openai_compatible/claude_convert.go`

**核心功能**: OpenAI → Claude Messages 反向转换（用于响应处理）

| 函数 | 功能 | 说明 |
|------|------|------|
| `ConvertOpenAIResponseToClaudeResponse` | 主转换函数 | 支持 Response API 和 ChatCompletions 两种格式 |
| `responseAPIResponseToClaude` | Response API → Claude | 处理 `output` 数组中的 message/reasoning/function_call |
| `chatResponseToClaude` | ChatCompletions → Claude | 处理 `choices` 数组 |

### 4.3 `openai_compatible/handler.go` (推测)

**核心功能**: 响应处理

| 函数 | 功能 | 说明 |
|------|------|------|
| `HandleClaudeMessagesResponse` | Claude 转换响应的统一入口 | 判断是否为 Claude 转换，路由到对应处理器 |
| `StreamHandler` | 流式响应处理 | SSE 格式 |
| `Handler` | 非流式响应处理 | JSON 格式 |

---

## 5. 问题汇总与优先级

### P0 - 必须修复

| 问题 | 影响范围 | 描述 |
|------|---------|------|
| **Zhipu `ConvertClaudeRequest` 自实现** | Zhipu | 自实现与 `openai_compatible` 不同步，可能缺少关键功能（如 `buildToolUseNames`） |

**建议**: 将 Zhipu 的 `ConvertClaudeRequest` 改为先调用 `openai_compatible.ConvertClaudeRequest`，再应用 Zhipu 特定逻辑（类似 DeepSeek 的做法）。

### P1 - 强烈建议修复

| 问题 | 影响范围 | 描述 |
|------|---------|------|
| **Response API 未注册** | DeepSeek, Moonshot, Zhipu | 三个适配器未在 `main.go` 中注册 `relaymode.ResponseAPI` |
| **MiniMax 缺少 tool message 名称回填** | MiniMax | 未实现 `backfillToolMessageNamesFromToolCalls` 逻辑 |
| **MiniMax `DefaultToolingConfig` 命名错误** | MiniMax | 返回 `MoonshotToolingDefaults`，应为 `MinimaxToolingDefaults` |

### P2 - 建议改进

| 问题 | 影响范围 | 描述 |
|------|---------|------|
| **DeepSeek 结构化输出被禁用** | DeepSeek | `structuredPromotionDisabled` 对 DeepSeek 返回 true，需确认是否合理 |
| **Response API 实际处理逻辑缺失** | MiniMax | 虽然注册了 ResponseAPI，但 `DoResponse` 未区分 Response API 和 ChatCompletions 的响应格式 |
| **Zhipu `DoResponse` 逻辑复杂** | Zhipu | 大量条件分支，建议重构 |
| **缺少集成测试** | 所有 | 未找到针对 Claude Messages → Response API → Claude Messages 的完整闭环测试 |

---

## 6. 修复建议

### 6.1 Zhipu `ConvertClaudeRequest` 重构

**当前实现**:
```go
func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
    // 自实现转换逻辑 (111-253 行)
    openaiRequest := &model.GeneralOpenAIRequest{...}
    // ... 大量自实现代码 ...
    return a.ConvertRequest(c, relaymode.ChatCompletions, openaiRequest)
}
```

**建议实现**:
```go
func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
    // 1. 使用共享的 openai_compatible 转换
    converted, err := openai_compatible.ConvertClaudeRequest(c, request)
    if err != nil {
        return nil, errors.Wrap(err, "convert claude request")
    }
    
    openaiReq, ok := converted.(*model.GeneralOpenAIRequest)
    if !ok {
        return converted, nil
    }
    
    // 2. 应用 Zhipu 特定逻辑（从原实现保留）
    // TopP [0.0, 1.0]
    openaiReq.TopP = helper.Float64PtrMax(openaiReq.TopP, 1)
    openaiReq.TopP = helper.Float64PtrMin(openaiReq.TopP, 0)
    
    // Temperature [0.0, 1.0]
    openaiReq.Temperature = helper.Float64PtrMax(openaiReq.Temperature, 1)
    openaiReq.Temperature = helper.Float64PtrMin(openaiReq.Temperature, 0)
    
    a.SetVersionByModelName(openaiReq.Model)
    if a.APIVersion == "v4" {
        return openaiReq, nil
    }
    return ConvertRequest(*openaiReq), nil  // v3 需要额外转换
}
```

### 6.2 注册 Response API 支持

**以 DeepSeek 为例** (`relay/adaptor/deepseek/main.go`):

```go
func GetRequestURL(meta *meta.Meta) (string, error) {
    requestPath := meta.RequestURLPath
    if idx := strings.Index(requestPath, "?"); idx >= 0 {
        requestPath = requestPath[:idx]
    }
    switch requestPath {
    case "/v1/messages":
        return openai_compatible.GetFullRequestURL(meta.BaseURL, "/v1/chat/completions", 0), nil
    case "/v1/responses":  // 新增
        return openai_compatible.GetFullRequestURL(meta.BaseURL, "/v1/chat/completions", 0), nil
    }
    return openai_compatible.GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
}
```

或者在 `main.go` 中添加 `relaymode.ResponseAPI`:

```go
// relay/adaptor/deepseek/main.go
func init() {
    adaptor.Register(channeltype.DeepSeek, func() adaptor.Adaptor {
        return &Adaptor{}
    })
    
    // 添加 ResponseAPI 支持
    relaymode.RegisterEndpoint(channeltype.DeepSeek, relaymode.ResponseAPI, "/v1/responses")
}
```

### 6.3 MiniMax Tool Message 名称回填

参考 `openai_compatible/claude_messages.go` 中的实现，在 MiniMax 的 `ConvertRequest` 中添加:

```go
func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
    // 移除 reasoning_effort
    if request.ReasoningEffort != nil {
        request.ReasoningEffort = nil
    }
    
    // 添加 tool message 名称回填
    backfillToolMessageNamesFromToolCalls(request)
    
    return request, nil
}

func backfillToolMessageNamesFromToolCalls(request *model.GeneralOpenAIRequest) {
    // 实现参考 openai_compatible/claude_messages.go 中的逻辑
    // ...
}
```

---

## 7. 测试建议

### 7.1 单元测试

| 测试场景 | 适配器 | 优先级 |
|---------|--------|:-:|
| Claude Messages → OpenAI 转换 | 所有 | P0 |
| Response API → OpenAI 转换 | MiniMax (已注册) | P0 |
| Thinking 处理 | DeepSeek | P0 |
| Tool message 名称回填 | MiniMax, Zhipu | P1 |
| v3/v4 API 分发 | Zhipu | P1 |

### 7.2 集成测试

1. **完整闭环测试**: Claude Messages → OpenAI → Claude Messages
2. **Response API 测试**: 直接向各适配器发送 `/v1/responses` 请求
3. **多轮对话测试**: 确保 thinking/reasoning_content 正确处理
4. **工具调用测试**: 测试 tool_use → tool_result 完整流程

---

## 8. 总结

### 8.1 做得好的地方

1. **DeepSeek 适配器**: Thinking 处理非常完善，考虑了多种边缘情况
2. **MiniMax 适配器**: 唯一注册了 Response API 支持（虽然处理逻辑待完善）
3. **共享转换逻辑**: `openai_compatible` 包提供了高质量的共享转换逻辑
4. **Zhipu 多版本支持**: 同时支持 v3 和 v4 API，功能丰富

### 8.2 需要改进的地方

1. **Zhipu 自实现风险**: `ConvertClaudeRequest` 自实现导致与标准不同步
2. **Response API 支持不完整**: 仅 MiniMax 注册，但处理逻辑未完善
3. **代码重复**: Zhipu 的自实现与 `openai_compatible` 存在大量重复
4. **缺少集成测试**: 未找到针对三种格式互转的完整测试

### 8.3 行动建议

| 优先级 | 行动项 | 负责人 | 预计工作量 |
|--------|--------|--------|------------|
| P0 | 重构 Zhipu `ConvertClaudeRequest` 使用 `openai_compatible` | 待定 | 2-3 小时 |
| P1 | 为 DeepSeek、Moonshot、Zhipu 注册 Response API | 待定 | 各 0.5 小时 |
| P1 | 修复 MiniMax tool message 名称回填 | 待定 | 1 小时 |
| P2 | 完善 Response API 响应处理逻辑 | 待定 | 3-4 小时 |
| P2 | 添加集成测试 | 待定 | 4-6 小时 |

---

**报告生成时间**: 2026-05-25 17:02  
**审查人**: 砖家 (AI Assistant)
