# 宣纸主题与正文排版设置设计

日期：2026-11-24
规模：小改动

## 意图

让长文阅读更接近小说阅读器：一套偏暖的“宣纸”配色，正文可以切换字体和字符间距，设置跨刷新、跨软导航保持。

## 目标

- 新增第 5 套配色 `paper`（宣纸），含明暗两版。
- 设置面板新增“字体”和“字间距”两组预设。
- 设置保存在 `localStorage`，刷新和软导航后保持。
- 现有 4 套配色、默认字体、默认字距的观感不变。

## 非目标

- 不新增字号、行高、页边距设置。
- 不做自定义字体输入，不打包 webfont，不做纸张纹理或噪点。
- 不改变代码块、源码视图、双视图代码侧的等宽字体。
- 不修改现有 4 套配色的色值。

## 反目标

- 不落入“暖米色背景 + 陶土色强调”的通用搭配。
- 不把设置面板做成排版实验室：只保留可扫视的预设。
- 不为了“纸感”加纹理、阴影堆叠或装饰性动画。

## 约束

- 页面运行在 `file://`，受 CSP 限制，不能可靠加载网络字体；只用系统字体并带 `serif` 回退。
- `localStorage` 可能不可用，沿用现有 `readPref` / `writePref` 的容错。
- 文档页在 `openmd-ready` 前隐藏 body，排版需在 reveal 前应用，避免闪动。
- `reader.js` 随页面缓存，改动后必须提升 `cacheGeneration`。
- `docpage.html` 是生产模板；`page.html` 仅测试使用，但同步更新，避免两套模板分叉。

## 方案比较

1. **排版实现**
   - 在每个 `.doc-section` 上写 inline style：软导航替换内容后要重新套用，容易漏。
   - 用 `html[data-reader-font="song"]` 等属性规则：可行，但每加一个预设就要多一套规则。
   - **选中：在 `html` 上设置 CSS 变量 `--reader-font`、`--reader-letter-spacing`**。软导航后无需重套，默认值与回退集中，代码块可单独重置。
2. **配色方向**
   - A 宣纸、B 旧书页、C 素白。**选中 A**：最贴合“宣纸”，微黄不脏；强调色用黛青，避开暖米色配陶土色的默认组合。B 更旧更黄，长时间阅读偏闷；C 几乎白，离“宣纸”最远。
3. **控件形式**
   - `<select>` 紧凑但不可扫视；range 适合连续值，但字间距只需要几档。
   - **选中：胶囊按钮预设**，与现有配色网格一致，点一下即预览。

## 设计

### 颜色 token

`html[data-theme="paper"]`：

| token | 明 | 暗 |
|---|---|---|
| `--bg` | `#f6f0e3` | `#1a1814` |
| `--bg-sidebar` | `#efe7d6` | `#211e18` |
| `--bg-subtle` | `#f1ebdd` | `#26221c` |
| `--bg-hover` | `#e7ddc9` | `#2e2922` |
| `--fg` | `#2f2a23` | `#e6ddcb` |
| `--fg-muted` | `#7c705e` | `#a89b85` |
| `--border` | `#ddd2bd` | `#3a342a` |
| `--accent` | `#3f5d6b` | `#93b3c0` |
| `--accent-fg` | `#ffffff` | `#10171b` |
| `--accent-soft` | `#e4ecef` | `#24343b` |

- 暗色通过 `@media (prefers-color-scheme: dark)` 覆盖，沿用现有主题的组织方式。
- 在 `docpage.html` / `page.html` 的头部内联样式中补 `html[data-theme="paper"]` 的背景色，避免样式加载前露白。
- 头部主题白名单和 `reader.js` 的 `validThemes` 都加入 `paper`。

### 排版 token 与预设

- `--reader-font`：默认沿用 `var(--font-body)`；预设栈：
  - song：`"SimSun", "宋体", "Songti SC", "Noto Serif CJK SC", serif`
  - kai：`"KaiTi", "楷体", "Kaiti SC", STKaiti, serif`
  - fang：`"FangSong", "仿宋", "FangSong_GB2312", STFangsong, serif`
- `--reader-letter-spacing`：标准 `0`、疏朗 `0.03em`、舒展 `0.06em`、宽松 `0.10em`。
- 正文覆盖：

```css
.doc-section.markdown-body {
  font-family: var(--reader-font, var(--font-body));
  letter-spacing: var(--reader-letter-spacing, 0);
}
.doc-section.markdown-body pre,
.doc-section.markdown-body code,
.doc-section.markdown-body kbd,
.doc-section.markdown-body samp {
  font-family: var(--font-mono);
  letter-spacing: normal;
}
```

`.doc-section.markdown-body` 的优先级高于 `github-markdown.css` 的 `.markdown-body`；代码块单独重置，源码视图不受影响。

- 存储键：`openmd-reader-font`（`default|song|kai|fang`）、`openmd-reader-spacing`（`0|0.03|0.06|0.1`）。
- 加载时校验预设值并应用，再 reveal；点击按钮立即应用到正文并持久化。

### 设置面板

```text
┌ 配色 ─────────────────────────────┐
│ ● 海盐蓝  ● 松石                   │
│ ● 紫藤    ● 石墨                   │
│ ● 宣纸（整行）                      │
├ 字体 ─────────────────────────────┤
│ [默认]      [宋体]                 │
│ [楷体]      [仿宋]                 │
├ 字间距 ───────────────────────────┤
│ [标准]      [疏朗]                 │
│ [舒展]      [宽松]                 │
└───────────────────────────────────┘
```

- 配色仍是 2 列网格；宣纸作为第 5 项横跨整行（`grid-column: 1 / -1`），与前面四套区分。
- 宣纸色点：明色用纸色底 + 黛青细描边；暗色用深纸色底 + 浅黛青描边。
- 字体和字间距使用与配色一致的胶囊按钮样式（`aria-pressed`），2×2 网格。
- 所有按钮都带 `data-reader-font` / `data-reader-spacing`，由同一段点击处理应用。

## 影响范围

- `internal/render/web/reader.css`：宣纸 token、色点、排版变量覆盖、设置控件样式。
- `internal/render/web/docpage.html`、`page.html`：主题白名单、头部背景、宣纸按钮、字体和字间距分组。
- `internal/render/web/reader.js`：`validThemes`、预设表、读取/应用/持久化、按钮事件。
- `internal/render/site.go`：提升 `cacheGeneration`。
- 测试：模板与 CSS/JS 字符串断言；浏览器里验证主题、字体、字间距、持久化、软导航和默认观感。

## 验收

- 不选新设置时，现有配色、字体、字距与现在完全一致。
- 宣纸主题明暗两版显示正确，刷新后保持。
- 四个字体预设应用到正文，代码块仍为等宽；刷新和软导航后保持。
- 四档字间距生效，代码块不跟随；刷新后保持。
- 现有 4 套主题和双视图、源码视图不受影响。
- `go vet ./...`、`go test ./... -count=1` 通过。

## 留给实现者

- CSS 组织方式与控件类的具体命名。
- 预设校验和不存在按钮时的降级路径。
- 测试的具体拆分。
