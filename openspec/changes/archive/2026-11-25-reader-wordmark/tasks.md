# Tasks

## 1. 模板字标

- [x] 1.1 在 `docpage.html` 的侧栏标题和 `settings-title` 位置加入 `.brand` 标记（内联 SVG M + `Open`/`MD` 文本），验证 `RenderDocPage` 输出包含 `brand-mark`、`Open`、`MD`，且原来两处 `<span>OpenMD</span>` 已被替换
- [x] 1.2 在 `page.html` 同步相同的字标标记，验证 `RenderPage` 输出与文档页一致

## 2. 字标样式

- [x] 2.1 在 `reader.css` 添加 `.brand`、`.brand-mark`、`.brand-wordmark`、`.brand-open`、`.brand-md` 样式，验证渲染结果的 computed style：文字使用衬线栈、M 使用 `var(--accent)`、文字使用 `var(--fg)`，字号分别继承 18px 和 14px
- [x] 2.2 确认 `.sidebar-header` 与 `.settings-title` 在 14px/18px 下不被关闭按钮挤压、不裁切，并用 290px 侧栏和窄视口截图检查

## 3. 测试与视觉检查

- [x] 3.1 更新 `page_test.go` 的 `settings-title` 断言，新增覆盖两个模板的品牌结构测试（SVG 折线、衬线栈、强调色），运行 `go test ./internal/render` 通过
- [x] 3.2 在默认 / 松石 / 紫藤 / 石墨 / 宣纸的明暗版本下截图检查 14px 和 18px 字标，确认 M 与文字不重叠、不裁切，并记录检查结果

## 4. 缓存与集成

- [x] 4.1 将 `cacheGeneration` 提升到 `2026-11-25-wordmark`，运行 `go test ./... -count=1` 通过；与 `reader-paper-typography` 的最终值合并时保持缓存失效
- [x] 4.2 运行 `go vet ./...` 和 `go test ./... -count=1`，重新渲染一个真实页面并确认侧栏与设置面板都显示新字标

## 验证记录

- 模板结构测试：`go test ./internal/render -count=1` 通过，`brand_test.go` 覆盖两个模板的字标结构。
- 视觉检查：海盐蓝、松石、紫藤、石墨、宣纸五套主题在明暗模式下，380px 与 320px 视口、18px 与 14px 字号均无重叠或裁切；桌面端字标与关闭按钮间距 112px / 198px，窄视口为 112px / 146px。M 的颜色随主题 `--accent` 变化，文字随 `--fg` 变化。
- 真实页面：用 `BuildFileSite` 渲染一篇 Markdown，宣纸主题下桌面 1400×900 与移动 390×844 的侧栏和设置面板都显示新字标，`scrollWidth` 未溢出。
- 全量校验：`go vet ./...`、`go test ./... -count=1` 通过。

## Workflow follow-up

- 在项目的 review 要求满足后归档该变更。
- 归档后确认 `openspec/specs/reader-branding/spec.md` 已生成。
