# Proposal

## Why

双视图下点击侧栏大纲时，阅读区已经滚动到目标标题，但大纲高亮仍停在前一节：高亮阈值按窗口顶部计算，没有把顶栏和滚动面板的偏移算进去。源码模式下点击大纲则被强制切回阅读模式，用户无法在源码里按大纲定位；这一行为来自 `source-view` 中“大纲导航离开源码模式”的既有要求，需要一并调整。

## What Changes

- 新增统一的「大纲定位」行为：点击大纲时按当前视图选择定位目标。
  - 阅读模式：滚动窗口到目标标题。
  - 双视图：先定位阅读区，再把源码区对齐到目标标题所在的源码行；大纲高亮 MUST 等于被点击的标题。
  - 源码模式：保持在源码模式，把源码滚动到目标标题行，不再强制切回阅读模式。
- 双视图下滚动阅读区或源码区时，大纲高亮都跟随当前滚动的一侧更新；源码模式下滚动页面也更新高亮。
- 模板向大纲项暴露标题在源码中的行号，供源码区定位使用。
- 移除 `source-view` 中“大纲导航离开源码模式”的要求，改由新的 `outline-navigation` 能力覆盖源码模式下的大纲行为。
- 模板、共享脚本变化后提升 `cacheGeneration`。

## Capabilities

### New Capabilities

- `outline-navigation`: 点击大纲和滚动各视图时，如何在阅读区、双视图源码区、源码模式之间定位并同步大纲高亮。

### Modified Capabilities

- `source-view`: 删除“大纲导航离开源码模式”的要求，源码模式下点击大纲改为留在源码视图并定位到对应标题行。

## Impact

- `internal/render/document.go`：大纲标题记录源码行号
- `internal/render/web/docpage.html`、`internal/render/web/page.html`：大纲项输出行号
- `internal/render/web/reader.js`：大纲点击、双区同步、源码定位与高亮更新
- `internal/render/web/reader.css`：仅在需要调整定位偏移时改动
- `internal/render/site.go`：提升缓存代次
- 相关 Go 测试与前端渲染验证
- 不改变文档内容渲染、文件树导航、主题或阅读排版设置
