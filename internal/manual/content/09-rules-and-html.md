# 第 9 章：分隔线、HTML 与安全边界

Markdown 可以做简单的排版，但它不是网页编辑器。这一章讲分隔线的正确用法，以及什么时候不该依赖 HTML。

## 你会学到

- 用 `---` 写分隔线，避免和标题下划线混淆。
- 理解 OpenMD 对原始 HTML 的处理方式。
- 在 Markdown 表达不了时，选择更稳妥的替代方案。

## 基本写法

分隔线由三个或更多短横线、星号组成，单独占一行，并且前后留空行：

```markdown
上一节的内容。

---

下一节的内容。
```

如果 `---` 紧贴在文字下面，它可能被识别成“标题下划线”，把上一行变成二级标题。前后留空行可以避免这种歧义。

Markdown 也允许写原始 HTML，但 OpenMD 会出于安全考虑忽略这些标签，只保留正常的 Markdown 排版。也就是说，从网页复制来的 `<details>`、`<div>` 之类内容不会按预期展开或布局。

## 在 OpenMD 里怎么看

阅读视图里，分隔线显示为一条水平线；源码视图里能看到 `---`。原始 HTML 在源码视图里仍然可见，但阅读视图不会执行它，也不会显示它。

如果确实需要折叠内容，优先改成“小标题 + 列表”或拆成多个小节。这样在任何 Markdown 阅读器里都能正常显示。

## 常见错误

- `---` 紧贴段落：上一行变成标题，或者分隔线不出现。
- 指望 `<details>` 折叠：OpenMD 不渲染原始 HTML。
- 从网页复制整段 HTML：阅读视图里通常什么都不剩。
- 用 HTML 做复杂布局：换一个阅读器就可能失效，Markdown 文档应保持可移植。

## 延伸阅读

- [CommonMark：Thematic breaks](https://spec.commonmark.org/0.31.2/#thematic-breaks)
- [Markdown Guide：Horizontal rules](https://www.markdownguide.org/basic-syntax/#horizontal-rules)

上一章：[表格](08-tables.md) · 下一章：[组织一篇文档](10-organize-documents.md)
