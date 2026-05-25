# DeepSeek TUI 中文使用说明书

> 版本：v0.8.37 | 仓库：github.com/Hmbown/DeepSeek-TUI

---

## 目录

1. [概述](#1-概述)
2. [安装](#2-安装)
3. [快速开始](#3-快速开始)
4. [交互模式（TUI）](#4-交互模式tui)
5. [CLI 命令参考](#5-cli-命令参考)
6. [配置管理](#6-配置管理)
7. [会话管理](#7-会话管理)
8. [认证与多提供商](#8-认证与多提供商)
9. [MCP 服务器](#9-mcp-服务器)
10. [Skills 技能系统](#10-skills-技能系统)
11. [工具与依赖](#11-工具与依赖)
12. [AI 编程工作流](#12-ai-编程工作流)
13. [高级功能](#13-高级功能)
14. [常用技巧](#14-常用技巧)
15. [故障排除](#15-故障排除)

---

## 1. 概述

**DeepSeek TUI** 是一款运行在终端中的 AI 编程助手，类似于 Claude Code 或终端版 GitHub Copilot。它直接在你的终端里提供 AI 驱动的代码生成、审查、重构、调试等能力。

### 核心能力

- **交互式 TUI**：在终端内提供全功能 AI 对话界面
- **非交互式 CLI**：支持单次命令执行，适合脚本和 CI/CD 集成
- **Agent 模式**：AI 可自主调用工具（读写文件、执行命令、搜索代码）
- **多提供商支持**：DeepSeek、OpenAI、OpenRouter、Ollama 等
- **MCP 协议**：通过 Model Context Protocol 扩展工具能力
- **Skills 系统**：可安装社区技能包，扩展 AI 的专业能力
- **会话持久化**：保存、恢复、分支 AI 对话
- **代码审查**：基于 git diff 的 AI 代码审查
- **指标统计**：使用量统计与审计日志

---

## 2. 安装

### macOS

```bash
# Homebrew（推荐）
brew install deepseek-tui

# 或通过 npm
npm install -g deepseek-tui
```

### Linux

```bash
# 通过 npm
npm install -g deepseek-tui

# 或从 GitHub Releases 下载二进制文件
# https://github.com/Hmbown/DeepSeek-TUI/releases
```

### 验证安装

```bash
deepseek --version
deepseek doctor    # 运行诊断，检查各项依赖和配置
```

---

## 3. 快速开始

### 3.1 首次配置

```bash
# 设置 API Key
deepseek auth set --api-key "sk-your-api-key"

# 或通过 stdin 输入（不回显）
deepseek auth set --api-key-stdin

# 运行诊断确认一切正常
deepseek doctor
```

### 3.2 发起第一次对话

```bash
# 直接进入交互式 TUI
deepseek

# 带初始提示词启动
deepseek "帮我分析这个项目的结构"

# 非交互式单次执行
deepseek exec "列出当前目录下所有 Go 文件"
```

---

## 4. 交互模式（TUI）

### 4.1 启动与退出

```bash
deepseek                  # 启动交互式 TUI
deepseek "我的问题"        # 带初始提示词启动
```

在 TUI 中：
- 输入问题后按 **Enter** 发送
- 按 **Ctrl+C** 退出
- 按 **Ctrl+S** 暂存当前输入到草稿箱

### 4.2 Agent 模式

Agent 模式下 AI 可以**自主调用工具**，包括读写文件、执行命令、搜索代码等。

在 TUI 中默认即为 Agent 模式。AI 会：
1. 理解你的需求
2. 制定执行计划
3. 自动调用工具完成任务
4. 每次工具调用前请求你的批准（可配置为 YOLO 模式自动批准）

### 4.3 审批策略

| 模式 | 行为 |
|------|------|
| **默认（Suggest）** | 读操作静默执行，写操作需要用户批准 |
| **YOLO 模式** | 自动批准所有操作（`--yolo` 参数） |

```bash
deepseek --yolo                              # YOLO 模式
deepseek --approval-policy auto              # 自动批准
deepseek --approval-policy suggest           # 默认：写操作需批准
```

---

## 5. CLI 命令参考

### 5.1 核心命令总览

| 命令 | 功能 |
|------|------|
| `deepseek` | 启动交互式 TUI |
| `deepseek exec` | 非交互式单次执行 |
| `deepseek review` | AI 代码审查（基于 git diff） |
| `deepseek apply` | 应用补丁文件 |
| `deepseek -p "提示词"` | 带提示词快速启动 |

### 5.2 `exec` — 非交互式执行

适合集成到脚本或 CI/CD 流水线：

```bash
# 基本用法
deepseek exec "解释这个函数的逻辑"

# Agent 模式（允许工具调用）
deepseek exec --auto "修复 test/user_test.go 中的编译错误"

# JSON 输出
deepseek exec --json "列出项目依赖"

# 流式 JSON 输出
deepseek exec --output-format stream-json "分析代码"
```

### 5.3 `review` — 代码审查

对当前 git 工作区的变更进行 AI 审查：

```bash
deepseek review                    # 审查未暂存的改动
deepseek review HEAD~3..HEAD       # 审查最近 3 次提交
```

### 5.4 `apply` — 应用补丁

```bash
deepseek apply < patch_file.diff   # 从文件应用补丁
git diff | deepseek apply          # 从管道应用补丁
```

### 5.5 `models` — 模型列表

```bash
deepseek models                    # 列出可用模型
deepseek models --provider openai  # 按提供商筛选
```

### 5.6 `update` — 更新

```bash
deepseek update                    # 检查并应用更新
```

### 5.7 `metrics` — 用量统计

```bash
deepseek metrics                   # 查看用量汇总
deepseek metrics --since 7d        # 最近 7 天的用量
deepseek metrics --json --since 30d # 最近 30 天 JSON 输出
```

### 5.8 `completions` — Shell 补全

```bash
# 生成 Shell 补全脚本
deepseek completions bash  > ~/.deepseek-completion.bash
deepseek completions zsh   > ~/.deepseek-completion.zsh
deepseek completions fish  > ~/.deepseek-completion.fish

# 在 .zshrc 中添加
source ~/.deepseek-completion.zsh
```

---

## 6. 配置管理

### 6.1 配置文件位置

- **主配置文件**：`~/.deepseek/config.toml`
- **MCP 配置**：`~/.deepseek/mcp.json`
- **Skills 目录**：`~/.deepseek/skills/`
- **工具输出缓存**：`~/.deepseek/tool_outputs/`

### 6.2 配置命令

```bash
deepseek config list               # 列出所有配置项
deepseek config get <key>          # 获取指定配置项
deepseek config set <key> <value>  # 设置配置项
deepseek config unset <key>        # 删除配置项
deepseek config path               # 显示配置文件路径
```

### 6.3 常用配置项

```toml
# ~/.deepseek/config.toml 示例

[provider.deepseek]
api_key = "sk-xxxxxxxxxxxxxxxx"
base_url = "https://api.deepseek.com/beta"
model = "deepseek-v4-pro"

[search]
provider = "bing"          # 搜索引擎：bing / duckduckgo / tavily

[notifications]
method = "os"              # 通知方式：os / terminal_bell / off

[subagents]
max_concurrent = 10        # 最大并行子代理数
```

### 6.4 运行时参数

```bash
deepseek --config ~/custom-config.toml    # 指定配置文件
deepseek --provider openai                # 指定提供商
deepseek --model gpt-4o                   # 指定模型
deepseek --base-url https://api.example.com  # 自定义 API 地址
deepseek --log-level debug                # 日志级别：trace/debug/info/warn/error
deepseek --sandbox-mode readonly          # 沙箱模式
deepseek --mouse-capture                  # 启用鼠标支持
deepseek --no-mouse-capture               # 禁用鼠标支持
deepseek --skip-onboarding                # 跳过新手引导
deepseek --telemetry false                # 关闭遥测
```

---

## 7. 会话管理

### 7.1 会话列表

```bash
deepseek sessions                        # 列出所有已保存的会话
```

### 7.2 恢复会话

```bash
deepseek resume <session_id>             # 按 ID 恢复
deepseek resume <prefix>                 # 按 ID 前缀恢复（只需前几位）
deepseek exec --resume <session_id>      # 非交互式恢复
```

### 7.3 继续最近会话

```bash
deepseek exec --continue                 # 继续当前工作区的最近会话
```

### 7.4 分支会话

```bash
deepseek fork <session_id>               # 从某个会话分支创建新会话
```

---

## 8. 认证与多提供商

### 8.1 认证管理

```bash
# 查看当前认证状态
deepseek auth status

# 设置 API Key
deepseek auth set --api-key "sk-xxx"
deepseek auth set --api-key-stdin        # 通过管道输入（不回显）

# 查看 Key 状态（不显示值）
deepseek auth get deepseek

# 列出所有提供商的认证状态
deepseek auth list

# 删除 Key
deepseek auth clear deepseek
```

### 8.2 支持的提供商

| 提供商 | 说明 |
|--------|------|
| `deepseek` | DeepSeek 官方 API |
| `openai` | OpenAI API |
| `openrouter` | OpenRouter 聚合 API |
| `nvidia-nim` | NVIDIA NIM |
| `novita` | Novita AI |
| `fireworks` | Fireworks AI |
| `sglang` | SGLang |
| `vllm` | vLLM |
| `ollama` | 本地 Ollama 模型 |
| `atlascloud` | Atlas Cloud |

### 8.3 认证优先级

认证凭据的查找顺序（高到低）：

1. `~/.deepseek/config.toml` 配置文件
2. 操作系统密钥链（Keyring）
3. 环境变量

```bash
# 环境变量方式
export DEEPSEEK_API_KEY="sk-xxx"
```

### 8.4 切换提供商

```bash
# 临时切换
deepseek --provider openai --model gpt-4o

# CLI 模式下指定
deepseek models --provider openrouter
```

---

## 9. MCP 服务器

MCP（Model Context Protocol）允许 DeepSeek TUI 连接外部工具服务。

### 9.1 初始化 MCP

```bash
deepseek setup --mcp                     # 创建默认 MCP 配置
deepseek mcp init                        # 初始化 MCP 配置
```

### 9.2 管理 MCP 服务器

```bash
deepseek mcp <command>                   # MCP 服务器管理
```

配置文件位于 `~/.deepseek/mcp.json`，可以添加各种 MCP 服务（如文件系统、数据库、API 等）。

### 9.3 作为 MCP 服务器运行

DeepSeek TUI 本身也可以作为 MCP 服务器对外提供服务：

```bash
deepseek mcp-server                      # 以 MCP 服务器模式运行（stdio）
```

---

## 10. Skills 技能系统

Skills 是可安装的扩展包，让 AI 获得特定领域的专业能力。

### 10.1 技能目录

| 位置 | 说明 |
|------|------|
| `~/.deepseek/skills/` | 全局技能（所有项目可用） |
| `<project>/skills/` | 项目级技能（当前项目专用） |
| `<project>/.agents/skills/` | 项目 .agents 目录下的技能 |

### 10.2 安装技能

```bash
# 从 GitHub 安装社区技能
deepseek skill-installer install <github-url>

# 初始化技能目录
deepseek setup --skills
```

### 10.3 内置技能示例

当前版本内置的技能包括：

- **delegate**：策略性任务委派
- **documents**：Word 文档处理
- **pdf**：PDF 读写与操作
- **spreadsheets**：Excel/CSV 表格处理
- **presentations**：PPT 演示文稿
- **feishu**：飞书/Lark 集成
- **mcp-builder**：MCP 服务器构建
- **plugin-creator**：插件脚手架
- **skill-creator**：技能创建器
- **skill-installer**：技能安装器
- **v4-best-practices**：V4 模型最佳实践

---

## 11. 工具与依赖

### 11.1 内置工具

DeepSeek TUI 为 AI 提供以下内置工具：

| 工具 | 功能 | 依赖 |
|------|------|------|
| `read_file` | 读取文件内容 | 无 |
| `write_file` | 写入文件 | 无 |
| `edit_file` | 编辑文件 | 无 |
| `apply_patch` | 应用补丁 | 无 |
| `exec_shell` | 执行 Shell 命令 | 无 |
| `grep_files` | 搜索文件内容 | 无 |
| `file_search` | 搜索文件名 | 无 |
| `web_search` | 网页搜索 | 无 |
| `code_execution` | 执行 Python 代码 | Python3 |
| `js_execution` | 执行 JavaScript 代码 | Node.js |
| `git_status/diff/log` | Git 操作 | Git |

### 11.2 可选依赖

| 工具 | 安装方式 |
|------|----------|
| Pandoc（文档转换） | `brew install pandoc` |
| Tesseract（图片 OCR） | `brew install tesseract` |
| Poppler（PDF 提取） | `brew install poppler`（可选，v0.8.32+ 已内置纯 Rust 提取器） |

### 11.3 检查依赖状态

```bash
deepseek doctor                         # 完整诊断，会列出所有工具状态
```

---

## 12. AI 编程工作流

### 12.1 基本工作流

```bash
# 1. 启动交互模式
deepseek

# 2. 描述你的需求
> 在 controller/user.go 中添加一个修改用户邮箱的接口

# 3. AI 会制定计划并逐步执行，每次操作前请求你的批准
# 4. 审查 AI 的修改，确认或拒绝
```

### 12.2 代码审查工作流

```bash
# 开发完成后，审查变更
deepseek review

# 审查指定范围的提交
deepseek review main..feature-branch
```

### 12.3 CI/CD 集成

```bash
# 在 CI 中自动修复 lint 错误
deepseek exec --auto --yolo "修复所有 golangci-lint 报错"

# 生成变更日志
deepseek exec --json "总结最近 10 次提交的变更内容"
```

### 12.4 项目初始化

```bash
# 为当前项目生成 AGENTS.md（AI 项目指令文件）
deepseek init
```

这会创建包含项目类型、构建命令、测试命令等信息的 `AGENTS.md`，帮助 AI 更好地理解你的项目。

---

## 13. 高级功能

### 13.1 插件系统

```bash
deepseek setup --plugins                # 初始化插件目录
```

插件位于 `~/.deepseek/plugins/`，可以扩展 TUI 的底层能力。

### 13.2 本地服务器模式

```bash
deepseek serve                          # 启动本地 TUI 服务器
deepseek app-server                     # 启动 App 服务器传输层
```

### 13.3 评估框架

```bash
deepseek eval                           # 运行离线评估框架
```

### 13.4 Eval 模式

用于对 AI 输出进行系统性评估和测试。

### 13.5 功能标志

```bash
deepseek features                       # 查看当前启用的功能标志
```

### 13.6 沙箱与审批

```bash
deepseek sandbox                        # 评估沙箱/审批策略决策
```

---

## 14. 常用技巧

### 14.1 提高效率

- **并行工具调用**：AI 会自动并行执行独立的工具调用，大幅提升速度
- **子代理**：复杂任务会自动拆分为多个子代理并行处理
- **上下文管理**：长对话中使用 `/compact` 压缩上下文，释放空间
- **草稿暂存**：在 TUI 输入框中按 `Ctrl+S` 暂存当前输入

### 14.2 提示词技巧

```
# 好的提示词
"在 controller/user.go 中添加 UpdateEmail 接口，要求：
 - POST /api/user/email
 - 需要验证新邮箱格式
 - 发送验证邮件
 - 添加单元测试"

# 避免模糊
"修改一下用户功能"  ❌
```

### 14.3 工作区管理

```bash
# 在不同项目中运行 DeepSeek TUI
cd ~/project-a && deepseek
cd ~/project-b && deepseek

# 每个项目有独立的会话历史
```

### 14.4 代理与网络

DeepSeek TUI 会使用系统代理设置。如需自定义：

```bash
export HTTPS_PROXY=http://proxy:port
deepseek
```

---

## 15. 故障排除

### 15.1 运行诊断

```bash
deepseek doctor
```

这会检查：
- 版本信息
- 配置文件状态
- API Key 配置
- API 连接性
- MCP 服务器状态
- 技能系统状态
- 工具依赖（Python、Node.js、Pandoc 等）
- 平台信息

### 15.2 常见问题

**Q: API 连接失败**
```bash
# 检查 API Key 是否配置
deepseek auth status

# 检查连接
deepseek doctor
```

**Q: 命令找不到**
```bash
# 确认安装
which deepseek
deepseek --version

# 重新安装
npm install -g deepseek-tui
```

**Q: 配置文件问题**
```bash
# 查看配置文件位置
deepseek config path

# 查看所有配置
deepseek config list
```

**Q: 沙箱不可用**
```bash
# macOS 上沙箱功能可能受限，这是正常的
# 命令仍会以 best-effort 方式执行
```

**Q: 日志调试**
```bash
deepseek --log-level debug              # 启用详细日志
```

### 15.3 更新到最新版

```bash
deepseek update

# 或通过 npm
npm update -g deepseek-tui
```

---

## 附录：全局参数速查

```
deepseek [OPTIONS] [PROMPT]

核心选项：
  --config <PATH>           指定配置文件
  --profile <NAME>          指定配置 Profile
  --provider <NAME>         选择 AI 提供商
  --model <NAME>            选择模型
  --api-key <KEY>           设置 API Key
  --base-url <URL>          自定义 API 地址
  --yolo                    自动批准所有工具调用
  --approval-policy <POL>   审批策略
  --sandbox-mode <MODE>     沙箱模式
  --log-level <LEVEL>       日志级别
  --mouse-capture           启用鼠标捕获
  --no-mouse-capture        禁用鼠标捕获
  --skip-onboarding         跳过新手引导
  --telemetry <BOOL>        遥测开关
  -p, --prompt <TEXT>       设置提示词
  -h, --help                显示帮助
  -V, --version             显示版本

常用子命令：
  exec        非交互式执行
  review      代码审查
  apply       应用补丁
  sessions    会话列表
  resume      恢复会话
  fork        分支会话
  doctor      系统诊断
  models      模型列表
  config      配置管理
  auth        认证管理
  mcp         MCP 服务管理
  setup       初始化配置
  init        生成 AGENTS.md
  update      更新程序
  metrics     用量统计
  completions 生成 Shell 补全
```

---

> 本文档基于 DeepSeek TUI v0.8.37 编写。更多信息请访问 [GitHub 仓库](https://github.com/Hmbown/DeepSeek-TUI)。
