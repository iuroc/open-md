# Proposal

## Why

侧栏的“文件列表 / 大纲”Tab 当前位于滚动容器内部：标题或文件很多时，列表滚到中间后 Tab 会被推出视野，必须先滚回顶部才能切换。另一方面，滚动正文时大纲高亮虽然会更新，但侧栏大纲列表本身不滚动，长目录下当前条目常常落在视野之外，读者看不出高亮去了哪里。

## What Changes

- 侧栏视图 Tab 固定在侧栏顶部，只有当前 Tab 的列表区域滚动；文件列表和大纲共用这一行为。
- 滚动正文、源码或双视图任一侧导致大纲高亮切换时，侧栏大纲列表以最小幅度滚动，让新的高亮项保持可见。
- 当前高亮项已经在视野内时不做任何滚动；切换到“大纲”Tab 时，当前高亮项应立即可见。
- 自动滚动只作用于侧栏列表，MUST NOT 改变正文滚动位置、当前高亮标题或大纲点击定位行为。

## Capabilities

### New Capabilities

- `reader-sidebar`: 阅读页侧栏的视图切换与滚动行为，包括固定 Tab、列表独立滚动，以及大纲高亮项在侧栏中的可见性。

### Modified Capabilities

无。

## Impact

- `internal/render/web/reader.css`：侧栏 Tab 的 sticky 定位、背景和全宽处理。
- `internal/render/web/reader.js`：大纲高亮变化时自动显示当前条目，切换 Tab 和切换文档时重置状态。
- `internal/render/site.go`：提升缓存代次。
- 测试：扩展 `internal/render/page_test.go` 的 CSS/JS 契约测试。
- 无新增依赖，不改变正文滚动、大纲定位和高亮判定逻辑。
