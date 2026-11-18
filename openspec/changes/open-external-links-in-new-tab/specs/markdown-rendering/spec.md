# Spec Delta

## ADDED Requirements

### Requirement: External destinations open in a new tab

系统 SHALL 把指向 `http` / `https` 或本地 `file:` 非 Markdown 目标的链接标记为新标签页打开，并 MUST 带 `rel="noopener noreferrer"`。页内锚点、已预渲染的 Markdown 文档和 `mailto:` / `tel:` 目标 MUST 保持当前页行为。

#### Scenario: External HTTP link

- **WHEN** 文档包含 `[示例](https://example.com)`
- **THEN** 渲染出的链接带 `target="_blank"` 和 `rel="noopener noreferrer"`

#### Scenario: Local file link

- **WHEN** 文档包含 `[手册](manual.pdf)`，且该文件被改写为 `file:` URL
- **THEN** 链接带 `target="_blank"` 和 `rel="noopener noreferrer"`

#### Scenario: Internal anchor

- **WHEN** 文档包含 `[跳到安装](#安装)` 或指向已预渲染文档的 `#doc=...` 链接
- **THEN** 链接不带 `target`，点击后仍在当前阅读页切换或滚动

#### Scenario: Mailto link

- **WHEN** 文档包含 `[邮件](mailto:a@example.com)`
- **THEN** 链接不带 `target`，交给系统邮件客户端处理
