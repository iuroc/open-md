package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const pixelPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="

func fixtures(opts Options) []Fixture {
	files := LightFixtures()
	if opts.Big {
		files = append(files, BigFixtures(opts.Sizes)...)
	}
	return files
}

func fixture(path, content string) Fixture {
	return Fixture{Path: path, Data: []byte(content)}
}

func LightFixtures() []Fixture {
	png, err := base64.StdEncoding.DecodeString(pixelPNGBase64)
	if err != nil {
		panic(err)
	}
	longLine := "# 超长单行\n\n" + strings.Repeat("abcdefghij", 10240) + "\n"
	longToken := strings.Repeat("abcdefghijklmnopqrstuvwxyz0123456789", 20)

	files := []Fixture{
		fixture("01-headings.md",
			"# 一级标题\n\n"+
				"## 二级标题\n\n"+
				"### 三级标题\n\n"+
				"#### 四级标题\n\n"+
				"##### 五级标题\n\n"+
				"###### 六级标题\n\n"+
				"## 重复标题\n\n"+
				"## 重复标题\n\n"+
				"## 中文标题：安装与使用\n\n"+
				"## 标题 🚀 与 emoji\n\n"+
				"## 带 `行内代码` 与 [链接](09-links-target.md) 的标题\n\n"+
				"## 这是一个很长的标题，用来观察大纲在窄侧栏中的换行与截断表现，标题会一直延伸下去直到超过一行\n\n"+
				"正文段落。\n"),
		fixture("02-text.md",
			"# 文本样式\n\n"+
				"**加粗**、*斜体*、~~删除线~~、`行内代码`、***加粗斜体***。\n\n"+
				"> 引用第一行\n>\n> > 嵌套引用\n\n"+
				"---\n\n"+
				"普通段落，行尾两个空格产生硬换行。  \n这是硬换行后的内容。\n\n"+
				"实体：&amp; &lt; &gt; &copy;。\n\n"+
				"转义：\\*不是斜体\\*、\\`不是代码\\`。\n\n"+
				"表情与中文混排：好的 👍。\n\n"+
				"结尾段落。\n"),
		fixture("03-lists.md",
			"# 列表\n\n"+
				"- 无序一\n- 无序二\n  - 嵌套二点一\n  - 嵌套二点二\n    - 三层\n\n"+
				"1. 有序一\n2. 有序二\n   1. 嵌套有序\n\n"+
				"5. 从五开始\n6. 六\n\n"+
				"- [x] 已完成任务\n- [ ] 未完成任务\n- [ ] 嵌套任务\n  - [x] 子任务\n\n"+
				"- 松散列表第一段\n\n- 松散列表第二段\n"),
		fixture("04-tables.md",
			"# 表格\n\n"+
				"| 左对齐 | 居中 | 右对齐 |\n| :--- | :---: | ---: |\n| a | b | c |\n| 中文长内容 | `code` | 123 |\n| 含竖线 \\| 的单元格 | 普通 | 456 |\n\n"+
				"| 很宽的列 1 | 很宽的列 2 | 很宽的列 3 | 很宽的列 4 | 很宽的列 5 |\n| --- | --- | --- | --- | --- |\n| 内容 | 内容 | 内容 | 内容 | 内容 |\n"),
		fixture("05-code.md",
			"# 代码\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n\n"+
				"```text\n无语言标记的围栏代码\n第二行 <tag> 与 & 实体\n```\n\n"+
				"    // 缩进代码\n    indented := true\n\n"+
				"```unknownlang\nunknown language block\n```\n\n"+
				"~~~bash\necho \"tilde fence\"\n~~~\n"),
		fixture("06-footnotes.md",
			"# 脚注\n\n正文引用[^1]，再引用一个中文脚注[^长注]。\n\n"+
				"[^1]: 第一条脚注内容。\n[^长注]: 中文脚注内容，带 **加粗**。\n"),
		fixture("07-footnotes-peer.md",
			"# 另一篇脚注\n\n这篇也有自己的[^1]。\n\n"+
				"[^1]: 第二篇文档的脚注一，内容不同。\n"),
		fixture("08-anchors.md",
			"# 锚点\n\n[安装](#install)、[中文标题](#中文标题)、[重复标题](#重复标题)、[百分号文件](100%.md)、[井号文件](<hash#name.md>)、[跨文档英文](09-links-target.md#usage)、[跨文档中文](09-links-target.md#中文标题)。\n\n"+
				"## Install\n\n安装章节。\n\n"+
				"## 中文标题\n\n中文锚点目标。\n\n"+
				"## 重复标题\n\n第一个重复标题。\n\n"+
				"## 重复标题\n\n第二个重复标题。\n"),
		fixture("09-links-target.md",
			"# 链接目标\n\n[回到锚点用例](08-anchors.md#锚点)\n\n"+
				"## Usage\n\n英文用法章节。\n\n"+
				"# 中文标题\n\n中文标题章节。\n"),
		fixture("10-frontmatter-yaml.md",
			"---\ntitle: Hi\ndate: 2026-11-21\ntags: [测试, frontmatter]\n---\n\n"+
				"# 正文标题\n\nfront matter 不应出现在正文和大纲中。\n"),
		fixture("11-frontmatter-toml.md",
			"+++\ntitle = \"Hi\"\ndate = \"2026-11-21\"\ntags = [\"测试\", \"frontmatter\"]\n+++\n\n"+
				"# TOML 正文\n\nTOML front matter 不应出现在正文和大纲中。\n"),
		fixture("12-hr-mid.md",
			"# 正文中的分隔线\n\n第一段。\n\n---\n\n第二段。\n\n---\n\n第三段。\n"),
		fixture("13-links.md",
			"# 链接\n\n"+
				"- [外部](https://example.com)\n"+
				"- [邮件](mailto:a@example.com)\n"+
				"- [电话](tel:+8610000)\n"+
				"- [PDF](common/手册.pdf)\n"+
				"- [预渲染文档](09-links-target.md)\n"+
				"- [页内锚点](#链接)\n"+
				"- [网络路径](<\\\\server\\share\\doc.md>)\n"),
		fixture("14-images.md",
			"# 图片\n\n同目录：\n\n![根目录图片](pixel-root.png)\n\n"+
				"子目录：\n\n![子目录图片](common/pixel.png)\n\n"+
				"带空格：\n\n![空格图片](<common/pixel one.png>)\n\n"+
				"中文文件名：\n\n![中文图片](<common/中文图.png>)\n\n"+
				"缺失图片（不应影响其余正文）：\n\n![缺失](common/不存在.png)\n\n"+
				"远程图片（离线可能不显示）：\n\n![远程](https://example.com/x.png)\n\n"+
				"图片后的正文仍然显示。\n"),
		fixture("15-broken.md",
			"# 失效目标\n\n[坏链接](101%.md)\n\n![坏图片](bad%zz.png)\n\n[空链接]()\n\n"+
				"其余正文应当正常显示。\n"),
		fixture("16-empty.md", ""),
		fixture("17-bom.md", "\ufeff# BOM 文档\n\n标题应正常显示，BOM 不应出现在正文中。\n"),
		fixture("18-unicode.md",
			"# Unicode\n\n"+
				"中文段落：的一是了我不人在他有这个上们来到时大地为子中你说生国年着就那和要她出也得里后自以会家可下而过天去能对小多然于心学么之都好看起发当没成只如事把还用第样道想作种开美总从无情己面最女但现前些所同日手又行意动方期它头经长儿回位分爱老因很给名法间斯知世什两次使身者被高已亲其进此话常与活正感。\n\n"+
				"Emoji：😀 🚀 🧪 ✅\n\n"+
				"组合字符：é 与 é\n\n"+
				"RTL：مرحبا بالعالم\n\n"+
				"超长不可断 token：\n\n"+longToken+"\n"),
		fixture("19-html.md",
			"# 原始 HTML\n\n块级：\n\n<script>alert('xss')</script>\n\n"+
				"行内：<b>加粗标签</b> 与 <img src=\"x\" onerror=\"alert(1)\">。\n\n"+
				"预期：HTML 不被执行，按普通文本显示。\n"),
		fixture("20-long-line.md", longLine),

		fixture("100%.md", "# 百分号文件\n\n文件名带 %，应能通过 08-anchors.md 中的链接打开。\n"),
		fixture("foo%bar.md", "# foo%bar\n\n字面文件名解析。\n"),
		fixture("a b.md", "# 空格文件名\n\n根目录中的带空格文件。\n"),
		fixture("中文文件名.md", "# 中文文件名\n\n中文路径与信息面板复制。\n"),
		fixture("UPPER.MARKDOWN", "# 大写扩展名\n\n.MARKDOWN 也应被识别。\n"),
		fixture("hash#name.md", "# 井号文件名\n\n链接需要按字面文件名解析。\n"),

		fixture("fallback/link.md", "# 回退用例\n\n[链接 a%20b.md](a%20b.md)\n\n目录里只有 `a b.md`，应回退解码打开它。\n"),
		fixture("fallback/a b.md", "# 回退目标\n\n这是 a b.md。\n"),
		fixture("literal/link.md", "# 字面优先\n\n[链接 a%20b.md](a%20b.md)\n\n应打开字面的 `a%20b.md`，不是 `a b.md`。\n"),
		fixture("literal/a%20b.md", "# 字面百分号文件\n\n这个文件才应被打开。\n"),
		fixture("literal/a b.md", "# 空格文件\n\n不应被打开。\n"),
		fixture("sub/nested.md", "# 子目录文档\n\n当前目录模式不应在根列表中显示本文件。\n"),

		fixture("tree/root.md",
			"# 文件树根\n\n"+
				"- [入门](guides/intro.md)\n"+
				"- [指南](guides/guide.md)\n"+
				"- [深层文档](guides/deep/deep.md)\n"+
				"- [叶子文档](guides/deep/deeper/leaf.md)\n"+
				"- [中文说明](<中文目录/说明.md>)\n"+
				"- [混合目录](mixed/readme.md)\n"+
				"- [同名 A](duplicates/a.md)\n"+
				"- [同名 B](duplicates/b/a.md)\n"),
		fixture("tree/guides/intro.md", "# 入门\n\n介绍。\n"),
		fixture("tree/guides/guide.md", "# 指南\n\n## Usage\n\n用法。\n\n# 中文标题\n\n中文。\n"),
		fixture("tree/guides/deep/deep.md", "# 深层文档\n\n位于 guides/deep/。\n"),
		fixture("tree/guides/deep/deeper/leaf.md", "# 叶子文档\n\n四层路径。\n"),
		fixture("tree/中文目录/说明.md", "# 中文目录说明\n\n目录名为中文。\n"),
		fixture("tree/中文目录/子目录/更多.md", "# 更多\n\n更深一层的中文路径。\n"),
		fixture("tree/no-md/note.txt", "这个目录没有 Markdown 文件。\n"),
		fixture("tree/mixed/readme.md", "# 混合目录\n\n同目录还有 note.txt 和 logo.png。\n"),
		fixture("tree/mixed/note.txt", "普通文本附件。\n"),
		{Path: "tree/mixed/logo.png", Data: png},
		fixture("tree/duplicates/a.md", "# 同名 A\n\n这个 a.md 在 tree/duplicates/。\n"),
		fixture("tree/duplicates/b/a.md", "# 同名 B\n\n这个 a.md 在 tree/duplicates/b/。\n"),
		fixture("tree/sort-order/2.md", "# 2\n"),
		fixture("tree/sort-order/10.md", "# 10\n"),
		fixture("tree/sort-order/a.md", "# a\n"),
		fixture("tree/sort-order/B.md", "# B\n"),
		fixture("tree/sort-order/中文.md", "# 中文\n"),
		{Path: "tree/empty", Dir: true},
		{Path: "tree/retry", Dir: true},

		{Path: "pixel-root.png", Data: png},
		{Path: "common/pixel.png", Data: png},
		{Path: "common/pixel one.png", Data: png},
		{Path: "common/中文图.png", Data: png},
		fixture("common/手册.pdf", "%PDF-1.4\n% OpenMD 人工验收占位附件\n%%EOF\n"),
	}

	for i := 1; i <= 501; i++ {
		files = append(files, fixture(
			fmt.Sprintf("limits-count/doc-%03d.md", i),
			fmt.Sprintf("# doc-%d\n\n占位内容。\n", i),
		))
	}
	return files
}
