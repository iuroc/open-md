package main

import (
	"strings"
)

func Checklist(big bool) []byte {
	var b strings.Builder
	b.WriteString("# OpenMD 人工验收清单\n\n")
	if big {
		b.WriteString("生成模式：含大文件（-big）。\n\n")
	} else {
		b.WriteString("生成模式：轻量。大小上限小节需要 `-big` 重新生成后执行。\n\n")
	}

	b.WriteString("## 1. 准备\n\n")
	b.WriteString("- [ ] 已安装 OpenMD（双击 exe，在对话框选择“是”）。\n")
	b.WriteString("- [ ] 在仓库根目录执行 `go run ./tools/cases`（或加 `-big`），输出目录为 `cases/`。\n")
	b.WriteString("- [ ] 知道 `cases/` 可随时整目录删除，不会提交进 Git。\n\n")

	b.WriteString("## 2. 基础渲染\n\n")
	b.WriteString("- [ ] `01-headings.md`：H1–H6 层级正确；重复标题可点击；中文、emoji、带代码/链接、超长标题显示正常；大纲与正文标题一致。\n")
	b.WriteString("- [ ] `02-text.md`：加粗、斜体、删除线、行内代码、引用、嵌套引用、水平线、硬换行、实体、转义均符合预期。\n")
	b.WriteString("- [ ] `03-lists.md`：有序/无序/嵌套/混合列表正确；任务列表显示复选框；从 5 开始的有序列表编号正确。\n")
	b.WriteString("- [ ] `04-tables.md`：三列对齐正确；单元格内代码正常；转义竖线不拆列；宽表可横向滚动查看。\n")
	b.WriteString("- [ ] `05-code.md`：Go 代码高亮；无语言、未知语言、缩进代码、波浪线围栏都按代码块显示；代码中的 `<tag>` 不被当作 HTML。\n")
	b.WriteString("- [ ] `05a-code-systems.md`：Go、Rust、C、C++、C#、Java、Kotlin、Swift、Scala、汇编；检查泛型、宏、指针、注解、生命周期、模式匹配等语法着色。\n")
	b.WriteString("- [ ] `05b-code-scripting.md`：Python、JavaScript、TypeScript、Ruby、PHP、Perl、Lua、Bash、PowerShell、R、Julia、Dart；检查装饰器、模板字符串、正则、heredoc、管道、泛型等语法着色。\n")
	b.WriteString("- [ ] `05c-code-functional.md`：Haskell、Elixir、Erlang、Clojure、F#、OCaml、Scheme、Common Lisp；检查模式匹配、宏、类型签名、列表推导等语法着色。\n")
	b.WriteString("- [ ] `05d-code-data-web.md`：JSON、YAML、TOML、SQL、HTML、CSS、SCSS、XML、GraphQL、Protobuf、INI、Diff、Dockerfile、Makefile、Nginx、Terraform、LaTeX；检查键、字符串、注释、标签、属性、指令等语法着色。\n")
	b.WriteString("- [ ] 所有语言块的注释、字符串、关键字、数字、类型颜色可区分；未识别的语言标记不高亮；代码中的 `<` `>` `&` 按文本显示而非 HTML。\n")
	b.WriteString("- [ ] `06-footnotes.md`：点击脚注引用跳到脚注，回链回到引用位置；中文脚注可用。\n")
	b.WriteString("- [ ] 同时打开 `06-footnotes.md` 与 `07-footnotes-peer.md`（放入同一目录浏览）：两篇的 `[^1]` 互不串，回链留在各自文档。\n")
	b.WriteString("- [ ] `08-anchors.md`：页内英文/中文锚点、重复标题锚点都定位到当前文档正确标题。\n")
	b.WriteString("- [ ] `09-links-target.md`：从 `08-anchors.md` 点击跨文档英文/中文锚点，切换文档并滚动到对应标题。\n")
	b.WriteString("- [ ] `10-frontmatter-yaml.md`：正文从 `# 正文标题` 开始；front matter 字段和分隔线不显示，大纲无 front matter 标题。\n")
	b.WriteString("- [ ] `11-frontmatter-toml.md`：TOML front matter 同样被剥离。\n")
	b.WriteString("- [ ] `12-hr-mid.md`：正文中间的 `---` 仍显示为水平线，不被当作 front matter。\n")
	b.WriteString("- [ ] `13-links.md`：外部 http 链接新标签打开；mailto/tel 交给系统；PDF 新标签打开；本地 md 预渲染并在当前页切换；网络路径不预加载。\n")
	b.WriteString("- [ ] `14-images.md`：根目录、子目录、带空格、中文名图片都显示；缺失图片不破坏其余正文；远程图片联网时显示。\n")
	b.WriteString("- [ ] `15-broken.md`：坏链接、坏图片、空链接只影响自身，其余正文正常显示。\n")
	b.WriteString("- [ ] `16-empty.md`：空文档正常打开，页面不报错。\n")
	b.WriteString("- [ ] `17-bom.md`：BOM 不出现在正文，首个标题正常。\n")
	b.WriteString("- [ ] `18-unicode.md`：中文、emoji、组合字符、RTL 文本显示正常；超长 token 不撑破页面。\n")
	b.WriteString("- [ ] `19-html.md`：`<script>` 与行内 HTML 不执行，按普通文本显示。\n")
	b.WriteString("- [ ] `20-long-line.md`：约 100 KB 单行能换行或横向滚动，页面不卡死。\n\n")

	b.WriteString("## 3. 文档间链接与锚点\n\n")
	b.WriteString("- [ ] 从目录模式打开 `08-anchors.md`：链接文档在侧栏“文档”Tab 中出现，并带 `↗` 标记。\n")
	b.WriteString("- [ ] 点击跨文档链接后，侧栏文档列表和阅读区状态一致；可以跳回原文档。\n")
	b.WriteString("- [ ] 指向已存在 `100%.md` 的链接能打开；指向 `foo%bar.md`、`hash#name.md` 的链接按字面文件名解析。\n\n")

	b.WriteString("## 4. Front matter\n\n")
	b.WriteString("- [ ] YAML、TOML、正文中的 `---` 三种情况分别对照 `10`、`11`、`12`。\n")
	b.WriteString("- [ ] 大纲中不出现 `title`、`date`、`tags` 等 front matter 字段。\n\n")

	b.WriteString("## 5. 文件名与路径\n\n")
	b.WriteString("- [ ] 直接打开 `100%.md`、`foo%bar.md`、`a b.md`、`中文文件名.md`、`hash#name.md`、`UPPER.MARKDOWN`，均能阅读。\n")
	b.WriteString("- [ ] `fallback/link.md` 点击链接后打开同目录的 `a b.md`（URL 解码回退）。\n")
	b.WriteString("- [ ] `literal/link.md` 点击链接后打开字面的 `literal/a%20b.md`，而不是 `literal/a b.md`。\n")
	b.WriteString("- [ ] 信息面板显示中文/百分号文件的完整路径，复制路径可用。\n\n")

	b.WriteString("## 6. 图片与本地附件\n\n")
	b.WriteString("- [ ] `14-images.md` 覆盖同目录、子目录、空格、中文名、缺失、远程图片。\n")
	b.WriteString("- [ ] `13-links.md` 中的 `common/手册.pdf` 在新标签页打开。\n")
	b.WriteString("- [ ] `tree/mixed/logo.png` 等图片在网页中可见，且不在 md 文件旁生成任何新文件。\n\n")

	b.WriteString("## 7. 目录模式与侧栏（当前行为）\n\n")
	b.WriteString("- [ ] 右键 `cases/` 选择“用 OpenMD 浏览 Markdown”：侧栏有“文件列表”“大纲”两个 Tab，默认“文件列表”。\n")
	b.WriteString("- [ ] 文件列表只显示当前层：`fallback/`、`sub/`、`tree/`、`limits-count/` 等子目录中的文档不出现。\n")
	b.WriteString("- [ ] `sub/nested.md` 不出现在 `cases/` 根列表（当前不递归）。\n")
	b.WriteString("- [ ] 单文件包含链接文档时，侧栏为“大纲”“文档”两个 Tab，默认“大纲”。\n")
	b.WriteString("- [ ] 侧栏可收起；阅读区可在固定宽度和全宽之间切换。\n")
	b.WriteString("- [ ] 500 篇以内自然排序正确：`01` 在 `10` 前，`2` 在 `10` 前。\n\n")

	b.WriteString("## 8. 大小上限（-big）\n\n")
	if big {
		b.WriteString("- [ ] 打开 `cases/limits-count/`：加载 500 篇并提示截断，第 501 篇不加载。\n")
		b.WriteString("- [ ] 打开 `cases/limits/skip/`：`big-33mib.md` 被跳过并显示警告，`small.md` 仍可阅读。\n")
		b.WriteString("- [ ] 直接打开 `cases/limits/skip/big-33mib.md`：显示文件超过大小限制的错误，进程保持响应。\n")
		b.WriteString("- [ ] 打开 `cases/limits/total/`：累计超过 25 MiB 后停止扫描并提示截断。\n")
		b.WriteString("- [ ] 打开 `cases/limits/all-big/`：显示“均超过大小限制”的空状态，而不是“没有 Markdown 文件”。\n")
		b.WriteString("- [ ] 直接打开 `cases/limits/at-limit/big-32mib.md`：32 MiB 临界文件能渲染；页面很大、耗时较长，属预期。\n")
	} else {
		b.WriteString("- [ ] 使用 `go run ./tools/cases -big -force` 重新生成后，再执行本小节。\n")
	}
	b.WriteString("\n")

	b.WriteString("## 9. 界面与信息面板\n\n")
	b.WriteString("- [ ] 四套配色（海盐蓝、松石、紫藤、石墨）切换生效，并跟随系统明暗。\n")
	b.WriteString("- [ ] 顶栏圆圈 + i 按钮打开信息面板，显示当前文档完整路径、所在目录、配色、作者、版本、构建时间。\n")
	b.WriteString("- [ ] 信息面板的复制路径对中文、空格、`%`、`#` 文件名都正确。\n")
	b.WriteString("- [ ] 临时页面写入安装目录的 `openmd-temp`，浏览器打开后程序退出，不在 md 旁生成文件。\n\n")

	b.WriteString("## 10. 安装与右键菜单（可选）\n\n")
	b.WriteString("- [ ] 安装后 `.md` / `.markdown` 的“打开方式”列表出现 OpenMD，默认打开方式未被修改。\n")
	b.WriteString("- [ ] 右键 Markdown 文件、文件夹、文件夹空白处，三个菜单项都可用。\n")
	b.WriteString("- [ ] 卸载后右键菜单、安装副本和临时目录被清理。\n\n")

	b.WriteString("## 11. 未来功能：递归扫描 + 文件树（待变更落地后验收）\n\n")
	b.WriteString("本小节在“目录递归扫描 + 文件树”变更实现前不执行；当前先确认第 7 节的“不递归”行为。\n\n")
	b.WriteString("- [ ] 用 OpenMD 打开 `cases/`：文件树显示根层文件与 `fallback/`、`sub/`、`tree/`、`limits-count/` 等文件夹。\n")
	b.WriteString("- [ ] 展开 `tree/`：能看到 `guides/`、`中文目录/`、`empty/`、`no-md/`、`mixed/`、`duplicates/`、`sort-order/`。\n")
	b.WriteString("- [ ] `tree/guides/deep/deeper/leaf.md` 出现在四层路径中，展开/收起状态正确。\n")
	b.WriteString("- [ ] `tree/empty/` 空目录和 `tree/no-md/` 无 Markdown 目录的显示符合变更设计。\n")
	b.WriteString("- [ ] `tree/duplicates/a.md` 与 `tree/duplicates/b/a.md` 在树中可区分，点击后打开各自文档。\n")
	b.WriteString("- [ ] `tree/sort-order/` 按自然序显示：`2.md`、`10.md`、`a.md`、`B.md`、`中文.md`。\n")
	b.WriteString("- [ ] 选择文件夹与选择文档时，阅读区和侧栏状态符合变更设计。\n")
	b.WriteString("- [ ] 点击跨目录链接后，文件树定位到目标文档。\n")
	b.WriteString("- [ ] 递归扫描遵守 500 篇 / 25 MiB 上限并给出清晰提示；`limits-count/` 单独打开时仍能验证截断。\n")
	b.WriteString("- [ ] 之前“子目录文档不出现”的行为被新树取代，且没有重复文档或漏显示。\n\n")

	b.WriteString("## 12. 观察项\n\n")
	b.WriteString("- [ ] 记录与规格不符但不阻塞验收的现象，作为后续变更输入：\n")
	b.WriteString("  - 现象：\n  - 复现文件/操作：\n  - 期望：\n\n")

	return []byte(b.String())
}
