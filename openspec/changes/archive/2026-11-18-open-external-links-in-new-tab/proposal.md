# Proposal

## Why

阅读页是临时 HTML 页面，当前点击外部链接会覆盖阅读器标签页，丢失当前文档、滚动位置和侧栏状态；外站的重定向和弹窗也会打断阅读。外部链接和本地非 Markdown 文件更适合在新标签页打开。

## What Changes

- `http` / `https` 链接默认在新标签页打开，并带 `rel="noopener noreferrer"`。
- 本地 `file:` 非 Markdown 文件（例如 PDF、图片、压缩包）默认在新标签页打开。
- 页内锚点、已预渲染的 Markdown 文档、脚注回链继续在当前页切换。
- `mailto:` / `tel:` 不强制新标签页，交给系统处理。
- 更新 README，说明上述行为。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `markdown-rendering`: 增加链接打开目标的约定，区分外部/本地文件链接与阅读器内部链接。

## Impact

- 代码：`internal/render/safe_links.go` 及相关测试。
- 行为：外部链接和本地非 Markdown 文件的默认打开方式变化；不涉及安装、注册表或浏览器启动方式。
- 文档：README 的链接行为说明。
- 依赖：不新增依赖。
