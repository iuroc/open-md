# OpenMD

一个只读的 Windows Markdown 阅读器：把 Markdown 转成 HTML，用默认浏览器打开，用于快速阅读。

## 特性

- 单 exe，自安装，不需要管理员权限。
- 注册为 `.md` / `.markdown` 的“打开方式”，并增加带图标的右键菜单。
- 支持右键文件夹和文件夹空白处，递归浏览目录内所有层级的 Markdown 文档。
- 每篇文档生成独立的 HTML 页面，文件树和跨文档链接使用普通页面跳转。
- 目录模式的侧栏以文件树展示文档，分为“文件列表”和“大纲”两个 Tab，默认显示文件树。
- 单文件包含链接文档时，侧栏分为“大纲”和“文档”两个 Tab，默认显示大纲。
- 侧栏可收起；阅读区可在固定宽度和全宽模式之间切换。
- 支持海盐蓝（默认）、松石、紫藤、石墨四套配色，跟随系统明暗。
- 正文使用 `github-markdown-css`，GFM、代码高亮。
- 自动剥离文档开头的 YAML / TOML front matter，元数据不进入正文和大纲。
- 顶栏的阅读模式按钮以下拉菜单提供“阅读 / 源码 / 双视图”三种模式；源码只读、Markdown 语法高亮、默认自动换行，可一键复制完整源码（含 front matter）。
- 宽表格的表头保持单行，表格在自身区域内横向滚动，正文单元格正常换行。
- 顶栏右侧的圆圈 + i 按钮打开“信息与设置”面板，分为“文件 / 设置 / 关于”三个 Tab，分别提供当前文档路径、配色、源码自动换行和版本信息。
- 本地图片使用绝对 `file:///` 地址，不在 md 文件旁生成任何文件。
- 临时 HTML 写入安装目录下的 `openmd-temp`，启动浏览器后程序立即退出。
- 无 HTTP 服务、无常驻后台进程、无浏览器内编辑。

## 构建

需要 Go 1.26+，在项目根目录执行：

```bash
go test ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```

也可以直接双击 `build.bat`：自动运行测试、注入版本号和构建时间、构建 `OpenMD.exe`，并生成 `source.git.bundle`。

只想要源码包时，双击 `bundle.bat` 即可。

产物是单个 `OpenMD.exe`。手动执行上面的 `go build` 时页面会显示“开发版”；用 `build.bat` 构建才会写入正式版本号和构建时间。

## 持续集成与发布

GitHub Actions 工作流见 `.github/workflows/build.yml`：

- Pull Request：只运行 `go vet ./...` 和 `go test ./... -count=1`。
- push 到 `main`：只运行测试，不构建。
- push `v*` 标签：测试通过后构建并创建 GitHub Release，附加 `OpenMD.exe` 和 `OpenMD.exe.sha256`，同时上传 workflow artifact。
- 也可以在 Actions 页面手动触发构建（workflow_dispatch），产物作为 workflow artifact 保留 14 天。

版本号规则：tag 构建去掉前缀 `v` 作为版本号；手动触发的构建使用 `yyyy.MM.dd-<run_number>`。`build.bat` 会识别外部传入的 `BUILD_VERSION`、`BUILD_TIME`；设置 `CI=true` 时跳过 `source.git.bundle` 和 `pause`，设置 `OPENMD_SKIP_TESTS=true` 时跳过测试（CI 中测试已由前一个 job 完成）。

发布正式版本：

```bash
git tag v2026.11.18
git push origin v2026.11.18
```

私有仓库的 Actions 和 Release 附件都需要有权限的账号访问；Windows runner 按 2 倍分钟计费。

## 阅读模式

- 点击顶栏的阅读模式按钮展开下拉菜单，可在“阅读”“源码”“双视图”之间切换。
- 双视图左侧显示源码、右侧显示正文，两侧各自滚动并保持同步；窗口宽度不足 900px 时自动回退为阅读视图。
- 双视图下顶栏的宽度模式按钮会隐藏，因为它对分栏布局没有作用；切回阅读/源码后恢复显示，全宽偏好保留。
- 通过文件树或跨文档链接打开另一篇时保持当前模式，刷新页面也会恢复上次选择的模式。
- 源码视图显示原始 Markdown，包括文档开头的 front matter，并按 Markdown 语法高亮。
- 围栏代码块按语言高亮；列表项里缩进的围栏同样支持。
- 源码默认自动换行；可在“设置”Tab 中关闭，关闭后保留原始行宽，双视图左侧会出现横向滚动条。
- 点击“复制源码”可复制当前文档的完整源码；焦点在源码区域时按 `Ctrl/Cmd + A` 只选中源码。
- 点击侧栏大纲会自动回到阅读视图，并跳到对应标题；双视图中则直接定位右侧正文和左侧源码。

## 打印

- 阅读模式打印渲染正文，源码模式打印源码，双视图只打印渲染正文。
- 打印时代码块临时自动换行，表格收缩到页面宽度并允许表头和单元格换行，不会出现横向滚动条或被裁切。
- 源码自动换行设置只影响屏幕显示；打印源码时始终换行。

## 版本与更新

- 仓库地址：<https://github.com/iuroc/open-md>（私有仓库，需要对应的 GitHub 访问权限）。
- 顶栏右侧的圆圈 + i 按钮打开“信息与设置”面板。
- “文件”Tab 显示当前文档的完整路径和所在目录，并支持复制路径；“设置”Tab 提供配色；“关于”Tab 显示作者、完整版本号、构建时间和发布页。
- 发布页指向 GitHub Releases：<https://github.com/iuroc/open-md/releases>。最新版本的构建时间晚于页面显示的构建时间，就说明有更新。
- 发布页链接只在正式构建的“关于”Tab 中出现；开发构建不会展示。
- 页面加载时不会访问 GitHub，也不做自动更新检查。

## 安装

1. 双击 `OpenMD.exe`（放在 Downloads 等任意可执行目录都可以）。
2. 程序会直接安装到：

```text
%LOCALAPPDATA%\Programs\OpenMD\OpenMD.exe
```

3. 安装完成后弹出一个对话框；选择“完成”即可，需要卸载时在同一对话框选择“卸载”。
4. 下载目录里的原 exe 之后可以删除，右键菜单仍然可用。

不会修改 `.md` 的默认打开方式，只把它加入“打开方式”列表和右键菜单。

## 使用

- 右键 Markdown 文件 → “用 OpenMD 阅读”。
- 右键文件夹 → “用 OpenMD 浏览 Markdown”。
- 在文件夹空白处右键 → “用 OpenMD 浏览 Markdown”。
- Markdown 文件的“打开方式”列表中也可以选择 OpenMD。

目录模式递归扫描所有子目录，并把结果组织成可展开、折叠的文件树。文件节点显示文件名，悬停可查看文档标题和完整相对路径。根目录存在 `README.md` / `index.md` 时优先打开它，否则打开层级最浅的第一篇。

每篇 Markdown 生成独立的 HTML 页面。扫描会跳过 `.git`、`node_modules`、`vendor` 等目录，也不会跟随符号链接或 junction。默认不限制文档数量和总量；如果本次需要生成的文档超过 3,000 篇或 8 MiB，会先询问：

- 加载全部：取消数量、总量、深度、目录项和单文件限制；
- 快速模式：只加载 3,000 篇 / 8 MiB；
- 取消：不打开。

未变化文档会复用已有页面，只重建新增或修改的内容。单文件模式超过 32 MiB 时也会先询问。

## 文档间链接

- 文档里链接到本地 `.md` / `.markdown` 时，会生成目标文档的独立页面，点击后在同一标签页导航。
- 页内锚点 `[跳转](#章节)` 会定位到当前文档的对应标题；跨文档锚点 `[说明](other.md#章节)` 会打开目标页面并滚动到标题。
- 中文标题同样支持锚点，例如 `[说明](other.md#中文标题)`。
- 只跟随一层链接，最多 20 个链接文档，每个不超过 2 MB，总计不超过 10 MB。
- 链接文档在文件列表中用 `↗` 标记；目录模式中，扫描根之外的链接文档单独放在“链接文档”分组，不混入文件树。
- 本地相对路径优先按字面文件名解析；字面文件不存在时才尝试 URL 解码，因此 `100%.md`、`foo%bar.md` 这类文件名也能正常打开。
- 链接到网络路径（UNC 或映射网络盘）的文档**不会在首屏预加载**，避免拖慢打开速度；点击时仍交给浏览器处理原始文件。
- 外部 `http(s)` 链接和本地非 Markdown 文件（PDF、图片等）默认在新标签页打开；页内锚点和已预渲染的 Markdown 文档仍在当前页切换。

## 临时文件

渲染缓存写入：

```text
%LOCALAPPDATA%\Programs\OpenMD\openmd-temp\rendered\<root-hash>\
  manifest.json
  tree.js
  reader.css
  reader.js
  doc-<hash>.html
```

同一源文件始终使用同一个输出路径，内容变化时原子覆盖。源文件、版本和文档集合未变化时直接复用已有页面。不会写到 Markdown 文件所在目录。旧缓存按 7 天 / 最近 50 个 root 清理。

如果安装目录不可写，会回退到 `%TEMP%\OpenMD`。

## 卸载

双击安装目录里的 `OpenMD.exe`，在对话框里选择“卸载”。

也可以执行：

```bash
OpenMD.exe --uninstall
```

卸载会移除右键菜单、临时目录和安装副本。当前运行的安装副本由临时 helper 进程在退出后删除；极少数情况下需要重启后手动清理残留目录。

## 已知限制

- 只读，不支持浏览器内编辑或保存回原文件。
- Markdown 中的原始 HTML 不渲染，避免本地脚本注入。
- 超大目录首次打开需要预生成所有页面，确认“加载全部”后不受数量、总量、深度或目录项限制，但生成时间会随文档量增加。
- 快速模式仍受 3,000 篇 / 8 MiB 限制，并在文件树底部提示截断。
- 单文件超过 32 MiB 时会先询问，确认后读取。
- 本地链接优先按字面文件名解析；如果同时存在 `a%20b.md` 和 `a b.md`，会优先打开字面文件。
- 网络路径的 md 链接不会预渲染成阅读页，点击后由浏览器打开原始文件。
- 本地图片使用绝对路径。Chrome / Edge 正常；Firefox 的 `file://` 同源策略可能拦截跨目录图片。
- 远程 `http(s)` 图片保持原地址，离线时可能不显示。
- exe 未做代码签名，首次运行可能触发 SmartScreen 提示。
- 安装与错误对话框使用 Windows 原生 TaskDialog，不跟随阅读页配色和深色模式。
- 不做 PDF 导出、数学公式、Mermaid、文件监听或实时预览。

## 图标

- 图标源文件：`assets/openmd.ico`（白底、青色描边和 M 标记）
- 网页 favicon：`internal/render/web/favicon.png`，由同一个生成器输出并以内嵌 data URI 写入页面
- 重新生成：`go run ./tools/icon`
- 嵌入 exe：`rsrc -ico assets/openmd.ico -manifest assets/app.manifest -o rsrc_windows_amd64.syso`
- `assets/app.manifest` 提供 Common Controls v6、`asInvoker` 和 PerMonitorV2 DPI；修改后需重新生成 `.syso`
- 已提交的 `rsrc_windows_amd64.syso` 会被 `go build` 自动使用，日常构建不需要再安装 `rsrc`

## 开发

```bash
go test ./...
go vet ./...
```

测试不会修改真实的 Markdown 关联；注册表测试只使用 `HKCU\Software\OpenMDTest-<随机>` 临时键并自动清理。

生成人工验收案例：

```bash
go run ./tools/cases        # 轻量案例，输出到 ./cases
go run ./tools/cases -big   # 追加大小上限用例（约 100 MB）
```

`cases/` 不提交进仓库，可随时整目录删除。验收步骤见生成后的 `cases/CHECKLIST.md`。

正文样式使用 [github-markdown-css](https://github.com/sindresorhus/github-markdown-css)（MIT），许可声明见 `THIRD_PARTY_NOTICES.md`。
