# 第 3 章：强调与行内代码

强调让关键词更醒目，行内代码让命令、文件名和短代码保持等宽。两者都在一行文字里完成，不需要另起段落。

## 你会学到

- 粗体、斜体、粗斜体和删除线。
- 用反引号写行内代码。
- 用反斜杠转义会被 Markdown 吃掉的符号。

## 基本写法

```markdown
这是 **粗体**，这是 *斜体*，这是 ***粗斜体***。

这是 ~~删除线~~，这是 `行内代码`。

想显示星号本身时，用反斜杠转义：\*不是斜体\*。
```

符号要紧贴文字，不要在里面加空格。`** 粗体 **` 中间的星号和文字之间有空隙时，往往不会生效。

## 在 OpenMD 里怎么看

阅读视图里会直接看到加粗、斜体和等宽效果；切到源码视图，就能看到这些效果对应的 `**`、`*` 和反引号。双视图适合对照：左边改符号，右边马上看效果。

行内代码适合写 `npm install`、`config.yaml`、`Ctrl + C` 这类内容。真正的多行代码请用第 7 章的围栏代码块。

## 常见错误

- 星号或下划线内侧多了空格：强调不生效。
- 英文单词里的下划线被当成斜体：`file_name` 一般没问题，但 `file_name_here` 容易误判，必要时用反斜杠。
- 用中文全角的 `＊` 或 `～`：Markdown 不识别。
- 把长代码塞进行内代码：阅读体验差，改用代码块。

## 延伸阅读

- [Markdown Guide：Emphasis](https://www.markdownguide.org/basic-syntax/#emphasis)
- [CommonMark：Emphasis and strong emphasis](https://commonmark.org/help/tutorial/04-emphasis.html)

上一章：[标题、段落与换行](02-headings-paragraphs.md) · 下一章：[列表](04-lists.md)
