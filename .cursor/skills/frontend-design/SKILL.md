---
name: frontend-design
description: Create distinctive, production-grade frontend interfaces with high design quality. Use this skill when the user asks to build web components, pages, or applications (e.g. websites, dashboards, React/Vue components, HTML/CSS layouts) or when styling/beautifying any web UI.
---

# Frontend Design

## Purpose

This skill guides creation of **distinctive、接近生产级** 的前端界面，而不是千篇一律的“AI 风格”界面。  
当用户让你实现或美化界面（组件 / 页面 / 落地页 / 仪表盘 / App UI 等）时，使用本技能。

核心目标：

- **有清晰的审美方向**，而不是折中、平庸的样子
- **实现真实可运行的代码**（而不是只讲概念）
- **细节精致**，包括排版、色彩、动效、空间布局等

参考源文档见 GitHub：`https://github.com/anthropics/skills/blob/3d59511518591fa82e6cfcf0438d68dd5dad3e76/skills/frontend-design/SKILL.md`

---

## 使用时机

在以下场景中，应主动使用本技能：

- 用户说“做一个页面 / 组件 / 前端 UI / 管理后台 / Landing Page / Dashboard / 移动端页面”等
- 用户要求“界面好看 / 现代化 / 具有设计感 / 像正式产品”
- 用户要求“重构页面样式 / 优化 UI / 美化 CSS”
- 用户希望“避免千篇一律的 AI 生成 UI”

---

## 设计思维流程

**在写任何代码之前，先回答这几个问题：**

1. **Purpose（目的）**
   - 这个界面解决什么问题？
   - 典型用户是谁？（开发者、运营、普通用户、高管等）

2. **Tone（审美基调，必须足够明确且偏极端）**
   从下面或自定义的方向中选择一个 **强烈** 的审美，而不是“有点啥都有一点”的折中：

   - 残酷极简（brutally minimal）
   - 极繁主义（maximalist chaos）
   - 复古未来（retro-futuristic）
   - 自然有机（organic / natural）
   - 奢华精致（luxury / refined）
   - 玩具感 / 游戏感（playful / toy-like）
   - 杂志 / 编辑风（editorial / magazine）
   - 工业 / 机能风（industrial / utilitarian）
   - 其他你明确命名的风格

   **关键：**无论选哪种，都要在代码里贯彻到底，而不是半途而废。

3. **Constraints（技术与体验约束）**
   - 哪个框架？例如：Vue、React、Svelte、纯 HTML/CSS/JS 等
   - 是否要考虑移动端 / 响应式？
   - 是否有性能或可访问性（a11y）的要求？

4. **Differentiation（记忆点）**
   问自己：**“用户看完后，会记住什么？”**
   - 特别的排版？
   - 夸张但克制的色彩？
   - 很有趣的 hover/transition？
   - 独特的背景 / 纹理 / 形状？

在回答清楚上面的问题之后，再开始动手写代码。

---

## 实现原则

- **生产级 & 可运行**
  - 代码必须能跑（不要求接后端，但结构要真实）
  - 组件/页面结构合理，语义化 HTML，合理的 class 命名

- **视觉突出 & 记忆点明确**
  - 整体风格统一，不要“这里一点霓虹、那里一点玻璃态”
  - 对于极简：用少量元素做极致打磨
  - 对于极繁：用有组织的“混乱”，避免纯堆叠

- **约束与统一**
  - **颜色、间距、圆角、阴影** 尽量用 CSS 变量或 Token 控制
  - 组件内不要随意发明新尺寸，尽量复用已有 Token

---

## 排版（Typography）

**关键原则：**

- 避免使用：`Arial`、`Inter`、系统默认字体等高度通用字体
- 使用字体来体现个性，可以：
  - 选择一个独特的 **标题 / Display 字体**
  - 搭配一个可靠、易读的 **正文字体**
- 在代码层面：
  - 用 `font-family` 明确指定，而不是依赖浏览器默认
  - 可配合 `letter-spacing`、`line-height`、`text-transform` 做出更有张力的排版

示例（示意，不强制要求使用此组合）：

```css
:root {
  --font-display: "DM Sans", system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-body: "IBM Plex Sans", system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

body {
  font-family: var(--font-body);
}

.page-title {
  font-family: var(--font-display);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
```

当用户没有特别指定字体时，**你可以主动选择一组独特但实用的组合**，并在说明中简单解释选择理由。

---

## 颜色与主题（Color & Theme）

**强制要求：**

- **必须** 选择一个清晰的配色策略，而不是平均分布的中庸色板
- 拒绝：
  - 大路货的紫色渐变 + 白底 + 圆角卡片
  - 毫无目的的多色彩虹

**推荐做法：**

- 选一个**主色**，一个 **强调色**，外加强度不同的中性色（背景 / 文本 / 分割线）
- 使用 CSS 变量统一管理：

```css
:root {
  --color-bg: #050816;
  --color-bg-subtle: #0b1020;
  --color-fg: #f5f5f7;
  --color-muted: #8b8fa3;
  --color-accent: #ffb347;
  --color-accent-soft: rgba(255, 179, 71, 0.16);
  --color-danger: #ff4b81;
}
```

当需要深色 / 浅色模式时，按风格延展，而不是简单反转颜色。

---

## 动效（Motion）

**目标：**用少量但有记忆点的动画，而不是到处乱飞。

优先级：

- **优先 CSS 动画 / 过渡**，在 React/Vue 环境下才考虑动画库
- 对于 HTML/CSS 方案：
  - 利用 `transition`、`transform`、`opacity`、`filter` 等属性
  - 利用 `animation-delay` 做“分段登场”的入场效果

建议的动效场景：

- 页面加载时的模块级入场（从下方轻微上移 + 渐显）
- 卡片 / 按钮 hover 时的轻微缩放和阴影变化
- 标签 / Tab 切换时的滑动下划线或背景块动画

**反模式（要避免）：**

- 所有东西都在动
- 动画持续时间过长导致操作迟缓（一般控制在 150–350ms）

---

## 空间与布局（Spatial Composition）

**核心思想：**通过空间组织形成节奏和层次，而不是简单的栅格堆叠。

可考虑的布局策略：

- 使用 **不对称布局**：一侧为巨大标题和说明，另一侧为内容卡片 / 预览区
- 允许 **元素轻微重叠**，配合阴影 / 模糊创造纵深感
- 利用 **超大内边距** 或 **极窄边距** 营造“呼吸感”或“密集感”

在 CSS 层面，可多使用：

- `display: grid` + `grid-template-columns` 做主布局
- `gap` 来控制整体节奏
- flex/grid 的对齐方式 (`align-items`, `justify-content`) 精细调整视觉重心

---

## 背景与视觉细节（Backgrounds & Details）

不要只用纯色背景。可以考虑：

- 渐变网格（gradient mesh）感的背景（可以用多层 `radial-gradient` 叠加）
- 轻微噪点纹理（使用 background-image + data-uri 或伪元素实现）
- 简洁的几何图案、边框、装饰线
- 玻璃态 / 模糊层，但注意**适度**和性能

示例（多层渐变背景示意）：

```css
.app-background {
  background:
    radial-gradient(circle at 0% 0%, rgba(255, 179, 71, 0.16), transparent 55%),
    radial-gradient(circle at 100% 100%, rgba(111, 66, 193, 0.2), transparent 50%),
    linear-gradient(145deg, #050816, #050816 40%, #10172f);
}
```

---

## 需要主动避免的“AI 审美”

当你设计时，要刻意回避以下模式：

- **字体**：大量使用 `Inter`、`Roboto`、`Arial`、`system-ui` 且不给出任何理由
- **配色**：
  - 白底 + 浅灰卡片 + 紫色主色 + 蓝色次要色
  - 大面积无目的的渐变背景
- **布局**：
  - 标准“三段式”：顶部大标题 + 三列卡片 + 底部 CTA，且毫无变形
  - 所有元素同样大小、同样圆角、同样阴影，没有主次

相反，你应该：

- 结合业务特点，选择**独特、命名得出的风格**
- 用文字一句话概括你设计的“性格”，并在代码中贯彻

---

## 根据复杂度匹配实现量

**复杂风格（极繁 / 重动效 / 丰富背景）**

- 需要：
  - 更复杂的布局
  - 更多动效（但仍需组织良好）
  - 一定量的自定义装饰元素（例如 SVG、伪元素、图案）

**极简 / 精致风格**

- 需要：
  - 极其精确的对齐与间距
  - 非常讲究的排版细节（字号层级、字重、行高）
  - 微妙却高级的色差和阴影

**总结：**优雅来自于“执行到位”，而不是“堆很多东西”。

---

## 实战使用步骤（给未来的自己用）

当你收到一个“做 UI / 页面 / 组件”的请求时，按以下流程来：

1. **澄清需求（在心里或用一句话写出来）**
   - 这是给谁用的？目的是什么？核心使用场景？
2. **选择审美方向并命名**
   - 例如：“冷调工业风监控面板”、“玩具感的阅读面板”、“复古杂志感博客”
3. **确定技术栈与约束**
   - 例如：“用 Vue 3 + `<script setup>`”、“保持单文件组件”、“移动端优先”等
4. **设计视觉系统（在代码中显式定义变量/Token）**
   - 字体变量
   - 颜色变量
   - 间距 / 圆角 / 阴影 Token
5. **先搭骨架再加细节**
   - 先用简陋样式完成布局
   - 再逐步增强字体、颜色、背景、动效
6. **自检：是否“有记忆点”？**
   - 如果没有，至少增加一个非常明确的特色点（特定布局、背景、动效或排版）

---

## 附加建议

- 对于复杂设计，可以在注释或说明文字中简要描述设计理念，方便用户理解和后续维护
- 如果用户已有设计体系（如某个 UI 库或 Design System），优先在其框架内做“有个性”的变化，而不是完全另起炉灶
- 在不影响设计的前提下，尽量保持语义化、可访问性和响应式友好

