# DeepSeek/MiniMax 兼容性测试记录（2026-05-29）

## 0. 执行信息

- 执行时间：2026-05-29 11:08-11:39 (UTC+8)
- 执行环境：localhost UniAPI
- UniAPI BaseURL：http://localhost:3000
- 原始输出目录（第一轮）：`/tmp/compat_20260529`
- 原始输出目录（第二轮）：`/tmp/compat_20260529_round2`
- 原始输出目录（第三轮，MiniMax 直连补测）：`/tmp/compat_20260529_round3_mm`
- 原始输出目录（第四轮，重启后最终复测）：`/tmp/compat_20260529_post_env_final`

## 1. 测试输入

- DeepSeek 直连：已提供厂商 Key
- MiniMax 直连：已提供厂商 Key（已执行补测）
- UniAPI 中转：已提供 BaseURL 与 Token

## 2. 先决问题排查（数据库“看起来为空”）

### 2.1 现象

1. 中转请求统一返回 `token not found`。
2. 手工查看 `data/uniapi.db` 时，数据并不为空。

### 2.2 根因

运行中的 UniAPI 实例实际连接的是仓库根目录数据库：

- `/Users/sunm15/Documents/uniapi/uniapi.db`

而不是：

- `/Users/sunm15/Documents/uniapi/data/uniapi.db`

该根目录数据库中 `tokens/channels` 计数为 0，导致鉴权失败。

### 2.3 与备份库对比

备份库：`/Users/sunm15/Documents/uniapi/data/db-backups/20260529-084113/root-uniapi.db`

关键计数：

1. users=1
2. tokens=5
3. channels=4
4. options=3

且提供的中转 token 在备份库存在（精确匹配）。

### 2.4 处理动作

已将运行实例改为显式使用备份库启动：

- `SQLITE_PATH=/Users/sunm15/Documents/uniapi/data/db-backups/20260529-084113/root-uniapi.db ./uniapi`

验证结果：

1. `/v1/models` 使用中转 token 返回 HTTP 200。
2. 模型列表可见 `deepseek-v4-pro` 与 `MiniMax-M2.7`。

## 3. 用例执行结果（第二轮有效结果）

### 3.1 DeepSeek

| 用例               | 路径 | 结果 | 关键信息                                                       |
| ------------------ | ---- | ---- | -------------------------------------------------------------- |
| 非流式基础问答     | 直连 | 通过 | HTTP 200，`content=OK-DS-BASIC`，见 `ds_direct_chat_basic.out` |
| 非流式多轮上下文   | 直连 | 通过 | HTTP 200，见 `ds_direct_chat_multiturn.out`                    |
| 非流式基础问答     | 中转 | 通过 | HTTP 200，`content=OK-DS-BASIC`，见 `ds_relay_chat_basic.out`  |
| 非流式多轮上下文   | 中转 | 通过 | HTTP 200，见 `ds_relay_chat_multiturn.out`                     |
| Responses/fallback | 中转 | 通过 | HTTP 200，`status=completed`，见 `ds_relay_responses.out`      |
| 流式输出           | 中转 | 通过 | HTTP 200，收到连续 SSE `data:` chunk，见 `ds_relay_stream.out` |

第一轮中转失败记录（保留）：HTTP 401 `token not found`，见 `/tmp/compat_20260529`。

### 3.2 MiniMax

| 用例                             | 路径      | 结果 | 关键信息                                                                                         |
| -------------------------------- | --------- | ---- | ------------------------------------------------------------------------------------------------ |
| 非流式基础问答                   | 直连      | 通过 | HTTP 200，见 `mm_direct_chat_basic.out`                                                          |
| 非法 role 输入（developer+user） | 直连      | 失败 | HTTP 400，`invalid role: developer`，见 `mm_direct_chat_role_edge.out`                           |
| Responses 接口                   | 直连      | 失败 | HTTP 404，`/v1/responses` 不可用，见 `mm_direct_responses.out`                                   |
| 流式输出                         | 直连      | 通过 | HTTP 200，SSE chunk 正常，见 `mm_direct_stream.out`                                              |
| 非流式基础问答                   | 中转      | 通过 | HTTP 200，返回可用回答，见 `mm_relay_chat_basic.out`                                             |
| 非法 role 输入（developer+user） | 中转      | 通过 | HTTP 200，未出现 role 校验报错，见 `mm_relay_role_edge.out`                                      |
| Responses/fallback               | 中转      | 通过 | HTTP 200，返回 `status=incomplete` 但结构有效，见 `mm_relay_responses.out`                       |
| 流式输出                         | 中转      | 通过 | HTTP 200，收到 SSE chunk，见 `mm_relay_stream.out`                                               |
| 同提示词对比（basic）            | 直连/中转 | 通过 | 直连与中转均 HTTP 200（语义可用），中转响应结构与直连略有差异，见 `mm_relay_chat_basic_sync.out` |

## 4. 对比结论

1. DeepSeek：已完成直连 vs 中转对比，核心路径均为 HTTP 200，可认为兼容通过。
2. MiniMax：中转路径核心能力通过，包括 `developer` role 输入场景未触发历史报错。
3. MiniMax 直连补测显示：厂商直连不接受 `developer` role（HTTP 400），而中转可兼容处理（HTTP 200），与预期修复方向一致。
4. MiniMax 厂商直连 `responses` 端点为 HTTP 404；中转 `responses/fallback` 可用（HTTP 200）。
5. 本次主要阻塞根因不是适配器逻辑，而是运行实例误连到了空数据库文件。

## 5. 原始响应摘录

1. `ds_direct_chat_basic.out`: HTTP 200，`content=OK-DS-BASIC`。
2. `ds_relay_chat_basic.out`: HTTP 200，`content=OK-DS-BASIC`。
3. `ds_relay_responses.out`: HTTP 200，`status=completed`。
4. `mm_relay_role_edge.out`: HTTP 200，`developer+user` 输入可处理。
5. `mm_relay_responses.out`: HTTP 200，`status=incomplete`（建议后续继续观察）。
6. `mm_relay_stream.out`: HTTP 200，SSE chunk 正常。
7. `mm_direct_chat_role_edge.out`: HTTP 400，`invalid role: developer`。
8. `mm_direct_responses.out`: HTTP 404（直连端点不可用）。

## 6. 下一步

1. 建议固定启动命令中的 `SQLITE_PATH`，避免再次连到仓库根目录默认库。
2. 建议后续补一个启动前自检：打印当前生效 `SQLitePath` 到启动日志。
3. 针对 MiniMax 增加自动化回归：断言 `developer/tool/function/system` 输入在中转侧不会触发上游 role 400。

## 7. 第四轮最终复测（按建议执行后）

### 7.1 执行背景

1. 已完成代码修复并重新编译、重启服务。
2. 启动日志确认生效 SQLite 路径为：`/Users/sunm15/Documents/uniapi/data/uniapi.db`。

### 7.2 中间波动与根因

1. 重启后一度出现 HTTP 503：`No available channels for Model ... under Group ...`。
2. 根因是 `tokens/users` 所属组与 `channels/abilities` 分组映射短暂不一致（token 实际归属 `default`，能力映射曾被改到 `rootgroup`）。
3. 统一恢复为 `default` 后，路由恢复正常。

### 7.3 最终复测结果（第四轮）

| 用例                                 | 路径 | 结果 | 关键信息                        |
| ------------------------------------ | ---- | ---- | ------------------------------- |
| DeepSeek chat (`deepseek-v4-pro`)    | 中转 | 通过 | HTTP 200，见 `ds_chat.out`      |
| MiniMax role-edge (`developer+user`) | 中转 | 通过 | HTTP 200，见 `mm_role_edge.out` |
| MiniMax responses (`MiniMax-M2.7`)   | 中转 | 通过 | HTTP 200，见 `mm_resp.out`      |

附：第四轮执行目录为 `/tmp/compat_20260529_post_env_final`。

## 8. 最终结论

1. MiniMax 兼容性修复（`developer` 等非上游允许 role 的中转兼容）在最终复测中维持有效。
2. DeepSeek 与 MiniMax 在当前实例上的关键中转路径（chat/responses/role-edge）均已恢复并验证为 HTTP 200。
3. 本次故障闭环确认：主要风险来自运行时数据库路径与分组映射一致性，而非适配器主逻辑回归。
