# OpenMD 设计文档

日期：2026-11-17  
状态：待批准  
目标平台：Windows 10/11，单 exe

## 1. 背景与目标

用户需要一个轻量的 Markdown 阅读工具：

- 用 Go 编写，最终产物是一个 exe。
- 第一次双击 exe 可以自安装，不要求管理员权限。
- 注册为 `.md` 文件的一种打开方式，并出现在右键菜单。
- 双击 / 右键打开 md 后，转成 HTML，用系统默认浏览器阅读。
- 也支持右键文件夹，展示当前目录下的 Markdown 文档列表，并在网页内切换。
- 阅读页提供文档内目录（TOC）、GFM 语法、代码高亮、跟随系统明暗模式。
- 临时 HTML 写到程序安装目录，不写 md 所在目录。
- 不启动 HTTP 服务，不阻塞终端，不启动常驻进程。
- 只读，不做浏览器内编辑。

## 2. 用户场景

1. 从下载目录双击 `OpenMD.exe`，选择“安装”。程序复制自身到用户级安装目录并注册右键菜单。下载目录里的原 exe 可以删除。
2. 右键任意 md 文件，选择“用 OpenMD 阅读”。浏览器打开渲染后的页面。
3. 右键一个文件夹（或文件夹空白处），选择“用 OpenMD 浏览 Markdown”。浏览器打开带文档列表的页面，点击切换文档。
4. 关闭浏览器标签即可；程序进程已经退出。旧的临时页面在后续启动时清理。
5. 需要移除时，双击任意 OpenMD.exe，选择“卸载”。

## 3. 范围

### 3.1 本期做

- 单 exe 自安装 / 更新 / 卸载。
- `HKCU` 注册表注册，不需要管理员权限。
- `.md`、`.markdown` 的“打开方式”和右键菜单。
- 文件夹右键与文件夹空白处右键。
- 单文件阅读页。
- 当前目录多文档阅读页，侧栏文档列表和网页内切换。
- 文档 TOC、GFM、Chroma 代码高亮、自动明暗主题。
- 本地图片使用绝对 `file:///` 地址加载。
- 临时运行目录管理、旧目录清理。
- 单元测试和人工验收清单。

### 3.2 非目标

- 不做编辑、保存回原文件。
- 不做本地 HTTP 服务。
- 不做 base64 图片内嵌（保留为后续选项）。
- 不做递归扫描子目录（本期只扫描右键所在目录的当前层）。
- 不支持 Linux / macOS。
- 不接管 `.md` 默认打开方式。
- 不做文件监听、实时刷新、PDF 导出、数学公式、Mermaid 渲染。
- 不渲染 Markdown 中的原始 HTML（出于安全考虑，见 9.4）。

### 3.3 反目标

- 绝不在 md 文件所在目录写临时文件。
- 不修改 `.md` 的默认关联。
- 不要求管理员权限。
- 不留下常驻后台进程或监听端口。
- 不因为单个坏文件导致整个目录阅读页无法打开。

## 4. 方案选择

### 方案 A：Go 原生渲染 + 临时 HTML + 默认浏览器（采用）

Go 用 goldmark 渲染 Markdown，Chroma 生成代码高亮，CSS/JS 全部嵌入 HTML。临时 HTML 写到安装目录下的随机运行目录，交给默认浏览器后进程退出。目录模式把当前目录所有文档预渲染进同一 HTML，由少量原生 JS 切换。

优点：单 exe、离线、无服务、无端口、启动快、行为可控。  
代价：只读；目录文档很多时 HTML 体积较大，因此限制文档数量与总大小。

### 方案 B：临时本地 HTTP 服务（未采用）

浏览器访问随机端口，可动态读取目录、支持保存。需要管理端口、进程寿命、防火墙与安全边界，违背“不启服务”的约束。

### 方案 C：WebView2 桌面窗口（未采用）

不依赖默认浏览器，编辑更容易，但引入 WebView2 运行时，且不符合“用默认浏览器打开”的目标。

### 方案 D：HTML 内嵌 JS Markdown 解析器（未采用）

需要把原文塞进页面由浏览器解析，体积更大、行为更不可控；Go 已经能可靠完成转换。

### 方案 E：Electron / Tauri（未采用）

产物不再是单 Go exe，体积和依赖明显更重。

## 5. 关键决策

| 决策 | 结论 |
|---|---|
| 语言 | Go 1.26+ |
| 产物 | 单个 `OpenMD.exe`，GUI 子系统，无控制台窗口 |
| 安装方式 | 自安装，复制到用户级目录 |
| 安装位置 | `%LOCALAPPDATA%\Programs\OpenMD\OpenMD.exe` |
| 注册范围 | 仅 `HKCU`，不要求管理员 |
| 默认关联 | 不修改 `.md` 默认打开方式 |
| Markdown | goldmark + GFM 扩展 |
| 代码高亮 | Chroma，类名 + 明暗两套 CSS |
| 图片 | 本地图片改写成绝对 `file:///` URL，不转 base64 |
| 临时 HTML | `<安装目录>\openmd-temp\run-<时间>-<随机>\index.html` |
| 浏览器启动 | `rundll32.exe url.dll,FileProtocolHandler <html>` |
| 服务 | 无 |
| 编辑 | 不支持 |
| 目录扫描 | 当前目录，非递归 |
| 目录文档上限 | 500 篇或原始 Markdown 总计 25 MB，先到先限 |

## 6. 系统结构

```text
OpenMD.exe
├── main.go                 参数分发、安装对话框、错误弹窗
├── internal/install        自安装、注册表、卸载、Shell 通知
├── internal/render         Markdown 解析、TOC、URL 改写、HTML 模板
├── internal/scan           当前目录 Markdown 扫描
├── internal/tempdir        运行目录创建、旧目录清理
├── internal/browser        默认浏览器启动
├── internal/assets         嵌入的 CSS / JS / HTML 模板
└── internal/ui             Windows MessageBox 封装
```

依赖：

- `github.com/yuin/goldmark`
- `github.com/alecthomas/chroma/v2`
- `github.com/yuin/goldmark-highlighting/v2`（如实现时更适合自定义渲染器，可替换）
- `golang.org/x/sys/windows`

所有 HTML/CSS/JS 使用 `//go:embed` 嵌入，不依赖外部文件。

## 7. 安装与注册

### 7.1 启动行为

- 无参数启动：显示系统 MessageBox，提供“安装/更新、卸载、取消”。
  - MessageBox 使用“是 / 否 / 取消”，文案明确映射：是 = 安装/更新，否 = 卸载，取消 = 退出。
- 带文件路径启动：阅读该 Markdown。
- `--dir <path>`：打开目录阅读页。
- 内部保留 `--install`、`--uninstall`、`--uninstall-helper`，供自动化或清理进程使用，不对外宣传。

### 7.2 安装流程

1. 计算安装路径：`%LOCALAPPDATA%\Programs\OpenMD\OpenMD.exe`。
2. 如果当前运行的就是安装路径，跳过复制，只重新注册。
3. 否则复制自身到安装目录下的临时文件名，再原子重命名到目标路径。
4. 创建 `openmd-temp` 目录。
5. 写入注册表并通知 Shell 关联已改变。
6. 弹出成功提示，说明右键菜单已可用。

复制失败、目标被占用、注册表写入失败时，弹窗说明原因，不留下半安装状态。安装目录中的旧版本 exe 被占用时提示先关闭 OpenMD。

### 7.3 注册表布局

所有键都在 `HKCU\Software\Classes` 下。

ProgID：

```text
HKCU\Software\Classes\OpenMD.Markdown
    (Default) = Markdown 文档
    DefaultIcon\(Default) = "<安装路径>",0
    shell\open\command\(Default) = "<安装路径>" "%1"
```

“打开方式”：

```text
HKCU\Software\Classes\.md\OpenWithProgids
    OpenMD.Markdown = <空>
HKCU\Software\Classes\.markdown\OpenWithProgids
    OpenMD.Markdown = <空>
```

应用注册（补充“打开方式”列表识别）：

```text
HKCU\Software\Classes\Applications\OpenMD.exe
    FriendlyAppName = OpenMD
    shell\open\command\(Default) = "<安装路径>" "%1"
    SupportedTypes\.md = <空>
    SupportedTypes\.markdown = <空>
```

文件右键菜单：

```text
HKCU\Software\Classes\SystemFileAssociations\.md\shell\OpenMD
    MUIVerb = 用 OpenMD 阅读
    Icon = "<安装路径>",0
    command\(Default) = "<安装路径>" "%1"
```

`.markdown` 做同样注册。

文件夹右键与文件夹空白处右键：

```text
HKCU\Software\Classes\Directory\shell\OpenMD
    MUIVerb = 用 OpenMD 浏览 Markdown
    Icon = "<安装路径>",0
    command\(Default) = "<安装路径>" --dir "%1"

HKCU\Software\Classes\Directory\Background\shell\OpenMD
    MUIVerb = 用 OpenMD 浏览 Markdown
    Icon = "<安装路径>",0
    command\(Default) = "<安装路径>" --dir "%V"
```

不修改 `.md` 或 `.markdown` 的默认值。注册完成后调用 `SHChangeNotify(SHCNE_ASSOCCHANGED)`。

### 7.4 卸载流程

1. 删除本程序创建的 ProgID、应用注册、文件右键菜单、文件夹右键菜单、`OpenWithProgids` 中对应的值。只删除自己的项，不删除整个 `.md` 键。
2. 删除 `openmd-temp` 和安装目录。
3. 如果当前运行的 exe 就是安装目录内的副本，无法直接删除自身：复制自身到 `%TEMP%` 的随机临时文件，启动 `--uninstall-helper <安装目录>`，父进程退出后由 helper 删除安装目录并清理自身。
4. 删除失败时，注册表仍然已经清理；弹窗提示残留路径，可重启后手动删除。
5. 通知 Shell 关联已改变。

## 8. 启动与临时目录

### 8.1 临时目录位置

主位置：

```text
<安装目录>\openmd-temp\run-20261117-153012-<随机串>\index.html
```

用 `os.MkdirTemp` 生成唯一目录，不手写随机数。

如果安装目录不可写，回退到：

```text
%TEMP%\OpenMD\run-20261117-153012-<随机串>\index.html
```

绝不回退到 md 文件所在目录。两个位置都失败时弹窗报错。

### 8.2 生成与打开

1. 创建运行目录。
2. 渲染 HTML，先写临时文件再重命名，避免浏览器读到半个文件。
3. 用 `rundll32.exe url.dll,FileProtocolHandler <html>` 打开默认浏览器。
4. 不等待浏览器，进程立即退出。
5. 打开失败时尝试 `explorer.exe <html>`，仍失败则弹窗。

### 8.3 清理

- 每次打开阅读页前，扫描 `openmd-temp` 下的运行目录。
- 删除修改时间超过 7 天的目录。
- 最多保留最近 50 个目录，超出时从旧到新删除。
- 清理失败静默忽略，不影响本次阅读。
- 浏览器已经加载的页面不依赖进程；删除旧目录只会影响重新加载旧标签页。

## 9. 渲染与页面

### 9.1 Markdown 能力

- CommonMark。
- GFM：表格、删除线、任务列表、自动链接。
- 脚注。
- 自动标题 ID。
- 标题级别 1–6，TOC 取 h1–h3。
- 代码块按语言高亮；未知语言按纯文本显示。
- 原始 HTML 默认不渲染，作为安全边界；文本会正确转义。

### 9.2 本地 URL 改写

对 Markdown 中的图片和链接做 AST 后处理：

- `http://`、`https://`、`mailto:`、`tel:`、`data:image/*`、纯 `#fragment` 保持不变。
- 本地相对路径按当前 Markdown 文件所在目录解析成绝对路径。
- Windows 绝对路径同样处理。
- 生成 `file:///C:/...` 形式，使用 `net/url` 编码空格、中文、`#`、`%` 等字符。
- UNC 路径按 `file://server/share/...` 处理。
- 图片不复制、不转 base64。
- 链接也做同样改写；指向已加载 Markdown 集合内文档的链接改为页内切换。
- 本地 Markdown 链接目标会按 9.6 的规则预渲染；网络路径不预加载。
- 禁止 `javascript:`、`vbscript:` 等危险 URL scheme。

### 9.3 页面结构

单文件模式：

```text
┌────────────────────────────────────────────┐
│ 顶部：文档标题                      [菜单] │
├───────────────┬────────────────────────────┤
│ 本篇目录       │                            │
│ - 一级标题     │   Markdown 正文             │
│   - 二级标题   │                            │
│   - 二级标题   │   代码块 / 表格 / 图片       │
│               │                            │
└───────────────┴────────────────────────────┘
```

目录模式：

```text
┌────────────────────────────────────────────┐
│ 顶部：文件夹名                      [菜单] │
├────────────────────┬───────────────────────┤
│ [文件列表] [大纲]  │                       │
│ ▸ README.md        │   当前文档正文          │
│ ▸ setup.md         │                       │
│                    │                       │
│ （或大纲）         │                       │
│ - 一级标题         │                       │
│   - 二级标题       │                       │
└────────────────────┴───────────────────────┘
```

说明：

- 目录模式只扫描当前层，所以文件列表显示的是当前目录内的文件名，不包含子目录路径。
- 目录模式侧栏分为“文件列表”和“大纲”两个 Tab，默认显示文件列表；单文件模式只显示“大纲”。
- 正文样式使用 `github-markdown-css`，侧栏和页面外壳使用自定义样式。

视觉基线：

- 正文使用 `github-markdown-css`，自定义 CSS 只负责页面外壳。
- 外壳使用冷灰底色和深青绿色强调色；正文链接仍保持 GitHub 样式。
- 正文最大宽度约 920px，可切换为全宽模式。
- 侧栏固定宽度约 260–300px，桌面可收起，窄屏以抽屉形式展开。
- 图片自适应宽度，不超出正文。
- 表格、引用、任务列表、代码块样式完整。
- 自动跟随 `prefers-color-scheme`，不要求手动主题切换。
- 代码高亮提供明暗两套 Chroma CSS。
- 页面含语义化结构、键盘焦点样式、TOC 当前项高亮。

### 9.4 安全

- 原始 HTML 不渲染，避免本地 Markdown 注入脚本。
- 所有文本、文件名、文档标题做 HTML 转义。
- 链接 scheme 白名单；图片仅允许 `file:`、`data:image/*`、`http:`、`https:`。
- 页面加 CSP：

```text
default-src 'none';
img-src file: data: https: http:;
style-src 'unsafe-inline';
script-src 'unsafe-inline';
font-src data:;
base-uri 'none';
form-action 'none';
```

- 不加载远程字体、远程脚本、iframe。
- 远程图片仍按 Markdown 原地址加载；离线或隐私环境下可能不显示，这是已知限制。
- 不在页面中显示用户的安装路径等无关信息。

### 9.5 文档切换与 TOC

- 目录页面把每篇文档渲染成一个 `<section hidden>`，切换时只改变可见性。
- 侧栏文档列表为按钮，记录目标 section。
- 每篇文档渲染一份 TOC，放入 `<template>`，切换时克隆到侧栏“本篇目录”。
- 使用 `IntersectionObserver` 或滚动位置计算 TOC 当前项。
- 支持 URL hash（如 `#doc=README.md`）恢复当前文档；刷新页面仍停留在同一篇。
- 文档内标题锚点正常跳转。
- 点击指向集合内其他 Markdown 的链接时，切换到对应文档；集合外或未预渲染的链接按 `file:///` 打开。
- 切换文档时正文回到顶部，保留每篇文档的滚动位置不是本期要求。


### 9.6 链接文档预渲染

- 首屏从当前文档（或目录文档）收集本地 Markdown 链接，只跟随一层，不递归。
- 链接目标在首屏预渲染并加入页面；点击链接直接切换，不重新启动程序。
- 本地链接目标最多 20 个，每个不超过 2 MB，总计不超过 10 MB；超出部分保留原始 `file:///` 链接。
- 重复目标去重；目标已在当前文档集合中时不重复渲染。
- 链接目标在侧栏“链接的文档”分组展示，方便返回。
- **网络路径不参与预渲染**：UNC 路径、`file://server/...`、映射网络盘在首屏跳过，不执行 stat/read，避免拖慢打开。点击时仍按原始 `file:///` 链接交给浏览器处理。

### 9.7 版本与更新

- `build.bat` 构建时注入 `Version` 和 `BuildTime`（本地时间、带时区）。
- 顶栏右侧显示版本号；点击打开版本面板，展示完整版本号和构建时间。
- 发布页链接只放在版本面板里，不直接出现在正文、侧栏或顶栏；仅在正式构建（有版本号和构建时间）时渲染。
- 页面加载时不访问网盘、不做自动更新检查。用户打开版本面板后，比较网盘 EXE 上传时间与构建时间判断是否需要更新。
- 开发构建显示“开发版”，不展示发布页链接。

## 10. 目录扫描

- 只扫描传入目录的当前层。
- 扩展名大小写不敏感：`.md`、`.markdown`。
- 跳过无读取权限的条目，不让整个页面失败。
- 按文件名自然排序。
- 文档标题优先取第一个 h1，没有则用去掉扩展名的文件名。
- 空目录或没有 Markdown 时弹窗提示。
- 总量限制：最多 500 篇，原始 Markdown 合计最多 25 MB。超过时停止扫描并在页面顶部提示只加载了前 N 篇。
- 单篇读取失败时在该文档位置显示错误占位，不影响其他文档。

## 11. 错误处理

使用 Windows MessageBox（GUI 子系统下没有控制台）显示用户可理解的错误：

- 文件不存在或无法读取。
- 目录不存在、没有 Markdown、扫描超限。
- 安装目录不可写。
- 复制自身失败、目标 exe 被占用。
- 注册表写入/删除失败。
- 临时目录创建失败。
- 默认浏览器启动失败。

错误文案避免堆砌技术细节；详细错误可写入安装目录下的 `openmd-error.log`（追加写，失败则忽略）。

## 12. 测试策略

自动化测试不修改真实注册表，也不启动浏览器。用接口隔离：

- `Associator`：注册 / 卸载，测试使用内存 fake。
- `BrowserOpener`：测试验证收到的 HTML 路径，不真正打开。
- `Installer`：使用临时目录测试复制、更新、同路径跳过、失败回滚。
- `TempManager`：允许注入根目录与当前时间，测试唯一目录、7 天清理、50 目录上限。

单元测试覆盖：

- GFM 表格、任务列表、脚注、标题 ID。
- 代码块语言高亮；未知语言回退纯文本。
- TOC 层级与锚点。
- 相对图片路径改写：空格、中文、`#`、`%`、`..`、绝对路径、UNC。
- 远程 URL、锚点、`data:image`、危险 scheme。
- 原始 HTML 被转义/移除。
- 目录扫描：大小写扩展名、排序、跳过坏文件、数量与大小上限。
- 目录链接改写为页内切换。
- 临时目录创建、命名唯一性、旧目录清理。
- 安装注册表键模型：安装后应有的键，卸载后应删除的键，且不碰 `.md` 默认值。
- 无参数分支、文件分支、`--dir` 分支的参数分发。
- HTML 转义与 CSP/meta 基础结构。

构建与静态检查：

```text
go test ./...
go vet ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```

## 13. 人工验收清单

1. 从 `Downloads` 双击 `OpenMD.exe`，选择“安装”，确认：
   - `%LOCALAPPDATA%\Programs\OpenMD\OpenMD.exe` 存在。
   - 下载目录的原 exe 可以删除，删除后功能仍正常。
2. 右键 md 文件，能看到“用 OpenMD 阅读”。
3. 在“打开方式”列表中能看到 OpenMD。
4. 打开 md 后浏览器出现页面，标题、TOC、代码高亮、表格、任务列表正常。
5. 包含相对图片的 md 能正确显示图片；图片位于 md 所在目录或子目录。
6. 路径包含空格、中文的 md 能正常打开。
7. 右键文件夹和文件夹空白处，均能打开目录阅读页。
8. 目录页侧栏列出当前目录所有 md，点击可直接切换，TOC 跟随切换。
9. 窄窗口下侧栏可收起，正文可读。
10. 系统切换深浅色后，页面自动跟随。
11. 临时 HTML 只出现在安装目录下的 `openmd-temp`，md 目录没有任何新增文件。
12. 连续多次打开生成不同的运行目录；旧目录按规则清理。
13. 直接双击已安装的 OpenMD.exe，选择“卸载”，确认：
    - 右键菜单消失。
    - `.md` 默认打开方式未被改变。
    - 安装目录与临时目录被删除，或明确提示残留路径。
14. 打开不存在的 md、没有 md 的目录、无权限文件时，有清晰提示且不崩溃。
15. Chrome / Edge 下完成上述验收；Firefox 下记录图片显示限制。

## 14. 风险与限制

- **Firefox 的 `file://` 同源策略**：跨目录本地图片可能被拦截。本期按 Chrome/Edge 为主；后续可加 `--embed-images` 或复制资源。
- **File association 缓存**：注册后 Explorer 可能短暂不刷新，已调用 `SHChangeNotify`；必要时重启 Explorer。
- **自删除**：Windows 不允许删除正在运行的 exe，卸载使用 helper 进程处理；极端情况下需要重启后删除。
- **大目录**：单页内嵌所有文档可能很大，已有 500 篇 / 25 MB 上限；仍可能影响打开速度。
- **原始 HTML 不渲染**：使用 `<details>`、badge、居中图片等原始 HTML 的文档会缺失这部分内容。这是 v1 的安全取舍。
- **远程图片**：离线环境或严格隐私设置下可能不显示。
- **SmartScreen / 杀毒软件**：未签名的自复制 exe 可能触发提示；本期不做代码签名。
- **默认浏览器异常**：`rundll32` 无法打开时回退 `explorer.exe`；两者都失败时给出提示。

## 15. 留给实现者的空间

- 具体 CSS 视觉细节、间距、字体栈和响应式断点。
- 错误文案的具体措辞。
- Chroma 明暗主题选型。
- 目录扫描的自然排序实现。
- 卸载 helper 的具体等待/重试策略。
- 是否在实现时用自定义 Chroma 渲染器替代 `goldmark-highlighting`。
- 是否为 exe 增加自定义图标（不影响功能验收）。

## 16. 验收定义

满足以下条件即视为本期完成：

- 单 exe 可在任意可执行目录双击并自安装。
- 安装后可阅读单文件，也可浏览当前目录并切换文档。
- 右键菜单、打开方式、文件夹菜单均可用。
- 页面具备 TOC、GFM、代码高亮、明暗跟随。
- 本地图片通过绝对 `file:///` 地址正常显示（Chrome/Edge）。
- 临时文件只写入安装目录，不写 md 目录。
- 卸载能移除菜单与安装文件。
- `go test ./...`、`go vet ./...`、构建命令全部通过。
