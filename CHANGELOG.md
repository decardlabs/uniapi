# Changelog

## v3.9.10 (2026-05-29)

- 计费口径一致性修复：独立上报 `reasoning_tokens` 合并计入 completion 计费，避免 DeepSeek V4 等低估计费
- 修复 Zhipu `glm-*` v4 鉴权头兼容性，按官方文档改用 `Bearer <api-key>`，保留 v3 JWT 兼容
- 新增 Zhipu base URL 归一化（同时支持 `open.bigmodel.cn` 与 `open.bigmodel.cn/api/paas/v4`）
- GLM 渠道映射到 Zhipu 适配器，补齐常用多模态模型默认列表
- 修正部署文档 Dockerfile 路径说明

## v3.9.9 (2026-05-29)

- MiniMax Chat/Response fallback 消息角色归一化（`system/tool/developer/function` → `user`）
- 新增适配器契约测试与 Response API fallback 端到端回归
- 明确多模态升级文档为历史提案状态
- 补充 `MULTIMODAL_ROUTE_MODE` 与 `MULTIMODAL_VISION_FALLBACK_MODEL` 说明
- 新增部署建议文档

## v3.9.8 (2026-05-28)

- 前端渠道编辑页 `other` 字段兼容历史字符串格式
- text-only 模型图片输入前置校验（鉴权层 + OpenAI 适配层双重防线）
- 自动视觉模型路由：按 `MULTIMODAL_ROUTE_MODE` 切换固定回退/按渠道能力选择
- `cmd/migrate -backfill-vision-capabilities` 批量回填历史渠道视觉能力

## v3.9.7 (2026-05-28)

- 修复 `/api/user/cache-analytics` 被 `/:id` 动态路由误匹配
- 修复缓存分析 SQL `created_at` 字段歧义（SQLite）

## v3.9.6 (2026-05-27)

- 新增管理员 Cache Analytics 页面（前端）
- 后端新增 `/api/user/cache-analytics` 聚合接口

## v3.9.5 (2026-05-25)

- 对齐后端 `common.Version`、前端 `package.json` 与锁文件版本元数据

## v3.9.3 (2026-05-25)

- 修正 Zhipu 适配器响应处理
- DeepSeek/MiniMax/Moonshot/Zhipu 渠道兼容性 review

## v3.9.2 (2026-05-25)

- 后端版本常量、前端包版本与锁文件、迁移工具版本一致性对齐

## v3.9.1 (2026-05-25)

- MiniMax 工具调用历史 `tool` 消息 `name` 回填逻辑
- MiniMax 默认 Base URL 升级为 `https://api.minimaxi.com/v1`
- Response API fallback `prompt`/`background` 不支持参数契约测试

## v3.8.11 (2026-05-25)

- MiniMax 默认 Base URL 更新为 `https://api.minimaxi.com/v1`

## v3.8.10 (2026-05-18)

- OpenAI-compatible 统一流路径 SSE 心跳保活（5s 心跳帧）

## v3.8.9 (2026-05-14)

- 渠道限流预检（`WouldChannelRateLimitBlock`）
- 能力暂停精确至 `(group, model, channel_id)` 三元组，差异化冷却窗口
- 413 错误自动搜索更大 `max_tokens` 渠道
- `STICKY_SESSION_ENABLED` 开关
- `go run ./cmd/test live` 真实渠道探测命令

## v3.8.7 (2026-05-14)

- 缩短能力暂停默认窗口（429: 15s, 5xx: 10s）
- `go run ./cmd/test live` 跳过 DeepSeek 强制 `tool_choice`

## v3.8.6 (2026-05-14)

- `STICKY_SESSION_ENABLED` 默认开启
- 选路观测日志

## v3.8.5 (2026-05-14)

- `go run ./cmd/test live` 真实渠道探测命令

## v3.8.4 (2026-05-14)

- 按用户+模型的 sticky session 账号绑定
- Response API `response_id`/`previous_response_id` 跨节点绑定
- 选路阶段账号限流预检

## v3.8.3 (2026-05-05)

- 渠道编辑页模型区重设计：推荐模型名 + 供应商目录导入卡片
- 加载供应商默认配置自动补齐占位定价

## v3.8.2 (2026-05-05)

- 请求模型映射 JSON 自动补齐
- 设置页"检查更新"优先读取 GitHub 最新 tag

## v3.8.1 (2026-05-04)

- 修正渠道工具测试过时 channel type 常量

## v3.8.0 (2026-05-04)

- 新增 Doubao、TencentTokenHub、GLM、Kimi 完整注册
- 统一各渠道 base URL 默认值
- 修复 DeepSeek 工具消息内容归一化

## v3.6.1 (2026-04)

- DeepSeek reasoning_content 注入修复
- Claude→OpenAI tool_use 名称回填

## v3.6.0 (2026-04)

- 动态渠道类型模板系统
- 全局动态渠道注册机制

## v3.5

- MCP 聚合代理
- Response API 完整支持

## v3.1.x

- 实时计费
- 多轮 tool_use 兼容性

## v3.0.0

- Modern UI 全量重构
- 移除旧模板
