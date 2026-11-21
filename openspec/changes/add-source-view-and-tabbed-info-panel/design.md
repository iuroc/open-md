# Design

## Context

OpenMD 是一个 Go 程序：读取 Markdown，用 goldmark 渲染成 HTML，把 HTML 写入临时文件并用默认浏览器打开，进程随后退出。页面没有 HTTP 服务，浏览器运行期间无法再向 Go 请求新的渲染结果。因此所有源码视图内容必须在生成页面时一起写入。

现有相关实现：

- `internal/render/document.go`：用 goldmark + Chroma 渲染正文，代码块已经使用 Chroma 的 `github` / `github-dark` 样式和明暗类切换。
- `internal/render/links.go`：`buildPageFromSources` 统一加载、渲染目录文档和链接文档；原始字节在渲染前仍然可用。
- `internal/render/web/page.html`：每个文档是一个 `.doc-section`，首屏渲染活动文档，其他文档 `hidden`，由 `reader.js` 切换。
- `internal/render/web/reader.css`：主题 token、顶栏、侧栏 Tab、浮动信息面板；表格样式来自 `github-markdown-css`，表格本身已经有 `display: block; overflow: auto`。
- `internal/render/web/reader.js`：文档切换、侧栏 Tab、信息面板、复制路径和明暗主题切换。

完整的交互与视觉设计见 `docs/superpowers/specs/2026-11-19-source-view-table-scroll-panel-tabs-design.md`；本文记录实现层面的技术决定。

## Goals / Non-Goals

**Goals:**

- 在同一个静态页面内切换渲染视图和源码视图，不新增页面、服务或依赖。
- 源码视图复用现有 Chroma 样式和明暗切换，保持和代码块一致的视觉语言。
- 源码复制直接复用现有剪贴板写入与回退逻辑。
- 信息面板的 Tab 结构与未来设置项增长兼容。
- 宽表格修复只改 CSS，不改变 Markdown 解析结果。

**Non-Goals:**

- 不做源码编辑、保存、搜索、行号、差异对比。
- 不新增源码独立页、左右分栏或 iframe。
- 不把复制结果做成逐字节还原（CRLF 会归一化为 LF）。
- 不持久化源码模式或信息面板 Tab 到 localStorage。
- 不修改目录扫描、文件大小限制、链接解析和配色 token。

## Decisions

### 源码视图是整页模式，不是单篇文档属性

采用 `body.source-mode` 类控制整页视图。每个 section 同时包含 `.doc-rendered` 和 `.source-view`，CSS 决定显示哪一个：

```css
.source-view { display: none; }
body.source-mode .doc-section .doc-rendered { display: none; }
body.source-mode .doc-section .source-view { display: block; }
```

切换文档时 `showDoc` 逻辑不变，源码模式自然延续到新文档；刷新页面不写 localStorage，回到渲染视图。

备选方案是每篇文档单独记录源码状态，但当前页面同一时间只显示一篇文档，按整页建模更简单，也避免切换文档时出现混合状态。

### 源码在 Go 端用 Chroma 预先生成

`Document` 增加：

```go
SourceHTML template.HTML // 不含外层 <pre> 的 token HTML
HasSource  bool          // 文件读取成功即为 true
```

`buildPageFromSources` 在 `RenderMarkdown` 之后调用 `highlightSource(sourceDoc.Source)`，原始字节未经过 `stripFrontMatter`，所以源码视图包含 front matter。

`highlightSource` 的关键配置：

```go
formatter := chromahtml.New(
    chromahtml.WithClasses(true),
    chromahtml.WithModeClasses(true),
    chromahtml.PreventSurroundingPre(true),
)
```

- `PreventSurroundingPre(true)` 让 Chroma 只输出 token span，模板提供外层 `<pre class="chroma light source-code">`，避免嵌套 `<pre>`。
- `WithModeClasses(true)` 保持和现有代码块相同的 `.chroma.light` / `.chroma.dark` 规则，`applyChromaTheme()` 继续负责运行时切换。
- 使用 `lexers.Get("markdown")` 和 `chroma.Coalesce`，减少 token 数量和输出体积。
- 高亮失败时回退为 `html.EscapeString` 的纯文本，源码视图仍然可用。

备选方案是浏览器端高亮，但需要另写 Markdown 词法分析，且很难和现有 Chroma 样式保持一致。

### 源码块外层由模板提供

每个 section 模板从：

```html
<section class="doc-section markdown-body" data-...>
  <template class="toc-template">...</template>
  {{$doc.HTML}}
</section>
```

改为：

```html
<section class="doc-section markdown-body" data-has-source="{{$doc.HasSource}}" data-...>
  <template class="toc-template">...</template>
  <div class="doc-rendered">{{$doc.HTML}}</div>
  <div class="source-view">
    <div class="source-toolbar">
      <span class="source-filename">{{$doc.RelPath}}</span>
      {{if $doc.HasSource}}<button class="source-copy panel-action">复制源码</button>{{end}}
      <span class="source-copy-status" role="status" aria-live="polite"></span>
    </div>
    {{if $doc.HasSource}}
    <pre class="chroma light source-code" tabindex="0">{{$doc.SourceHTML}}</pre>
    {{else}}
    <p class="source-empty">该文档没有可显示的源码。</p>
    {{end}}
  </div>
</section>
```

`.doc-rendered` 只是显示开关，不改变 `markdown-body` 作用域；现有标题锚点、脚注和链接样式仍由 section 上的 `markdown-body` 控制。

### 源码视图与正文同列排版

源码视图不额外渲染卡片。`.source-toolbar` 只保留无边框的一行元信息，`.source-code` 去掉边框、圆角、背景和内边距，直接落在 `.doc-section` 的阅读列里；字号和行高与正文一致（16px / 1.5），字体保持等宽，Chroma 语法高亮保留。

### 复制从源码 DOM 读取，不额外嵌入原始文本

复制按钮读取当前 section 中 `.source-code` 的 `textContent`。软件换行只影响显示，不会在文本节点中插入换行，所以复制结果保持原始段落和代码行结构。

- 优先 `navigator.clipboard.writeText`，失败回退到现有的临时 `textarea` + `execCommand`。
- 复制结果包含 front matter、缩进和空行；换行被 Chroma/浏览器归一化为 LF。
- 空内容不写入剪贴板，提示“没有可复制的源码”。

备选方案是再嵌入一份原始文本或 Base64 以保留 CRLF，但会让目录模式的页面体积再增加约一份源码，收益不足以抵消成本。

### 源码内全选由键盘事件限定范围

`.source-code` 增加 `tabindex="0"`。全局 `keydown` 监听只在事件目标位于 `.source-code` 内时处理 `Ctrl/Cmd + A`，用 `Range.selectNodeContents` 选中整个 `<pre>` 并阻止默认行为；焦点在源码外时保持浏览器默认全选。

### 切换视图保留数值滚动位置

切换前记录 `window.scrollY`，切换后按新内容高度截断并恢复：

```js
const y = window.scrollY;
document.body.classList.toggle("source-mode");
syncSourceButton();
const max = document.documentElement.scrollHeight - window.innerHeight;
window.scrollTo(0, Math.min(y, Math.max(0, max)));
```

不记录每篇文档、每种视图的历史位置，避免状态膨胀；切换文档时仍沿用现有 `showDoc` 的滚动逻辑。

### 大纲点击先退出源码模式

`#toc` 的点击监听只做一件事：如果当前是源码模式，移除 `body.source-mode` 并同步按钮状态。随后浏览器执行默认的锚点跳转；标题上的 `scroll-margin-top` 继续提供顶栏偏移。

### 信息面板改为固定头部 + 固定 Tab 栏 + 独立滚动内容区

面板结构改为：

```html
<div id="settings-panel" class="settings-panel" role="dialog">
  <div class="settings-header">OpenMD + 关闭按钮</div>
  <div class="settings-tabs" role="tablist">三个 role="tab" 按钮</div>
  <div class="settings-body">
    <section role="tabpanel">...</section>
    <section role="tabpanel" hidden>...</section>
    <section role="tabpanel" hidden>...</section>
  </div>
</div>
```

CSS 关键点：

```css
.settings-panel {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 80px);
  padding: 0;
  overflow: hidden;
}
.settings-body {
  min-height: 0;
  padding: 16px;
  overflow: auto;
}
```

`reader.js` 增加 `activateSettingsTab(name)`，同步 `hidden`、`aria-selected` 和 roving `tabindex`；左右方向键、Home/End 移动并激活。Tab 状态只保存在 DOM 中，关闭再打开保留，刷新后回到默认的“文件”。

### 宽表格只增加表头不折行

保留 `github-markdown-css` 已有的 `display: block; overflow: auto`，在 `reader.css` 增加：

```css
.markdown-body table th {
  white-space: nowrap;
}
```

表头的最小内容宽度恢复为完整标题，自动布局把表格撑到自然宽度，超过阅读区时由表格自身滚动。正文单元格仍按列宽换行。

备选方案是给所有单元格设置 `min-width` 或增加 wrapper；前者会让短列也无谓变宽，后者在现有滚动机制之上重复一层容器。

### 测试策略

Go 测试覆盖数据与模板：

- `highlightSource` 输出非空 token HTML；空输入返回空 HTML；高亮失败回退为转义文本。
- `buildPageFromSources` 为成功读取的文档设置 `HasSource=true` 和 `SourceHTML`，零字节文件也为 `true`；读取失败为 `false`。
- `RenderPage` 包含源码切换按钮、`.source-code`、复制按钮、三个 Tab 的 ARIA 结构和表头样式。
- 更新现有面板测试的标题断言。

浏览器验证覆盖布局和交互：

- 宽表格 fixture 在 760px 和桌面宽度下截图，确认表头单行、可横向滚动。
- 源码视图截图和交互检查：高亮、软换行、front matter、复制、`Ctrl+A`、滚动位置、大纲跳转。
- 面板截图和交互检查：三个 Tab、固定头部、内容滚动、键盘导航、配色切换。

## Risks / Trade-offs

- **页面体积增加**：所有已加载文档的源码高亮 HTML 都会写入页面。目录模式上限 25 MB，页面会明显变大 → 先按当前限制实现并实测打开速度；若不可接受，后续给超大文档降级为纯文本源码或改为按需生成。
- **复制不是逐字节一致**：CRLF 归一化为 LF，末尾换行可能被补齐 → 对阅读和粘贴足够；若未来需要严格还原，再增加原始源码嵌入。
- **Chroma 额外一遍词法分析**：生成页面时对所有文档多跑一次 Markdown 词法分析 → 与现有 goldmark 渲染同量级，目录模式需实测；必要时只对超过阈值的文档降级。
- **源码模式与大纲**：源码中标题元素不可见，点击大纲会退出源码模式 → 这是有意行为，避免跳到不可见位置。
- **面板 Tab 状态不持久化**：刷新后回到“文件” → 符合“信息面板优先”的定位；如果用户更需要记住设置页，可以后续改成 localStorage。
- **移动端顶栏**：多一个图标按钮会压缩标题空间 → 标题已有省略号，操作按钮仍可容纳；窄屏可通过截图验证。

## Migration Plan

无数据迁移和兼容性变更。实现完成后按项目约定先提交，再执行 `build.bat` 生成 `OpenMD.exe` 和 `source.git.bundle`。回滚只需还原本次提交。
