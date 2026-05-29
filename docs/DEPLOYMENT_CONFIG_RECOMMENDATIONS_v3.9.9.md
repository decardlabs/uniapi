# UniAPI v3.9.9 部署与配置建议

## 1. 文档目标

本文面向 v3.9.9 发布后的实际部署，提供一套可执行的配置基线与上线流程，重点覆盖：

1. 数据库路径与运行实例一致性
2. DeepSeek / MiniMax 兼容性优先的通道配置
3. 多模态路由相关开关的安全默认值
4. 发布验证、回滚与故障排查

## 2. 适用范围

1. 单机部署（systemd / 直接进程）
2. 小规模中转服务（SQLite 或 MySQL）
3. 需要优先保障 DeepSeek / MiniMax 稳定可用的场景

## 3. 新版本部署关键点（v3.9.9）

本版本上线时建议重点确认以下变化已经在运行环境生效：

1. MiniMax 请求角色兼容：不被上游接受的角色会在中转侧归一处理，避免 role 校验失败。
2. 默认 SQLite 路径：建议使用 data 目录下数据库文件，避免误连仓库根目录历史库。
3. 启动日志可见 SQLite 实际路径：用于快速确认实例连接的是预期数据库。
4. 用户内容抓取超时与流式空闲超时支持显式配置，建议在生产明确设置。

## 4. 推荐配置基线（.env）

建议基线：

1. 开发/验证环境默认 SQLite。
2. 生产优先 MySQL（若继续使用 SQLite，必须固定绝对路径并做备份）。
3. 显式设置关键超时，避免依赖隐式默认值。

示例（兼容优先）：

```env
PORT=3000
GIN_MODE=release
DEBUG=false

# 会话
SESSION_SECRET=replace-with-long-random-string

# 数据库（二选一，优先 SQL_DSN）
SQLITE_PATH=./data/uniapi.db
# SQL_DSN=root:password@tcp(127.0.0.1:3306)/uniapi?charset=utf8mb4&parseTime=True&loc=Local

# Redis（可选）
# REDIS_CONN_STRING=redis://127.0.0.1:6379/0

# 超时（建议显式）
RELAY_TIMEOUT=300
USER_CONTENT_REQUEST_TIMEOUT=30
IDLE_TIMEOUT=30

# 多模态路由（文本模型收到图片时）
MULTIMODAL_ROUTE_MODE=fixed_fallback
MULTIMODAL_VISION_FALLBACK_MODEL=gpt-4o
```

## 5. 通道与分组一致性建议

中转可用性依赖以下四处一致：

1. users.group
2. tokens.user_id 对应用户所属 group
3. channels.group
4. abilities.group + abilities.model

上线前建议检查：

1. 目标 token 所属用户 group 是否正确。
2. DeepSeek / MiniMax 对应 model 是否在该 group 下存在 enabled abilities。
3. 渠道状态为 enabled，且未被 suspend_until 长时间冻结。

## 6. 标准发布流程（建议）

### 6.1 构建

```bash
make build-release GIT_TAG=v3.9.9
```

### 6.2 发布前门禁

```bash
go vet ./...
go test -race ./...
```

### 6.3 启动后健康检查

```bash
curl -sS http://127.0.0.1:3000/api/status
curl -sS http://127.0.0.1:3000/v1/models -H "Authorization: Bearer <RELAY_KEY>"
```

### 6.4 兼容性冒烟（最小集合）

1. DeepSeek chat 非流式
2. MiniMax chat（developer role 输入）
3. MiniMax responses

通过标准：三项均返回 HTTP 200，且响应结构符合对应端点规范。

## 7. 运行期可观测性建议

1. 保留结构化日志，重点关注：
   1. using SQLite database path
   2. selected_channel_id / selected_by
   3. No available channels for Model ... under Group ...
2. 对 401、403、503 建立最小告警。
3. 灰度阶段建议开启更高日志级别并保留 24-48 小时。

## 8. 回滚与应急建议

1. 二进制回滚：保留上一版本可执行文件与启动参数。
2. 数据回滚：发布前备份 data/uniapi.db 或 MySQL 快照。
3. 配置回滚：保留上一版 .env 备份，变更采用增量提交。
4. 快速止血策略：
   1. 固定 SQLITE_PATH 到已验证数据库
   2. 统一 group 映射（users/tokens/channels/abilities）
   3. 重新执行最小兼容性冒烟

## 9. 发布验收清单

1. 构建成功：build-release 完成，版本注入正确。
2. 门禁通过：go vet、go test -race 全通过。
3. 健康通过：/api/status、/v1/models 可用。
4. 兼容通过：DeepSeek / MiniMax 三项冒烟通过。
5. 文档同步：README 版本说明、DOCS_INDEX 索引、测试记录更新。

## 10. 相关文档

1. docs/CONFIG_GUIDE.md
2. docs/COMPATIBILITY_TEST_PLAN_2026-05-29.md
3. docs/COMPATIBILITY_TEST_RESULTS_2026-05-29.md
4. docs/DOCS_INDEX.md
