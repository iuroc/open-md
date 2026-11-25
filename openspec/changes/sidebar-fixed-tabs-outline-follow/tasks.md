# Tasks

## 1. 固定侧栏 Tab

- [x] 1.1 在 `docpage.html` 和 `page.html` 把 `.sidebar-tabs` 移到 `.sidebar-scroll` 之前，让 Tab 成为侧栏固定的一行；在 `reader.css` 为 `.sidebar-tabs` 设置 `flex: 0 0 auto` 和背景，保留 `.sidebar-scroll` 的 `flex: 1; overflow: auto`；在 `internal/render/page_test.go` 增加 `TestSidebarTabsStayPinned` 断言 CSS 和两个模板的结构顺序，运行 `go test ./internal/render -run TestSidebarTabsStayPinned -count=1` 通过。

## 2. 大纲高亮跟随

- [x] 2.1 在 `reader.js` 增加 `lastActiveTOC`、`tocPanelVisible` 和 `revealTOCLink`（按侧栏可视区和固定 Tab 底部计算超出量，必要时调整 `sidebarScroll.scrollTop`），在 `setActiveTOC` 中仅在 id 变化时调用；增加 `TestReaderJSRevealsActiveOutline` 断言这些函数和调用点，运行通过。
- [x] 2.2 在 `activateTab("toc")` 中强制显示当前高亮项，在 `replaceTOC` 中重置 `lastActiveTOC`；更新 `TestReaderJSRevealsActiveOutline` 覆盖切换 Tab 和切换文档的重置，运行通过。

## 3. 集成与验证

- [x] 3.1 提升 `internal/render/site.go` 的 `cacheGeneration`；运行 `go vet ./...` 和 `go test ./... -count=1` 全部通过。
- [x] 3.2 生成包含大量标题和文件的页面做浏览器验证：文件列表和大纲滚动后 Tab 仍可点击；滚动正文时大纲列表跟随高亮、已可见时不抖动；源码和双视图滚动同样生效；切换 Tab 后当前项可见；侧栏滚动不影响正文。记录每项结果。
- [x] 3.3 提交全部改动，提交信息说明固定侧栏 Tab 和大纲跟随；确认 `git status` 干净。

## 4. 记住侧栏 Tab

- [x] 4.1 在 `reader.js` 增加 `openmd-sidebar-tab` 键、`rememberSidebarTab` 和 `restoreSidebarTab`：点击 Tab 写入 `sessionStorage`，初始化时校验 Tab 存在后恢复，程序回退不覆盖；扩展 `internal/render/page_test.go` 的 `TestReaderJSRestoresSidebarTab`，运行通过。
- [x] 4.2 提升 `cacheGeneration`，运行 `go vet ./...` 和 `go test ./... -count=1` 通过；浏览器验证刷新保留 Tab、无记录默认文件列表、无文件列表时回退大纲、关闭浏览器后不保留。

## 验证记录

- `go vet ./...`、`go test ./... -count=1` 通过。
- 用无头 Chrome 打开真实渲染页面验证侧栏 Tab 记忆：
  - 目录模式首次打开：文件列表；点击大纲后 `sessionStorage` 写入 `toc`，刷新后仍为大纲；再切回文件列表并刷新，仍为文件列表。
  - 同一浏览器中新建标签页打开目录页面：无历史记录，默认文件列表，存储键为空。
  - 单文件（没有文件列表 Tab）页面：始终显示大纲；预设存储值为文件列表时，刷新后不覆盖该值。
  - 带链接文档的单文件页面：无历史记录时默认大纲；点击文档 Tab 后刷新，仍为文档 Tab。

## Workflow follow-up

- 实现完成并验证后归档 change，同步 `reader-sidebar` 主 spec。
