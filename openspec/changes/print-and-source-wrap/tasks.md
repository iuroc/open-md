# Tasks

## 1. 打印布局与模式分派

- [x] 1.1 在 `reader.css` 的 `@media print` 中拆平双视图：重置 `body.split-mode`、`#doc-content`、`.doc-section` 的高度和溢出，隐藏 `.source-view`，让 `.doc-rendered` 回到普通文档流；运行 `go test ./internal/render` 验证打印 CSS 断言通过
- [x] 1.2 在打印样式中重置 `.app` 和 `#content` 的 `min-height`，避免末尾空白页；验证打印相关测试通过
- [x] 1.3 在 `page_test.go` 增加双视图打印用例，断言打印块包含分栏拆平、源码栏隐藏和正文完整流式输出的规则

## 2. 打印时的代码块与表格溢出

- [x] 2.1 打印时对 `.markdown-body pre` 和 `pre > code` 强制 `white-space: pre-wrap`、`overflow: visible` 和断词规则；验证 CSS 测试覆盖这些属性
- [x] 2.2 打印时把 `.markdown-body table` 改为 `display: table`、`width: 100%`、`overflow: visible`，并让表头/单元格允许换行；验证 CSS 测试覆盖表格适配规则
- [x] 2.3 打印源码模式时强制源码换行，即使 `body.source-no-wrap` 存在；验证打印覆盖规则优先级测试通过

## 3. 源码自动换行设置

- [x] 3.1 在 `docpage.html` 和 `page.html` 的“设置”Tab 增加“源码视图 → 自动换行”开关，使用 `role="switch"` 和 `aria-checked`；验证模板测试能看到开关标记
- [x] 3.2 在 `reader.js` 增加 localStorage 读写（`openmd-source-wrap`）和 `source-no-wrap` 类同步，初始化时在页面显示前应用；验证 JS 测试断言持久化键和状态切换
- [x] 3.3 增加关闭换行后的 CSS：源码 `white-space: pre`，双视图源码栏可横向滚动并显示默认横向滚动条，竖向滚动条仍隐藏；验证 CSS 测试和 headless Edge 横向滚动检查
- [x] 3.4 更新 `README.md`，说明打印模式分派、源码自动换行设置和双视图隐藏宽度按钮；验证文档描述与实现一致

## 4. 双视图宽度按钮

- [x] 4.1 在 `reader.css` 增加 `body.split-mode #width-toggle { display: none; }`；验证 CSS 测试断言该规则存在
- [x] 4.2 验证双视图切回阅读/源码后宽度按钮恢复，且 `body.full-width` 状态没有被清除

## 5. 集成验证

- [x] 5.1 运行 `go test ./...` 和 `go vet ./...`，确认全部通过
- [x] 5.2 用临时站点在 headless Edge 中验证：阅读打印、源码打印、双视图打印均符合预期，关闭换行后双视图出现横向滚动条
- [x] 5.3 按仓库约定提交改动，然后执行 `printf '\n' | cmd.exe //c build.bat`，确认 `OpenMD.exe` 和 `source.git.bundle` 构建成功
