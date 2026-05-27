# uniapi 缓存命中率提升 — 最终方案 (v3)

> 基于 v2 方案 + 代码级 review 修正后的可落地版本

**适用范围**：Claude Messages、Response API、Chat Completion 三类入口统一优化，只作用于它们最终经 uniapi 中转到 DeepSeek 的路径。DeepSeek adaptor 的原生兼容逻辑不改。

## 链路

```
Claude Code → Claude Messages API
  → uniapi 格式转换 (Claude→OpenAI)
  → DeepSeek API (自动 KV Context Cache)
  → 响应含 prompt_cache_hit_tokens / prompt_cache_miss_tokens
```

## 核心发现

DeepSeek 的缓存机制是**自动的、基于 token 级前缀单元的智能匹配**，不需要任何显式注入。

**真正的问题是数据管道断了**：DeepSeek 返回了 `prompt_cache_hit_tokens`，但 `Usage` 结构体没有解析这个字段，导致：

1. **缓存命中也按全价计费**（v4-flash: $0.14/MTok 而非 $0.0028/MTok，差了 50 倍）
2. **缓存命中率完全不可见**

而计费层 `relay/quota/quota.go` 已经通过 `PromptTokensDetails.CachedTokens` + `CachedInputRatio` 完整支持了缓存定价——**数据没传进去，管道白铺了。**

## 前提与限制

这份方案只优化 uniapi 最终收到的标准化请求，不控制上游智能体如何选择工具、重排消息、插入时间戳或切换模型。

因此：

- 如果上游 agent 的请求前缀稳定，三类入口的缓存命中率和 sticky 效果会明显提升。
- 如果上游 agent 的调用方式不稳定，uniapi 仍然能做计费映射和部分响应缓存，但命中率只能尽力而为，不能写成确定性保证。
- 所有缓存相关优化都以“不改变 DeepSeek adaptor 兼容行为”为前提；若上游没有返回缓存字段，行为仍回退为按全价计费。

---

## 方案总览

| #      | 改造项                                    | 类别      | 工作量 | 预期收益                    |
| ------ | ----------------------------------------- | --------- | ------ | --------------------------- |
| **P0** | 解析 DeepSeek 缓存指标 + 三类入口路由粘性 | 计费+路由 | ~1天   | 缓存命中时立即降费 50x-120x |
| **P1** | 响应缓存                                  | 新增功能  | ~1天   | 捕获重试场景                |
| **—**  | 格式转换稳定性测试                        | 质量保障  | ~2h    | 防止回归                    |
| **—**  | Prometheus 缓存计数器                     | 可观测    | ~30min | 量化效果                    |

**已取消项**（从 v1 中移除）：

| 原方案项                 | 原优先级 | 取消原因                             |
| ------------------------ | -------- | ------------------------------------ |
| cache_control 自动注入   | P0       | DeepSeek 不认，是 Anthropic 专用机制 |
| Claude Code 下游优化文档 | P2       | 用户没有可操作的旋钮                 |
| Anthropic 头部转发       | P2       | DeepSeek 不需要，转发了也无意义      |
| 管理后台缓存面板         | P2       | 过度建设，Prometheus 计数器足够      |
| 缓存分析报告             | P2       | 无人看的定期报告，不做事后文档       |

---

## P0 — 解析缓存指标并正确计费

### 问题

DeepSeek v4 响应返回的 usage：

```json
{
  "prompt_tokens": 1000,
  "completion_tokens": 200,
  "prompt_cache_hit_tokens": 800,
  "prompt_cache_miss_tokens": 200
}
```

当前 `Usage` 结构体没有 `prompt_cache_hit_tokens` / `prompt_cache_miss_tokens`，这两个字段被 JSON 反序列化静默丢弃。

### 改造

#### 1.1 扩展 Usage 结构体

**文件**：`relay/model/misc.go`

```go
type Usage struct {
    // ... 现有字段不变 ...

    // DeepSeek Context Cache（返回于 usage 顶层）
    // https://api-docs.deepseek.com/guides/kv_cache
    PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens,omitempty"`
    PromptCacheMissTokens int `json:"prompt_cache_miss_tokens,omitempty"`
}
```

#### 1.2 在三类计费入口做统一映射

**文件**：`relay/quota/cache_usage.go`（新增）

Claude Messages、Response API、Chat Completion 的最终计费分别走 `relay/controller/claude_messages_billing.go`、`relay/controller/response_billing.go`、`relay/controller/helper.go`。这里改为新增一个共享辅助函数，只在这三条路径调用 `quotautil.Compute` 之前做同一层映射：

```go
func ApplyDeepSeekCacheUsage(usage *model.Usage) {
    // === 新增：DeepSeek 缓存指标统一映射 ===
    if usage == nil || usage.PromptCacheHitTokens <= 0 {
        return
    }
    if usage.PromptTokensDetails == nil {
        usage.PromptTokensDetails = &model.UsagePromptTokensDetails{}
    }
    if usage.PromptTokensDetails.CachedTokens == 0 {
        usage.PromptTokensDetails.CachedTokens = usage.PromptCacheHitTokens
    }
}
```

**为什么放在这里**：这条优化只针对三类入口最终都指向 DeepSeek 的场景，因此把映射提炼成共享函数后，只需在三条最终计费入口各调用一次即可。

在实现时，`helper.go`、`response_billing.go`、`claude_messages_billing.go` 分别先调用这个辅助函数，再进入 `quotautil.Compute`。

#### 1.3 DeepSeek 兼容性边界

这个映射只是在 `Usage` 上补充 `PromptTokensDetails.CachedTokens`，不会改写 DeepSeek 的请求体、响应体解析或 adaptor 兼容逻辑；如果上游不返回缓存字段，现有行为仍然会退化为按全价计费。

#### 1.4 error 场景保护

如果上游超时或返回异常，`usage` 为 nil，辅助函数自动跳过，不会 panic。此外加一个 Prometheus 计数器记录解析成功/失败次数，用于监控。

#### 1.5 验证

```go
// relay/quota/quota_test.go 新增
TestPostConsumeQuota_DeepSeekFullCache     → prompt_cache_hit_tokens = prompt_tokens, 按 cache 价
TestPostConsumeQuota_DeepSeekNoCache       → prompt_cache_hit_tokens = 0, 按全价
TestPostConsumeQuota_DeepSeekPartialCache  → 混合计费，只减缓存部分
TestPostConsumeQuota_DeepSeekNilUsage      → 上游异常时，不 panic
```

### 预期效果

| 模型                          | 当前全价/MTok | 缓存命中/MTok | 60% 命中实际成本 | 节省    |
| ----------------------------- | ------------- | ------------- | ---------------- | ------- |
| deepseek-v4-flash             | $0.14         | $0.0028       | **$0.058**       | **59%** |
| deepseek-v4-pro               | $1.74         | $0.0145       | **$0.702**       | **60%** |
| long session (90% hit, flash) | $0.14         | $0.0028       | **$0.017**       | **88%** |

---

## P0 — 粘性路由加强

### 问题

当前 sticky key：`sticky_session:user:{id}:model:{name}`

两个问题：

1. 现有实现只有 user/model 粘性，没有可复用的会话标识，因此方案不能假装已经能自动识别 Claude Code session
2. TTL 默认 30 分钟太短，Claude Code 单次会话通常 1-3 小时

### 改造

#### 2.1 保持现有粘性键结构，先增强三类入口路由 TTL

```go
// 当前
stickyKey = fmt.Sprintf("sticky_session:user:%d:model:%s", userId, modelName)

// 目标
// 仍然沿用 user/model 粘性键，但只在 Claude Messages、Response API、Chat Completion 三类入口最终路由到 DeepSeek 时提高 TTL。
stickyKey = fmt.Sprintf("sticky_session:user:%d:model:%s", userId, modelName)
```

#### 2.2 TTL 策略

```go
// 当前：固定
ttl = time.Duration(config.StickySessionTimeoutSeconds) * time.Second

// 目标：滑动窗口，最低 4 小时
ttl = max(stickySessionTTL(), 4 * time.Hour)
```

每次命中自动续期。

#### 2.3 降级

绑定 channel 不可用时：

1. 尝试等待恢复（当前 `ChannelPoolRecoveryWaitMax` 已有）
2. 恢复失败 → 选新 channel，清除旧绑定
3. 不重置 sticky 绑定键，新 channel 上重新积累缓存

#### 2.4 会话级隔离的后续扩展

如果后续需要真正的会话级隔离，再单独引入客户端协作字段，例如 `X-Session-Id` 或相似的请求头；这不应与当前三类入口统一优化绑定到同一个交付里。

### 配置项变更

```go
// common/config/config.go 新增
StickySessionMinTTL int  // 最低 TTL（秒），默认 14400 (4h)
```

---

## P1 — 响应缓存 (Response Cache)

### 问题

重试或重复请求时，每次都走完整链路（格式转换 + 上游调用）。Claude Code 的重试间隔通常在秒级。

### 改造

#### 3.1 缓存键

```go
cacheKey = SHA256(
    req.Method + ":" +          // POST
    req.URL.Path + ":" +        // /v1/messages
    string(reqBody) + ":" +     // 原始请求 body（Claude 格式）
    mappedModel + ":" +         // 转换后的 model
    apiKeyPrefix                // API Key 前 8 位
)
```

#### 3.1.1 适用范围

只缓存 Claude Messages、Response API、Chat Completion 三类入口的非流式请求：

- `meta.IsStream == false`
- 不含 tool call 的纯响应
- 非 MCP tool-search loop
- 非 background / websocket 类长连接响应

这样可以避免把流式事件缓存成一次性快照，也避免把幂等性不明确的交互误判为重试。

#### 3.2 存储

两级，线程安全：

```go
// L1: 内存 sync.Map
type memCache struct {
    store sync.Map  // map[string]*cacheEntry
    ttl    time.Duration
    max    int
}

// L2: Redis（分布式）
// Key: "response_cache:{sha256}"
// TTL: 300s
```

#### 3.3 集成位置

在三类入口对应的请求处理链路中，格式转换之前查询，上游响应之后写入：

```
Request → tryServeFromCache
  ├─ HIT → 直接返回缓存响应
  └─ MISS → ConvertClaudeRequest → DoRequest → tryCacheResponse
```

其中 Claude Messages 走 Claude 请求结构，Response API 和 Chat Completion 走各自的请求结构，但缓存键和命中条件保持一致，只针对最终会被送往 DeepSeek 的标准化请求体生效。

如果上游 agent 不稳定导致这些标准化请求体本身频繁变化，缓存命中率也会随之下降；uniapi 只能复用已经稳定下来的请求，而不能把不稳定的前缀“修”成稳定前缀。

#### 3.4 缓存规则

```go
func shouldCache(resp *http.Response, body []byte) bool {
    if resp.StatusCode != http.StatusOK { return false }
    if len(body) > 1<<20 { return false }        // >1MB 不缓存
    if containsToolCall(body) { return false }    // 含 tool call 不缓存
    return true
}
```

同时增加一条前置规则：请求本身是 stream 时直接跳过缓存，不进入命中判断。

#### 3.5 配置

```yaml
# Channel Config
response_cache:
  enabled: true
  ttl_seconds: 60
  max_entry_bytes: 1048576
  max_entries: 1000
```

---

## 质量保障 — 格式转换稳定性测试

### 不再单独成项

经过分析，当前格式转换已经是确定性的：

- Go `json.Marshal` 对 struct 确定，对 map 按 key 排序
- `ConvertClaudeRequest` 是纯函数，无随机变量
- 当前不是瓶颈

### 只需加两个回归测试

```go
// relay/adaptor/openai_compatible/claude_convert_regression_test.go

// 1. 相同输入 → 相同输出
func TestConvertClaudeRequest_Deterministic(t *testing.T) {
    input := loadFixture("claude_request_fixture.json")
    r1, _ := ConvertClaudeRequest(ctx, input)
    r2, _ := ConvertClaudeRequest(ctx, input)
    b1, _ := json.Marshal(r1)
    b2, _ := json.Marshal(r2)
    require.Equal(t, b1, b2)
}

// 2. Session 内 append-only，前缀稳定
func TestConvertClaudeRequest_AppendOnlyPrefix(t *testing.T) {
    turn1 := loadFixture("session_turn_1.json")
    turn2 := loadFixture("session_turn_2.json")
    c1, _ := ConvertClaudeRequest(ctx, turn1)
    c2, _ := ConvertClaudeRequest(ctx, turn2)
    b1, _ := json.Marshal(c1)
    b2, _ := json.Marshal(c2)
    require.True(t, bytes.HasPrefix(b2, b1))
}
```

---

## 可观测性 — Prometheus 计数器（30 分钟顺带做）

做完 P0 的计费映射后在 `PostConsumeQuota` 顺带加三个 counter，不单独建文件、不做面板：

```go
// 在三类入口的 PostConsumeQuota 逻辑旁
cacheHitTokens := 0
if usage.PromptTokensDetails != nil {
    cacheHitTokens = max(usage.PromptTokensDetails.CachedTokens, 0)
}
cacheMissTokens := max(promptTokens - cacheHitTokens, 0)

metrics.CacheHitTokensTotal.WithLabelValues(channelName, modelName).Add(float64(cacheHitTokens))
metrics.CacheMissTokensTotal.WithLabelValues(channelName, modelName).Add(float64(cacheMissTokens))
```

---

## 实施路线图

```
P0（1天）：
  relay/model/misc.go          +2 个字段
  relay/quota/cache_usage.go   +共享映射函数
  relay/controller/helper.go   +调用共享映射
  relay/controller/response_billing.go +调用共享映射
  relay/controller/claude_messages_billing.go +调用共享映射
  relay/quota/quota_test.go    +4 个测试用例
  model/sticky_session.go      +三类入口路由 TTL 调整
  common/config/config.go      +1 个配置项

P1（1天）：
  middleware/response_cache.go 新建 ~150 行
  relay/controller/claude_messages.go  +查询/写入钩子

质量保障（2小时）：
  relay/adaptor/openai_compatible/claude_convert_regression_test.go 新建

可观测（30分钟）：
  common/metrics/cache.go 新建 ~30 行
  relay/quota/cache_usage.go   +Prometheus 埋点
```

---

## 文件变更清单（最终版）

| 文件                                                                | 变更                               | 行数估计 |
| ------------------------------------------------------------------- | ---------------------------------- | -------- |
| `relay/model/misc.go`                                               | Usage 加 2 个字段                  | +5       |
| `relay/quota/cache_usage.go`                                        | DeepSeek 缓存映射共享函数          | +15      |
| `relay/controller/helper.go`                                        | 调用共享映射函数                   | +3       |
| `relay/controller/response_billing.go`                              | 调用共享映射函数                   | +3       |
| `relay/controller/claude_messages_billing.go`                       | 调用共享映射函数                   | +3       |
| `relay/quota/quota_test.go`                                         | 4 个 DeepSeek 缓存计费测试         | +80      |
| `model/sticky_session.go`                                           | 三类入口路由 TTL 调整              | +10      |
| `common/config/config.go`                                           | StickySessionMinTTL                | +3       |
| `middleware/response_cache.go`                                      | **新建** — 内存+Redis 双层缓存     | +150     |
| `relay/controller/claude_messages.go`                               | 响应缓存查询/写入                  | +25      |
| `relay/adaptor/openai_compatible/claude_convert_regression_test.go` | **新建** — 确定性测试              | +60      |
| `common/metrics/cache.go`                                           | **新建** — 3 个 Prometheus counter | +30      |

**共 12 文件，约 +383 行。** 其中新增 4 个文件，修改 8 个文件。无文档、无面板、无报告。

---

## 风险对照

| 风险                        | 影响                   | 应对                                                                                                                          |
| --------------------------- | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| DeepSeek 缓存指标字段名变化 | 计费不准               | `PromptCacheHitTokens` 字段不命中时 json 值为 0，自动走全价，不会多收只会少收                                                 |
| session_id 自动生成误合并   | 不同 session 抢 sticky | 方案不再引入自动 session_id；仅对三类入口最终路由到 DeepSeek 的请求提高 sticky TTL，真正的 session 隔离留给后续客户端协作版本 |
| 响应缓存误命中              | 返回过时响应           | 只缓存非流式、无 tool call 的成功响应，TTL 60s，且请求必须显式走缓存路径                                                      |
| 格式转换非确定性回归        | 前缀不稳定             | 两个回归测试在 CI 中拦截                                                                                                      |
