# Design

## Context

当前渲染链路在 `internal/render/document.go` 中先解析 Markdown，再遍历 AST 重写链接和图片，最后交给 goldmark 渲染。标题 ID 目前依赖 `parser.WithAutoHeadingID()`，脚注使用 goldmark 默认的全局 `fn:` / `fnref:` ID；目录扫描在 `internal/scan/scan.go` 中按文件顺序累加大小。现有单测覆盖主路径，但没有覆盖锚点、脚注、front matter、特殊文件名和超大文件。详见 `proposal.md` 和本变更的两份 spec。

## Goals / Non-Goals

**Goals:**

- 在现有“单 exe + 临时 HTML + 浏览器”架构内修复第一批正确性问题，不引入 HTTP 服务或常驻进程。
- 保持英文标题的现有锚点尽量兼容，同时为中文和带标点标题生成稳定 ID。
- 让异常输入只影响局部，不把整篇文档变成错误页。
- 让目录扫描在面对超大文件时继续工作，并给出可理解的警告。

**Non-Goals:**

- 不处理远程/UNC 图片拦截、CSP nonce、`openmd://` 按需加载、临时文件保留策略、卸载 `UserChoice`。
- 不增加 GBK/UTF-16 编码识别；本变更只处理 UTF-8（含 BOM）。
- 不引入 YAML/TOML 解析依赖；front matter 只做边界识别和剥离，不解释字段。
- 不改变安装、注册表、浏览器启动和现有目录 500 篇 / 25 MB 的产品上限。

## Decisions

### 1. 标题 ID：解析后统一生成，而不是依赖 goldmark 默认值

在 AST walk 中，用已经提取出的纯文本 `textValue` 生成 slug，并写入 heading 的 `id` 属性；不再依赖 `parser.WithAutoHeadingID()` 的结果。

- slug 规则：Unicode 小写；保留字母、数字和中文；空格、`-`、`_` 折叠成单个 `-`；去掉其他标点；去掉首尾 `-`；空结果用 `section`。
- 同一文档内用 map 去重，重复标题依次加 `-2`、`-3`。
- 最终 ID 仍由 `IDPrefix` 加文档前缀，避免跨文档冲突。

理由：`nodeText` 已经能拿到标题的可见文本，能正确处理行内代码、强调和链接；goldmark 的默认实现会丢掉中文并返回 `heading`，而且对带链接的标题会产生 `linkxmd` 这类不稳定 ID。  
备选方案：实现自定义 `parser.IDs`。它只能拿到原始行文本，仍需自己解析行内语法，复杂度更高。

兼容性：英文标题继续得到 `hello-world` 这类 ID；中文标题从 `heading` 变为基于文本的 ID，属于有意的行为变化。

### 2. 锚点：统一改成 `#doc=<id>&anchor=<fragment>`

在 `RenderOptions` 中增加当前文档 ID，`RewriteURL` 对以 `#` 开头的目标不再原样返回，而是生成：

```text
#doc=doc-0&anchor=install
```

`reader.js` 现有的 `a[href^="#doc="]` 拦截器已经能处理这种链接，会切换文档并滚动到 `doc-0-install`。  
备选方案：直接把 `#install` 改成 `#doc-0-install`。这样能利用浏览器原生锚点，但同文档和跨文档会使用两套格式，且 TOC 链接与正文链接的行为会分叉；统一成 `#doc=` 更一致。

生成的大纲链接仍使用 `href="#<真实ID>"`，保持浏览器原生跳转，不受影响。

### 3. 脚注：使用 goldmark 自带的 ID 前缀能力

`extension.NewFootnote(extension.WithFootnoteIDPrefix(docID+"-"))` 会给脚注引用、脚注定义和回链统一加前缀，例如：

```text
doc-0-fnref:1
doc-0-fn:1
```

`RenderMarkdown` 已经为每篇文档创建独立的 `newMarkdown()`，把 `docID` 传进去即可，不需要字符串后处理，也不需要 fork goldmark 的 renderer。  
备选方案：渲染后对 HTML 做 `strings.ReplaceAll`。它更简单，但依赖脆弱的字符串匹配，未来 goldmark 输出变化会失效。

### 4. 本地路径：字面路径优先，URL 解码回退

新增一个共享的本地目标解析函数，供链接、图片和链接文档预加载共用。解析顺序：

1. 绝对路径（含 UNC）直接按文件系统路径处理。
2. 对相对目标，先拼接 `baseDir` + 原始字面路径；如果文件存在，使用字面路径。
3. 字面路径不存在时，再用 `url.Parse` 解析，使用解码后的 `Path`，并保留 query/fragment。
4. `url.Parse` 失败时，回退到字面路径并构造 `file://` URL，而不是返回错误。

关键约束：在第 2 步之前必须用 `isNetworkPath` 排除 UNC 和网络盘，避免为了判断“是否存在”而阻塞首屏。

这样可以同时满足：

- `100%.md`：字面文件存在，直接命中，不再触发非法 URL 转义。
- `foo%bar.md`：字面文件存在，直接命中，不会被解码成错误字节。
- `a%20b.md`：字面文件不存在时，回退到 URL 解码，命中 `a b.md`。
- 字面文件名带 `#`：文件存在时优先按字面路径处理；不存在时才把 `#` 当 fragment。

备选方案：始终按 URL 语义解析，再对失败目标做特殊处理。它无法正确处理字面 `%` 文件名，也无法消除 `foo%bar.md` 被误解码的问题。

### 5. 异常链接：局部降级，不中断文档

`RenderMarkdown` 的 AST walk 不再因为单个链接/图片解析失败返回 error。解析失败时：

- 链接：保留可读文本，目标置空或保留原始文本，不生成可执行的跳转。
- 图片：不生成 `src`，只保留 alt 文本。
- 其余节点继续渲染。

这需要把 `RewriteURL` 的错误语义从“返回 error”改成“返回可用的降级目标 + 可选诊断信息”。如果后续需要提示，可把诊断收集到 `Document.Warnings`，但本变更先不做 UI。

### 6. Front matter：轻量边界识别，不引入解析器

在 `RenderMarkdown` 的最前面调用 `stripFrontMatter`：

- 仅当第一行（允许 BOM 已剥离）是 `---` 或 `+++` 时开始。
- 在文件前部有限范围内查找同种结束行，避免扫描整个大文件。
- 块内至少出现一个 `key:`（YAML）或 `key =` / `[section]`（TOML）形状的行，降低把普通水平线误判为 front matter 的概率。
- 找到结束行则移除整个块；否则保留原文，按普通 Markdown 渲染。

不解析字段，不显示元数据。备选方案是引入 YAML 库做完整解析，但收益低、依赖和错误处理成本高。

### 7. 超大文件：单文件上限 + 目录逐项跳过

- 在 `render` 包定义单文件默认上限（32 MiB），`BuildFilePage` 先 `Stat` 再读取；超限返回带文件名和上限的错误。
- 目录扫描使用 `min(MaxFileBytes, MaxTotalBytes)` 作为单文件预算。单个文件超过预算时记录 warning 并 `continue`；累计总量达到 `MaxTotalBytes` 时才 `break`。
- `scan.Result` 增加 skipped 信息（或 warning 字符串），页面顶部沿用现有 `.warnings` 展示。
- 如果目录里所有 Markdown 都被跳过，返回新的“存在 Markdown 但都因大小限制被跳过”错误，区别于 `ErrNoMarkdown`。
- `buildDirPage` 读取文件时使用带 limit 的读取函数，防止 `Stat` 与 `ReadFile` 之间文件变大。

备选方案：只把 `break` 改成 `continue`。它不能阻止单个 30 MB 文件被读入，也不能给出“全部被跳过”的准确提示。

### 8. 测试策略

- `internal/render`：标题 slug、同文档/跨文档锚点、中文锚点、脚注多文档、front matter、非法 URL、字面 `%` 文件名、`a%20b.md` 回退、单文件超限。
- `internal/scan`：超大文件在正常文件之前/之后、全部超大、数量截断、总量截断、warning 内容。
- 断言尽量落在用户可见的 HTML/错误信息上，而不是内部函数细节。
- 保持现有测试不回归；新增测试使用小尺寸临时文件，不生成 32 MB 真实文件（通过可注入的 limit 或小常量测试）。

## Risks / Trade-offs

- **字面路径优先会改变 URL 语义**：如果同一目录同时存在 `a%20b.md` 和 `a b.md`，会优先打开字面文件。→ 在 README 的已知限制中说明；对普通用户而言，字面文件名更符合 Windows 直觉。
- **每个相对链接都要 `Stat`**：大量链接可能增加 IO。→ 在 `Stat` 前排除网络路径；同一页面的解析结果可以按目标缓存，典型文档的链接数量很小。
- **中文标题 ID 变化**：旧的外部链接如果写的是 `#heading` 会失效。→ 旧 ID 本来就不表达标题语义；在变更说明中标注这是有意的修正。
- **front matter 误判**：文档开头用 `---` 做水平线，后面又出现 `---`，可能被误剥离。→ 要求块内有 key/value 形状，并限制在前部；无法完全消除歧义，但能覆盖绝大多数真实文档。
- **32 MiB 单文件上限**：极少数合法大文档会被拒绝。→ 错误信息明确给出上限；后续如有需求可做成设置项。
- **脚注前缀破坏用户手写的 `#fn:1` 链接**：这类链接不是标准用法，且跨文档本来就是错的。→ 接受，必要时在 README 说明。

## Migration Plan

没有数据迁移。发布后旧临时 HTML 会在下次启动时按现有规则清理；用户重新打开文档即可得到新行为。若需回滚，直接回退本变更的代码提交，不需要修改注册表或安装目录。
