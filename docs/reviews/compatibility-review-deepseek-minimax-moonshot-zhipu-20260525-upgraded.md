# UniAPI 模型渠道兼容性 Review 报告

**生成时间：** 2026-05-25  
**审查范围：** DeepSeek、MiniMax、Moonshot (Kimi)、Zhipu (GLM)  
**审查基准：** 各厂商官方 API 文档 + 当前仓库实现

---

## 1. 结论摘要

本次 review 只认官方文档明确支持的能力，并对照当前仓库实现判断 UniAPI 的渠道兼容性。

### 总体结论

| 渠道            | 官方文档口径                                                                                                                                                                | 当前实现状态                                                                                                   | 结论                                                                             |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| DeepSeek        | OpenAI/Anthropic 兼容的 chat/completions；支持 thinking、tools、json/object/json_schema；未看到官方 Responses API 文档                                                      | ChatCompletions 与 Claude Messages 路径较完整；代码里额外做了 ResponseAPI 路由                                 | **Chat/Claude 兼容良好，Responses API 属于实现层兼容，不是官方文档明确承诺能力** |
| MiniMax         | 提供 OpenAI/Anthropic 兼容的语言模型接口；支持 tools；支持 JSON Mode；未看到官方 Responses API 文档                                                                         | ChatCompletions 与 Claude Messages 可用；代码里有 ResponseAPI 路由，但请求转换并未按 Response API 语义单独处理 | **Chat/Claude 可用，Response API 目前是高风险的“名义支持”**                      |
| Moonshot (Kimi) | 官方文档明确为 `POST /v1/chat/completions`；支持 tools、JSON Mode、JSON Schema、thinking / preserved thinking；未看到 Responses API 文档                                    | ChatCompletions 与 Claude Messages 路径可用；无 Responses API 支持                                             | **Chat/Claude 兼容良好，Responses API 不支持**                                   |
| Zhipu (GLM)     | 官方文档提供 OpenAI API 兼容 `https://open.bigmodel.cn/api/paas/v4/`，Claude API 兼容 `https://open.bigmodel.cn/api/anthropic/v1/messages`；支持 tools、JSON mode、thinking | v4 OpenAI 路径和 Claude 转换路径基本对齐；v3 专有路径属于仓库内兼容层                                          | **OpenAI/Claude 兼容良好，v3 是仓库兼容层，Responses API 不在官方文档范围内**    |

### 关键判断

1. 四家官方文档都明确支持的核心能力是 ChatCompletions / Claude-style messages、工具调用、JSON 输出或结构化输出、thinking / reasoning（不同厂商表述不同）。
2. 只有 Zhipu 的官方文档明确给出 OpenAI 与 Claude 两套兼容入口；DeepSeek、MiniMax、Kimi 的官方材料都更偏向“OpenAI/Anthropic 兼容 chat/completions”，而不是 OpenAI Responses API。
3. 当前仓库里对 `/v1/responses` 的支持，更多是实现层的桥接策略，不是这些厂商官方文档明确承诺的兼容面，因此应该按“需要额外回归测试”的高风险能力来对待。

---

## 2. 官方文档证据

### 2.1 DeepSeek

官方文档要点：

| 项         | 文档结论                                                                                               |
| ---------- | ------------------------------------------------------------------------------------------------------ |
| Base URL   | `https://api.deepseek.com`                                                                             |
| 兼容格式   | OpenAI / Anthropic API 格式                                                                            |
| 聊天接口   | `chat/completions` 示例清晰可见                                                                        |
| 模型       | `deepseek-v4-flash`、`deepseek-v4-pro`；`deepseek-chat` / `deepseek-reasoner` 标注将于 2026/07/24 废弃 |
| Thinking   | 文档示例直接包含 `thinking` 与 `reasoning_effort`                                                      |
| Tools      | 官方文档提到 Agent tools / Anthropic API 指南                                                          |
| 结构化输出 | 官方文档示例包含 JSON / Anthropic 兼容能力；未见 Responses API                                         |

### 2.2 MiniMax

官方文档要点：

| 项            | 文档结论                                                                               |
| ------------- | -------------------------------------------------------------------------------------- |
| Base URL      | 文档中心提供 OpenAI / Anthropic 兼容入口                                               |
| 兼容格式      | OpenAI API 兼容、Anthropic API 兼容                                                    |
| 聊天接口      | `HTTP API（OpenAI API 兼容）`、`HTTP API（Anthropic API 兼容）` 都指向语言模型对话能力 |
| Tools         | 明确支持工具调用 / agentic workflow                                                    |
| 结构化输出    | 官方文档给出 OpenAI/Anthropic 兼容调用示例；语言模型接口强调对话与工具调用             |
| Responses API | 文档索引中未看到官方 Responses API 页面                                                |

### 2.3 Moonshot / Kimi

官方文档要点：

| 项                | 文档结论                                                                                             |
| ----------------- | ---------------------------------------------------------------------------------------------------- |
| Base URL          | `https://api.moonshot.cn/v1`                                                                         |
| 兼容格式          | OpenAI `chat.completions`；并提供 `Anthropic` 兼容入口说明                                           |
| 聊天接口          | 明确为 `POST /v1/chat/completions`                                                                   |
| Tools             | 官方页面完整说明 `tools`、`tool_calls`、`tool_choice=auto`、流式 `delta.tool_calls` 处理             |
| JSON Mode         | 官方文档明确支持 `response_format={"type":"json_object"}`                                            |
| Structured Output | 文档明确支持 `response_format={"type":"json_schema"}`，并说明 `json_schema` 结构与 `strict` 行为     |
| Thinking          | `kimi-k2-thinking`、`kimi-k2.6`、`kimi-k2.5` 支持思考；`thinking.keep="all"` 支持 preserved thinking |
| Responses API     | 官方文档索引和页面中未见 Responses API                                                               |

### 2.4 Zhipu / GLM

官方文档要点：

| 项                       | 文档结论                                                                                                                                  |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------- |
| OpenAI 兼容入口          | `https://open.bigmodel.cn/api/paas/v4/`                                                                                                   |
| Claude 兼容入口          | `https://open.bigmodel.cn/api/anthropic/v1/messages`                                                                                      |
| Tools                    | `tools` + `tool_choice="auto"`，`tool_calls` 返回 `function.name` / `function.arguments`                                                  |
| JSON / Structured Output | `response_format={"type":"json_object"}` 明确支持                                                                                         |
| Thinking                 | `extra_body={"thinking": {"type": "enabled"}}`；GLM-5.1 / GLM-5 / GLM-4.7 默认开启 thinking；`clear_thinking: False` 可保留历史 reasoning |
| Claude 兼容              | 官方文档明确推荐 Anthropic SDK 迁移到智谱的 `api/anthropic` 入口                                                                          |
| Responses API            | 官方文档检索结果中未见 Responses API                                                                                                      |

---

## 3. 代码实现对照

### 3.1 DeepSeek

**文件：** [relay/adaptor/deepseek/adaptor.go](../../relay/adaptor/deepseek/adaptor.go)

当前实现与官方文档的匹配点：

| 能力              | 实现情况                                                                                     | 评语                                                    |
| ----------------- | -------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| ChatCompletions   | 使用 `openai_compatible.ConvertClaudeRequest`、`HandleClaudeMessagesResponse`、thinking 处理 | 与官方文档一致                                          |
| Tools             | 依赖共享的 Claude/OpenAI 转换，支持 `tool_calls`                                             | 与官方文档一致                                          |
| Thinking          | 会从原始 Claude 请求恢复 `thinking`，并注入 `reasoning_content`                              | 与官方文档方向一致                                      |
| Structured Output | 对 `response_format.json_schema` 做系统指令降级                                              | 属于兼容层策略，方向合理                                |
| Responses API     | `GetRequestURL` 把 `relaymode.ResponseAPI` 路由到 `/v1/chat/completions`                     | **实现层支持，但官方文档未明确给出 Responses API 能力** |

### 3.2 MiniMax

**文件：** [relay/adaptor/minimax/adaptor.go](../../relay/adaptor/minimax/adaptor.go)

当前实现与官方文档的匹配点：

| 能力                     | 实现情况                                                                   | 评语                                                |
| ------------------------ | -------------------------------------------------------------------------- | --------------------------------------------------- |
| ChatCompletions          | 使用 `openai_compatible.ConvertClaudeRequest`                              | 与官方文档一致                                      |
| Tools                    | 已做 `BackfillToolMessageNamesFromToolCalls`                               | 与工具调用文档方向一致                              |
| JSON / Structured Output | 对 `response_format.json_schema` 做 `EnsureInstruction` 降级               | 与官方文档的 JSON/兼容能力方向一致                  |
| Responses API            | `GetRequestURL` 接受 `relaymode.ResponseAPI` 并转到 `/v1/chat/completions` | **高风险：URL 上接受了，但请求/响应语义未单独建模** |

### 3.3 Moonshot / Kimi

**文件：** [relay/adaptor/moonshot/adaptor.go](../../relay/adaptor/moonshot/adaptor.go)

当前实现与官方文档的匹配点：

| 能力                          | 实现情况                                                                        | 评语                                               |
| ----------------------------- | ------------------------------------------------------------------------------- | -------------------------------------------------- |
| ChatCompletions               | `GetRequestURL` 明确路由到 `/v1/chat/completions`                               | 与官方文档一致                                     |
| Tools                         | 复用 `openai_compatible.ConvertClaudeRequest` 与 `HandleClaudeMessagesResponse` | 与官方文档一致                                     |
| JSON Mode / Structured Output | 对 `response_format.json_schema` 做 `EnsureInstruction` 降级                    | 与官方文档一致                                     |
| Thinking                      | 当前代码未显式操控 `thinking.keep` / preserved thinking                         | **与官方文档相比，保留思考内容的能力未被完整利用** |
| Responses API                 | 代码未注册                                                                      | 与官方文档索引一致（官方未见此能力）               |

### 3.4 Zhipu / GLM

**文件：** [relay/adaptor/zhipu/adaptor.go](../../relay/adaptor/zhipu/adaptor.go)

当前实现与官方文档的匹配点：

| 能力                     | 实现情况                                                           | 评语                                                                     |
| ------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| OpenAI 兼容 v4           | `base_url` 对应的 OpenAI-compatible 路径与官方文档方向一致         | 与官方文档一致                                                           |
| Claude 兼容              | 改用 `openai_compatible.ConvertClaudeRequest`，再做 Zhipu 特定处理 | 与官方文档一致                                                           |
| Tools                    | 通过 `tool_calls` 与 `tool_choice` 工作                            | 与官方文档一致                                                           |
| JSON / Structured Output | `response_format.json_schema` 降级为系统指令                       | 与官方文档一致                                                           |
| Thinking                 | 会清理 `reasoning_effort` / `response_format`，并保留 v4 兼容路径  | 部分对齐；官方文档里 `thinking` 是明确能力，当前实现主要依赖上游请求格式 |
| v3 专有路径              | 代码仍保留 `api/paas/v3/model-api/...`                             | **这是仓库兼容层，不是官方 OpenAI/Claude 文档主线**                      |

---

## 4. 兼容性评审

### 4.1 ChatCompletions / Claude Messages

这部分是四家最稳定的共同能力。

| 渠道            | 结论                                  |
| --------------- | ------------------------------------- |
| DeepSeek        | 兼容良好                              |
| MiniMax         | 兼容良好                              |
| Moonshot / Kimi | 兼容良好                              |
| Zhipu / GLM     | 兼容良好（v4 OpenAI + Claude 双入口） |

### 4.2 Tools / Function Calling

| 渠道            | 官方文档                                                              | 当前实现 |
| --------------- | --------------------------------------------------------------------- | -------- |
| DeepSeek        | 有 agent tools / Anthropic 指南                                       | 可用     |
| MiniMax         | 文档明确支持工具调用                                                  | 可用     |
| Moonshot / Kimi | 文档明确描述 `tool_calls` / `tool_choice=auto` / 流式 tool_calls 处理 | 可用     |
| Zhipu / GLM     | 文档明确支持 `tools` / `tool_choice=auto` / `tool_calls`              | 可用     |

### 4.3 JSON Mode / Structured Output

| 渠道            | 官方文档                                        | 当前实现                                      |
| --------------- | ----------------------------------------------- | --------------------------------------------- |
| DeepSeek        | 兼容 OpenAI/Anthropic；示例与实际支持场景可对照 | 通过 `response_format.json_schema` 做指令降级 |
| MiniMax         | OpenAI/Anthropic 兼容文档；无独立 Responses API | 通过 `EnsureInstruction` 降级                 |
| Moonshot / Kimi | 明确支持 `json_object` 和 `json_schema`         | 通过 `EnsureInstruction` 降级                 |
| Zhipu / GLM     | 明确支持 `json_object`                          | 通过 `EnsureInstruction` 降级                 |

### 4.4 Thinking / Reasoning

| 渠道            | 官方文档                                                         | 当前实现风险                                                           |
| --------------- | ---------------------------------------------------------------- | ---------------------------------------------------------------------- |
| DeepSeek        | 明确支持 thinking / reasoning_effort                             | 低                                                                     |
| MiniMax         | 官方索引里未见 explicit thinking 说明                            | 中：实现中清理了不支持字段，属于稳妥兼容                               |
| Moonshot / Kimi | 明确支持 `kimi-k2-thinking`、`thinking.keep`、preserved thinking | 中：代码没有单独建模 `keep`，但基础兼容可用                            |
| Zhipu / GLM     | 明确支持 `thinking` / `clear_thinking` / reasoning_content 保留  | 中：当前实现依赖上游兼容路径，未显式把所有 thinking 变体都做成专用分支 |

### 4.5 Responses API

这是本次 review 的核心风险点。

| 渠道            | 官方文档证据                  | 当前实现                                                                 | 结论                                     |
| --------------- | ----------------------------- | ------------------------------------------------------------------------ | ---------------------------------------- |
| DeepSeek        | 未找到官方 Responses API 文档 | 代码把 `relaymode.ResponseAPI` 路由到 chat/completions                   | **实现层可尝试，但不能当作官方承诺能力** |
| MiniMax         | 未找到官方 Responses API 文档 | 代码接受 `relaymode.ResponseAPI`，但没有单独的 request/response 语义转换 | **高风险，建议默认视为未完整支持**       |
| Moonshot / Kimi | 未找到官方 Responses API 文档 | 代码未注册                                                               | **不支持**                               |
| Zhipu / GLM     | 未找到官方 Responses API 文档 | 代码未注册                                                               | **不支持**                               |

---

## 5. 关键发现

### P0 - 必须修正的认知

1. **不要把 `/v1/responses` 当成四家官方都支持的能力。**
   - 官方文档的共同交集是 chat/completions、Claude messages、tools、JSON/structured output、thinking。
   - Responses API 在这四家的官方文档里没有形成统一承诺。

2. **MiniMax 的 ResponseAPI 目前是“代码接得上”，不是“官方语义完整支持”。**
   - `GetRequestURL` 只把路径转到了 `/v1/chat/completions`。
   - 但 `ConvertRequest` / `DoResponse` 并没有按 Response API 请求体和响应体单独建模。

### P1 - 建议修正的实现差距

1. **Moonshot 的 preserved thinking 没有被完整建模。**
   - 官方文档明确支持 `thinking.keep="all"`。
   - 当前实现只做了基础 OpenAI/Claude 转换，没有把 `keep` 作为显式兼容能力暴露。

2. **Zhipu 的 v3 路径应当被标注为仓库兼容层，不应和官方 OpenAI 兼容文档混为一谈。**
   - 官方文档主线是 `api/paas/v4` 和 `api/anthropic/v1/messages`。
   - v3 专有路径更多是历史兼容面。

3. **DeepSeek、MiniMax、Moonshot、Zhipu 的 structured output 都属于“兼容层降级实现”，而不是每家都在同一层面承诺了完全一致的 JSON Schema 约束。**
   - 当前 `EnsureInstruction` 的做法合理，但需要单独的回归测试确认不破坏 tool calling / thinking。

---

## 6. 建议的后续动作

### 6.1 文档口径

1. 把仓库对外文档里的“Responses API 支持”收窄成“实现层桥接能力”，不要写成四家官方都原生支持。
2. 对 MiniMax 的 Response API 标记为实验性或受限支持，除非补齐请求 / 响应转换闭环。
3. 对 Moonshot 的 `thinking.keep` 与 Zhipu 的 `clear_thinking` 增加专门的兼容说明。

### 6.2 代码口径

1. 为 MiniMax 补一个明确的 Response API 请求转换层，或者直接去掉 ResponseAPI 路由，避免“看似支持、实则半截子”。
2. 为 Moonshot 增加 preserved thinking 的显式处理和测试。
3. 为 Zhipu 补充 `thinking` / `clear_thinking` 的专门测试，确认 v4 兼容路径与官方行为一致。

### 6.3 测试口径

建议至少补这些回归用例：

| 用例                              | 覆盖对象                               |
| --------------------------------- | -------------------------------------- |
| ChatCompletions + tools           | 四家全部                               |
| Claude Messages + tool_calls 回填 | 四家全部                               |
| JSON Mode / json_schema           | Kimi / Zhipu / DeepSeek / MiniMax      |
| thinking / preserved thinking     | DeepSeek / Kimi / Zhipu                |
| Response API 桥接                 | 仅标记为桥接层测试，不写成官方原生支持 |

---

## 7. 总结

这次 review 的核心结论是：

1. **ChatCompletions / Claude Messages / Tools / JSON Mode / Thinking 才是四家官方文档真正重叠的兼容核心。**
2. **Responses API 不是这四家官方文档的共同支持面，当前仓库里的相关实现应当按高风险兼容层处理。**
3. **DeepSeek、MiniMax、Moonshot、Zhipu 的当前实现整体可用，但应把文档宣称范围收紧到官方真正承诺的能力，不要把实现桥接误写成官方原生能力。**

**报告生成：** 砖家 🔬  
**依据：** DeepSeek / MiniMax / Moonshot / Zhipu 官方文档 + 当前仓库实现
