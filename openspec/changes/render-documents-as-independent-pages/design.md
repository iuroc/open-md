# Design

## Context

当前实现把所有文档渲染进同一个 `index.html`，因此必须有 500 篇上限。渲染、锚点、脚注和跨文档链接都依赖“所有文档已在同一页”。现有扫描、链接解析和源码高亮可以复用，但页面输出、导航和缓存需要重构。

## Goals / Non-Goals

**Goals:**

- 每篇文档一个独立 HTML，稳定路径，可覆盖复用。
- 同一 root 共享 CSS、JS 和 `tree.js`。
- 未变化文档跳过读取和渲染。
- 首次打开按待重建工作量决定是否弹窗；确认后可完全放开限制。
- 文件树使用方案 1 的视觉层级，并支持懒渲染和跨页面展开状态。
- 源码模式跨页面保持，刷新重置。

**Non-Goals:**

- 不做 iframe 外壳。
- 不做本地 HTTP 服务、按需渲染或全文搜索。
- 不保留单页包含全部文档的旧输出模式。
- 不实现跨 root 的总入口页。

## Decisions

### 独立页面 + 普通导航，不使用 iframe

每篇文档生成 `doc-<path-hash>.html`，树链接和跨文档链接都是普通 `<a href>`。实测 Chrome 的 `file://` iframe 为跨源：父页面无法读取 `contentDocument`，`contentWindow.location` 抛 `SecurityError`，只有 `postMessage` 可用。普通导航更简单，浏览器前进/后退、锚点和查找都按原生行为工作。

### 稳定缓存与 manifest

缓存目录：

```text
<base>/rendered/<root-key>/
  manifest.json
  tree.js
  reader.css
  reader.js
  github-markdown.css
  chroma.css
  doc-<hash>.html
```

`root-key` 由扫描根绝对路径 hash 加可读前缀组成。文档输出名由源文件绝对路径 hash 决定。

`manifest.json` 保存：版本、known-set hash、每篇文档的路径、输出名、mtime、size、标题和错误状态。输出使用临时文件 + rename 原子覆盖。

### 缓存复用规则

- 版本或 known-set hash 变化：全量重建。
- 单篇 mtime/size 变化或输出缺失：只重建该篇。
- 其余文档：跳过读取、解析、高亮和写入，直接复用页面和标题。
- 每次运行重新生成 `tree.js` 和 manifest。
- 生成后删除当前 root 中不在 manifest 的 `doc-*.html`。
- 旧 root 按 7 天 / 最近 50 个清理。
- 开发构建禁用复用，避免模板变化后读到旧页面。

### 软阈值与确认

扫描元数据后，结合 manifest 计算本次待重建文档数和字节数。待重建超过 3,000 篇或 8 MiB 时弹窗：

- 加载全部：使用无限制扫描，取消数量、总量、深度、目录项和单文件限制；
- 快速模式：应用 3,000 篇 / 8 MiB 限制；
- 取消：不打开。

缓存命中不计入待重建量，因此大目录重复打开不会反复弹窗。按当前机器实测，3,000 篇小文档首次生成约 4 秒，8 MiB 文档约 5 秒。

### 并行渲染

使用 `min(runtime.NumCPU(), 8)` 个 worker 渲染待重建文档。单篇失败生成错误页，不终止整批。manifest 和 tree.js 在所有 worker 完成后统一写入。

### 共享 tree.js 与懒渲染

`tree.js` 写入文档树数据，所有页面加载同一份。`reader.js` 初始只渲染根节点和当前路径；文件夹首次展开时才创建子节点。展开状态按 root 存入 `sessionStorage`。

### 链接重写

所有可生成页面的文档（根内文档和链接文档）先分配稳定的 `doc-<hash>.html`。跨文档链接重写为相对页面 URL + 锚点；页内锚点保持 `#anchor`；未知目标回退为原始 `file://` 文件。生成页面没有 `file:` scheme，因此不会被标记为新标签页。

每篇页面只渲染一篇文档，标题 ID 和脚注 ID 不再需要文档前缀。

### 源码模式跨页面

源码模式状态按 root 存入 `sessionStorage`。页面加载时读取 `performance.getEntriesByType("navigation")[0].type`：`reload` 时忽略并清除状态，其他导航类型时恢复。这样文件树和跨文档链接保持模式，刷新回到渲染视图。

### 树形视觉方案

采用方案 1：固定 16px 引导槽、每层缩进、竖向引导线、文件夹 600 字重、文件 400 字重且颜色更淡、当前项使用 `--accent-soft` 背景和左侧 accent 线。不使用目录计数。

## Risks / Trade-offs

- [首次打开仍需等待全部页面生成] → 并行渲染 + 软阈值确认；缓存命中后接近 stat 速度。
- [新增/删除文档导致全量重建] → 首版接受；后续可按链接依赖缩小重建范围。
- [mtime/size 不能发现所有内容变化] → 开发构建禁用缓存；发布版本每次构建都会使版本变化并全量重建。
- [缓存长期增长] → 每 root 清理旧页面，旧 root 按 TTL / 数量清理。
- [file:// 外部 CSS/JS 受 CSP 影响] → 已验证 `style-src file:` / `script-src file:` 可用，页面使用相对资源。
- [共享 tree.js 被缓存] → 文件内容每次重写，文件名固定；页面加载时浏览器可能复用旧文件，必要时在 URL 加版本参数。

## Migration Plan

1. 旧的 `run-*` 临时目录继续按现有逻辑清理。
2. 新输出写入 `rendered/<root-key>`；首次运行不依赖旧页面。
3. 回滚方式：恢复旧 `BuildDirPage` / `RenderPage` 和单页模板；新缓存目录可整体删除，无数据迁移。
