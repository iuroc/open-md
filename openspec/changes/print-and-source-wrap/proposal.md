# Proposal

## Why

当前 `Ctrl + P` 打印没有跟随阅读模式：双视图会把左右两栏一起送入打印，并且受固定高度和内部滚动容器影响，只能打印当前可视区域；渲染后的代码块和宽表格在打印时仍保留横向滚动，内容容易被裁切。源码视图的自动换行目前也不可配置。

## What Changes

- 打印按当前阅读模式分派：阅读模式打印渲染正文；源码模式打印源码；双视图只打印渲染正文。
- 双视图打印时临时拆平为单栏渲染视图，不打印左侧源码栏，也不受屏幕滚动位置影响。
- 打印时代码块临时强制自动换行，取消横向滚动。
- 打印时表格临时适配页面宽度，表头和单元格允许换行，取消横向滚动条。
- “设置”Tab 新增“源码视图 → 自动换行”开关，默认开启并持久化；关闭后源码保留原始行宽、允许横向滚动；打印时始终换行。
- 双视图下隐藏顶栏的宽度模式按钮。

## Capabilities

### New Capabilities

- `reader-printing`: 定义各阅读模式下的打印内容、双视图打印布局，以及打印时代码块和表格的溢出处理。
- `reader-view-controls`: 定义顶栏阅读模式与宽度控件在不同阅读模式下的可用性和可见性。

### Modified Capabilities

- `source-view`: 源码换行由固定行为改为可配置，并明确打印时始终换行。

## Impact

- `internal/render/web/reader.css`：打印媒体查询、源码换行状态、双视图宽度按钮隐藏、滚动条状态。
- `internal/render/web/reader.js`：源码自动换行设置的读取、持久化和开关同步。
- `internal/render/web/docpage.html`、`internal/render/web/page.html`：设置面板新增自动换行开关。
- `internal/render/page_test.go`、`internal/render/site_test.go`：覆盖打印 CSS、换行设置和控件可见性。
- `README.md`：补充打印和源码换行设置说明。
