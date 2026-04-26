# one-api 版本管理指南

> 适用项目: [decardlabs/one-api](https://github.com/decardlabs/one-api)
> 最后更新: 2026-04-26

---

## 当前版本状态

| 版本 | Tag | Commit | 说明 | 日期 |
|------|-----|--------|------|------|
| **v2.0** | `v2.0` | `c83021f` | 充值审批 + USD展示 + i18n + UI增强 | 2026-04-26 |
| v1.0 (base) | — | `17b7596` | 原始音频token计费增强 | — |

---

## 版本号规则 (Semver)

```
主版本.次版本.修订号 (MAJOR.MINOR.PATCH)
```

| 类型 | 触发条件 | 示例 |
|------|---------|------|
| **MAJOR** | 不兼容的重大重构、数据库结构变更 | `2.0` → `3.0` |
| **MINOR** | 新功能模块，向下兼容 | `2.0` → `2.1` |
| **PATCH** | Bug修复、小改动，向下兼容 | `2.0` → `2.0.1` |

**Tag 命名格式**: `vX.Y.Z`（必须带 `v` 前缀）

---

## 开发流程

### 正常开发流程

```
main (稳定版)
 │
 ├── git checkout -b feature/xxx    ← 从最新tag或main开分支
 │   ├── 开发...
 │   ├── git commit -am "feat: ..."
 │   └── git checkout main && git merge feature/xxx
 │
 ├── 测试验证
 │
 ├── git tag -a v2.1 -m "版本说明"   ← 打tag定版
 │
 └── git push origin main --tags     ← 推送（含tag）
```

### 具体操作步骤

#### ① 开始新版本开发

```bash
# 1. 确认当前在 main 分支，代码干净
git status                    # 应该是 clean
git checkout main

# 2. 拉取远程最新
git pull origin main

# 3. 基于当前版本创建功能分支（建议每个功能一个分支）
git checkout -b feature/v3-sidebar-layout      # v3.0 的侧边栏功能
git checkout -b feature/v3-type-safe           # v3.0 的类型安全治理

# 或者如果只有一个大任务，直接用：
git checkout -b release/v3.0                   # v3.0 整体开发分支
```

#### ② 开发过程中的日常操作

```bash
# 修改文件后提交
git add -A
git commit -m "feat: 新增左侧边栏组件"
git commit -m "fix: 修复表格滚动问题"
git commit -m "style: 统一按钮圆角样式"

# Commit message 格式:
# feat:     新功能
# fix:      bug修复
# style:    格式调整（不影响功能）
# refactor: 重构（不新增功能、不修bug）
# docs:     文档更新
# test:     测试相关
# chore:    构建/工具链变更
# release:  版本发布
```

#### ③ 查看进度和差异

```bash
# 查看从上一个版本以来的所有提交
git log v2.0..HEAD --oneline

# 查看与上一个版本的文件差异
git diff v2.0 --stat

# 查看具体改了什么
git diff v2.0 -- web/modern/src/

# 当前分支状态
git branch -vv
```

---

## ⚠️ 版本回退操作（重点）

### 场景 1：还没打 tag，想放弃最近的修改

```bash
# 查看最近3次提交
git log --oneline -3

# 放弃最近1次提交（保留文件修改）
git reset --soft HEAD~1

# 放弃最近1次提交（同时丢弃文件修改）⚠️ 危险
git reset --hard HEAD~1

# 回退到某个指定commit（保留文件修改）
git reset --soft c83021f    # 回到v2.0
```

### 场景 2：已打 tag，发现重大 Bug 需要回退发布版

```bash
# 1. 先删除远程和本地的错误tag
git push origin --delete v2.0    # 删除远程tag
git tag -d v2.0                  # 删除本地tag

# 2. 回退代码到目标版本
git reset --hard c83021f         # 硬回到v2.0的commit

# 3. 强制推送 ⚠️ 需要谨慎
git push --force origin main

# 4. 重新打tag（修复bug后）
git add -A && git commit -m "fix: 修复紧急bug"
git tag -a v2.0 -m "v2.0 (hotfix)"
git push origin main --tags
```

### 场景 3：想查看/恢复历史版本代码

```bash
# 只是想看看某个版本的代码（不影响当前工作）
git checkout v2.0        # 进入"分离HEAD"状态浏览
git checkout main        # 看完后切回来

# 基于旧版本创建热修复分支
git checkout -b hotfix/v2.0.1 v2.0
# ... 修复bug ...
git commit -am "fix: 修复审批拒绝后状态未更新"
git tag -a v2.0.1 -m "v2.0.1 hotfix"

# 合并回main
git checkout main
git merge hotfix/v2.0.1
git push origin main --tags
```

### 场景 4：误操作恢复（安全网）

```bash
# 如果 reset --hard 后后悔了，找回丢失的commit
git reflog                    # 查看所有操作记录
# 显示类似:
# c83021f HEAD@{0}: reset: moving to v2.0
# abc1234 HEAD@{1}: commit: 我要找的这个提交

# 恢复到之前的commit
git reset --hard abc1234
```

**核心原则**: `git reflog` 是你的救命稻草，只要 commit 过就不会丢。

---

## 发布定版流程（以 v3.0 为例）

### 发布前 Checklist

- [ ] 前端构建通过: `cd web/modern && npm run build`
- [ ] Go 编译通过: `go build -o one-api`
- [ ] 功能测试完成（手动或自动化）
- [ ] 更新 CHANGELOG（如有）
- [ ] Commit message 格式规范

### 发布命令序列

```bash
#!/bin/bash
# 一键发布脚本 —— 复制到终端执行

set -e  # 任何一步失败就停止

VERSION="v3.0"

echo "=== 1. 确保在 main 分支 ==="
git checkout main
git pull origin main

echo "=== 2. 确保所有修改已提交 ==="
git status
# 如果有未提交的修改，先 commit

echo "=== 3. 运行构建验证 ==="
cd /Users/macairm5/WorkBuddy/20260423193254/one-api/web/modern && npm run build
cd /Users/macairm5/WorkBuddy/20260423193254/one-api && go build -o /dev/null .

echo "=== 4. 创建 Tag ==="
git tag -a "$VERSION" -m "$VERSION - 版本说明"

echo "=== 5. 推送到远程 ==="
git push origin main
git push origin "$VERSION"

echo "✅ $VERSION 发布完成!"
echo "   Commit: $(git rev-parse HEAD)"
echo "   Tag: $(git describe --tags --always)"
```

---

## Git 工作流速查表

| 操作 | 命令 |
|------|------|
| 查看当前状态 | `git status` |
| 查看提交历史 | `git log --oneline -10` |
| 查看所有标签 | `git tag -l --sort=-version:refname` |
| 查看某版本详情 | `git show v2.0` |
| 比较两个版本 | `git diff v2.0..v3.0 --stat` |
| 创建分支 | `git checkout -b feature/xxx` |
| 切换分支 | `git checkout main` |
| 合并分支 | `git merge feature/xxx` |
| 删除分支 | `git branch -d feature/xxx` |
| 暂存修改 | `git stash` |
| 恢复暂存 | `git stash pop` |
| 撤销未提交的修改 | `git checkout -- <file>` |
| 撤销最近一次提交(保留修改) | `git reset --soft HEAD~1` |
| 找回丢失的提交 | `git reflog` |

---

## 下一步：开始 v3.0 开发

根据 UPGRADE-PLAN.md 的 Top 3 建议，v3.0 可以包含：

1. **P0**: 左侧边栏布局改造
2. **P0**: TypeScript 类型安全深度治理
3. **P1**: Auth 中间件公共逻辑抽取

启动命令:

```bash
cd /Users/macairm5/WorkBuddy/20260423193254/one-api
git checkout -b release/v3.0
# 开始开发...
```

完成后执行上述「发布定版流程」即可定版 v3.0。
