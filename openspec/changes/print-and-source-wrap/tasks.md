# Tasks

## 1. 打印布局与模式分派

- [x] 1.1 在 `reader.css` 的 `@media print` 中拆平双视图：重置 `body.split-mode`、`#doc-content`、`.doc-section` 的高度和溢出，隐藏 `.source-view`，让 `.doc-rendered` 回到普通文档流；运行 `go test ./internal/render` 验证打印 CSS 断言通过
- [x] 1.2 在打印样式中重置 `.app` 和 `#content` 的 `min-height`，避免末尾空白页；验证打印相关测试通过
- [x] 1.3 在 `page_test.go` 增加双视图打印用例，断言打印块包含分栏拆平、源码栏隐藏和正文完整流式输出的规则

## 2. 打印时的代码块与表格溢出

- [x] 2.1 打印时对 `.markdown-body pre` 和 `pre > code` 强制 `white-space: pre-wrap`、`overflow: visible` 和断词规则；验证 CSS 测试覆盖这些属性
- [x] 2.2 打印时把 `.markdown-body table` 改为 `display: table`、`width: 100%`、`overflow: visible`，并让表头/单元格允许换行；验证 CSS 测试覆盖表格适配规则
- [x] 2.3 打印源码模式时强制源码换行；验证打印覆盖规则优先级测试通过

## 3. 源码视图固定自动换行

- [x] 3.1 从 `docpage.html` 和 `page.html` 移除“自动换行”开关；验证模板不再包含 `source-wrap-toggle`
- [x] 3.2 从 `reader.js` 移除源码换行设置状态、localStorage 读写和事件监听；验证源码视图使用固定换行
- [x] 3.3 从 `reader.css` 移除 `source-no-wrap` 分支和横向滚动规则，恢复双视图源码栏无条件隐藏滚动条；验证测试通过
- [x] 3.4 更新 `README.md`，移除自动换行设置说明，保留“源码始终自动换行”的描述
- [x] 3.5 源码视图启用 `word-break: break-all`，确保超长单词也能强制断开；验证 CSS 测试和 headless Edge 无横向滚动

## 4. 双视图宽度按钮

- [x] 4.1 在 `reader.css` 增加 `body.split-mode #width-toggle { display: none; }`；验证 CSS 测试断言该规则存在
- [x] 4.2 验证双视图切回阅读/源码后宽度按钮恢复，且 `body.full-width` 状态没有被清除

## 5. 集成验证

- [x] 5.1 运行 `go test ./...` 和 `go vet ./...`，确认全部通过
- [x] 5.2 用临时站点在 headless Edge 中验证：阅读打印、源码打印、双视图打印均符合预期
- [x] 5.3 按仓库约定提交改动，然后执行 `printf '\n' | cmd.exe //c build.bat`，确认 `OpenMD.exe` 和 `source.git.bundle` 构建成功

## 6. 打印背景与源码工具栏

- [x] 6.1 在打印样式中设置页面底色为纯白，并覆盖应用与 Markdown 颜色变量，保证深色模式下仍为深色可读文本；验证打印 CSS 测试通过
- [x] 6.2 在打印样式中隐藏 `.source-toolbar`；验证源码打印 PDF 不包含文件名或“复制源码”
- [x] 6.3 更新 `README.md` 打印说明；验证文档描述与实现一致
- [x] 6.4 用临时站点重新生成打印 PDF，确认白底、源码工具栏消失、正文与代码仍完整
- [x] 6.5 收窄打印白色背景：只设置页面画布为纯白，代码块、表格、行内代码保留浅色主题背景；验证打印 CSS 测试和 PDF 内容
