# UniAPI 三端点兼容性测试白皮书

> 版本: 2.0 | 日期: 2026-05-29 | 状态: Draft
> 适用范围: UniAPI 三端点格式转换兼容性 + 中转站 vs 原厂对比测试

---

## 1. 概述

### 1.1 背景与目标

UniAPI 的核心使命是实现 **Chat Completions**、**Response API**、**Claude Messages** 三种 API 格式的透明互转。用户使用任意格式请求，UniAPI 自动路由到上游渠道并转换响应，保证用户始终收到与请求格式一致的响应。

本白皮书旨在：

1. 建立三端点兼容性的系统化测试框架
2. 识别当前测试覆盖的缺口与风险
3. 提供可执行的测试计划与用例设计
4. 为持续集成提供回归测试策略
5. 建立 UniAPI 中转站与原厂 API 的对比测试体系，验证中转透明性

### 1.2 两大测试维度

### 1.3 三端点定位

| 端点 | 路径 | 生态 | 代表客户端 |
|---|---|---|---|
| Chat Completions | `/v1/chat/completions` | OpenAI 生态，业界事实标准 | Cursor, Copilot, Continue |
| Response API | `/v1/responses` | OpenAI 新版有状态 API | Codex, OpenAI Agents SDK |
| Claude Messages | `/v1/messages` | Anthropic 生态 | Claude Code, Claude Desktop |

### 1.4 转换架构总览

```
                        ┌──────────────────┐
                        │   UniAPI 入口     │
                        └────────┬─────────┘
                                 │
                    ┌────────────┼────────────┐
                    ▼            ▼            ▼
            /v1/chat/      /v1/responses   /v1/messages
          completions                           │
                    │            │            │
                    ▼            ▼            ▼
            ┌───────────────────────────────────────┐
            │        格式检测 & 自动路由              │
            │   relay/format + middleware/autodetect │
            └───────────────┬───────────────────────┘
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
        ┌─────────┐  ┌──────────┐  ┌──────────┐
        │ CC 请求  │  │ RA 请求   │  │ CM 请求   │
        └────┬────┘  └────┬─────┘  └────┬─────┘
             │            │             │
             └────────────┼─────────────┘
                          ▼
              ┌───────────────────────┐
              │  请求转换 (Request)    │
              │  CC↔RA / CC↔CM / RA↔CM │
              └───────────┬───────────┘
                          ▼
              ┌───────────────────────┐
              │  上游渠道 (Adaptor)    │
              │  OpenAI/Claude/Gemini  │
              │  DeepSeek/阿里/百度...  │
              └───────────┬───────────┘
                          ▼
              ┌───────────────────────┐
              │  响应转换 (Response)   │
              │  含流式 SSE 转换       │
              └───────────┬───────────┘
                          ▼
              ┌───────────────────────┐
              │  返回用户原始格式响应   │
              └───────────────────────┘
```

---

## 2. 测试矩阵

### 2.1 格式转换 9 宫格

请求入站格式 × 上游渠道格式，共 9 种组合。其中 3 种为直通（同格式），6 种需要格式转换。

| 请求格式 \ 渠道格式 | → Chat Completions | → Response API | → Claude Messages |
|---|---|---|---|
| **Chat Completions 入站** | 直通 ✅ | CC→RA ⚠️ | CC→CM ⚠️ |
| **Response API 入站** | RA→CC ⚠️ | 直通 ✅ | RA→CM ⚠️ |
| **Claude Messages 入站** | CM→CC ⚠️ | CM→RA ⚠️ | 直通 ✅ |

### 2.2 四维测试模型

每个转换组合需覆盖 4 个维度：

| 维度 | 说明 | 风险等级 |
|---|---|---|
| **纯文本对话** | 基本消息收发，含多轮 | 🟢 低 |
| **Tool Calling** | 函数定义、调用、结果回传 | 🔴 高 |
| **多模态内容** | 图片（URL/Base64）、文件 | 🟡 中 |
| **流式 SSE** | 实时事件流转换 | 🔴 高 |

**总测试用例**：9 × 4 = 36，减去 3 个直通 = **33 个转换用例**。

### 2.3 扩展维度

除四维模型外，还需关注以下横切关注点：

| 关注点 | 说明 |
|---|---|
| **Thinking/推理** | Extended thinking 块的保留与转换，signature 完整性 |
| **Structured Output** | `response_format` / `json_schema` 跨格式传递 |
| **Usage/Billing** | Token 计数在转换后是否准确 |
| **错误处理** | 上游错误信息跨格式转换 |
| **长上下文** | 大量消息（100+轮）的转换性能 |
| **并发/重试** | 转换在 failover 路径中的一致性 |

---

## 3. 现有测试覆盖分析

### 3.1 已覆盖场景

| 测试文件 | 覆盖内容 | 格式路径 |
|---|---|---|
| `relay/format/detect_test.go` | API 格式自动检测 | 所有格式 |
| `relay/adaptor/openai/channel_conversion_test.go` | CC→RA URL 转换决策 | CC→RA |
| `relay/adaptor/openai/claude_conversion_test.go` | Claude tool_choice 规范化 | CM→CC |
| `relay/adaptor/openai/claude_streaming_test.go` | Claude 流式 SSE 转换 | CM→CC (stream) |
| `relay/adaptor/openai/response_api_handler_test.go` | RA→CC 非流式响应转换 | RA→CC |
| `relay/adaptor/openai/response_api_stream_handler_test.go` | RA→CC 流式转换 | RA→CC (stream) |
| `relay/adaptor/openai_compatible/claude_convert_test.go` | CC→CM / RA→CM 响应转换 | CC→CM, RA→CM |
| `relay/controller/claude_messages_request_test.go` | Claude thinking/signature 保留 | CM 内部 |
| `relay/model/tool_test.go` | Tool Call 序列化/反序列化 | CC 内部 |

### 3.2 覆盖缺口

| 缺口 | 风险 | 说明 |
|---|---|---|
| **CM→RA 转换** | 🔴 高 | 无任何测试覆盖，Claude 请求转 Response API 出站 |
| **RA→CM 转换（非流式）** | 🟡 中 | 仅 `claude_convert_test.go` 覆盖了 RA→CM 响应，但非端到端 |
| **CC→CM 请求转换** | 🟡 中 | 有响应转换测试，缺少请求转换测试 |
| **多模态跨格式** | 🔴 高 | 无 base64 图片跨格式转换测试 |
| **Tool Call 跨格式（CM→RA/RA→CM）** | 🔴 高 | 仅有 CC↔CM 的 tool_use 转换，RA 侧缺失 |
| **流式中断/错误跨格式** | 🟡 中 | 仅有正常流式转换，无错误场景 |
| **Thinking block 跨格式** | 🟡 中 | 仅 CM 内部测试，无 CC→CM thinking 转换 |
| **端到端链路** | 🔴 高 | 单测覆盖 adaptor 内部，无 controller→adaptor 全链路 |

### 3.3 覆盖缺口热力图

```
              CC出站    RA出站    CM出站
CC入站        ✅✅✅     🟡🟡⚠️    🟡🟡⚠️
RA入站        ✅✅🟡     ✅✅✅     🔴🔴🔴
CM入站        ✅✅🟡     🔴🔴🔴    ✅✅✅

图例: ✅=已覆盖  🟡=部分覆盖  ⚠️=缺流式  🔴=未覆盖
(每格三个符号依次为: 纯文本/Tool Call/流式SSE)
```

---

## 4. 关键场景测试设计

### 4.1 P0：流式 SSE 事件映射

三种格式的流式事件结构完全不同，是转换中**最容易出 bug** 的地方。

#### 4.1.1 SSE 事件对照表

**文本增量输出**：

| 阶段 | CC 格式 | RA 格式 | CM 格式 |
|---|---|---|---|
| 流开始 | `data: {"choices":[{"delta":{"role":"assistant"}}]}` | `event: response.created` + `event: response.output_item.added` | `event: message_start` + `event: content_block_start` |
| 内容增量 | `data: {"choices":[{"delta":{"content":"Hel"}}]}` | `event: response.output_text.delta` + `data: {"delta":"Hel"}` | `event: content_block_delta` + `data: {"delta":{"type":"text_delta","text":"Hel"}}` |
| 内容完成 | — | `event: response.output_text.done` | `event: content_block_stop` |
| 流结束 | `data: [DONE]` | `event: response.completed` + `data: [DONE]` | `event: message_delta` + `event: message_stop`（**无 [DONE]**） |

**Tool Call 增量输出**：

| 阶段 | CC 格式 | RA 格式 | CM 格式 |
|---|---|---|---|
| 调用开始 | `delta.tool_calls[0]` 含 `id` + `function.name` | `event: response.function_call_arguments.delta` | `event: content_block_start` + `type: tool_use` |
| 参数增量 | `delta.tool_calls[0].function.arguments` 分片 | `event: response.function_call_arguments.delta` | `event: content_block_delta` + `type: input_json_delta` |
| 调用完成 | `finish_reason: "tool_calls"` | `event: response.output_item.done` | `event: content_block_stop` |

#### 4.1.2 流式转换测试用例

| ID | 转换路径 | 场景 | 验证点 |
|---|---|---|---|
| SSE-01 | CC stream → CM stream | 纯文本 | `message_start` → `content_block_delta` → `message_stop`，无 `[DONE]` |
| SSE-02 | CC stream → CM stream | Tool Call | `content_block_start(tool_use)` → `input_json_delta` → `content_block_stop` |
| SSE-03 | CC stream → RA stream | 纯文本 | `response.output_text.delta` → `response.completed` + `[DONE]` |
| SSE-04 | RA stream → CC stream | 纯文本 | `delta.content` 分片 + 最终 `data: [DONE]` |
| SSE-05 | RA stream → CC stream | Tool Call | `delta.tool_calls` 增量合并正确 |
| SSE-06 | RA stream → CM stream | 纯文本 | RA 事件 → CM 事件完整映射 |
| SSE-07 | CM stream → CC stream | 纯文本 | `content_block_delta` → `delta.content` + `[DONE]` |
| SSE-08 | CM stream → CC stream | Tool Call | `tool_use` → `tool_calls` 增量，含 `index` 字段 |
| SSE-09 | CM stream → RA stream | 纯文本 | CM 事件 → RA 事件映射 |
| SSE-10 | 任意格式 | 流式中断 | 上游断连时，客户端收到格式正确的错误事件 |
| SSE-11 | CC stream → CM stream | Thinking | `thinking` content block 保留，signature 不截断 |
| SSE-12 | RA stream → CC stream | 并行 Tool Call | 多个 `tool_calls` 的 `index` 字段正确映射 |

### 4.2 P0：Tool Calling 跨格式转换

#### 4.2.1 字段映射对照

**工具定义（Request）**：

```
CC:  tools[].function:       {name, description, parameters, strict}
RA:  tools[]:                 {type:"function", name, description, parameters, strict}
CM:  tools[]:                 {name, description, input_schema, type:"custom"}
```

**工具调用（Response）**：

```
CC:  choices[].message.tool_calls[]:  {id, type:"function", function:{name, arguments}}
RA:  output[]:                         [{type:"function_call", call_id, name, arguments}]
CM:  content[]:                        [{type:"tool_use", id, name, input}]
```

**工具结果（Request - 下一轮）**：

```
CC:  messages[{role:"tool", tool_call_id, content}]
RA:  input[{type:"function_call_output", call_id, output}]
CM:  messages[{role:"user", content:[{type:"tool_result", tool_use_id, content}]}]
```

**tool_choice**：

```
CC:  {type:"function", function:{name:"x"}}     / "auto" / "required" / "none"
RA:  {type:"function", name:"x"}                / 同 CC
CM:  {type:"tool", name:"x"}                    / "auto" / "any" / none
```

> 注意：CM 的 `"any"` 对应 CC/RA 的 `"required"`。

#### 4.2.2 Tool Call 测试用例

| ID | 转换路径 | 场景 | 验证点 |
|---|---|---|---|
| TC-01 | CC→CM | 单工具调用 | `function.name` → `name`，`function.arguments` → `input`，`id` → `id` |
| TC-02 | CC→CM | 并行工具调用 | 多个 `tool_calls` → 多个 `tool_use` content block，顺序保持 |
| TC-03 | CM→CC | 单工具调用 | `tool_use.id` → `tool_calls[].id`，`input` → `function.arguments` |
| TC-04 | CM→CC | 并行工具调用 | `index` 字段正确设置 |
| TC-05 | RA→CC | 函数调用 | `call_id` → `tool_calls[].id`，`arguments` → `function.arguments` |
| TC-06 | CC→RA | 函数调用 | `id` → `call_id`，`function.name` → `name` |
| TC-07 | RA→CM | 函数调用 | `call_id` → `id`，`arguments`(string) → `input`(JSON) |
| TC-08 | CM→RA | 函数调用 | `id` → `call_id`，`input`(JSON) → `arguments`(string) |
| TC-09 | 任意→任意 | tool_choice | `"required"` ↔ `"any"` 互转正确 |
| TC-10 | CM→CC | 工具结果 | `tool_result.tool_use_id` → `tool_call_id`，content 映射 |
| TC-11 | CC→CM | 工具结果 | `tool_call_id` → `tool_use_id`，role 转换 |
| TC-12 | RA→CC | 工具结果 | `function_call_output.call_id` → `tool_call_id` |
| TC-13 | 任意→任意 | arguments 含特殊字符 | JSON 中的引号、换行、Unicode 正确传递 |
| TC-14 | 任意→任意 | arguments 为空对象 | `{}` 不丢失 |

### 4.3 P1：多模态内容转换

#### 4.3.1 图片字段映射

**输入侧（Request messages）**：

```
CC:  content: [{type:"image_url", image_url: {url: "https://..." / "data:image/png;base64,..."}}]
RA:  input:   [{type:"image_url", image_url: {url: "https://..." / "data:image/png;base64,..."}}]
CM:  content: [{type:"image", source: {type:"url", url: "https://..."}}]
                or [{type:"image", source: {type:"base64", media_type: "image/png", data: "..."}}]
```

**输出侧（Response）**：

```
CC:  content: [{type:"image_url", image_url: {url: "..."}}]   （部分模型支持）
RA:  output:  [{type:"image", image_url: "..."}}]
CM:  content: [{type:"image", source: {type:"base64", ...}}]
```

#### 4.3.2 多模态测试用例

| ID | 转换路径 | 场景 | 验证点 |
|---|---|---|---|
| MM-01 | CC→CM | 图片 URL | `image_url.url` → `source.type:"url"` + `source.url` |
| MM-02 | CC→CM | 图片 Base64 | `image_url.url="data:..."` → `source.type:"base64"` + 正确拆分 media_type/data |
| MM-03 | CM→CC | 图片 URL | `source.url` → `image_url.url` |
| MM-04 | CM→CC | 图片 Base64 | `source.data` → `image_url.url="data:{media_type};base64,{data}"` |
| MM-05 | RA→CC | 图片 URL | 同 CC→CC 透传 |
| MM-06 | RA→CM | 图片 URL | RA 格式 → CM `source` 格式 |
| MM-07 | 任意→任意 | 多图 | 多个图片 content block 顺序保持 |
| MM-08 | CC→CM | 混合内容 | 同一 message 中 text + image 混合，顺序和类型正确 |
| MM-09 | 任意→任意 | 大 Base64 | 1MB+ 图片 base64 在 JSON 转换中不被截断或转义 |
| MM-10 | 任意→任意 | 图片+Tool Call | 多模态输入 + tool_call 输出同帧 |

### 4.4 P1：Thinking/推理块转换

| ID | 转换路径 | 场景 | 验证点 |
|---|---|---|---|
| TH-01 | CC→CM | thinking_content | `reasoning_content` → `thinking` content block，含 `signature` |
| TH-02 | CM→CC | thinking block | `thinking` content block → `reasoning_content` |
| TH-03 | CM→CC | 多轮 thinking | assistant message 中的 thinking + signature 完整保留 |
| TH-04 | RA→CC | reasoning output | RA `reasoning` output item → CC `reasoning_content` |
| TH-05 | RA→CM | reasoning output | RA `reasoning` output item → CM `thinking` block |
| TH-06 | 任意→任意 | HTML 字符 | thinking 文本中的 `<` `>` `&` 不被 JSON 转义为 `\u003c` |

### 4.5 P2：Usage/Billing 准确性

| ID | 转换路径 | 场景 | 验证点 |
|---|---|---|---|
| UB-01 | CC→CM | 非流式 usage | `prompt_tokens` → `input_tokens`，`completion_tokens` → `output_tokens` |
| UB-02 | CC→CM | 缓存 token | `prompt_tokens_details.cached_tokens` → `cache_read_input_tokens` |
| UB-03 | CC→CM | 缓存写入 | `cache_write_5m_tokens` / `cache_write_1h_tokens` → `cache_creation_input_tokens` |
| UB-04 | RA→CC | 流式 usage | RA `response.completed` 中的 usage 映射到 CC usage |
| UB-05 | 任意→任意 | 流式无 usage | 上游未返回 usage 时，UniAPI 估算值合理 |
| UB-06 | 任意→任意 | 流式中断 | 中断时 usage 不计或按已消费部分计 |

---

## 5. 端到端集成测试

### 5.1 测试环境

| 组件 | 说明 |
|---|---|
| UniAPI 实例 | 本地启动，连接 SQLite 测试库 |
| 上游模拟 | httptest.Server 模拟各格式上游响应 |
| 测试客户端 | Go HTTP 客户端发送各格式请求 |

### 5.2 端到端测试用例

每个 E2E 用例覆盖完整链路：**客户端请求 → UniAPI 路由 → 转换 → 模拟上游 → 转换响应 → 客户端验证**。

| ID | 入站格式 | 上游格式 | 流式 | 内容 | 验证 |
|---|---|---|---|---|---|
| E2E-01 | CC | CM (Anthropic) | 否 | 纯文本 | 响应为 CC 格式，内容正确 |
| E2E-02 | CC | CM (Anthropic) | 是 | 纯文本 | SSE 事件序列正确，含 `[DONE]` |
| E2E-03 | CM | CC (OpenAI) | 否 | 纯文本 | 响应为 CM 格式，含 `message_start/stop` |
| E2E-04 | CM | CC (OpenAI) | 是 | 纯文本 | SSE 事件为 CM 格式 |
| E2E-05 | RA | CC (DeepSeek) | 否 | 纯文本 | 响应为 RA 格式，含 `response.completed` |
| E2E-06 | RA | CC (DeepSeek) | 是 | 纯文本 | SSE 事件为 RA 格式 |
| E2E-07 | CC | CM (Anthropic) | 否 | Tool Call | tool_calls → tool_use 转换正确 |
| E2E-08 | CM | CC (OpenAI) | 是 | Tool Call | tool_use → tool_calls 流式增量正确 |
| E2E-09 | CC | CM (Anthropic) | 是 | 图片 | image_url → source 转换正确 |
| E2E-10 | RA | CM (Anthropic) | 是 | Thinking | reasoning → thinking block 保留 |

### 5.3 渠道类型覆盖

不同渠道类型的 adaptor 实现不同，需重点测试的渠道：

| 优先级 | 渠道类型 | Adaptor | 原因 |
|---|---|---|---|
| P0 | OpenAI | `relay/adaptor/openai` | 主力渠道，CC↔RA 转换最复杂 |
| P0 | Anthropic | `relay/adaptor/anthropic` | Claude 原生渠道，CM 格式主路径 |
| P0 | OpenAI-Compatible | `relay/adaptor/openai_compatible` | CC→CM 转换主路径，覆盖 DeepSeek/阿里等 |
| P1 | Gemini | `relay/adaptor/gemini` | 独立 adaptor，有 `adaptor_claude_test.go` |
| P1 | AWS Claude | `relay/adaptor/aws/claude` | Bedrock 路径，签名方式不同 |
| P2 | Vertex AI | `relay/adaptor/vertexai` | Google Cloud 路径 |

---

## 6. 格式自动检测测试

UniAPI 支持**格式自动检测**（`AUTO_DETECT_API_FORMAT=true`），当客户端将请求发到错误端点时自动修正。此功能对兼容性至关重要。

### 6.1 检测逻辑

| 检测条件 | 判定格式 |
|---|---|
| 含 `input` 字段（无 `messages`），或含 `instructions`，或含 `max_output_tokens` | Response API |
| 含 `messages` + `system` 字段，或 `tool_use`/`tool_result` content，或 `input_schema` in tools | Claude Messages |
| 含 `messages` + 标准 OpenAI 格式 | Chat Completions |

### 6.2 自动检测测试用例

| ID | 场景 | 请求路径 | 实际格式 | 预期行为 |
|---|---|---|---|---|
| AD-01 | Cursor 发 RA 请求到 CC 路径 | `/v1/chat/completions` | Response API | 透明重路由到 RA handler |
| AD-02 | Claude 请求发到 CC 路径 | `/v1/chat/completions` | Claude Messages | 透明重路由到 CM handler |
| AD-03 | CC 请求发到 CM 路径 | `/v1/messages` | CC | 透明重路由到 CC handler |
| AD-04 | RA 请求发到 CM 路径 | `/v1/messages` | RA | 透明重路由到 RA handler |
| AD-05 | 重定向模式 | 任意错配 | — | HTTP 302 到正确路径 |
| AD-06 | 检测关闭 | 任意错配 | — | 返回错误 |

---

## 7. 中转站 vs 原厂对比测试

### 7.1 核心问题

UniAPI 的核心承诺是**透明中转**——调用方不应感知到中间层的存在。但中转过程中有多个环节可能引入差异。对比测试的价值在于**验证中转链路的透明性**。

**前提认知**：LLM 输出是非确定性的，同样的 prompt 发两次结果可能不同。因此对比测试**不比内容，比结构和精度**。

| 对比维度 | 能否直接对比 | 原因 |
|---|---|---|
| 响应文本内容 | ❌ 不能逐字对比 | 非确定性 |
| JSON 结构完整性 | ✅ 可以 | 结构是确定的 |
| Token 用量 | ✅ 可以 | 上游返回的 usage 是确定值 |
| SSE 事件序列 | ✅ 可以 | 事件类型和顺序是确定的 |
| Tool Call 参数结构 | ✅ 可以 | 字段名/类型是确定的 |
| HTTP 状态码 | ✅ 可以 | 错误码是确定的 |
| 延迟 | ⚠️ 需统计 | 单次不可比，需多次采样 |

### 7.2 中转链路中的 5 大差异源

根据代码分析，UniAPI 中转链路中存在以下差异源：

#### 差异源 1：Token 计数偏差（🔴 高）

UniAPI 使用本地 tiktoken 估算 prompt token 数量，与上游实际计数可能不一致：

```go
// relay/adaptor/openai/token.go:78-81
func getTokenNum(tokenEncoder *tiktoken.Tiktoken, text string) int {
    if config.ApproximateTokenEnabled {
        return int(float64(len(text)) * 0.38)  // 粗估，与实际偏差可达 ±30%
    }
    return len(tokenEncoder.Encode(text, nil, nil))
}
```

即使不用近似模式，tiktoken 本地计数与上游 API 返回的 `usage.prompt_tokens` 也可能不同（上游有独立的 tokenizer 实现，对图片/音频 token 计算方式不同）。

**影响**：计费偏差，用户被多扣或少扣 quota。

#### 差异源 2：格式转换丢字段（🔴 高）

以 Claude Messages → Chat Completions 为例，Claude 的 `thinking` block、`signature` 字段在转换后需要特殊处理才能保留。如果转换逻辑不完整，这些字段会丢失。

**影响**：功能性 bug，多轮对话时 Claude 可能因缺少 signature 而拒绝请求。

#### 差异源 3：流式 SSE 事件映射不完整（🟡 中）

三种格式的流式事件结构完全不同，映射关系复杂。例如 Claude 不发 `[DONE]`，CC 要发；Claude 的 `content_block_delta` 需要映射为 CC 的 `delta.content`。

**影响**：客户端解析失败。

#### 差异源 4：Fallback 路由（🟡 中）

```go
// relay/controller/response_fallback.go:33
func relayResponseAPIThroughChat(c *gin.Context, meta *metalib.Meta, ...) {
    chatRequest, err := openai.ConvertResponseAPIToChatCompletionRequest(responseAPIRequest)
    meta.Mode = relaymode.ChatCompletions  // 模式被改写
}
```

当上游渠道不支持 Response API 时，UniAPI 自动降级为 CC。**同一请求通过 UniAPI 和原厂，走的处理链路可能完全不同**。

**影响**：降级路径的转换质量未经充分验证。

#### 差异源 5：预扣费/退款（🟡 中）

UniAPI 使用 `PreConsumedQuota` 机制，请求前预扣 quota，完成后按实际 usage 退款或补扣。

**影响**：在流式中断、fallback 等场景下，预扣值与实际消费可能不一致。

### 7.3 对比测试方法

#### 方法 1：结构保真度测试（Structural Fidelity）

**原理**：不比内容，比结构。两个响应的 JSON schema 必须一致。

```python
def test_structural_fidelity():
    """验证 UniAPI 中转后响应结构与原厂一致"""
    # 同一请求分别发 UniAPI 和原厂
    uniapi_resp = request(uniapi_baseurl, uniapi_key, payload)
    origin_resp = request(origin_baseurl, origin_key, payload)

    # 对比 JSON 结构（忽略值）
    assert same_keys(uniapi_resp.json(), origin_resp.json())
    assert same_types(uniapi_resp.json(), origin_resp.json())
    assert same_nested_structure(uniapi_resp.json(), origin_resp.json())
```

**验证点**：

| 字段 | 验证内容 |
|---|---|
| `usage` | 包含 `prompt_tokens`、`completion_tokens`、`total_tokens` 三个字段 |
| `choices[0].message` | 包含 `role`、`content`（如有）、`tool_calls`（如有） |
| `choices[0].finish_reason` | 值在合法集合内（`stop`/`tool_calls`/`length`/`content_filter`） |
| `id` | 非空字符串 |
| `model` | 与请求 model 一致 |
| `object` | 值为 `chat.completion` 或 `chat.completion.chunk` |

#### 方法 2：Usage 精度测试（Usage Accuracy）

**原理**：对比 UniAPI 报告的 token 用量与原厂实际值。由于输出非确定性，`completion_tokens` 会不同，但 `prompt_tokens` 应几乎一致。

```python
def test_usage_accuracy():
    """验证 UniAPI 报告的 token 用量与原厂一致"""
    results = []
    for prompt in test_prompts:
        uniapi_resp = request(uniapi_baseurl, uniapi_key, prompt)
        origin_resp = request(origin_baseurl, origin_key, prompt)

        uniapi_usage = uniapi_resp.json()["usage"]
        origin_usage = origin_resp.json()["usage"]

        results.append({
            "uniapi_prompt": uniapi_usage["prompt_tokens"],
            "origin_prompt": origin_usage["prompt_tokens"],
            "uniapi_completion": uniapi_usage["completion_tokens"],
            "origin_completion": origin_usage["completion_tokens"],
        })

    # prompt_tokens 偏差 < 5%（同样输入，计数应一致）
    for r in results:
        prompt_diff = abs(r["uniapi_prompt"] - r["origin_prompt"]) / r["origin_prompt"]
        assert prompt_diff < 0.05, f"prompt_tokens 偏差 {prompt_diff:.1%}"

    # completion_tokens 量级相当（允许非确定性差异，但不应差一个数量级）
    for r in results:
        ratio = r["uniapi_completion"] / max(r["origin_completion"], 1)
        assert 0.2 < ratio < 5.0, f"completion_tokens 比率 {ratio:.1f} 异常"
```

**关键指标**：

| 指标 | 期望阈值 | 原因 |
|---|---|---|
| `prompt_tokens` 偏差 | < 5% | 相同输入，计数应一致 |
| `completion_tokens` 量级 | 同数量级 | 输出非确定性，但量级应相当 |
| `cached_tokens` 传递 | 字段存在且 > 0 | 缓存 token 影响计费，必须传递 |
| `total_tokens` | = prompt + completion | 算术一致性 |

#### 方法 3：流式保真度测试（Streaming Fidelity）

**原理**：对比 SSE 事件类型和序列完整性，忽略事件内容。

```python
def test_streaming_fidelity():
    """验证流式 SSE 事件序列与原厂一致"""
    uniapi_events = stream_request(uniapi_baseurl, uniapi_key, stream_payload)
    origin_events = stream_request(origin_baseurl, origin_key, stream_payload)

    # 提取事件类型序列
    uniapi_types = [e.event_type for e in uniapi_events]
    origin_types = [e.event_type for e in origin_events]

    # CC 格式：事件类型序列应一致（忽略 delta 内容）
    assert uniapi_types == origin_types, \
        f"事件序列不一致\nUniAPI: {uniapi_types}\n原厂: {origin_types}"

    # 终止事件验证
    assert uniapi_events[-1].data.strip() == "[DONE]", \
        f"流式终止事件异常: {uniapi_events[-1].data}"

    # 首事件必须含 role
    first_data = json.loads(uniapi_events[0].data)
    assert first_data["choices"][0]["delta"].get("role") == "assistant"
```

**验证矩阵**：

| 格式 | 首事件 | 末事件 | 关键约束 |
|---|---|---|---|
| CC | `delta: {role: "assistant"}` | `data: [DONE]` | `[DONE]` 必须存在 |
| RA | `event: response.created` | `event: response.completed` + `[DONE]` | completed 事件含完整 usage |
| CM | `event: message_start` | `event: message_stop` | **无 `[DONE]`** |

#### 方法 4：Tool Call 完整性测试

**原理**：验证 tool_call 的 id、name、arguments 三个关键字段无损传递。

```python
def test_tool_call_completeness():
    """验证 Tool Call 跨中转无损"""
    payload = {
        "model": "gpt-4o",
        "messages": [{"role": "user", "content": "北京天气怎么样"}],
        "tools": [{
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "获取天气",
                "parameters": {
                    "type": "object",
                    "properties": {"city": {"type": "string"}},
                    "required": ["city"]
                }
            }
        }],
        "tool_choice": "auto"
    }

    uniapi_resp = request(uniapi_baseurl, uniapi_key, payload)
    origin_resp = request(origin_baseurl, origin_key, payload)

    uniapi_tc = uniapi_resp.json()["choices"][0]["message"].get("tool_calls", [])
    origin_tc = origin_resp.json()["choices"][0]["message"].get("tool_calls", [])

    # 字段完整性：每个 tool_call 必须包含 id、type、function
    for tc in uniapi_tc:
        assert "id" in tc, "tool_call 缺少 id"
        assert "type" in tc, "tool_call 缺少 type"
        assert "function" in tc, "tool_call 缺少 function"
        assert "name" in tc["function"], "function 缺少 name"
        assert "arguments" in tc["function"], "function 缺少 arguments"

        # arguments 必须是合法 JSON
        args = json.loads(tc["function"]["arguments"])
        assert isinstance(args, dict), "arguments 不是 JSON 对象"

    # 工具名称应一致（不丢字段）
    uniapi_names = {tc["function"]["name"] for tc in uniapi_tc}
    origin_names = {tc["function"]["name"] for tc in origin_tc}
    assert uniapi_names == origin_names, \
        f"工具名称不一致: UniAPI={uniapi_names}, 原厂={origin_names}"
```

#### 方法 5：延迟开销测试（Latency Overhead）

**原理**：统计 UniAPI 中转带来的额外延迟。需多次采样，取统计值。

```python
def test_latency_overhead(n=50):
    """验证 UniAPI 中转延迟在可接受范围"""
    uniapi_latencies = []
    origin_latencies = []

    for _ in range(n):
        t0 = time.time()
        request(uniapi_baseurl, uniapi_key, payload)
        uniapi_latencies.append(time.time() - t0)

        t0 = time.time()
        request(origin_baseurl, origin_key, payload)
        origin_latencies.append(time.time() - t0)

    uniapi_p50 = sorted(uniapi_latencies)[n // 2]
    origin_p50 = sorted(origin_latencies)[n // 2]
    overhead = uniapi_p50 - origin_p50
    overhead_pct = overhead / origin_p50 * 100

    # 中转 P50 延迟增量应 < 100ms 或 < 10%
    assert overhead < 0.1 or overhead_pct < 10, \
        f"延迟开销过大: +{overhead*1000:.0f}ms (+{overhead_pct:.1f}%)"
```

**参考阈值**：

| 指标 | 可接受 | 警告 | 不可接受 |
|---|---|---|---|
| P50 延迟增量 | < 50ms | 50-100ms | > 100ms |
| P99 延迟增量 | < 200ms | 200-500ms | > 500ms |
| 延迟增量百分比 | < 5% | 5-10% | > 10% |

#### 方法 6：错误传播测试（Error Propagation）

**原理**：验证上游的错误信息被正确传递，不被吞掉或改写。

```python
def test_error_propagation():
    """验证错误信息正确传播"""
    test_cases = [
        {"payload": {"model": "nonexistent-model", "messages": [{"role": "user", "content": "hi"}]},
         "expected_status": 404, "desc": "不存在的模型"},
        {"payload": {"model": "gpt-4", "messages": []},
         "expected_status": 400, "desc": "空消息列表"},
        {"payload": {"model": "gpt-4", "messages": [{"role": "user", "content": "hi"}], "max_tokens": -1},
         "expected_status": 400, "desc": "非法 max_tokens"},
        {"payload": {"model": "gpt-4", "messages": [{"role": "user", "content": "hi" * 100000}]},
         "expected_status": 400, "desc": "超出上下文长度"},
    ]

    for case in test_cases:
        uniapi_resp = request(uniapi_baseurl, uniapi_key, case["payload"])
        origin_resp = request(origin_baseurl, origin_key, case["payload"])

        # HTTP 状态码应一致
        assert uniapi_resp.status_code == origin_resp.status_code, \
            f"[{case['desc']}] 状态码不一致: UniAPI={uniapi_resp.status_code}, 原厂={origin_resp.status_code}"

        # 错误响应结构应包含 error 字段
        uniapi_body = uniapi_resp.json()
        assert "error" in uniapi_body, \
            f"[{case['desc']}] 缺少 error 字段: {uniapi_body}"

        # error.message 应包含原厂错误关键信息
        uniapi_msg = uniapi_body["error"].get("message", "")
        assert len(uniapi_msg) > 0, \
            f"[{case['desc']}] error.message 为空"
```

### 7.4 对比测试用例矩阵

| ID | 方法 | 格式 | 流式 | 场景 | 验证点 | 通过标准 |
|---|---|---|---|---|---|---|
| CMP-01 | 结构保真 | CC | 否 | 纯文本 | JSON 结构一致 | 所有 key 和 type 匹配 |
| CMP-02 | 结构保真 | CC | 是 | 纯文本 | SSE 事件类型序列 | 事件类型序列匹配 |
| CMP-03 | 结构保真 | CC | 否 | Tool Call | tool_calls 结构完整 | id/name/arguments 均存在 |
| CMP-04 | 结构保真 | CM | 否 | 纯文本 | Claude 格式结构 | content/stop_reason/usage 均存在 |
| CMP-05 | 结构保真 | CM | 是 | Thinking | thinking block 保留 | type/signature 存在 |
| CMP-06 | 结构保真 | RA | 否 | 纯文本 | RA 格式结构 | output/usage/status 均存在 |
| CMP-07 | Usage 精度 | CC | 否 | 纯文本 | prompt_tokens 偏差 | < 5% |
| CMP-08 | Usage 精度 | CC | 否 | 多轮 | prompt_tokens 偏差 | < 5% |
| CMP-09 | Usage 精度 | CC | 否 | 图片输入 | prompt_tokens 偏差 | < 10%（图片 token 估算差异大） |
| CMP-10 | Usage 精度 | CC | 是 | 纯文本 | 流式 usage 完整性 | 末事件含完整 usage |
| CMP-11 | Usage 精度 | CM | 否 | 缓存 token | cached_tokens 传递 | 字段存在且值 > 0 |
| CMP-12 | 流式保真 | CC | 是 | 纯文本 | 事件序列 | 事件类型序列匹配原厂 |
| CMP-13 | 流式保真 | CC | 是 | Tool Call | tool_call 增量 | index/name 增量正确 |
| CMP-14 | 流式保真 | CC | 是 | 长输出 | 流不截断 | 收到 `[DONE]` |
| CMP-15 | 流式保真 | CM | 是 | 纯文本 | CM 事件序列 | 无 `[DONE]`，message_stop 存在 |
| CMP-16 | Tool Call | CC | 否 | 单工具 | 字段完整 | id/name/arguments 均存在 |
| CMP-17 | Tool Call | CC | 否 | 并行工具 | 多工具完整 | 所有 tool_calls 字段完整 |
| CMP-18 | Tool Call | CC | 否 | arguments 含特殊字符 | arguments 合法 | JSON 可解析，含引号/换行 |
| CMP-19 | Tool Call | CM | 否 | 单工具 | Claude 格式 | tool_use id/name/input 完整 |
| CMP-20 | Tool Call | RA | 否 | 单工具 | RA 格式 | function_call call_id/name/arguments 完整 |
| CMP-21 | 延迟开销 | CC | 否 | 纯文本 | P50 延迟增量 | < 100ms |
| CMP-22 | 延迟开销 | CC | 是 | 纯文本 | 首 token 延迟 | TTFB 增量 < 100ms |
| CMP-23 | 延迟开销 | CM | 否 | 纯文本 | P50 延迟增量 | < 100ms |
| CMP-24 | 错误传播 | CC | 否 | 不存在模型 | 状态码 + 错误信息 | 状态码一致，error.message 含关键信息 |
| CMP-25 | 错误传播 | CC | 否 | 超上下文 | 状态码 + 错误信息 | 状态码一致 |
| CMP-26 | 错误传播 | CC | 是 | 流式中断 | 客户端收到正确终止 | 不挂起，收到错误或 [DONE] |
| CMP-27 | 错误传播 | CM | 否 | max_tokens=0 | 状态码 + 错误信息 | 状态码一致 |

### 7.5 对比测试意义评估

#### 高价值场景（必做）

| 场景 | 意义 | 原因 |
|---|---|---|
| **格式转换保真度** | ⭐⭐⭐⭐⭐ | UniAPI 核心价值，转换丢字段 = 功能性 bug |
| **Token 计费精度** | ⭐⭐⭐⭐⭐ | 直接影响用户钱，偏差大 = 计费错误 |
| **Tool Call 完整性** | ⭐⭐⭐⭐⭐ | 编程智能体依赖 tool_call，丢字段 = 工作流中断 |
| **流式事件序列** | ⭐⭐⭐⭐ | 客户端依赖事件序列解析，错序 = 解析失败 |
| **错误传播** | ⭐⭐⭐⭐ | 错误被吞 = 调试困难，用户无法定位问题 |

#### 中等价值场景（建议做）

| 场景 | 意义 | 限制 |
|---|---|---|
| **响应内容质量** | ⭐⭐ | 非确定性导致无法直接对比，只能做粗粒度判断 |
| **延迟开销** | ⭐⭐ | 受网络波动影响大，需大量采样才有统计意义 |
| **并发能力** | ⭐⭐ | 受 UniAPI 自身资源限制，与原厂无直接可比性 |

#### 低价值场景（不建议做）

| 场景 | 意义 | 原因 |
|---|---|---|
| **逐字内容对比** | ⭐ | 非确定性决定了这不可能通过 |
| **价格对比** | ⭐ | UniAPI 有自己的计费体系，不是"原价转售" |
| **模型能力对比** | ⭐ | 同模型同 prompt，能力无差异（除非路由到了不同模型） |

### 7.6 对比测试实施要点

#### 测试环境

| 组件 | 说明 |
|---|---|
| UniAPI 实例 | 真实部署的 UniAPI，配置待测渠道 |
| 原厂 API | 直接调用上游厂商 API（如 api.openai.com、api.anthropic.com） |
| 测试脚本 | Python/Go 编写的自动化对比脚本 |
| 测试模型 | 选取 2-3 个代表性模型（如 gpt-4o、claude-sonnet-4-20250514、deepseek-chat） |

#### 测试时机

| 时机 | 测试范围 | 频率 |
|---|---|---|
| 渠道新增/修改 | CMP-01 ~ CMP-27 | 每次变更 |
| UniAPI 版本发布 | CMP-01 ~ CMP-27 | 每次发布 |
| 定期回归 | CMP-07 ~ CMP-11 (Usage) + CMP-12 ~ CMP-15 (流式) | 每周 |
| 上游 API 变更 | 全量 | 按需 |

#### 控制变量

为保证对比有效性，需控制以下变量：

1. **同一模型**：UniAPI 和原厂请求相同的 model 名称
2. **同一 prompt**：使用相同的 messages/temperature/max_tokens
3. **temperature=0**：降低输出非确定性（但注意不保证完全确定性）
4. **非流式优先**：结构对比先做非流式，流式额外验证事件序列
5. **排除缓存**：避免 prompt cache 导致 usage 差异（或专门测试缓存场景）

---

## 8. 回归测试策略

### 8.1 测试分层

```
┌──────────────────────────────────────┐
│         CI/CD 流水线                  │
│  go test -race ./relay/...           │
│  (单元测试，每次提交)                  │
├──────────────────────────────────────┤
│         每日构建                      │
│  E2E 测试 + 格式兼容矩阵              │
│  (完整 9 宫格 + 流式)                 │
├──────────────────────────────────────┤
│         版本发布前                    │
│  全渠道集成测试                       │
│  (含真实上游 API 调用)                │
├──────────────────────────────────────┤
│         手动触发                      │
│  性能测试 + 压力测试                  │
│  (长上下文、并发、大文件)              │
└──────────────────────────────────────┘
```

### 8.2 回归测试命令

```bash
# 1. 快速单测（CI 每次提交）
go test -race -count=1 ./relay/...

# 2. 格式转换专项
go test -race -run "TestFormatConversion" ./relay/...

# 3. 流式转换专项
go test -race -run "TestSSE" ./relay/...

# 4. Tool Call 转换专项
go test -race -run "TestToolCall" ./relay/...

# 5. 完整 E2E（需启动 UniAPI 实例）
go test -race -tags=e2e ./test/...
```

### 8.3 测试文件组织建议

```
relay/
  adaptor/
    openai/
      format_conversion_matrix_test.go     ← 9 宫格非流式转换
      streaming_cross_format_test.go       ← 流式跨格式 SSE
      tool_call_cross_format_test.go       ← Tool Call 跨格式
      multimodal_cross_format_test.go      ← 多模态跨格式
      thinking_cross_format_test.go        ← Thinking 块跨格式
    openai_compatible/
      claude_messages_matrix_test.go       ← Claude 涉及的全矩阵
    anthropic/
      native_to_cross_format_test.go       ← Anthropic 原生→其他格式
  controller/
    format_conversion_e2e_test.go          ← Controller 层端到端
  format/
    detect_edge_cases_test.go              ← 格式检测边界用例
test/
  e2e_format_compat_test.go               ← 完整 E2E 测试
```

---

## 9. 风险矩阵

### 9.1 高风险场景

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| CM→RA 转换无测试 | Claude Desktop 用户通过 RA 渠道调用时静默出错 | P0 补充 CM→RA 转换测试 |
| 流式 Tool Call 增量丢失 | 并行函数调用时参数不完整，智能体执行错误 | P0 补充流式 TC 增量测试 |
| Base64 图片转换截断 | 截图分析功能失败 | P0 补充大 Base64 测试 |
| Thinking signature 丢失 | 多轮对话时 Claude 拒绝请求 | P1 补充跨格式 thinking 测试 |
| RA fallback 递归 | RA→CC→RA 死循环 | 已有 `meta.ResponseAPIFallback` 保护，但缺测试 |
| Token 计费偏差 | UniAPI 本地估算 vs 原厂实际计数不一致，用户被多扣或少扣 | 对比测试 CMP-07~CMP-11 验证 |
| 对比测试误报 | LLM 非确定性导致 completion_tokens 不一致被误判为 bug | 采用量级判断（同数量级即通过），仅 prompt_tokens 严格对比 |

### 9.2 中风险场景

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| Usage 跨格式不准 | 计费偏差 | P1 补充 usage 映射测试 |
| 格式检测误判 | 请求路由到错误 handler | P1 补充边界用例 |
| 长上下文转换超时 | 大消息列表转换 OOM | P2 性能测试 |
| 错误响应格式不一致 | 客户端解析失败 | P2 补充错误场景测试 |
| 延迟突增 | UniAPI 中转延迟突然增大，用户体感差 | 定期运行 CMP-21~CMP-23 监控 |
| 上游 API 变更未感知 | 原厂修改响应格式后 UniAPI 转换出错 | 对比测试作为早期预警 |

---

## 10. 实施路线图

### Phase 1（P0，1-2 周）：补齐核心转换测试 + 对比基线

1. 创建 `format_conversion_matrix_test.go`，覆盖 9 宫格纯文本转换
2. 创建 `streaming_cross_format_test.go`，覆盖 SSE-01 ~ SSE-12
3. 创建 `tool_call_cross_format_test.go`，覆盖 TC-01 ~ TC-14
4. 重点补 CM→RA 和 RA→CM 转换测试
5. **建立对比测试基线**：实现 CMP-01 ~ CMP-07（结构保真 + Usage 精度），确立通过标准

### Phase 2（P1，2-3 周）：补齐扩展维度 + 对比测试完善

1. 创建 `multimodal_cross_format_test.go`，覆盖 MM-01 ~ MM-10
2. 创建 `thinking_cross_format_test.go`，覆盖 TH-01 ~ TH-06
3. 补充 Usage/Billing 映射测试
4. 补充格式检测边界用例
5. **扩展对比测试**：CMP-08 ~ CMP-20（多场景 Usage + Tool Call 完整性）

### Phase 3（P2，3-4 周）：端到端与回归 + 对比自动化

1. 创建 `format_conversion_e2e_test.go`
2. 配置 CI 流水线分层运行
3. 建立兼容性仪表盘（测试通过率 × 格式组合）
4. 定期全渠道集成测试
5. **对比测试自动化**：CMP-21 ~ CMP-27（延迟 + 错误传播），集成到 CI

---

## 11. 附录

### A. 三格式消息结构对比

#### A.1 请求消息

```json
// Chat Completions
{
  "model": "gpt-4",
  "messages": [
    {"role": "system", "content": "You are helpful"},
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": "Hi!"},
    {"role": "user", "content": [
      {"type": "text", "text": "What is this?"},
      {"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}
    ]}
  ],
  "tools": [{"type": "function", "function": {"name": "get_weather", "parameters": {...}}}],
  "tool_choice": "auto",
  "stream": true
}

// Response API
{
  "model": "gpt-4",
  "input": [
    {"role": "system", "content": "You are helpful"},
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": "Hi!"},
    {"role": "user", "content": [
      {"type": "text", "text": "What is this?"},
      {"type": "image_url", "image_url": {"url": "data:image/png;base64,..."}}
    ]}
  ],
  "tools": [{"type": "function", "name": "get_weather", "parameters": {...}}],
  "tool_choice": {"type": "function", "name": "get_weather"},
  "stream": true
}

// Claude Messages
{
  "model": "claude-3-opus",
  "system": "You are helpful",
  "messages": [
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": "Hi!"},
    {"role": "user", "content": [
      {"type": "text", "text": "What is this?"},
      {"type": "image", "source": {"type": "base64", "media_type": "image/png", "data": "..."}}
    ]}
  ],
  "tools": [{"name": "get_weather", "input_schema": {...}}],
  "tool_choice": {"type": "tool", "name": "get_weather"},
  "stream": true,
  "max_tokens": 4096
}
```

#### A.2 响应消息

```json
// Chat Completions
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "model": "gpt-4",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "Hi!", "tool_calls": [...]},
    "finish_reason": "stop"
  }],
  "usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
}

// Response API
{
  "id": "resp-123",
  "object": "response",
  "model": "gpt-4",
  "output": [
    {"type": "message", "role": "assistant", "content": [{"type": "output_text", "text": "Hi!"}]},
    {"type": "function_call", "call_id": "call_1", "name": "get_weather", "arguments": "{...}"}
  ],
  "usage": {"input_tokens": 10, "output_tokens": 5, "total_tokens": 15},
  "status": "completed"
}

// Claude Messages
{
  "id": "msg_123",
  "type": "message",
  "model": "claude-3-opus",
  "role": "assistant",
  "content": [
    {"type": "thinking", "thinking": "Let me think...", "signature": "ErQCC..."},
    {"type": "text", "text": "Hi!"},
    {"type": "tool_use", "id": "toolu_1", "name": "get_weather", "input": {...}}
  ],
  "stop_reason": "tool_use",
  "usage": {"input_tokens": 10, "output_tokens": 5}
}
```

### B. 关键代码路径参考

| 功能 | 代码位置 |
|---|---|
| 格式检测 | `relay/format/detect.go` |
| 自动路由中间件 | `middleware/api_format_detect.go` |
| CC→RA 请求转换 | `relay/adaptor/openai/responseapi_convert_request.go` |
| RA→CC 请求转换 | `relay/controller/response_fallback.go` |
| CC→CM 请求转换 | `relay/adaptor/openai_compatible/claude_messages.go` |
| CC→CM 响应转换 | `relay/adaptor/openai_compatible/claude_messages.go` |
| CM→CC 请求转换 | `relay/controller/claude_messages.go` |
| RA→CC 流式转换 | `relay/adaptor/openai/response_api_stream_handler.go` |
| CC→CM 流式转换 | `relay/adaptor/openai_compatible/claude_messages.go` (`ConvertOpenAIStreamToClaudeSSE`) |
| CM→CC 流式转换 | `relay/adaptor/anthropic/main.go` |
| RA 非流式→CC | `relay/adaptor/openai/response_api_handler.go` |
| Tool choice 规范化 | `relay/adaptor/openai/claude_conversion.go` |
| Thinking normalization | `relay/adaptor/common/deepseekcompat/` |
| 端点配置 | `relay/channeltype/endpoints.go` |
| 能力检测 | `relay/adaptor/openai/response_model.go` |

### C. 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `AUTO_DETECT_API_FORMAT` | `true` | 启用格式自动检测 |
| `AUTO_DETECT_API_FORMAT_ACTION` | `transparent` | 检测后行为：`transparent` 内部重路由 / `redirect` 302 |
| `APPROXIMATE_TOKEN_ENABLED` | `false` | 启用 token 近似计数（`len(text)*0.38`），影响计费精度 |

### D. 对比测试脚本参考

#### D.1 最小可运行对比脚本（Python）

```python
#!/usr/bin/env python3
"""UniAPI vs 原厂对比测试 - 最小可运行版本"""
import json, time, requests

# ===== 配置 =====
UNIAPI_BASE = "http://localhost:3000"
UNIAPI_KEY = "sk-uniapi-xxx"
ORIGIN_BASE = "https://api.openai.com"
ORIGIN_KEY = "sk-origin-xxx"
MODEL = "gpt-4o"

# ===== 测试用例 =====
TEST_PROMPTS = [
    {"messages": [{"role": "user", "content": "Hello"}], "max_tokens": 50},
    {"messages": [
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "1+1=?"},
    ], "max_tokens": 50},
    {"messages": [{"role": "user", "content": "hi"}], "max_tokens": 50,
     "tools": [{"type": "function", "function": {
         "name": "get_weather", "description": "获取天气",
         "parameters": {"type": "object", "properties": {"city": {"type": "string"}}}
     }}], "tool_choice": "auto"},
]

def request_api(base_url, api_key, payload, stream=False):
    headers = {"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"}
    payload_copy = {**payload, "model": MODEL, "stream": stream}
    resp = requests.post(f"{base_url}/v1/chat/completions", headers=headers, json=payload_copy, timeout=60)
    return resp

def test_structural_fidelity():
    """CMP-01: 结构保真度"""
    for i, payload in enumerate(TEST_PROMPTS):
        uniapi_resp = request_api(UNIAPI_BASE, UNIAPI_KEY, payload)
        origin_resp = request_api(ORIGIN_BASE, ORIGIN_KEY, payload)
        u, o = uniapi_resp.json(), origin_resp.json()
        assert set(u.keys()) == set(o.keys()), f"Case {i}: 顶层 key 不一致"
        assert "choices" in u, f"Case {i}: 缺少 choices"
        assert "usage" in u, f"Case {i}: 缺少 usage"
        print(f"  ✅ Case {i}: 结构保真通过")

def test_usage_accuracy():
    """CMP-07: Usage 精度"""
    for i, payload in enumerate(TEST_PROMPTS[:2]):  # 仅纯文本
        uniapi_resp = request_api(UNIAPI_BASE, UNIAPI_KEY, payload)
        origin_resp = request_api(ORIGIN_BASE, ORIGIN_KEY, payload)
        u, o = uniapi_resp.json(), origin_resp.json()
        u_pt = u["usage"]["prompt_tokens"]
        o_pt = o["usage"]["prompt_tokens"]
        diff = abs(u_pt - o_pt) / max(o_pt, 1) * 100
        status = "✅" if diff < 5 else "⚠️"
        print(f"  {status} Case {i}: prompt_tokens UniAPI={u_pt} 原厂={o_pt} 偏差={diff:.1f}%")

def test_streaming_fidelity():
    """CMP-12: 流式保真度"""
    payload = {**TEST_PROMPTS[0], "stream": True}
    for label, base, key in [("UniAPI", UNIAPI_BASE, UNIAPI_KEY), ("原厂", ORIGIN_BASE, ORIGIN_KEY)]:
        headers = {"Authorization": f"Bearer {key}", "Content-Type": "application/json"}
        payload_copy = {**payload, "model": MODEL}
        resp = requests.post(f"{base}/v1/chat/completions", headers=headers,
                             json=payload_copy, stream=True, timeout=60)
        events = []
        has_done = False
        for line in resp.iter_lines():
            line = line.decode("utf-8").strip()
            if line.startswith("data: "):
                data = line[6:]
                if data == "[DONE]":
                    has_done = True
                else:
                    events.append(json.loads(data).get("choices", [{}])[0].get("delta", {}))
        print(f"  {'✅' if has_done else '❌'} {label}: 收到 {len(events)} 个事件, [DONE]={'有' if has_done else '无'}")

def test_error_propagation():
    """CMP-24: 错误传播"""
    bad_payload = {"model": "nonexistent-model-xyz", "messages": [{"role": "user", "content": "hi"}]}
    uniapi_resp = request_api(UNIAPI_BASE, UNIAPI_KEY, bad_payload)
    origin_resp = request_api(ORIGIN_BASE, ORIGIN_KEY, bad_payload)
    match = "✅" if uniapi_resp.status_code == origin_resp.status_code else "❌"
    print(f"  {match} 状态码: UniAPI={uniapi_resp.status_code} 原厂={origin_resp.status_code}")
    if "error" in uniapi_resp.json():
        print(f"  ✅ UniAPI 返回 error 字段")
    else:
        print(f"  ❌ UniAPI 未返回 error 字段")

if __name__ == "__main__":
    print("=== UniAPI vs 原厂对比测试 ===\n")
    print("--- 1. 结构保真度 (CMP-01) ---")
    test_structural_fidelity()
    print("\n--- 2. Usage 精度 (CMP-07) ---")
    test_usage_accuracy()
    print("\n--- 3. 流式保真度 (CMP-12) ---")
    test_streaming_fidelity()
    print("\n--- 4. 错误传播 (CMP-24) ---")
    test_error_propagation()
    print("\n=== 测试完成 ===")
```

#### D.2 对比测试文件组织

```
test/
  comparison/
    config.yaml                  ← 测试配置（baseurl、apikey、模型列表）
    structural_fidelity_test.py  ← CMP-01 ~ CMP-06
    usage_accuracy_test.py       ← CMP-07 ~ CMP-11
    streaming_fidelity_test.py   ← CMP-12 ~ CMP-15
    tool_call_completeness_test.py ← CMP-16 ~ CMP-20
    latency_overhead_test.py     ← CMP-21 ~ CMP-23
    error_propagation_test.py    ← CMP-24 ~ CMP-27
    run_all.sh                   ← 一键运行全部对比测试
```

### E. 对比测试与格式转换测试的关系

两类测试互补，各有侧重：

```
┌─────────────────────────────────────────────────────────────────┐
│                        测试全景                                   │
├─────────────────────────────┬───────────────────────────────────┤
│   格式转换兼容性 (§2-§6)     │   中转站 vs 原厂对比 (§7)          │
├─────────────────────────────┼───────────────────────────────────┤
│ 方法: 单元测试 + mock 上游     │ 方法: 真实 API 对比                │
│ 关注: 转换逻辑正确性           │ 关注: 中转链路透明性                │
│ 优势: 可重复、无成本、快速      │ 优势: 发现端到端问题                │
│ 劣势: mock 可能与真实行为不同   │ 劣势: 非确定性、有 API 成本         │
│ 覆盖: 9 宫格全量               │ 覆盖: 代表性场景抽样                │
├─────────────────────────────┼───────────────────────────────────┤
│        发现 "转换代码有 bug"     │        发现 "计费与原厂不一致"      │
│        发现 "SSE 事件丢失"      │        发现 "延迟突增"              │
│        发现 "字段映射错误"       │        发现 "错误被吞掉"            │
└─────────────────────────────┴───────────────────────────────────┘
```

**典型工作流**：
1. 对比测试发现 prompt_tokens 偏差 > 5%（如 CMP-07）
2. 定位到 `token.go` 中 `ApproximateTokenEnabled` 或 tiktoken 编码问题
3. 编写/补充格式转换单测覆盖此场景
4. 修复后对比测试验证

---

_本白皮书随 UniAPI 版本迭代持续更新。测试覆盖进展以 CI 仪表盘为准。_
