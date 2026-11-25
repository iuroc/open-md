# 第 5 章：链接与图片

链接把文档连起来，图片让说明更直观。这一章还会讲 OpenMD 特有的跨文档跳转和标题锚点。

## 你会学到

- 写普通链接、带标题的链接和相对路径链接。
- 链接到另一个 Markdown 文件，以及链接到某一节。
- 插入本地图片，处理带空格的路径。

## 基本写法

```markdown
[打开 Markdown Guide](https://www.markdownguide.org/ "外部教程")

[手册首页](index.md)

[跳到本章的“常见错误”](#常见错误)

![示意图](images/example.png)
```

方括号里是显示文字，圆括号里是目标地址。图片前面多一个感叹号，方括号里写替代文字，图片加载失败时它会派上用场。

## 在 OpenMD 里怎么看

在目录模式下，链接到另一个 `.md` 文件时会生成独立页面，左侧文件树里也能直接切换。链接到 `#标题` 会跳到当前文档的对应小节，中文标题同样支持。

本地图片会转换成 `file:///` 地址，不会在 Markdown 旁边生成任何缓存文件。路径里有空格时，可以用尖括号包起来：

```markdown
![架构图](<images/系统 架构.png>)
```

## 常见错误

- 使用 Windows 反斜杠路径：`images\a.png` 不可靠，改用正斜杠。
- 图片路径写错或大小写不一致：图片显示为破图。
- 把图片链接写成普通链接：看到的是文字而不是图片。
- 链接到文件夹而不是文件：浏览器打不开。
- 远程图片在断网时无法显示：需要离线阅读的文档尽量用本地图片。
- 中文标题锚点尽量保持标题稳定，改标题会让旧锚点失效。

## 延伸阅读

- [Markdown Guide：Links](https://www.markdownguide.org/basic-syntax/#links)
- [Markdown Guide：Images](https://www.markdownguide.org/basic-syntax/#images)
- [GitHub Docs：自动链接与相对链接](https://docs.github.com/zh/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax)

上一章：[列表](04-lists.md) · 下一章：[引用块](06-blockquotes.md)
