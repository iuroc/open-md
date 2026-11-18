# Design

## Context

`internal/render/safe_links.go` 已经接管 Markdown 链接的 HTML 输出；链接目标在 `RenderMarkdown` 中先经过 `RewriteURL`，所以到这里时：

- 已预渲染的 Markdown 文档是 `#doc=...`
- 页内锚点也是 `#doc=...`
- 本地非 Markdown 文件是 `file:///...`
- 外部链接是 `https://...` / `http://...`

因此可以在渲染器里按最终 `href` 前缀分类，不需要修改链接重写逻辑。

## Goals / Non-Goals

**Goals:**

- 外部 `http(s)` 和本地 `file:` 非 Markdown 链接默认在新标签页打开。
- 保留 `#` 内部链接、脚注链接和 TOC 链接的当前页行为。
- 为新标签页链接添加 `rel="noopener noreferrer"`。

**Non-Goals:**

- 不增加“外部链接在当前页打开”的设置项；默认行为先保持简单。
- 不改变 `mailto:` / `tel:` 的系统处理方式。
- 不处理未预渲染的网络 Markdown 文档如何在阅读器内打开（后续 `openmd://` 方案）。

## Decisions

### 1. 在 `safeLinkRenderer` 中按最终 href 前缀判断

对每个 `<a>`：

- `http://`、`https://`、`file://`、`//` 开头：输出 `target="_blank" rel="noopener noreferrer"`。
- `#` 开头：不输出 `target`。
- `mailto:`、`tel:`：不输出 `target`。
- 空目标或其他：不输出 `target`。

理由：不依赖 Markdown 原始目标，避免 `[x](other.md#a)` 这类已被重写为 `#doc=...` 的内部链接被误判为外部链接。

备选方案：使用 `<base target="_blank">`。它会影响 TOC、脚注回链和 `#doc=` 链接，无法只针对外部目标，因此不采用。

### 2. 只加 `target` 和 `rel`，不加视觉标记

外部链接已经是一个可识别的 `<a>`；本变更不增加 `↗` 图标，避免阅读界面的视觉噪音。若后续有可访问性需求，可再加 `title` 或图标。

## Risks / Trade-offs

- **标签页变多**：外部链接多时会产生更多标签。→ 这是阅读器保留当前上下文的代价；用户可随手关闭，必要时后续再加设置项。
- **无障碍预期**：屏幕阅读器不会自动提示“新标签页”。→ 本变更先保持默认行为；后续可补 `title` 或视觉标记。
- **本地文件可能是 Markdown**：已预渲染的 Markdown 是 `#doc=...`，不会走 `file:` 分支；未预渲染的网络 Markdown 会进入新标签页并显示原始文件，至少不会顶掉阅读器。

## Migration Plan

无数据迁移。重新打开文档即可生效；回滚只需回退渲染器改动。
