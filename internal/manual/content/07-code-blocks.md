# 第 7 章：代码

技术文档里最常见的 Markdown 内容就是代码。短的用行内代码，长的用围栏代码块。

## 你会学到

- 行内代码和围栏代码块的区别。
- 给代码块加语言标记，获得语法高亮。
- 处理长行和代码里的反引号。

## 基本写法

行内代码用一对反引号包住，适合命令、文件名和变量：

```markdown
运行 `go test ./...` 前先保存文件。

配置文件叫 `config.yaml`。
```

多行代码用三个反引号包起来，并在开头的反引号后面写语言名。下面就是一段带 `go` 标记的代码：

```go
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
```

在源码视图里，你会看到这段代码前后的三反引号围栏；阅读视图只显示代码本身。

语言名可以是 `go`、`javascript`、`python`、`json`、`bash`、`markdown` 等。写错语言名不会报错，只是没有高亮；不写语言名也可以显示，但不会有语法着色。

如果代码本身包含反引号，外层可以用更多反引号包住，例如用四个反引号开始和结束。

## 在 OpenMD 里怎么看

阅读视图会按语言给代码着色；源码视图会对整篇 Markdown 做高亮。打印时，长代码块会自动换行，避免被页面边缘裁切。

代码块里的内容不会被当成 Markdown 解析，所以 `#`、`*`、`-` 都可以原样显示。

## 常见错误

- 围栏没有闭合：后面的正文都会变成代码。
- 开头围栏后直接换行、没写语言名：能显示，但没有高亮。
- 代码块前后没有空行：有些解析器会把它和段落粘在一起。
- 用行内代码包多行内容：反引号不能跨行。
- 语言名写成大写或拼错：高亮可能失效。

## 延伸阅读

- [Markdown Guide：Code](https://www.markdownguide.org/basic-syntax/#code)
- [GitHub Docs：代码块](https://docs.github.com/zh/get-started/writing-on-github/working-with-advanced-formatting/creating-and-highlighting-code-blocks)

上一章：[引用块](06-blockquotes.md) · 下一章：[表格](08-tables.md)
