# Design

## Context

见 `proposal.md` 的 Why。当前阅读页通过 `body` 上的 `source-mode` / `split-mode` 类区分三种阅读模式，打印样式集中在 `reader.css` 末尾的 `@media print` 中。双视图使用固定高度和内部滚动容器（`#doc-content`、`.doc-section`、`.source-view`、`.doc-rendered`），因此打印时需要先拆掉这些布局。源码视图固定自动换行，不提供关闭换行或横向滚动的模式。

## Goals / Non-Goals

**Goals:**

- 打印内容与当前阅读模式一致，双视图只输出渲染正文
- 打印时消除代码块和表格的横向滚动，内容不被裁切
- 打印页面底色为纯白，同时保留代码块、表格等内容背景
- 源码模式打印时隐藏工具栏
- 双视图下隐藏无效的宽度模式按钮

**Non-Goals:**

- 不改变屏幕上的代码块和表格行为
- 不新增打印预览页或服务端渲染
- 不改变阅读模式的切换、滚动同步和文档导航逻辑
- 不提供关闭源码自动换行的设置，也不支持源码横向滚动

## Decisions

### 1. 打印分派使用纯 CSS，而不是 `beforeprint` 脚本

保留 `body.split-mode` / `body.source-mode` 作为唯一状态来源，在 `@media print` 中用相同或更高特异性的选择器覆盖布局：

```text
body.split-mode #doc-content       -> height:auto; overflow:visible
body.split-mode .doc-section       -> display:block; height:auto
body.split-mode .source-view       -> display:none
body.split-mode .doc-rendered      -> overflow:visible; padding:0
```

这样打印时不需要修改 DOM 或类名，也不会出现进入打印前的一帧闪烁，退出打印后无需恢复任何状态。替代方案是用 `beforeprint` 临时切换类名，但需要处理恢复时机，且会污染运行时状态。

### 2. 打印时强制代码块换行

在 `@media print` 中同时覆盖容器和内部代码：

```text
.markdown-body pre,
.markdown-body pre > code
  white-space: pre-wrap
  overflow: visible
  overflow-wrap: anywhere
```

只改 `pre` 不够，因为 GitHub Markdown 样式给 `pre > code` 设置了 `white-space: pre`。

### 3. 打印时让表格适配页面宽度

把 GitHub Markdown 的滚动表格布局临时改为普通表格：

```text
.markdown-body table
  display: table
  width: 100%
  overflow: visible

.markdown-body table th,
.markdown-body table td
  white-space: normal
  overflow-wrap: anywhere
```

`overflow: visible` 只隐藏滚动条还不够，`width: max-content` 仍可能让表格超出纸张；因此同时改为 `width: 100%` 并允许换行。表头原本的单行规则需要显式覆盖。

### 4. 双视图隐藏宽度按钮

宽度按钮只影响 `body.full-width`，而双视图已经取消了 920px 阅读宽度上限，按钮没有可见效果。直接在 CSS 中隐藏：

```text
body.split-mode #width-toggle { display: none; }
```

不修改 `body.full-width` 类和用户偏好，因此切回阅读模式后状态仍然保留。

### 5. 打印时只把页面底色改为纯白

打印只替换页面/画布底色，不抹平内容块自己的背景：

- 应用层变量（`--bg`、`--fg` 等）改为白底深色文本
- `.markdown-body` 的 `--bgColor-default` 改为纯白，作为页面内容底色
- `--bgColor-muted` 和 `--bgColor-neutral-muted` 保留浅色主题的值（`#f6f8fa`、`#818b981f`），让代码块、表格和行内代码保留可见的背景
- 不强制 `.chroma` 背景透明

这样 PDF 的白色页边距与页面底色一致，同时代码块和表格仍有内容背景，彩色 PDF 阅读不受影响。深色模式下 Chroma 的深色 token 颜色在白底上对比度不足，因此在打印且系统为深色时把 Chroma token 降级为深色文本；浅色模式下保留原有语法高亮颜色。黑白打印用户可以在浏览器打印面板关闭背景图形，让内容背景自然变白。

### 6. 源码工具栏不参与打印

源码工具栏包含文件名和“复制源码”按钮，属于屏幕操作控件。打印时直接隐藏整个 `.source-toolbar`，只保留源码正文。双视图打印本来就隐藏源码栏，因此该规则主要作用于单独源码模式。

### 7. 源码视图固定自动换行

源码视图继续使用 `.source-code` 的 `white-space: pre-wrap`、`overflow-wrap: anywhere` 和 `word-break: break-all`，即使单个超长单词也会在阅读宽度内强制断开。不提供设置开关，也不启用横向滚动。双视图左侧源码栏保持滚动条隐藏。移除已添加的设置 UI、localStorage 键和 `source-no-wrap` 状态。

## Risks / Trade-offs

- [超宽表格中不可断行的图片或固定宽度元素仍可能超出纸张] → 对表头和单元格启用 `overflow-wrap: anywhere`；在验收用例中覆盖长文本和宽表格。
- [打印覆盖规则与双视图屏幕规则特异性冲突] → 在 `@media print` 中使用与屏幕规则相同或更高的选择器，并置于样式表末尾。
- [深色模式打印时语法高亮会降级为单色] → 白底可读性优先；浅色模式打印仍保留原有 Chroma 配色。

## Migration Plan

无数据迁移。旧版本可能写入的 `openmd-source-wrap` 键不再被读取，保留在 localStorage 中也不影响任何行为；如需彻底清理，可由后续维护操作删除。
