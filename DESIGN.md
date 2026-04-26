# Design System — One API

## Product Context
- **What this is:** AI API 网关与管理平台，统一管理多个 LLM 渠道、令牌、用量统计与计费系统
- **Who it's for:** 开发者（API 调用者）、平台管理员（渠道/用户/运营）、团队负责人（数据看板）
- **Space/industry:** Developer Tools / API Gateway / AI Infrastructure
- **Project type:** Web Application (Dashboard + Admin + Consumer)

## Current State Assessment (现状评估)

### 技术栈（Modern 主题）
- **UI 框架:** shadcn/ui + Radix UI (new-york style)
- **CSS:** Tailwind CSS v3.4 + CSS Variables (HSL)
- **图标:** lucide-react
- **图表:** recharts
- **布局:** Top Navigation Bar (顶栏导航) + Grid 三行布局
- **主题:** Slate & Teal 双色调, 支持 light/dark/system

### 已有优势 ✅
1. shadcn/ui 组件体系成熟，可定制性强
2. CSS Variables 主题系统完善，支持亮/暗模式
3. 移动端优先设计（dvh/safe-area/touch-target）
4. 智能导航折叠（ResizeObserver 自动收进 More）
5. 代码分割完善（全部页面 lazy load）
6. 国际化支持（i18next）
7. 响应式断点体系完整（xs ~ 2xl）

### 需要改进的问题 ⚠️

| 问题 | 严重度 | 说明 |
|------|--------|------|
| **视觉层级扁平** | 🔴 高 | 全局使用石板灰(Slate)为主色，缺乏品牌辨识度，和无数 shadcn 默认项目撞脸 |
| **色彩性格模糊** | 🔴 高 | Teal 作为强调色在深色模式变主色，但整体偏保守工具感，没有"AI 平台"的科技感 |
| **Dashboard 信息密度低** | 🟡 中 | OverviewCards 可能缺乏关键指标的视觉权重区分 |
| **表格体验** | 🟡 中 | 数据表是核心场景（渠道/日志/令牌），需要更好的排序、筛选、批量操作 UX |
| **空状态/加载状态** | 🟡 中 | Skeleton 已有但可能不够精细 |
| **导航项过多** | 🟡 中 | 16 个导航项在顶栏中拥挤，"More" 菜单会频繁触发 |
| **Chat Playground 差异化** | 🟢 低 | 对话界面应该有独立的沉浸式体验，不应受全局约束 |

## Aesthetic Direction（美学方向）

### 推荐方向: "Industrial Precision — 工业精密风"

**一句话:** 深色优先的数据控制台美学，用精确的网格、克制的色彩和高对比度的数据可视化传达"可靠的基础设施"感。

**为什么选这个方向:**
1. One-API 是开发者工具，用户期待的是**效率和可信度**，不是花哨的营销页
2. 数据表格、日志查看、渠道监控是最高频操作 -> 需要高可读性
3. 与竞品（OpenAI Platform / Anthropic Console / Vercel Dashboard）拉开差异化
4. 当前 Slate & Teal 太"安全"，需要注入个性但不失专业

### Mood（情绪词）
> 冷静、精确、高效、可信赖、不喧哗但不容忽视

### Decoration Level（装饰级别）: **Intentional（克制装饰）**
微妙的背景纹理（噪点/网格）、精心设计的空状态插画、数据可视化的细节打磨。不做大面积装饰性元素。

## Color System（色彩系统）

### Approach: **Restrained with Semantic Accent（克制+语义强调）**
一个主色承载品牌，语义颜色承载状态，中性色承载层次。

### Proposed Palette（建议调色板）

#### Light Mode
```
--background:     210 20% 98%   /* 微蓝灰白底 - 比 pure white 更柔和 */
--foreground:     215 20% 10%   /* 近黑文字 */
--primary:        217 91% 60%   /* Royal Blue #3B82F6 - 科技感主色 */
--primary-fg:     210 20% 98%   /* 白色上文字 */
--accent:         162 85% 40%   /* Emerald 绿色 - 成功/增长语义 */
--secondary:      214 15% 94%   /* 浅灰辅助区 */
--muted:          213 12% 92%   /* 更浅的 muted */
--destructive:    0 84% 60%     /* 鲜红 - 危险 */
--border:         214 15% 90%   /* 柔和边框 */
```

#### Dark Mode（推荐为默认/首选模式）
```
--background:     222 47% 6%    /* 深邃蓝黑底 - 不是纯黑 */
--foreground:     210 10% 90%   /* 浅灰白文字 */
--primary:        217 91% 65%   /* Royal Blue 提亮 */
--primary-fg:     222 47% 6%    /* 深底色上文字 */
--accent:         160 80% 45%   /* Emerald 提亮 */
--secondary:      222 18% 14%   /* 深灰次级区 */
--muted:          220 13% 17%   /* 微妙的深 muted */
--destructive:    0 72% 55%     /* 暗红 */
--border:         220 14% 17%   /* 低对比边框 */
```

### Semantic Status Colors
| 状态 | Light | Dark | 用途 |
|------|-------|------|------|
| Success | `158 65% 42%` | `160 70% 45%` | 渠道可用、充值成功、API 正常 |
| Warning | `38 85% 50%` | `38 80% 48%` | 配额不足、响应慢、即将过期 |
| Error | `0 84% 60%` | `0 72% 55%` | 渠道宕机、API Key 失效、余额耗尽 |
| Info | `217 91% 60%` | `217 75% 58%` | 系统通知、提示信息 |

### Chart Colors (15色) - 高饱和度、易区分
保留现有 chart 色板基础，调整为更鲜明的版本用于深色背景。

### 关键决策
- **从 Slate&Teal → Royal Blue&Emerald**: 主色从灰蓝色变为鲜明蓝色，强调色从青绿变为翠绿
- **暗色模式为一级公民:** 目标用户是开发者，长时间盯着屏幕，暗色模式应作为推荐默认
- **保持 HSL 变量体系:** 不改变架构，只改数值

## Typography（字体排版）

### Font Stack
| 角色 | 字体 | 理由 |
|------|------|------|
| **Display/Hero** | **Geist Sans** (700/800) | 现代、几何感强、数字渲染优秀（仪表盘数字很重要）|
| **Body** | **Geist Sans** (400/500) | 同一字体家族保证一致性，可读性强 |
| **Data/Tables** | **Geist Mono** (400/500) | 表格中的代码/ID/时间戳等需要等宽字体 |
| **Code** | **JetBrains Mono** | Playground 代码块、JSON 编辑器 |
| **UI Labels** | Geist Sans (500 medium) | 按钮、标签、导航 |

### Why Geist?
- Vercel 出品的开源字体，天然适合开发者工具品类
- 数字字形（tabular-nums）完美对齐，对仪表盘/表格至关重要
- 支持中文（通过 Noto Sans CJK fallback）
- 不是 Inter/Roboto 等被滥用的默认字体

### Scale (Modular Type Scale, 1.25 ratio)
```
2xs:  0.625rem (10px)  — badge, tag
 xs:   0.75rem  (12px)  — table cell, caption
 sm:   0.875rem (14px)  — body small, form label
 base: 1rem     (16px)  — body default
 lg:   1.125rem (18px)  — card title, section heading
 xl:   1.25rem  (20px)  — page title
 2xl:  1.5rem   (24px)  — dashboard stat value
 3xl:  1.875rem (30px)  — hero number, big metric
 4xl:  2.25rem  (36px)  — landing hero (rare)
```

### Loading Strategy
```html
<link rel="preconnect" href="https://cdn.jsdelivr.net" />
<link href="https://cdn.jsdelivr.net/npm/geist@1.3.1/dist/fonts/geist-sans/style.css" rel="stylesheet" />
<link href="https://cdn.jsdelivr.net/npm/geist@1.3.1/dist/fonts/geist-mono/style.css" rel="stylesheet" />
<!-- Fallback: system-ui, -apple-system, sans-serif -->
<!-- Chinese: "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif -->
```

## Spacing（间距）

### Base Unit: **4px**
### Density: **Comfortable（舒适密度）**
> 不是最紧凑的（那是 CLI 工具），也不是最宽松的（那是 SaaS 营销页）。数据密集型应用的最佳平衡点。

### Spacing Scale
```
0:    0
1:    4px    — icon-text gap, inner padding tight
2:    8px    — standard gap between related elements
3:    12px   — form field spacing
4:    16px   — section internal padding, card padding
5:    20px   — between groups within a section
6:    24px   — section separation, card gap grid
8:    32px   — major section gaps (desktop)
10:   40px   — page-level sections
12:   48px   — major page areas
16:   64px   — hero / landing sections
```

## Layout（布局）

### Approach: **Grid-Disciplined（栅格规范）**
严格对齐的栅格系统，让数据表格和信息卡片自然有序。

### Grid System
- **Max content width:** 1440px (container max at 2xl breakpoint)
- **Columns:** 12-column grid for complex pages (Dashboard, Settings)
- **Gutter:** 24px (desktop), 16px (tablet), 12px (mobile)
- **Card grid:** 4-col (lg) / 2-col (md) / 1-col (sm) responsive

### Border Radius Scale
```
sm:   4px  — input, small button, tag
md:   6px  — card, dialog, dropdown
lg:   8px  — large card, modal
full: 9999px  — pill button, avatar
```
> 从当前的 0.5rem(8px) lg 降低到更精致的层级。小圆角 = 专业感。

## Motion（动效）

### Approach: **Minimal-Functional（最小功能动效）**
只做帮助理解状态的过渡动画，不做表演性动效。

### Easing
```css
/* Standard easing tokens */
--ease-out: cubic-bezier(0.16, 1, 0.3, 1);    /* exit/expand */
--ease-in: cubic-bezier(0.7, 0, 0.84, 0);       /* enter/collapse */
--ease-in-out: cubic-bezier(0.76, 0, 0.24, 1);  /* move */ 
```

### Duration
| 类型 | 时长 | 示例 |
|------|------|------|
| Micro | 50-100ms | Hover state, button press |
| Short | 150-250ms | Toast appear, dropdown open, tab switch |
| Medium | 250-400ms | Modal enter, page transition |
| Long | 400-700ms | Initial data load animation (stagger) |

### Recommended Animations
- **Table row hover:** bg transition 100ms（已有）
- **Modal/Dialog:** scale(0.95) → scale(1) + fade, 150ms（已有 tailwindcss-animate）
- **Stat counter:** number roll-up on Dashboard load（新增）
- **Skeleton pulse:** subtle shimmer, not harsh pulse（调整现有）
- **Navigation active indicator:** slide from left, 200ms（新增）

## Page-Specific Guidelines

### 1. Dashboard（仪表盘）— 最高优先级
```
┌─────────────────────────────────────────────────────┐
│  [4 stat cards: 总请求 | 成功率 | Token消耗 | 余额] │  ← 大号数字 + sparkline
├──────────┬──────────┬──────────┬────────────────────┤
│ Usage     │ Top      │ Channel  │ Trend              │
│ Chart     │ Models   │ Health   │ (time series)       │
│ (area)    │ (bar)    │ (status)│                    │
└──────────┴──────────┴──────────┴────────────────────┘
```
- Stat cards: 数字用 3xl Geist Sans Bold, 单位用 sm muted
- 图表统一使用 recharts, 配色从 --chart-* 取
- Channel Health: 状态灯 + 响应时间 + 错误率

### 2. Channels（渠道列表）— 核心页面
- 表格列: 名称 | 类型 | 状态(灯) | 模型数 | 请求量 | 成功率 | 响应时间(P50/P99) | 操作
- 状态灯: inline dot, green/yellow/red
- 操作按钮: Edit/Test/Delete + 更多菜单
- 批量操作 toolbar: 启用/禁用/删除选中
- 搜索/筛选: 全文搜索 + 类型筛选 + 状态筛选

### 3. Logs（日志）— 数据密集
- 时间戳用 Geist Mono, 相对时间显示（"3分钟前"）
- 日志级别 color-coded: success=green, error=red, warning=amber
- 详情弹窗: JSON pretty-printed, 可折叠
- 分页 + 无限滚动选项

### 4. Chat Playground（对话）— 沉浸模式
- 独立布局: 隐藏顶部导航或极简化（仅显示模型选择器）
- 左右分栏: 对话区(宽) | 参数面板(窄, 可折叠)
- Markdown 渲染: 代码高亮、数学公式支持（已具备）
- 流式输出: 打字机效果（已具备）

### 5. Settings（设置）— 分组表单
- Tab 导航: 个人设置 / 系统运营 / 运营配置 / 其他
- 表单分组: 用 Card 包裹每组相关设置
- 保存按钮: Sticky bottom bar 或每 group 独立保存

## Component Refinement（组件优化）

### Button Variants (扩展)
```
default  — primary blue background
secondary — outlined, border only
ghost     — text only, no bg
destructive — red (danger actions)
success   — green (confirm/create positive actions)  ← 新增
outline-success — green outline                      ← 新增
```

### Table Enhancements
- **Sticky header:** 滚动时表头固定
- **Row selection:** checkbox column, batch actions
- **Column visibility toggle:** 自定义显示列
- **Export:** CSV/JSON export button
- **Density control:** comfortable/compact/comfortable 切换

### Card Types
```
stat-card     — 大数字 + 标签 + 趋势线 (dashboard)
data-card     — 内容卡片, 有标题栏 + 操作
form-card     — 表单容器, 内部有 field groups
status-card   — 带状态指示器的卡片 (channel health)
```

### Empty States
每个主要列表页面都需要精心设计的空状态：
- Channels: "添加你的第一个渠道" + CTA 按钮 + 插图
- Tokens: "创建 API Key 开始调用" + 快捷链接
- Logs: "暂无日志记录" + 说明文字
- Dashboard: "等待数据..." + 首次设置引导

## Safe Choices（安全的选择 — 符合类别惯例）
1. ✅ 顶栏导航布局（不改侧边栏，改动太大）
2. ✅ shadcn/ui + Tailwind 技术栈不动
3. ✅ CSS Variables HSL 主题体系保持
4. ✅ lucide-react 图标库继续使用
5. ✅ light/dark/system 三种模式不变

## Risks（ deliberate departures from convention — 风险点）
1. **🎨 主色从 Slate Blue → Royal Blue**: 更鲜明更有辨识度，风险是与当前用户的习惯反差。收益是品牌记忆度大幅提升。
2. **🌙 暗色模式为推荐默认**: 开发者工具的常规做法，但部分用户可能更喜欢亮色。解决方案是尊重 system preference 但在 onboarding 引导尝试暗色。
3. **📐 圆角从 8px 降级到 4~8px 层级**: 更精致但可能与 shadcn 默认风格略有偏差。通过覆盖 radius 变量即可。
4. **🔤 引入 Geist 字体**: 额外的字体加载(~30KB)，但对数字渲染和整体气质提升显著。CDN 加载 + system fallback 保证性能。

## Implementation Priority（实施优先级）

### Phase 1: Design Foundation (设计基石) — 1-2天
- [ ] 更新 index.css CSS Variables（新配色）
- [ ] 引入 Geist 字体（CDN + fallback chain）
- [ ] 更新 tailwind.config.js（border-radius scale, font-family）
- [ ] 创建 base components override（Button variants, 新 Card types）

### Phase 2: Core Pages Optimization (核心页面) — 3-5天
- [ ] Dashboard 重构（stat cards + layout grid）
- [ ] Channels 页面优化（表格增强 + 批量操作）
- [ ] Logs 页面优化（时间戳格式 + 详情弹窗）

### Phase 3: Polish (打磨) — 2-3天
- [ ] Empty states 设计实现
- [ ] Settings 页面分组重构
- [ ] Chat Playground 沉浸模式
- [ ] Mobile 端精细化调整
- [ ] 动效微调（stagger reveal 等）

### Phase 4: Validation (验证) — 1天
- [ ] design-review 视觉 QA
- [ ] 跨浏览器测试
- [ ] 响应式全断点验证
- [ ] 无障碍(a11y) 基本检查

## Decisions Log
| Date | Decision | Rationale |
|------|----------|-----------|
| 2025-04-25 | Initial design system created | Based on full codebase audit of modern theme (shadcn/ui + Tailwind + Radix UI) |
| 2025-04-25 | Primary: Royal Blue (#3B82F6) | Replaces Slate gray for brand identity; matches developer tool aesthetic |
| 2025-04-25 | Font: Geist Sans/Mono | Optimal for data display; tabular-nums alignment; Vercel pedigree |
| 2025-04-25 | Dark mode as first-class citizen | Target users are developers who spend long hours in the app |
| 2025-04-25 | Keep top-nav layout | Sidebar migration is too high cost/ratio for current value |
