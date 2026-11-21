# Proposal

## Why

OpenMD 目前只提供渲染后的阅读视图，用户需要查看原始 Markdown 并把源码复制到别处时无能为力。宽表格还会先把表头文字挤成多行，再出现横向滚动，阅读体验被破坏。信息与设置面板把文件、配色和关于内容堆在一条长滚动区里，随着设置项增加会继续变长，缺少稳定的分类结构。

## What Changes

- 顶栏增加源码视图切换：当前文档在渲染视图和源码视图之间切换，源码只读、带 Markdown 语法高亮、自动换行。
- 源码视图包含 front matter，支持复制当前文档完整源码；源码内 `Ctrl/Cmd + A` 只选中源码。
- 单文件、目录、链接文档三种模式都支持源码视图；切换文档时保持当前视图模式。
- 宽表格的表头不再折行，表格整体横向滚动；正文单元格仍正常换行。
- 信息与设置面板改为“文件 / 设置 / 关于”三个 Tab；标题栏和 Tab 栏固定，内容区独立滚动。
- 面板标题改为 `OpenMD`，精简重复分组标题，并为未来设置项预留“设置”Tab 内的分组结构。
- 不改变 Markdown 渲染规则、目录扫描、链接打开方式和现有四套配色。

## Capabilities

### New Capabilities

- `source-view`: 当前文档的只读源码视图，包括顶栏切换、Markdown 语法高亮、自动换行、复制源码和源码区域内的全选行为。
- `reader-panel`: 阅读页的信息与设置面板，包括“文件 / 设置 / 关于”Tab、文件信息、配色设置、关于信息和键盘交互。

### Modified Capabilities

- `markdown-rendering`: 新增宽表格布局要求，规定列较多时表头不折行、表格整体横向滚动。

## Impact

- `internal/render/document.go`：增加源码高亮数据。
- `internal/render/links.go`：为加载成功的文档附上源码高亮结果。
- `internal/render/page.go`、`internal/render/web/page.html`：源码切换按钮、源码块、Tab 面板模板。
- `internal/render/web/reader.css`：源码视图、面板 Tab、宽表格表头样式。
- `internal/render/web/reader.js`：源码模式切换、复制源码、源码内全选、面板 Tab 交互、滚动位置处理。
- `internal/render/*_test.go`：源码高亮、页面结构、面板结构和表格样式回归测试。
- `README.md`：补充源码视图、宽表格和信息面板说明。
- 无新增依赖；继续使用现有的 Chroma、goldmark 和 github-markdown-css。
