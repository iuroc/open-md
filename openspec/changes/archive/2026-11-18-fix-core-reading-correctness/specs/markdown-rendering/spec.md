# Spec Delta

## Purpose

定义 Markdown 到阅读页的渲染契约，确保标题锚点、脚注、front matter 和本地链接在多文档页面中保持可预测、可导航，并且单个异常输入不会破坏整篇文档的可读性。

## ADDED Requirements

### Requirement: Heading anchors remain navigable across documents

系统 SHALL 为每个标题生成文档内唯一、可重复生成的 ID，并把页内锚点和已预渲染文档的跨文档锚点解析到真实标题。点击锚点后 MUST 切换到目标文档并滚动到对应标题。

#### Scenario: Same-document anchor

- **WHEN** 文档包含 `[安装](#install)` 和标题 `## Install`
- **THEN** 点击链接后仍停留在当前文档，并滚动到 `Install` 标题，而不是跳转到不存在的 `#install`

#### Scenario: Cross-document English anchor

- **WHEN** 文档包含 `[用法](other.md#usage)`，且 `other.md` 已被预渲染
- **THEN** 点击链接后切换到 `other.md` 的阅读区，并滚动到 `usage` 标题

#### Scenario: Cross-document Chinese anchor

- **WHEN** 文档包含 `[说明](other.md#中文标题)`，且 `other.md` 的标题是 `# 中文标题`
- **THEN** 点击链接后切换到 `other.md`，并滚动到该标题

#### Scenario: Duplicate headings

- **WHEN** 同一文档包含两个文本相同的标题
- **THEN** 两个标题获得不同的 ID，页面中不存在重复的标题 ID

### Requirement: Footnotes are scoped to their document

系统 SHALL 把脚注引用、脚注定义和回链的 ID 限定在所属文档内，使同一页面中的多篇文档可以各自使用 `[^1]` 而不发生冲突。

#### Scenario: Two documents with footnote one

- **WHEN** 目录模式或链接文档页面同时渲染两篇文档，且各自包含 `[^1]`
- **THEN** 点击第二篇文档的脚注引用后，页面停留在第二篇文档并显示第二篇的脚注内容

#### Scenario: Footnote backlink

- **WHEN** 用户在第二篇文档中点击脚注回链
- **THEN** 页面回到第二篇文档中对应的脚注引用位置，而不是第一篇文档

### Requirement: Front matter is excluded from rendered content

系统 SHALL 在渲染前识别并移除文档开头由 `---` 或 `+++` 包围的 front matter 块。被移除的元数据 MUST NOT 出现在正文、文档标题或大纲中。

#### Scenario: YAML front matter

- **WHEN** 文档以 `---` 开始，包含 `title: Hi` 等字段，并以 `---` 结束，随后是 `# 正文标题`
- **THEN** 页面正文从 `# 正文标题` 开始，不显示 `title: Hi`，也不显示 front matter 产生的分隔线

#### Scenario: TOML front matter

- **WHEN** 文档以 `+++` 包围元数据块
- **THEN** 元数据被移除，正文和大纲不包含 front matter 字段

#### Scenario: Horizontal rule after content

- **WHEN** `---` 出现在正文中间，而不是文档开头
- **THEN** 它仍按普通 Markdown 水平线渲染

### Requirement: A broken local destination does not fail the document

系统 SHALL 保证单个链接或图片的目标解析失败不会使整篇文档渲染失败。正文其余部分 MUST 继续显示，解析失败的目标 MUST NOT 被当作可执行链接。

#### Scenario: Invalid percent sequence

- **WHEN** 文档包含 `[bad](100%.md)`，且该目标无法按 URL 解析
- **THEN** 文档其余内容正常渲染，只有该链接失效，不出现整篇文档被替换为“无法读取”的结果

#### Scenario: Malformed image destination

- **WHEN** 文档包含一个无法解析的图片目标
- **THEN** 图片不加载，但文档其余内容仍然显示

### Requirement: Local destinations prefer filesystem paths

系统 SHALL 在解析相对本地链接和图片时优先按文件系统路径解析，再回退到 URL 语义。包含 `%`、空格、中文等字符的现有文件 MUST 能被正确解析。

#### Scenario: Literal percent in filename

- **WHEN** 文档链接到当前目录下真实存在的 `100%.md`
- **THEN** 该链接解析到 `100%.md`，而不是因为非法 URL 转义而失败

#### Scenario: Valid percent sequence in filename

- **WHEN** 文档链接到当前目录下真实存在的 `foo%bar.md`
- **THEN** 该链接解析到 `foo%bar.md` 本身，而不是被错误解码成其他路径

#### Scenario: URL-encoded space fallback

- **WHEN** 文档链接到 `a%20b.md`，当前目录不存在字面文件名 `a%20b.md`，但存在 `a b.md`
- **THEN** 在文件系统字面路径不存在时，系统回退到 URL 解码并解析到 `a b.md`
