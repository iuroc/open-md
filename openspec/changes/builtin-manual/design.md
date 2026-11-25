# Design

## Context

参见 `proposal.md`。当前实现的关键约束：

- OpenMD 没有主窗口。打开文档时，`app.openPage` 先把文档站点渲染到 `%LOCALAPPDATA%\Programs\OpenMD\openmd-temp\rendered\<root-hash>\`，再用浏览器打开 `site.InitialPage`；进程随后退出。
- 阅读页的「信息与设置」面板有「文件 / 设置 / 关于」三个 Tab，「关于」位于 `internal/render/web/docpage.html`。
- 页面通过 `file:///` 加载；CSP 只允许 `script-src file:` 的内联和经典脚本。动态 `fetch()` 本地文件在 Chromium / Firefox 下不可用，但普通 `<a href>` 顶层导航和 `<script src>` 可用。
- 站点渲染已有缓存与失效机制：`manifest.json` 按版本、源文件 mtime / size 和已知文档集合决定是否重建；模板或共享资源变化时手动提升 `cacheGeneration`。
- `tempdir.Manager.Cleanup()` 只清理 `run-*` 和 `rendered/*`；其他子目录不会被自动清理。

## Goals / Non-Goals

**Goals:**

- 新用户能在阅读页内打开一本完整的 Markdown 入门手册，离线可用。
- 手册页面与普通文档体验一致：文件树、大纲、阅读 / 源码 / 双视图、打印、主题和排版设置全部可用。
- 入口只出现在阅读页「关于」Tab；不新增安装器、对话框或右键菜单入口。
- 手册生成失败时不影响文档打开，入口自动隐藏。

**Non-Goals:**

- 不增加 `--manual` 命令行入口，也不在安装完成对话框里加入口。
- 不做手册搜索、目录折叠、阅读进度或交互式练习。
- 不摘录或改写第三方教程正文；外部内容只以链接形式出现。
- 不改变文档扫描、渲染、缓存清理和安装器的既有行为。

## Decisions

### 1. 手册内容以 Markdown 章节形式内嵌

在 `internal/manual/content/` 下放 `index.md`、`01-…md` 到 `11-…md`、`12-cheatsheet.md`、`13-faq.md`，通过 `//go:embed` 打进 exe。章节之间使用相对链接，首页 `index.md` 由既有 `initialSiteDocIndex` 自动选中。

不在仓库里手写 HTML，也不在构建时预生成：手写 HTML 会失去 Markdown 写作、复用不了阅读界面；构建时预生成会让 workflow 多一步且容易与 reader 资源不同步。

### 2. 打开文档时生成，而不是安装时生成

入口只在阅读页，安装时生成没有消费者。`app.openPage` 在渲染文档站点之前调用 `manual.Ensure(baseDir)`：

1. 计算内嵌内容的 SHA-256 指纹；
2. 把章节写到 `<baseDir>/manual/src-<指纹>/`（已存在则跳过）；
3. 以 `<baseDir>/manual/cache` 为 `CacheBase`，用现有 `render.BuildDirSite` 生成站点；
4. 返回手册首页的 `file:///` URL。

正常文档打开本身就要调用 `BuildDirSite`，手册只多 14 个小文件，耗时在毫秒级；后续打开通过 `manifest.json` 复用页面，只做一次目录扫描。

`Ensure` 调用放在 `openPage` 的 `build(base)` 闭包内，主缓存目录（安装目录下的 `openmd-temp`）不可写而回退到 `%TEMP%\OpenMD` 时，手册也会跟着在回退目录生成；两次尝试各自使用自己的 `baseDir`。

### 3. 指纹目录隔离版本，提取与生成都可恢复

手册内容变化时指纹变化，源目录和输出目录随之变化，不会读到旧页面。

- 章节先写入 `<baseDir>/manual/.staging-<随机>/`，写完后整体 `os.Rename` 为 `src-<指纹>`；目标已存在时丢弃暂存目录。这样进程中途退出不会留下看起来“已存在”的残缺章节。
- `src-<指纹>` 下写入 `.complete` 标记；只有标记存在且手册首页可读时才跳过提取，否则重新提取。
- 生成成功后清理 `<baseDir>/manual/src-*` 和 `cache/*`，但只删除**非当前版本且超过 24 小时未修改**的目录，避免删掉其他进程正在阅读的版本；清理失败只忽略。

不使用 `cacheGeneration` 之外的额外版本号：模板或共享资源变化时按既有约定提升 `cacheGeneration`，`BuildDirSite` 的 manifest 会使旧页面失效并重建。

### 4. 多进程同时打开时串行生成

OpenMD 每打开一个文档就启动一个进程，文件关联和“打开方式”都可能同时拉起多个进程，它们会争用同一个手册目录。因此：

- 快路径只读：先检查 `.complete` 标记和首页，命中就直接返回，不加锁。
- 需要生成时用 `<baseDir>/manual/.lock`（`O_CREATE|O_EXCL`）串行化，最多等待约 2 秒；拿到锁后重新检查，避免重复生成。
- 等不到锁时，如果已有完整手册就复用；否则返回错误，由调用方隐藏入口。
- 锁文件超过 5 分钟视为陈旧并尝试删除，避免进程崩溃后永久阻塞。

文档站点既有的缓存不加锁，本变更不改变其并发行为。

### 5. 入口放在「关于」Tab，用普通链接打开新标签页

在 `docpage.html` 和 `page.html` 的「关于」区域增加：

```html
{{if .ManualURL}}<a class="manual-link panel-action" href="{{.ManualURL}}" target="_blank" rel="noopener noreferrer">打开使用手册</a>{{end}}
```

`ManualURL` 为空时整段不渲染。手册页面自身在生成时不设置 `ManualURL`，因此手册的「关于」Tab 不会链接到自己。

入口使用顶层导航而不是脚本注入：手册已经是一组完整页面，跳转即可获得大纲、源码视图和打印能力；脚本注入只能在当前页里复刻一套较弱的阅读界面。

### 6. `ManualURL` 通过站点选项透传，并参与文档页面缓存失效

- `render.SiteOptions` 增加 `ManualURL`，由 `BuildFileSite` / `BuildDirSite` 透传到 `renderDocs` → `renderOneDoc` → `DocPageData`。
- `Page` / `pageData` 同样增加 `ManualURL`，让 `page.html` 与 `docpage.html` 保持同步；`BuildFilePage` 当前只被测试使用，也一并支持，避免两个模板行为分叉。
- `siteManifest` 增加 `ManualURL` 字段；打开文档时若清单中的 URL 与当前值不同，视为需要重建，避免缓存页面保留上一版本或已删除手册的链接。
- `app.Manual` 作为可注入的依赖，便于 `app` 包测试；生产路径设置为 `manual.Ensure`。

### 7. 生成失败时静默降级

`manual.Ensure` 返回错误时，`openPage` 以空的 `ManualURL` 继续渲染文档；「关于」Tab 不显示手册入口。不弹窗、不写入站点警告，因为手册是辅助内容，失败不应打断阅读。生成错误在单元测试中覆盖，必要时在开发构建里通过测试或日志定位。

不设置硬性生成超时：内容只有十几个小文件，同步生成即可；只有并发等待使用上面的 2 秒重试上限。

### 8. 内容结构、状态与出处

面向零基础读者，每章统一为「你会学到 / 基本写法 / 在 OpenMD 里怎么看 / 常见错误 / 延伸阅读」：

- `index.md`：手册说明、目录、如何配合源码 / 双视图学习；
- 01 从零开始：Markdown 是什么；
- 02 标题、段落与换行；
- 03 强调与行内代码；
- 04 列表；
- 05 链接与图片；
- 06 引用块；
- 07 代码块；
- 08 表格；
- 09 分隔线、折叠与 HTML 边界；
- 10 组织一篇文档（README / index、目录结构、front matter、交叉引用）；
- 11 五个实战配方（项目 README、学习笔记、接口说明、会议记录、长文打印）；
- 12 速查表；
- 13 常见问题（渲染没生效、中文标点、Windows 路径）。

正文全部原创；每章末尾的「延伸阅读」只放 Markdown Guide、CommonMark 规范、GitHub Docs 等外部链接，不复制原文，因此 `THIRD_PARTY_NOTICES.md` 不需要新增条目。手册不引用外部图片，保证离线完整；示例只用文本、代码块和表格。

## Risks / Trade-offs

- **每次打开文档都多一次手册检查** → 内容指纹和 `manifest.json` 命中后只做少量文件读取和一次目录扫描，可忽略；失败不影响主流程。
- **多个进程同时首次打开** → 锁文件串行化生成；等待超时时复用已有手册或隐藏入口，文档始终正常打开。
- **进程中断留下残缺手册** → 章节写到暂存目录后整体重命名，并用 `.complete` 标记判断完整；残缺目录下次重新提取。
- **手册链接指向的缓存目录被手动删除** → 下次打开文档会重新生成；已打开的旧标签页可能失效，可接受。
- **文档页面缓存保留旧手册 URL** → 通过 manifest 的 `ManualURL` 比较强制重建。
- **外部延伸阅读链接失效** → 手册核心内容离线可用，外链只是补充；后续维护时替换即可。
- **内容维护成本** → 章节按固定模板拆分，单章篇幅控制在入门可读完的范围，便于后续增补。

## Migration Plan

新增能力，无数据迁移。模板、脚本或共享资源变化时按既有约定提升 `cacheGeneration`，旧页面下次打开时重建。回滚时删除 `internal/manual` 与入口模板改动，并再次提升缓存代次即可。
