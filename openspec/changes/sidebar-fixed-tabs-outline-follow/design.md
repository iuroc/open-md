# Design

## Context

- 侧栏结构是 `#sidebar`（固定宽度）→ `.sidebar-header`（60px）→ `.sidebar-scroll`（`overflow: auto`，包含 Tab 和两个面板）。Tab 和列表共用一个滚动容器，所以列表变长后 Tab 会滚出视野。
- `docpage.html` 和 `page.html` 都有侧栏，`page.html` 还分“目录 / 链接文档 / 单文档”三种分支；改动应尽量不动模板结构。
- `reader.js` 的 `setActiveTOC` 只切换 `.active` 类；`updateActiveTOC` 在滚动事件里持续调用，最近的改动通过 `outlineScrollTarget` 保证点击后的高亮不被平滑滚动覆盖。
- 侧栏滚动位置已经按 root 保存在 `sessionStorage`，并有独立的滚动监听。
- 侧栏当前 Tab 只由模板里的 `active` 类决定，点击只切换 DOM，没有写入存储；刷新后回到“文件列表”。

## Goals / Non-Goals

**Goals:**

- Tab 在任意列表滚动位置都可点击，且不改变现有 DOM 顺序和键盘操作。
- 高亮项变化时只滚动侧栏列表，且仅在必要时滚动。
- 与现有的大纲点击、双视图滚动和高亮判定逻辑兼容。
- 刷新或在同一浏览器标签页内切换文档后，保留上次的侧栏 Tab。

**Non-Goals:**

- 不改变“哪个标题是高亮”的判定规则。
- 不调整文件树展开、侧栏滚动位置持久化和大纲点击定位。
- 不把文件列表和大纲拆成两个独立滚动容器。
- 不在关闭浏览器后继续保留 Tab 选择，也不按文档或目录分别记录。

## Decisions

### 1. Tab 移出滚动容器，作为侧栏的固定一行

```html
<div class="sidebar-tabs">...</div>
<div class="sidebar-scroll">
  <div class="tab-panel">...</div>
</div>
```

```css
.sidebar-tabs {
  flex: 0 0 auto;
  padding: 12px 16px 8px;
  border-bottom: 1px solid var(--border);
  background: var(--bg-sidebar);
}
```

- 侧栏本身是纵向 flex 容器，把 Tab 设为不可收缩的固定一行，`.sidebar-scroll` 继续 `flex: 1; overflow: auto`，只有列表滚动。
- 相比 `position: sticky`：sticky 在带 `padding` 的滚动容器里会相对内容区而不是滚动口对齐，顶部残留细条，且列表内容会从 Tab 上方露出；移出滚动容器后几何关系明确，高亮项的可见区计算也更简单。
- 两个模板各调整一次结构；`page.html` 的三种分支只把 Tab 移到 `.sidebar-scroll` 之前，面板内容顺序不变。

### 2. 高亮项用自定义滚动计算显示，而不用 `scrollIntoView`

```js
function revealTOCLink(link) {
  // 取侧栏可视区域，并扣除 sticky Tab 的高度
  // 只有高亮项超出上下边界时，才调整 sidebarScroll.scrollTop
}
```

- `scrollIntoView` 可能连带滚动外层窗口，且会把 sticky Tab 下方的区域当作可见范围，导致高亮项藏在 Tab 后面。
- 直接修改 `sidebarScroll.scrollTop` 只影响侧栏；滚动幅度按超出量计算，已可见时 `delta === 0`，不产生抖动。
- 用 Tab 的 `getBoundingClientRect().bottom` 作为可见区上边界，Tab 不存在或高度为 0 时退回侧栏滚动区顶部。

### 3. 只在“高亮 id 变化”和“切到大纲 Tab”时显示

- `setActiveTOC` 记录 `lastActiveTOC`；id 不变时不重复计算和滚动。滚动事件仍按现有频率调用高亮判定，但不会每次都动侧栏。
- `activateTab("toc")` 在 `updateActiveTOC()` 后强制显示当前项，覆盖“列表被滚到别处、但高亮 id 没变”的情况。
- `replaceTOC` 在切换文档时重置 `lastActiveTOC`，让新文档的首个高亮项也能显示。
- 大纲面板隐藏时直接返回，避免在“文件列表”Tab 中无意义地改动侧栏滚动，也避免和文件树的滚动位置恢复互相干扰。

### 4. 与点击定位和高亮保持解耦

- `setOutlineScrollTarget` 仍然直接设置点击项为高亮；`revealTOCLink` 只滚动侧栏，不修改 `outlineScrollTarget`，所以平滑滚动期间的高亮不会被覆盖。
- 点击项本身就在视野内，显示逻辑通常不产生额外滚动。

### 5. 缓存代次

`reader.css` 和 `reader.js` 嵌入页面缓存，提升 `cacheGeneration` 让已有页面重新生成。

### 6. 侧栏 Tab 选择用 sessionStorage 保留

- 键名 `openmd-sidebar-tab`，只接受 `files` 和 `toc`；与阅读模式、侧栏滚动位置同属当前浏览器标签页的会话状态。
- 用户点击 Tab 时写入；初始化时读取，只有对应 Tab 在当前页面存在才切换过去，否则保持模板默认（文件列表，或没有文件列表时的大纲）。
- 初始化恢复不写入，避免程序回退覆盖用户上次的选择。
- 页面启动时 body 由 `openmd-booting` 隐藏，恢复发生在显示之前，不会先闪一下文件列表。
- 不使用 `localStorage`：关闭浏览器后回到默认，符合“无历史记录才默认”的预期，也和现有阅读模式、滚动位置的生命周期一致。

## Risks / Trade-offs

- **Tab 与列表的间距变化** → Tab 移出滚动容器后保留 `12px` 顶部内边距和 `8px` 底部内边距，和原来的视觉间距接近；列表滚动不再影响 Tab。
- **自动滚动和用户手动浏览大纲冲突** → 仅在内容滚动导致高亮 id 变化时触发；用户单独滚动大纲列表不会触发自动滚动。
- **高亮项比可视区还高** → 仍按超出量滚动，优先露出顶部；标题通常不会超过侧栏高度。
- **`page.html` 单文档分支没有 Tab** → `revealTOCLink` 在找不到 Tab 时使用侧栏顶部作为边界，行为仍然正确。
- **隐藏面板下的布局数据为 0** → 先检查面板是否隐藏，隐藏时直接跳过。
- **记录的 Tab 在当前页面不存在** → 初始化时校验 Tab 是否存在，不存在就保持模板默认，不覆盖记录。

## Migration Plan

- 无需数据迁移。
- 提升缓存代次后，已有页面在下次打开时重新生成。
- 回滚时恢复旧版 `reader.css` / `reader.js` 即可，`sessionStorage` 中的侧栏滚动键和 `openmd-sidebar-tab` 键都不会造成影响。

## Open Questions

无。
