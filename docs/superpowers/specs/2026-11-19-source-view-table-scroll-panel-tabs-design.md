# OpenMD 源码视图、宽表格与信息面板 Tab 设计

## 背景

OpenMD 目前把 Markdown 渲染成阅读页面，用户只能看到渲染结果。实际使用中，用户还需要像用记事本打开 `.md` 文件一样查看原始内容，并把源码复制到别处。只提供渲染视图无法满足这个需求。

同时，宽表格的渲染有问题：`github-markdown-css` 给表格加了 `display: block; overflow: auto`，表格本身可以横向滚动，但自动列宽会先把列压窄，表头文字被挤成多行，滚动条出现时内容已经变形。

顶栏右侧的“信息与设置”面板目前把所有内容放在一条很长的滚动区域里：文件路径、配色、关于 OpenMD。随着设置项增加，这个面板会越来越长，入口和内容都需要重新组织。

## 目标

- 顶栏增加“查看源码”切换：当前文档在渲染视图和源码视图之间切换，源码只读、带 Markdown 语法高亮。
- 单文件、目录、链接文档三种模式都支持源码视图；切换文档时保持当前视图模式。
- 源码视图自动换行，不出现横向滚动。
- 源码视图提供“复制源码”，复制当前文档的完整原始 Markdown，包括 front matter。
- 源码区域支持局部选中复制；焦点在源码内时，`Ctrl/Cmd + A` 只选中源码，而不是整个页面。
- 宽表格的表头不换行，表格整体横向滚动。
- 信息面板改为三个 Tab：文件、设置、关于；标题栏和 Tab 栏固定，内容区独立滚动。
- 面板文案同步调整，为未来增加的设置项留出稳定的分组位置。

## 非目标

- 不做源码编辑、保存、搜索、行号或差异对比。
- 不新开源码窗口或独立标签页，不做左右分栏阅读。
- 不改变 Markdown 渲染规则、目录扫描、链接打开方式和现有四套配色。
- 不把“复制源码”做成逐字节还原：复制结果是当前浏览器中显示的源码文本，换行统一为 LF。
- 不持久化源码视图状态；刷新页面后回到渲染视图。
- 不新增顶级 Tab 来承载单个设置项；新增设置进入“设置”Tab 内的分组。

## 方案

### 源码视图：顶栏切换（采用）

顶栏增加一个 `</>` 图标按钮，切换当前页面的源码模式。源码模式是整页状态，不是单篇文档的属性；切换文档时继续显示新文档的源码。

未采用：

- **独立源码页**：需要再生成一套页面，切换时丢失阅读页的侧栏和大纲，和“边读边看源码”的用途不符。
- **左右分栏**：阅读区和源码各占一半，长文档两边都变窄，移动端不可用，超出当前阅读器的克制范围。

### 源码高亮：服务端 Chroma（采用）

复用已有的 Chroma 依赖和 `github` / `github-dark` 样式，在 Go 生成页面时用 Markdown 词法分析器高亮源码。现有代码块高亮已经在使用同一套 CSS 和明暗切换逻辑，源码视图可以直接复用。

未采用：

- **浏览器端高亮**：需要另写一套 Markdown 词法分析，质量和维护成本都不如 Chroma。
- **只高亮当前文档**：页面是静态 HTML，没有服务端；用户切换文档后无法再请求新的高亮结果，除非另开页面。

高亮在生成页面时对所有已加载文档执行一次。目录模式最多加载 25 MB Markdown，页面体积会增加，这是本方案的主要代价，见“风险与取舍”。

### 复制源码：从高亮后的源码节点读取（采用）

源码以 `<pre>` 的形式呈现，复制时读取该节点的 `textContent`。软换行只影响显示，不会在文本里插入换行，所以复制结果保持原始段落和代码行结构。

未采用：

- **额外嵌入一份原始文本或 Base64**：可以逐字节保留 CRLF，但会让页面再增加约一份源码的体积，目录模式代价更高。这里优先控制页面体积。
- **只支持鼠标选中后 `Ctrl+C`**：新用户找不到“复制全部”的入口，不如同时提供显式按钮。

复制结果说明：

- 包含 front matter、缩进、空行和 `git` 风格的换行结构。
- 换行符为 LF（Chroma 和浏览器文本节点都会把 CRLF 归一化为 LF），粘贴到编辑器或聊天工具不影响使用。
- 末尾是否有换行取决于原文件；高亮器可能补一个末尾换行，不做额外修正。

### 宽表格：表头不换行（采用）

保留现有表格的 `display: block; overflow: auto` 横向滚动机制，只增加：

```css
.markdown-body table th {
  white-space: nowrap;
}
```

表头不再被自动列宽压缩，整张表格按内容宽度撑开并在容器内横向滚动。正文单元格仍然自动换行，避免表格无限变宽。

未采用：

- **给所有单元格设置固定 `min-width`**：会把数字列、短文本列也撑宽，滚动距离变长，收益不如只处理表头。
- **额外包一层 `overflow-x` 容器**：现有表格自身已经可以滚动，加 wrapper 只会增加模板和样式复杂度。

### 信息面板：面板内 Tab（采用）

保留顶栏的信息入口和浮动面板，把内容分成三个同级 Tab：

```text
┌ OpenMD ─────────────── × ┐
│  文件  │  设置  │  关于   │
├──────────────────────────┤
│                          │
│  当前文档                │
│  C:\docs\a.md            │
│  所在目录                │
│  C:\docs                 │
│  [复制路径]  已复制       │
│                          │
└──────────────────────────┘
```

- **文件**：当前文档路径、所在目录、复制路径。
- **设置**：配色，以及未来新增的阅读设置；设置项按分组向下排列。
- **关于**：阅读器简介、作者、版本、构建时间、发布页。

未采用：

- **手风琴折叠**：所有分组仍在一列里，展开状态会让面板高度跳动，且用户要逐个展开才能找到设置。
- **左侧竖向 Tab 栏**：在 360px 宽的面板里会挤占内容宽度，三个中文标签横向排列已经足够。

### 面板视觉方向

OpenMD 的既有语言是安静的海盐蓝阅读界面、四套跟随明暗的配色、圆角浮动控件和等宽路径字体。面板重设计沿用现有 token，不引入新配色，也不使用卡片堆叠。

- **Tab 栏**：等宽三项，底部 2px 选中线，和侧栏的“文件列表 / 大纲”Tab 使用同一套语言。没有胶囊分段控件、没有渐变、没有全大写标签。
- **层级**：面板标题为品牌名 `OpenMD`，Tab 承担分类；Tab 内容不再重复“文件”“设置”“关于”的标题。文件路径用等宽字体，设置项用正文系统字体。
- **结构**：面板头部和 Tab 栏固定，内容区独立滚动；面板高度受限时不会把 Tab 推出视野。
- **焦点**：键盘焦点可见，Tab 支持左右方向键切换；复制结果和复制按钮保持明确的 `aria-live` 反馈。
- **克制**：源码视图是这次唯一“有存在感”的新表面；面板保持低对比、轻边框、无阴影叠加。

## 设计细节

### 数据与渲染

`Document` 增加：

```go
SourceHTML template.HTML // Chroma 生成的 token HTML，不含外层 <pre>
HasSource  bool          // 文件是否读取成功，源码是否可用
```

`HasSource` 表示“文件读取成功”，不表示“内容非空”。零字节或纯空白文件也算有源码，源码区显示空的 `<pre>`；点击复制时提示“没有可复制的源码”。只有读取失败的文档才是 `false`。

`buildPageFromSources` 在渲染 Markdown 后，对原始 `sourceDoc.Source` 生成源码 HTML。原始字节在 `RenderMarkdown` 之前保留，因此源码视图会显示 front matter，而正文仍然剥离 front matter。

新增 `highlightSource(source []byte) (template.HTML, error)`：

- 使用 `lexers.Get("markdown")`，找不到 lexer 时回退为 HTML 转义后的纯文本。
- 使用 `chroma.Coalesce` 合并相邻 token，减少输出体积。
- 使用 `chromahtml.New(WithClasses(true), WithModeClasses(true), PreventSurroundingPre(true))`，只输出 token span，由模板提供外层 `<pre>`。
- 使用 `styles.Get("github")`，和现有代码块保持一致。
- 高亮失败时回退为 `html.EscapeString` 的纯文本，源码视图仍然可用。

读取失败的文档保留现有错误视图，`HasSource` 为 `false`；源码模式下显示“该文档没有可显示的源码。”，不会出现空白页。

### 模板结构

顶栏在“全宽”和“信息”之间增加源码切换按钮：

```html
<button type="button" id="source-toggle" class="icon-button"
        aria-label="查看源码" title="查看源码" aria-pressed="false">
  <!-- 左右尖括号 SVG -->
</button>
```

每个文档 section 增加 `data-has-source`，把渲染内容包进 `.doc-rendered`，并增加源码块：

```html
<section class="doc-section markdown-body" data-has-source="{{$doc.HasSource}}" ...>
  <div class="doc-rendered">{{$doc.HTML}}</div>
  <div class="source-view">
    <div class="source-toolbar">
      <span class="source-filename" title="{{$doc.RelPath}}">{{$doc.RelPath}}</span>
      {{if $doc.HasSource}}
      <button type="button" class="source-copy panel-action">复制源码</button>
      {{end}}
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

对应的显示规则：

```css
.source-view { display: none; }
body.source-mode .doc-section .doc-rendered { display: none; }
body.source-mode .doc-section .source-view { display: block; }
```

信息面板改为：

```html
<div class="settings-header">
  <span class="settings-title">OpenMD</span>
  <button id="settings-close" ...>×</button>
</div>
<div class="settings-tabs" role="tablist" aria-label="信息与设置">
  <button id="settings-tab-file" role="tab" data-settings-tab="file"
          aria-selected="true" aria-controls="settings-panel-file">文件</button>
  <button id="settings-tab-settings" role="tab" data-settings-tab="settings"
          aria-selected="false" aria-controls="settings-panel-settings">设置</button>
  <button id="settings-tab-about" role="tab" data-settings-tab="about"
          aria-selected="false" aria-controls="settings-panel-about">关于</button>
</div>
<div class="settings-body">
  <section id="settings-panel-file" role="tabpanel" aria-labelledby="settings-tab-file">...</section>
  <section id="settings-panel-settings" role="tabpanel" aria-labelledby="settings-tab-settings" hidden>...</section>
  <section id="settings-panel-about" role="tabpanel" aria-labelledby="settings-tab-about" hidden>...</section>
</div>
```

### 状态与交互

源码模式是 `body.source-mode` 类：

- 顶栏按钮切换类，更新 `aria-pressed`，标题在“查看源码 / 返回阅读”之间切换。
- 源码模式下所有 section 都隐藏 `.doc-rendered`、显示 `.source-view`；`HasSource=false` 的文档在源码区显示“该文档没有可显示的源码。”，并隐藏复制按钮。读取失败的渲染错误视图只在渲染模式下可见。
- 切换文档时模式保持不变；页面刷新时回到渲染视图，不写 localStorage。
- 切换源码模式时保留当前滚动位置：切换前记录 `window.scrollY`，切换后按新内容高度截断并恢复。不额外记录每篇文档、每种视图的历史位置，行为保持可预测。
- 点击大纲链接时，如果当前处于源码模式，先退出源码模式，再让浏览器滚动到对应标题。
- `applyChromaTheme()` 继续负责给 `.chroma` 元素加 `light` / `dark` 类，源码块和代码块共用。

复制源码：

- 读取当前 section 的 `.source-code` 的 `textContent`；为空时提示“没有可复制的源码”。
- 优先 `navigator.clipboard.writeText`，失败回退到临时 `textarea` + `document.execCommand("copy")`。
- 按钮旁边的 `role="status"` 显示“已复制 / 复制失败”，2 秒后清空；不改变按钮文字，避免按钮宽度跳动。

源码内 `Ctrl/Cmd + A`：

- 仅当事件目标位于 `.source-code` 内时拦截。
- 用 `Range.selectNodeContents` 选中整个 `<pre>`；其他位置保持浏览器默认全选行为。

信息面板 Tab：

- 默认打开“文件”；同一次页面会话内关闭再打开，保留上次停留的 Tab，不写入 localStorage。
- 点击切换；左右方向键在三个 Tab 间移动并切换，Home/End 跳到首尾。
- 隐藏非活动 `role="tabpanel"`，同步 `aria-selected` 和 roving `tabindex`。
- 面板头部和 Tab 栏不滚动，`.settings-body` 独立滚动。
- 配色切换、复制路径、发布页链接行为保持不变。

### 文案调整

- 面板标题：`信息与设置` → `OpenMD`。
- Tab：`文件`、`设置`、`关于`。
- 文件区去掉重复的“文件”分组标题，保留“当前文档”“所在目录”“复制路径”。
- 设置区保留“配色”分组标题，未来设置项作为新的 `.settings-group` 加在同一 Tab 内。
- 关于区增加一句产品说明：“只读的 Windows Markdown 阅读器。”
- 关于区保留作者、版本、构建时间；发布提示改为“网盘里的 OpenMD.exe 比构建时间新，就说明有更新。”，链接仍为“打开发布页”。
- 源码按钮标题：`查看源码` / `返回阅读`；复制按钮：`复制源码`。

### 样式

源码视图：

- `.source-view`：`--bg-subtle` 背景、1px `--border` 边框、12px 圆角。
- `.source-toolbar`：`position: sticky; top: var(--chrome-height)`，显示文件名和复制按钮；文件名等宽、超长省略。
- `.source-code`：13px 等宽字体、1.7 行高、`white-space: pre-wrap; overflow-wrap: anywhere;`、`tab-size: 4`；覆盖 `markdown-body pre` 的外边距和背景，避免横向滚动。
- 复制状态使用 `--fg-muted`，不新增警告色。

面板：

- 面板宽度从 340px 调到 360px，仍受 `calc(100vw - 32px)` 限制。
- 面板改为纵向 flex，头部和 Tab 栏固定，`.settings-body` 滚动。
- Tab 栏横向等宽三项，底部选中线和侧栏 Tab 一致；hover 使用 `--bg-hover`。
- `.settings-group:first-child` 去掉顶部外边距；“关于”Tab 不再使用独立的分隔边框。
- `focus-visible` 覆盖 Tab、源码切换按钮、复制按钮。

表格：

- 只在 `reader.css` 增加表头 `white-space: nowrap`；不改 vendor CSS。
- 现有 `display: block; overflow: auto` 继续承担横向滚动。

### 测试与验证

Go 测试：

- `highlightSource` 对普通 Markdown 生成非空 token HTML；对空源码返回空 HTML；失败路径回退为转义文本。
- `buildPageFromSources` 为每篇成功读取的文档设置 `HasSource` 和 `SourceHTML`，零字节文件同样为 `true`；源码内容包含 front matter，渲染内容仍然不包含；读取失败的文档为 `false`。
- `RenderPage` 输出源码切换按钮、`.source-view`、`.source-code`、复制源码按钮和三个 Tab 的 ARIA 结构。
- 更新现有面板测试：标题从“信息与设置”改为“OpenMD”，仍保留“信息与设置”的 `aria-label`。
- CSS 回归：`reader.css` 包含表头 `white-space: nowrap` 规则。

浏览器验证：

- 用宽表格 fixture，在 760px 和桌面宽度下截图：表头单行、表格可横向滚动、正文单元格正常换行。
- 打开源码视图：高亮正常、长行自动换行、front matter 可见、复制按钮和 `Ctrl+A` 可用。
- 在长文档中部切换源码模式：滚动位置近似保留；切回后仍落在相近位置。
- 读取失败的文档在源码模式下显示空状态而不是渲染错误页；零字节文档显示空源码区。
- 打开信息面板：三个 Tab 切换正常、内容区滚动、配色切换正常、复制路径正常。
- 明暗模式和四套配色下检查源码块和信息面板的对比度。

## 评审后确定的小决定

以下决定由实现者直接采用，不需要再确认；如果和你的预期不同，告诉我就改：

- **无源码文档**：源码模式是全局视图，所有 section 都隐藏渲染内容并显示源码区。读取失败的文档显示“该文档没有可显示的源码。”，同时隐藏复制按钮；渲染错误页只在渲染模式下出现。
- **切换时的滚动位置**：保留切换前的数值滚动位置，并按新内容高度截断；不记录每篇文档在两种视图下的历史位置。
- **空文件语义**：`HasSource` 表示读取成功，零字节和纯空白文件也算有源码；源码区为空，复制时提示“没有可复制的源码”。

## 风险与取舍

- **页面体积增加**：源码视图需要把每篇文档的高亮 HTML 一起写入页面。目录模式上限 25 MB，源码高亮和现有渲染 HTML 叠加后页面会明显变大；这是顶栏同页切换的直接代价。若实测影响打开速度，后续可以给超大文档降级为纯文本源码，或只对当前文档生成源码。
- **复制不是逐字节一致**：CRLF 会变成 LF，末尾换行可能被补齐。对阅读、粘贴、二次编辑足够；若未来需要严格还原文件字节，需要另存原始源码。
- **Chroma 高亮成本**：所有文档在生成页面时都要额外走一次 Markdown 词法分析。现有渲染已经处理同一批字节，增加的 CPU 在可接受范围内；超大目录需要实测。
- **源码模式与大纲**：源码模式下标题元素被隐藏，点击大纲会退出源码模式。这是有意的取舍，避免大纲跳到不可见的位置。
- **面板标题改为品牌名**：`OpenMD` 比“信息与设置”更短，也更适合未来增加 Tab；但用户可能仍称它为“信息面板”，入口按钮的 `aria-label` 保持“信息与设置”以保留可理解性。
- **移动端顶栏**：增加一个图标按钮后，窄屏标题空间更紧；标题已有省略号，三个操作按钮仍可容纳。

## 验收

- 顶栏源码按钮可在渲染和源码之间切换，单文件、目录、链接文档都适用。
- 源码带 Markdown 高亮、自动换行、只读；front matter 可见。
- “复制源码”复制当前文档完整源码；源码内 `Ctrl/Cmd+A` 只选中源码。
- 切换文档后源码模式保持；刷新页面回到渲染视图。
- 宽表格表头不折行，表格可横向滚动，正文不溢出页面。
- 信息面板显示“文件 / 设置 / 关于”三个 Tab；默认文件，会话内记住上次 Tab。
- 面板标题、Tab 和文案按本文更新；文件、设置、关于内容完整可用。
- 现有配色、复制路径、发布页、目录/链接切换行为不回归。
- `go test ./...` 通过，`build.bat` 构建成功。
