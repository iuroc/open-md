# OpenMD 人工验收案例生成器设计

日期：2026-11-21
状态：待确认

## 目的

为 OpenMD 提供一套可手工逐项验收的 Markdown 案例。案例不提交进仓库，由仓库内的 Go 小工具按需生成到 `cases/`，验收完可整目录删除。内容不追求真实合理，只要求把渲染、链接、目录扫描、大小上限和界面行为覆盖全。

同时为稍后的「目录递归扫描 + 文件树」变更预留验收素材：嵌套目录、深层路径、中文目录、空目录、同名文件等。

## 目标

- `go run ./tools/cases` 在仓库根目录生成轻量案例集（约 2–3 MB）。
- `-big` 追加大小上限用例（合计约 100 MB）。
- 生成 `cases/CHECKLIST.md`：每条写明“操作 → 期望结果”。
- 覆盖 README / openspec 规格中的全部用户可见行为。
- 生成器可重复运行；只清理自己生成的目录。
- 生成器本身有快速单元测试，保证文件清单不漂移。

## 非目标

- 不做浏览器自动化或端到端测试；这是人工验收素材。
- 不把生成逻辑塞进 `OpenMD.exe`。
- 不把任何生成物提交进 Git。
- 不实现递归扫描和文件树——那是另一个 OpenSpec 变更。
- 不追求案例内容语义完整，只保证格式与功能触发点。

## 约束

- Windows、Go 1.26，沿用 `tools/icon` 的工具目录惯例。
- OpenMD 现有上限：目录模式 500 篇 / 25 MiB 总量 / 单文件 32 MiB；单文件模式 32 MiB；链接文档最多 20 篇、每篇 2 MiB、合计 10 MiB。
- 目录模式当前只扫描当前层，不递归。
- 文件名包含中文、空格、`%`、`#` 时仍需可生成、可在浏览器打开。

## 方案比较

- **A（选用）Go 工具 `tools/cases`**：与 `tools/icon` 一致；直接生成中文文件名、特殊字符文件名和大文件，行为稳定；可用 `go test` 保护结构。
- **B 批处理 / PowerShell 脚本**：不依赖 Go，但 UTF-8 中文名、百分号文件名和大文件生成容易出错，维护成本高。
- **C `OpenMD.exe` 隐藏参数**：少跑一个命令，但把测试能力混入产品，增加体积和风险，不采用。

## 设计

### 命令

```bash
go run ./tools/cases            # 轻量案例，输出到 ./cases
go run ./tools/cases -big       # 追加大小上限用例
go run ./tools/cases -out D:\qa\openmd   # 自定义输出目录
go run ./tools/cases -force     # 覆盖非本工具生成的目录
```

- 生成物根目录写入标记文件 `.openmd-cases`。
- 输出目录不存在或为空 → 直接生成。
- 已有标记 → 清空后重新生成。
- 非空且无标记 → 报错，除非显式 `-force`，避免误删。

### 生成器结构

```text
tools/cases/
  main.go        # 参数解析、清理与写盘
  fixtures.go    # 案例清单与内容
  bigfiles.go    # 33/32 MiB 等大文件的流式生成
  cases_test.go  # 文件清单、内容抽查、清理保护
```

- 案例用 `fixture{Path, Data}` 或 `fixture{Path, Write func(io.Writer) error}` 描述；大文件必须流式写入，不整块进内存。
- 内容固定、无时间戳、无随机数，重复生成结果一致。
- `limits/all-big/big.md` 优先对 33 MiB 文件建 NTFS 硬链接；失败时退化为复制并提示。
- 测试通过 `Options` 把大文件尺寸缩小到几十 KB，避免 CI 写 100 MB。

### 输出布局

```text
cases/
  CHECKLIST.md
  .openmd-cases

  # 平铺主用例（自然序编号）
  01-headings.md           H1–H6、重复标题、中文标题、长标题
  02-text.md               强调、删除线、行内代码、引用、水平线、硬换行
  03-lists.md              有序/无序/嵌套/混合/任务列表、非 1 起始编号
  04-tables.md             对齐、单元格内代码、转义竖线、宽表、中文
  05-code.md               围栏代码（有无语言）、缩进代码、长行、未知语言
  05a-code-systems.md      Go、Rust、C、C++、C#、Java、Kotlin、Swift、Scala、汇编
  05b-code-scripting.md    Python、JavaScript、TypeScript、Ruby、PHP、Perl、Lua、Bash、PowerShell、R、Julia、Dart
  05c-code-functional.md   Haskell、Elixir、Erlang、Clojure、F#、OCaml、Scheme、Common Lisp
  05d-code-data-web.md     JSON、YAML、TOML、SQL、HTML、CSS、SCSS、XML、GraphQL、Protobuf、INI、Diff、Dockerfile、Makefile、Nginx、Terraform、LaTeX
  06-footnotes.md          脚注 [^1]，与 07-footnotes-peer.md 同时打开
  07-footnotes-peer.md     独立 [^1]，验证多文档脚注作用域
  08-anchors.md            页内英文/中文锚点、重复标题锚点
  09-links-target.md       被链接文档，含 ## Usage 与 # 中文标题
  10-frontmatter-yaml.md   YAML front matter + 正文标题
  11-frontmatter-toml.md   TOML front matter
  12-hr-mid.md             正文中的 --- 仍按水平线渲染
  13-links.md              http、mailto、tel、本地 pdf、本地 md、UNC 链接
  14-images.md             同目录图片、子目录图片、缺失图片、中文/空格图片、远程图片
  15-broken.md             不存在的 101%.md、坏图片、空链接；其余正文正常
  16-empty.md              空文件
  17-bom.md                UTF-8 BOM + 标题
  18-unicode.md            中文、emoji、组合字符、RTL、超长不可断 token
  19-html.md               原始 HTML 块与行内 HTML（应不执行）
  20-long-line.md          约 100 KB 单行，验证换行与滚动

  # 文件名 / 路径用例
  100%.md                  存在的字面百分号文件
  foo%bar.md
  a b.md
  中文文件名.md
  UPPER.MARKDOWN
  hash#name.md
  fallback/link.md         链接 a%20b.md，仅存在 a b.md → 回退解码
  fallback/a b.md
  literal/link.md          同时存在 a%20b.md 与 a b.md → 字面文件优先
  literal/a%20b.md
  literal/a b.md
  sub/nested.md            当前不应出现在 cases 根目录列表

  # 未来递归扫描 + 文件树素材
  tree/root.md             链接到各子目录文档
  tree/guides/intro.md
  tree/guides/guide.md     带英文/中文标题锚点
  tree/guides/deep/deep.md
  tree/guides/deep/deeper/leaf.md
  tree/中文目录/说明.md
  tree/中文目录/子目录/更多.md
  tree/empty/              空目录
  tree/no-md/note.txt      无 Markdown 的目录
  tree/mixed/readme.md
  tree/mixed/note.txt
  tree/mixed/logo.png
  tree/duplicates/a.md
  tree/duplicates/b/a.md   同名不同目录
  tree/sort-order/2.md 10.md a.md A.md 中文.md

  # 大小上限（仅 -big）
  limits-count/            501 篇极小文档（默认轻量版也生成）
  limits/skip/big-33mib.md     33 MiB：单文件模式报错；目录模式跳过
  limits/skip/small.md        跳过超限文件后仍正常加载
  limits/total/01.md…05.md    5 × 5.5 MiB（各 11 << 19 字节），累计超过 25 MiB → 截断
  limits/all-big/big.md       只有超限文件 → 全部超限的空状态
  limits/at-limit/big-32mib.md 32 MiB 临界文件，应能打开（渲染很慢）

  # 资源
  common/pixel.png         同目录与子目录图片、中文/空格文件名图片
  common/手册.pdf          非 Markdown 本地附件
```

### 验收清单

`CHECKLIST.md` 由生成器写出，固定包含以下小节：

1. 准备：生成命令、安装/打开 OpenMD。
2. 基础渲染：逐篇打开 01–20。
3. 文档间链接与锚点：预渲染、跨文档锚点、失效链接、脚注作用域。
4. Front matter：YAML / TOML / 正文中的 `---`。
5. 文件名与路径：`%`、空格、中文、`#`、大小写扩展名、回退解码。
6. 图片与本地附件：同目录、子目录、缺失、远程、PDF。
7. 目录模式与侧栏：文件列表/大纲两个 Tab、目录不递归、`sub/nested.md` 不出现。
8. 大小上限（`-big`）：单文件 32 MiB、目录 500 篇 / 25 MiB、跳过警告、全部超限空状态。
9. 界面：侧栏收起、全宽切换、四套配色、系统明暗跟随、信息面板路径与复制。
10. 安装与右键菜单（可选）：安装、右键文件/文件夹/空白处、卸载。
11. 未来功能：递归扫描 + 文件树（标注为待变更落地后验收）。
12. 观察项：与规格不符但不阻塞验收的现象，留作后续变更参考。

未来功能小节会预先写入检查步骤，例如：展开 `tree/guides/deep/deeper/leaf.md`、区分 `tree/duplicates` 下两个 `a.md`、空目录与无 Markdown 目录的展示、递归打开 `cases/` 时 500 篇 / 25 MiB 上限是否生效、链接跳转后文件树是否定位。

### 仓库改动

- 新增 `tools/cases/`（生成器 + 测试，约 20–25 KB）。
- `.gitignore` 增加 `/cases/`。
- README「开发」小节增加生成命令说明。
- 生成物不提交。

## 需要确认的决定

- 默认轻量、`-big` 按需（已确认）。
- 保留 32 MiB 临界用例，体积不设上限（已确认）。
- 案例目录平铺 + 少量专用子目录（已确认）。
