# Tasks

## 1. 源码数据与高亮

- [x] 1.1 在 `internal/render/document.go` 增加 `SourceHTML`、`HasSource` 字段和 `highlightSource`，用 Chroma Markdown 词法分析器生成带明暗类的 token HTML；补充普通内容、空内容和高亮失败回退的测试，运行 `go test ./internal/render -run TestHighlightSource -count=1` 通过
- [x] 1.2 在 `internal/render/links.go` 的 `buildPageFromSources` 中为成功读取的文档附加源码 HTML，读取失败保持 `HasSource=false`；为零字节文件和高亮失败补充测试，运行 `go test ./internal/render -run TestBuildPageSource -count=1` 通过
- [x] 1.3 补充 front matter 断言：源码 HTML 包含 front matter，渲染 HTML 仍不含 front matter；运行对应测试通过

## 2. 源码视图界面与交互

- [x] 2.1 修改 `internal/render/web/page.html`：增加顶栏源码切换按钮、`.doc-rendered` 包裹、`.source-view` 工具栏、`.source-code` 和空状态，并在 section 上输出 `data-has-source`；补充 `RenderPage` 结构测试，运行 `go test ./internal/render -run TestRenderPageSource -count=1` 通过
- [x] 2.2 在 `internal/render/web/reader.css` 让源码视图与正文同列排版：无外层卡片、无边框工具栏、等宽字体、`pre-wrap` 自动换行、透明 Chroma 背景和焦点样式；生成带长行、front matter 和代码块的 fixture 页面，用 Chrome headless 截图确认换行和明暗样式
- [x] 2.3 在 `internal/render/web/reader.js` 实现源码模式切换、`aria-pressed` 同步、文档切换时保持模式、切换时保留滚动位置；在浏览器中切换长文档并确认滚动位置和按钮状态
- [x] 2.4 实现“复制源码”：从当前 `.source-code` 读取 `textContent`，优先 Clipboard API，失败回退现有 textarea 方案，显示“已复制 / 复制失败 / 没有可复制的源码”；在浏览器中复制含 front matter 的文档并核对剪贴板，空文件确认提示
- [x] 2.5 实现源码内 `Ctrl/Cmd + A` 只选中源码，以及点击大纲时先退出源码模式；在浏览器中分别验证源码内全选、页面其他位置全选和大纲跳转
- [x] 2.6 更新 `README.md` 的源码视图功能说明，确认按键、换行和复制行为与实现一致

## 3. 宽表格布局

- [x] 3.1 在 `internal/render/web/reader.css` 增加 `.markdown-body table th { white-space: nowrap; }`，并补充 CSS 规则回归测试；运行测试通过
- [x] 3.2 用 13 列宽表格 fixture 在 760px 和桌面宽度下生成页面截图，确认表头单行、表格可横向滚动、正文单元格正常换行；更新 `README.md` 中宽表格行为说明

## 4. 信息面板 Tab

- [x] 4.1 修改 `internal/render/web/page.html`：面板标题改为 `OpenMD`，增加“文件 / 设置 / 关于”Tab、三个 `role="tabpanel"`、文件路径与复制路径、配色和关于内容，保留发布页条件；更新现有面板测试并运行 `go test ./internal/render -run TestRenderPage -count=1` 通过
- [x] 4.2 在 `internal/render/web/reader.css` 增加面板 flex 布局、固定头部与 Tab 栏、独立滚动内容区、等宽三项 Tab、hover 和 `focus-visible`；生成页面并用截图确认三个 Tab、内容滚动和窄屏布局
- [x] 4.3 在 `internal/render/web/reader.js` 实现 Tab 切换、`aria-selected` 和 roving `tabindex`、左右方向键与 Home/End、会话内保留上次 Tab；在浏览器中逐项验证鼠标和键盘操作
- [x] 4.4 更新 `README.md` 的面板说明和文案，确认标题、Tab 名称、复制路径和发布页描述与实现一致

## 5. 集成验证

- [x] 5.1 运行 `go test ./... -count=1`，确认全部测试通过
- [x] 5.2 分别用单文件、目录和链接文档 fixture 打开页面，验证源码切换、复制、源码全选、宽表格滚动、三个面板 Tab 和四套配色无回归；用截图记录关键状态
- [x] 5.3 按仓库约定提交改动，然后执行 `printf '\n' | cmd.exe //c build.bat`，确认 `OpenMD.exe` 和 `source.git.bundle` 生成成功

## Workflow follow-up

- 等待你审阅本 change 的 proposal、specs、design 和 tasks。
- 审阅通过后执行 `/opsx:apply` 开始实现。
- 实现和验证完成后归档该 change。
