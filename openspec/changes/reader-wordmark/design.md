# Design

## Context

参见 `proposal.md` 的动机。现状：

- 文档页模板 `internal/render/web/docpage.html` 的侧栏标题是 `<span>OpenMD</span>`，信息与设置面板标题是 `<span class="settings-title">OpenMD</span>`；`page.html` 有同样的两处。
- 现有 `.sidebar-header` 为 18px/700 sans，`.settings-title` 为 14px/700 sans。
- 应用图标由 Go 代码绘制：圆头折线 M，点为 `(78,176) (78,80) (128,132) (178,80) (178,176)`，`stroke-width=30`，青色描边。
- 页面通过 `html[data-theme]` 和 CSS 变量切换配色，`--accent`、`--fg` 已覆盖明暗模式和四套主题。
- 两个页面模板各自用 `template.Must(template.New(...).Parse(source))` 独立解析，没有共享 partial。
- 页面是本地 `file://`，不能依赖网络字体；CSP 允许内联 SVG。

## Goals / Non-Goals

**Goals:**

- 用内联 SVG 的图标 M + 系统衬线字标替换纯文字 OpenMD。
- 字标在 14px 和 18px 下清晰、完整、不裁切。
- 字标完全复用现有主题变量，不引入新颜色或新字体文件。
- 保持 `OpenMD` 作为可访问文本，不改变页面标题、favicon 或交互。

**Non-Goals:**

- 不重绘整套拉丁字形，不嵌入字体子集。
- 不改 exe 图标、favicon、浏览器标签标题。
- 不加动画、纸张纹理、渐变或阴影。
- 不改设置面板结构、主题 token 或阅读排版设置。

## Decisions

### 1. 图标 M 用内联 SVG，而不是字体或位图

复用图标代码里的折线点，裁切到 `viewBox="63 65 130 126"`，`stroke-width="30"`，圆头圆角。内联 SVG 在离线页面里精确、可缩放、可用 `currentColor` 跟随主题，且没有字体授权或加载成本。

替代方案：

- CSS 只换衬线字体：改动最小，但和图标没有联系，识别度不足。
- 嵌入只含 OpenMD 的 WOFF2 子集：可控字形，但需要子集化、授权和额外的 data URI；对本变更收益不足。
- 位图 M：主题变色和高 DPI 下不如 SVG。

### 2. 文本使用系统衬线栈

字体栈：`"Palatino Linotype", "Book Antiqua", Palatino, Georgia, "Songti SC", SimSun, serif`。覆盖 Windows/macOS 常见衬线，最后回退到通用 `serif`。`Open` 用 400，`MD` 用 700，字标字距 `.005em`，M 与文字间距 `.42em`。

替代方案：继续用 sans。被否，因为衬线字标和阅读器的纸张/书籍气质更匹配，也能和 UI sans 形成清晰分工。

### 3. 两个模板各自保留一份小标记

`docpage.html` 和 `page.html` 的模板解析互相独立。字标只有约 6 行标记，直接在两处写入；测试同时覆盖两个模板。抽出 `brand.html` 共享 partial 需要改动两处 `template.Parse` 的装配方式，收益不足以抵消额外管道。

### 4. M 用强调色，文字用前景色

`.brand-mark { color: var(--accent); }`，文字继承 `var(--fg)`。这样 M 在蓝色/松石/紫藤/石墨主题下自动变化，未来宣纸主题只需定义 `--accent` 即可，不需要字标专用变量。

替代方案：固定青色 M。被否，在紫藤和石墨主题下会和界面冲突。

### 5. 可访问性保持纯文本

SVG 加 `aria-hidden="true"` 和 `focusable="false"`；`Open` 与 `MD` 两个 span 保留真实文本，不额外加 `aria-label`，避免品牌名被朗读两次。

### 6. 提升缓存代次

模板和共享 CSS 变化后，已有渲染缓存必须失效。`cacheGeneration` 提升到 `2026-11-24-wordmark`。另一个并行变更 `reader-paper-typography` 也会改同一常量，最终合并时使用两者之一或合成值；只要最终值不同于两个变更之前的缓存即可。

## Risks / Trade-offs

- **系统衬线跨机器差异** → 使用覆盖 Windows/macOS 的字体栈并以通用 `serif` 兜底；在 Windows 的 Chrome 和 Edge 上验证 14px/18px。
- **M 在 14px 下偏重** → 按 1em 实现后检查设置面板标题；如偏重，可把 M 调整到 0.92em 或降低描边，但不改变字标结构。
- **侧栏标题与关闭按钮挤压** → 字标保持 `white-space: nowrap`，`min-width: 0` 处理极窄视口；在移动端侧栏检查。
- **和 `reader-paper-typography` 的模板/CSS 冲突** → 已和另一会话确认只改各自区域；`cacheGeneration` 在 rebase 时合并。
- **文件面板、TOC 高亮等依赖标题文本的测试** → 保留 `OpenMD` 文本节点，只更新结构断言。
