# Spec Delta

## MODIFIED Requirements

### Requirement: Heading anchors remain navigable across documents

系统 SHALL 为每个标题生成文档内唯一、可重复生成的 ID，并把页内锚点和跨文档锚点解析到真实标题。点击跨文档锚点后 MUST 导航到目标文档页面并滚动到对应标题。

#### Scenario: Same-document anchor

- **WHEN** 文档包含 `[安装](#install)` 和标题 `## Install`
- **THEN** 点击链接后仍停留在当前页面，并滚动到 `Install` 标题，而不是跳转到不存在的 `#install`

#### Scenario: Cross-document English anchor

- **WHEN** 文档包含 `[用法](other.md#usage)`
- **THEN** 点击链接后导航到 `other.md` 的页面，并滚动到 `usage` 标题

#### Scenario: Cross-document Chinese anchor

- **WHEN** 文档包含 `[说明](other.md#中文标题)`
- **THEN** 点击链接后导航到 `other.md` 的页面，并滚动到该标题

#### Scenario: Duplicate headings

- **WHEN** 同一文档包含两个文本相同的标题
- **THEN** 两个标题获得不同的 ID，页面中不存在重复的标题 ID

### Requirement: External destinations open in a new tab

系统 SHALL 把指向 `http` / `https` 或本地 `file:` 非 Markdown 目标的链接标记为新标签页打开，并 MUST 带 `rel="noopener noreferrer"`。页内锚点、由 OpenMD 生成的 Markdown 文档页面和 `mailto:` / `tel:` 目标 MUST 保持当前页行为。

#### Scenario: External HTTP link

- **WHEN** 文档包含 `[示例](https://example.com)`
- **THEN** 渲染出的链接带 `target="_blank"` 和 `rel="noopener noreferrer"`

#### Scenario: Local file link

- **WHEN** 文档包含 `[手册](manual.pdf)`，且该文件被改写为 `file:` URL
- **THEN** 链接带 `target="_blank"` 和 `rel="noopener noreferrer"`

#### Scenario: Generated document page

- **WHEN** 文档包含指向已生成页面的 `[说明](doc-xxxx.html#usage)`
- **THEN** 链接不带 `target`，点击后在同一标签页导航

#### Scenario: Internal anchor

- **WHEN** 文档包含 `[跳到安装](#安装)`
- **THEN** 链接不带 `target`，点击后仍在当前页面滚动

#### Scenario: Mailto link

- **WHEN** 文档包含 `[邮件](mailto:a@example.com)`
- **THEN** 链接不带 `target`，交给系统邮件客户端处理
