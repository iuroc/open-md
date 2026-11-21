# Proposal

## Why

当前目录模式把所有文档预渲染进同一个 HTML，因此必须用 500 篇的数量上限保护浏览器页面。像 OpenSpec 这样包含大量规格和归档的目录，会因为这个上限而漏掉文档；单个文档页面也不会因为整页体积而变慢。

## What Changes

- 目录模式和单文件模式改为**每篇 Markdown 一个独立 HTML 页面**，文件树链接使用普通页面跳转，不再把所有文档放在同一页。
- 使用稳定的渲染缓存：
  - 扫描根目录和每篇文档都有稳定输出路径；
  - 相同源文件覆盖同一个 HTML；
  - `manifest.json` 记录 mtime、size、版本和文档集合 hash，未变化时跳过读取和渲染；
  - 删除已不存在的文档页面，按 TTL / 最近使用数量清理旧 root。
- 首次打开时并行渲染需要更新的文档；共享 CSS、JS 和 `tree.js` 只写一份。
- 默认不设业务数量上限。只在本次待重建文档超过 3,000 篇或 8 MiB 时弹窗确认：
  - 加载全部：取消数量、总量、深度、目录项和单文件限制；
  - 只加载 3,000 篇 / 8 MiB：应用快速模式限制；
  - 取消：不打开。
- 文件树改为方案 1：固定引导槽、缩进引导线、文件夹和文件字重区分、轻量选中态，不使用目录计数。
- 树数据由共享 `tree.js` 提供，侧栏按展开状态懒渲染，避免大目录一次性创建大量 DOM。
- 源码视图在文档页面之间保持，刷新页面时回到渲染视图。
- **BREAKING**：目录模式不再生成单个包含全部文档的 HTML；跨文档链接改为指向生成页面的相对 URL。

## Capabilities

### New Capabilities

- `document-pages`: 每篇文档的独立 HTML 输出、稳定缓存、并行生成、缓存复用和清理。

### Modified Capabilities

- `document-loading`: 硬性容量上限改为基于待重建工作量的软阈值确认；确认后可完全放开限制。
- `directory-browsing`: 文件树节点改为页面链接，采用新的层级视觉和懒渲染，并跨页面恢复展开状态。
- `markdown-rendering`: 跨文档链接指向生成页面；页内锚点和脚注不再需要文档前缀。
- `source-view`: 源码模式在页面导航时保持，刷新时重置。

## Impact

- 新增/修改：`internal/scan`、`internal/render`、`internal/tempdir`、`internal/ui`、`internal/app`。
- 页面资源：`internal/render/web/page.html`、`reader.css`、`reader.js`，新增树数据/共享资源配置。
- 测试：扫描、缓存、页面生成、链接重写、浏览器导航和源码模式。
- 文档：README、OpenSpec specs、验收清单。
- 不新增依赖，不引入 HTTP 服务或 iframe。
