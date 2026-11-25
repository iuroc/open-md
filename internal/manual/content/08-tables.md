# 第 8 章：表格

表格适合对比参数、列出选项、整理短信息。Markdown 表格用竖线分隔单元格，用一行分隔线定义表头。

## 你会学到

- 写一张最基本的表格。
- 用冒号控制左对齐、居中、右对齐。
- 处理宽表格和单元格里的竖线。

## 基本写法

```markdown
| 选项 | 默认值 | 说明 |
| --- | :---: | ---: |
| theme | blue | 配色名称 |
| size | 16 | 正文字号 |
| width | auto | 页面宽度 |
```

第二行是分隔行，不能省略。冒号的位置决定对齐方式：

- `:---` 左对齐
- `:---:` 居中
- `---:` 右对齐

单元格里要显示竖线本身时，用反斜杠转义：`\|`。

## 在 OpenMD 里怎么看

阅读视图里，宽表格会在正文区域横向滚动，不会把页面撑破；打印时表格会适配页宽。源码视图能看到每一行的竖线，双视图适合调整列宽和对齐。

表格单元格里不能直接放多行内容。需要多行说明时，拆成多张表或改成列表。

## 常见错误

- 漏掉分隔行：整张表变成普通段落。
- 表头和数据行的列数不一致：多出来的单元格会被忽略或错位。
- 用全角的 `｜`：Markdown 只认半角竖线。
- 在单元格里直接换行：表格会断掉。
- 表格太宽、文字太长：阅读时要横向滚动，尽量精简列。

## 延伸阅读

- [Markdown Guide：Tables](https://www.markdownguide.org/extended-syntax/#tables)
- [GitHub Docs：表格](https://docs.github.com/zh/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-tables)

上一章：[代码](07-code-blocks.md) · 下一章：[分隔线、HTML 与安全边界](09-rules-and-html.md)
