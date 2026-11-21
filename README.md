# OpenMD

一个只读的 Windows Markdown 阅读器：把 Markdown 转成 HTML，用默认浏览器打开，用于快速阅读。

## 特性

- 单 exe，自安装，不需要管理员权限。
- 注册为 `.md` / `.markdown` 的“打开方式”，并增加带图标的右键菜单。
- 支持右键文件夹和文件夹空白处，浏览当前目录的 Markdown 文档列表。
- 目录模式的侧栏分为“文件列表”和“大纲”两个 Tab，默认显示文件列表。
- 单文件包含链接文档时，侧栏分为“大纲”和“文档”两个 Tab，默认显示大纲。
- 侧栏可收起；阅读区可在固定宽度和全宽模式之间切换。
- 支持海盐蓝（默认）、松石、紫藤、石墨四套配色，跟随系统明暗。
- 正文使用 `github-markdown-css`，GFM、代码高亮。
- 自动剥离文档开头的 YAML / TOML front matter，元数据不进入正文和大纲。
- 顶栏右侧的圆圈 + i 按钮打开“信息与设置”面板，提供当前文档完整路径、所在目录、配色、作者、版本、构建时间和发布页。
- 信息面板显示当前文档的完整绝对路径和所在目录，并支持复制路径。
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

## 版本与更新

- 仓库地址：<https://github.com/iuroc/open-md>（私有仓库，需要对应的 GitHub 访问权限）。
- 顶栏右侧的齿轮按钮打开设置面板。
- 设置面板显示作者、完整版本号、构建时间、配色选择，以及一个不太显眼的发布页链接。
- 发布页指向 GitHub Releases：<https://github.com/iuroc/open-md/releases>。最新版本的构建时间晚于页面显示的构建时间，就说明有更新。
- 发布页链接只在正式构建的设置面板中出现；开发构建不会展示。
- 页面加载时不会访问 GitHub，也不做自动更新检查。

## 安装

1. 双击 `OpenMD.exe`（放在 Downloads 等任意可执行目录都可以）。
2. 在系统对话框里选择“是”：安装 / 更新。
3. 程序会把自己复制到：

```text
%LOCALAPPDATA%\Programs\OpenMD\OpenMD.exe
```

4. 下载目录里的原 exe 之后可以删除，右键菜单仍然可用。

不会修改 `.md` 的默认打开方式，只把它加入“打开方式”列表和右键菜单。

## 使用

- 右键 Markdown 文件 → “用 OpenMD 阅读”。
- 右键文件夹 → “用 OpenMD 浏览 Markdown”。
- 在文件夹空白处右键 → “用 OpenMD 浏览 Markdown”。
- Markdown 文件的“打开方式”列表中也可以选择 OpenMD。

目录模式只扫描当前层，不递归子目录。最多加载 500 篇、原始 Markdown 合计 25 MB，单个文件不超过 25 MB；超过的文件会跳过并给出提示。单文件模式的上限是 32 MiB，超限会直接报错。

## 文档间链接

- 文档里链接到本地 `.md` / `.markdown` 时，会预先渲染链接目标，点击后在同一页切换。
- 页内锚点 `[跳转](#章节)` 会定位到当前文档的对应标题；跨文档锚点 `[说明](other.md#章节)` 会切换到目标文档并滚动到标题。
- 中文标题同样支持锚点，例如 `[说明](other.md#中文标题)`。
- 只跟随一层链接，最多 20 个链接文档，每个不超过 2 MB，总计不超过 10 MB。
- 链接文档在文件列表中用 `↗` 标记，方便识别和跳回。
- 本地相对路径优先按字面文件名解析；字面文件不存在时才尝试 URL 解码，因此 `100%.md`、`foo%bar.md` 这类文件名也能正常打开。
- 链接到网络路径（UNC 或映射网络盘）的文档**不会在首屏预加载**，避免拖慢打开速度；点击时仍交给浏览器处理原始文件。
- 外部 `http(s)` 链接和本地非 Markdown 文件（PDF、图片等）默认在新标签页打开；页内锚点和已预渲染的 Markdown 文档仍在当前页切换。

## 临时文件

临时页面写入：

```text
%LOCALAPPDATA%\Programs\OpenMD\openmd-temp\run-<时间>-<随机>\index.html
```

不会写到 Markdown 文件所在目录。旧运行目录在后续启动时清理：超过 7 天或超过最近 50 个目录。

如果安装目录不可写，会回退到 `%TEMP%\OpenMD`。

## 卸载

双击任意 `OpenMD.exe`，在对话框里选择“否”：卸载。

也可以执行：

```bash
OpenMD.exe --uninstall
```

卸载会移除右键菜单、临时目录和安装副本。当前运行的安装副本由临时 helper 进程在退出后删除；极少数情况下需要重启后手动清理残留目录。

## 已知限制

- 只读，不支持浏览器内编辑或保存回原文件。
- Markdown 中的原始 HTML 不渲染，避免本地脚本注入。
- 目录模式不递归，文档多时受 500 篇 / 25 MB 上限限制；单个超大文件会被跳过并提示，不会阻止其他文档加载。
- 单文件模式最多读取 32 MiB，超过会提示文件过大。
- 本地链接优先按字面文件名解析；如果同时存在 `a%20b.md` 和 `a b.md`，会优先打开字面文件。
- 网络路径的 md 链接不会预渲染成阅读页，点击后由浏览器打开原始文件。
- 本地图片使用绝对路径。Chrome / Edge 正常；Firefox 的 `file://` 同源策略可能拦截跨目录图片。
- 远程 `http(s)` 图片保持原地址，离线时可能不显示。
- exe 未做代码签名，首次运行可能触发 SmartScreen 提示。
- 不做 PDF 导出、数学公式、Mermaid、文件监听或实时预览。

## 图标

- 图标源文件：`assets/openmd.ico`（白底、青色描边和 M 标记）
- 网页 favicon：`internal/render/web/favicon.png`，由同一个生成器输出并以内嵌 data URI 写入页面
- 重新生成：`go run ./tools/icon`
- 嵌入 exe：`rsrc -ico assets/openmd.ico -o rsrc_windows_amd64.syso`
- 已提交的 `rsrc_windows_amd64.syso` 会被 `go build` 自动使用，日常构建不需要再安装 `rsrc`

## 开发

```bash
go test ./...
go vet ./...
```

测试不会修改真实的 Markdown 关联；注册表测试只使用 `HKCU\Software\OpenMDTest-<随机>` 临时键并自动清理。

正文样式使用 [github-markdown-css](https://github.com/sindresorhus/github-markdown-css)（MIT），许可声明见 `THIRD_PARTY_NOTICES.md`。
