# Proposal

## Why

长篇 Markdown 阅读时，现有四套配色和固定正文排版偏工具化。用户希望像常见小说阅读器一样拥有偏暖的纸感底色，并能选择字体和字符间距，以降低长时间阅读的视觉疲劳。

## What Changes

- 新增第 5 套配色“宣纸”（`paper`），含明暗两版；纸底 `#F6F0E3`、墨色 `#2F2A23`、黛青强调 `#3F5D6B`，不添加纸张纹理或噪点。
- 设置面板的“设置”Tab 在配色之外增加“字体”和“字间距”两组预设：
  - 字体：默认 / 宋体 / 楷体 / 仿宋；
  - 字间距：标准 `0` / 疏朗 `0.03em` / 舒展 `0.06em` / 宽松 `0.10em`。
- 字体和字间距只作用于渲染后的正文；代码块和源码视图 MUST 保持等宽字体与正常字距。
- 阅读偏好写入 `localStorage`，刷新页面和在文档间软导航后保持；未选择新偏好时保持现有默认观感。
- 不引入字号、行高、页边距设置，不支持自定义字体输入，不打包 webfont。

## Capabilities

### New Capabilities

- `reader-appearance`: 阅读页的配色主题、正文字体与字符间距，以及这些偏好的默认值和持久化。

### Modified Capabilities

- `reader-panel`: “设置”Tab 的阅读设置从仅配色扩展为配色、字体、字间距三组，并保持即时生效与持久化。

## Impact

- `internal/render/web/reader.css`：宣纸配色 token、主题色点、正文排版变量、设置控件样式。
- `internal/render/web/reader.js`：主题白名单、排版预设、偏好读取/应用/持久化、设置按钮事件。
- `internal/render/web/docpage.html`、`page.html`：主题白名单、宣纸按钮、字体和字间距分组。
- `internal/render/site.go`：提升缓存代次，使已有页面重新生成。
- 测试：模板、CSS/JS 契约和浏览器行为验证。
- 无新增依赖；不加载网络字体，`file://` 与现有 CSP 约束不变。
