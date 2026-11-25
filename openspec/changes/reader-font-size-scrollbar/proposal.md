# Proposal

## Why

现有阅读设置已经支持配色、字体和字间距，但正文字号固定为 16px，长时间阅读时无法按屏幕距离和个人习惯调整。同时，滚动条仍使用浏览器默认样式，在五套配色（尤其宣纸）下显得突兀，破坏整页的视觉一致性。

## What Changes

- 设置面板的“设置”Tab 在“字体”和“字间距”之间增加“字号”预设：小 `14px` / 标准 `16px` / 大 `18px` / 特大 `20px`，默认标准。
- 字号 MUST 只作用于内容区域（渲染正文和源码视图），MUST NOT 改变侧边栏、顶栏、设置面板和目录的字号。
- 标题、列表、表格、行内代码和代码块随正文按现有 `em`/`%` 规则等比缩放；脚注和 CSV 表格的固定 `12px` 改为相对单位，默认仍为 `12px`。
- 字号偏好写入 `localStorage`，刷新页面和在文档间软导航后保持；未选择时保持现有 16px 观感。
- 滚动条改为主题化外观：颜色由当前配色的弱化文字色推导，五套配色和明暗模式自动适配；使用圆角、透明轨道的细滚动条，hover 时加深。
- 滚动条覆盖窗口主滚动区域、侧边栏、设置面板和分栏阅读的两栏；分栏源码栏保持现有隐藏滚动条行为。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `reader-appearance`: 增加字号预设（只影响内容区域）和随主题变化的滚动条外观两条要求。
- `reader-panel`: “设置”Tab 的阅读设置从配色、字体、字间距扩展为配色、字体、字号、字间距。

## Impact

- `internal/render/web/reader.css`：字号变量与规则、脚注/CSV 相对字号、滚动条 token 与样式。
- `internal/render/web/reader.js`：字号预设、偏好读取/应用/持久化、初始化与按钮事件。
- `internal/render/web/docpage.html`、`page.html`：“设置”Tab 增加字号分组。
- `internal/render/site.go`：提升缓存代次，使已有页面重新生成。
- 测试：扩展 `internal/render/page_test.go` 的 CSS/JS/模板契约测试。
- `README.md`：设置说明补充字号和主题滚动条。
- 无新增依赖；不改变滚动行为、锚点和滚动位置恢复逻辑。
