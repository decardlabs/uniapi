# UniAPI DeepSeek 多模态支持升级方案（历史提案）

> 文档状态：Historical Draft（部分内容已落地）
>
> 说明：本文档最初用于评审与规划。其后续部分能力已在版本发布中落地，请以 README 版本历史为准：`v3.9.8`（多模态请求容错升级）。本文保留用于记录设计思路与演进背景。

## 1. 背景与问题定义

当前编程智能体在同一会话中发送过截图（image 相关内容）后，即便后续只输入文本，也可能因历史消息仍包含图片内容而持续触发多模态校验路径，表现为重复报错或“循环失败”。

该问题本质是：

- 会话历史重放（客户端行为）
- 服务端图片内容检测按整段消息集合生效（非仅最后一轮）
- 模型能力、路由策略、Token 白名单三者可能不一致

## 2. 升级目标

1. DeepSeek 作为主模型时，文本请求稳定可用。
2. 截图请求行为可控：要么正确路由到可用视觉模型，要么一次性返回清晰错误。
3. 避免同一会话里“重复带图重放”导致连续失败。
4. 兼顾可观测性：能快速定位失败是权限、能力、还是重放行为引起。

## 3. 现状评估（基于当前代码）

- 服务端会检测请求体中的图片内容标记（image_url / input_image / image）并进入含图分支。
- DeepSeek 文本模型在现有逻辑中通常被识别为 text-only（特定名称模式例外）。
- 在 fixed_fallback 模式下，含图且 text-only 时可自动改路由到视觉模型（例如 gpt-4o）。
- Token 白名单会同时校验“请求模型”和“自动路由后的模型”；若未包含回退模型，仍会失败。

结论：仅靠“我这次没再上传图片”并不总能恢复，关键是请求是否仍带历史图片块。

## 4. 总体方案（分层治理）

采用“配置先止血 + 服务端策略增强 + 客户端防重放协同”三层方案。

### 4.1 配置层（立即止血）

目标：在不改代码前提下先提高成功率。

建议配置：

- `MULTIMODAL_ROUTE_MODE=fixed_fallback`
- `MULTIMODAL_VISION_FALLBACK_MODEL=gpt-4o`

同时满足：

1. rootgroup 下有可用 DeepSeek 文本渠道。
2. rootgroup 下有可用 gpt-4o 视觉渠道。
3. 编程智能体 Token 白名单至少包含：DeepSeek 主模型 + gpt-4o。

### 4.2 服务端策略层（中期增强）

目标：降低历史重放导致的“假文本、真含图”误判风险。

新增配置建议：

- `MULTIMODAL_IMAGE_DETECTION_SCOPE`
  - `all_history`（默认，当前行为）
  - `last_user_turn`（推荐，优先检查最后一轮用户输入）

新增策略建议：

- `MULTIMODAL_RETRY_GUARD_ENABLED=true`
- `MULTIMODAL_RETRY_GUARD_WINDOW_SECONDS=120`

行为定义：

1. 若同一会话在短时间内连续出现同类“含图能力错误”，触发一次性阻断。
2. 返回明确错误提示：请清理历史图片消息或新建会话再试。
3. 不再继续自动回退与重复校验，避免循环消耗。

### 4.3 客户端协同层（根治闭环）

目标：从源头减少“失败重放”。

建议客户端策略：

1. 遇到图片能力错误后，只允许一次自动修复重试。
2. 重试必须裁剪历史中的图片块（或新建会话）。
3. 第二次失败直接停止自动重试并提示人工处理。

## 5. DeepSeek 专项升级点（代码改造可能性）

### 5.1 能力判定从启发式升级为能力注册表

问题：通过模型名包含关系推断 text-only / vision，扩展性弱。

改造方向：

1. 新增模型能力注册结构：`supports_text`, `supports_vision`, `supports_tools`, `supports_response_api`。
2. DeepSeek 型号按能力显式登记。
3. 路由时优先读取能力表，不再依赖字符串猜测。

### 5.2 新增 DeepSeek 优先路由策略（可选）

新增模式建议：`deepseek_first`

含图请求流程：

1. 若请求为 DeepSeek 文本模型：先尝试 DeepSeek 视觉模型池。
2. 若不可用：再回退到 `MULTIMODAL_VISION_FALLBACK_MODEL`（如 gpt-4o）。
3. 若仍不可用：返回明确错误，不循环。

### 5.3 统一权限语义与错误语义

1. 先判断能力（模型是否支持含图输入）。
2. 再判断权限（Token 是否允许实际路由模型）。
3. 错误分层返回：
   - `capability_not_supported`
   - `token_model_not_allowed`
   - `multimodal_history_replay_detected`

## 6. 实施阶段与里程碑

### 阶段 A（1 天，配置落地）

- 固化多模态环境变量。
- 校验组、渠道、Token 白名单一致性。
- 完成文本/截图两类手工验收。

产出：业务可用性显著改善。

### 阶段 B（3-5 天，服务端增强）

- 增加检测范围配置（all_history / last_user_turn）。
- 增加 retry guard。
- 增加结构化日志字段：
  - `image_detected`
  - `image_detection_scope`
  - `auto_routed_model`
  - `retry_guard_hit`

产出：循环失败显著下降，问题可观测。

### 阶段 C（5-10 天，能力化改造）

- 引入模型能力注册表。
- 新增 `deepseek_first` 策略。
- 清理历史兼容分支中的重复判定逻辑。

产出：DeepSeek 多模态支持可扩展、可维护。

## 7. 测试与验收标准

### 7.1 测试矩阵

1. Chat Completions（文本、含图）
2. Claude Messages（文本、含图）
3. Responses API（文本、含图）
4. Token 白名单缺失/完整两组
5. all_history 与 last_user_turn 两种检测范围
6. retry guard 开关前后对比

### 7.2 通过标准

1. 文本请求成功率不低于当前基线。
2. 含图请求在可用路由下成功；不可用时一次性失败并提示可操作建议。
3. 同一会话中不再出现 3 次以上同类自动重试失败。
4. 日志可明确判定失败原因。

## 8. 风险与回滚

### 8.1 风险

1. `last_user_turn` 模式可能放宽过度，漏检某些依赖历史图片上下文的场景。
2. 新增策略分支后，路由与权限判断顺序不一致可能引入新边缘问题。

### 8.2 回滚策略

1. 任何异常可先回退为：
   - `MULTIMODAL_IMAGE_DETECTION_SCOPE=all_history`
   - `MULTIMODAL_ROUTE_MODE=fixed_fallback`
2. 保留 retry guard 开关可快速关闭。
3. 按阶段发布，避免一次性大改。

## 9. 建议的最小可行路线（推荐）

按优先级执行：

1. 先做阶段 A（配置一致性）
2. 再做阶段 B（检测范围 + retry guard）
3. 最后做阶段 C（能力注册表 + deepseek_first）

该顺序可在最短时间内降低生产问题，同时为后续 DeepSeek 多模态能力演进打好基础。

---

## 附录：执行边界

- 本文档是“升级方案”，不是“变更记录”。
- 本文档不代表任何代码已修改、配置已上线。
- 进入实施前，应单独产出任务拆解与回归测试清单。
