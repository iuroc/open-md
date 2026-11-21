# Proposal

## Why

OpenMD 的目录模式只扫描当前层。像 OpenSpec 这样把文档按能力、变更和归档分散到多层子目录的目录，顶层往往没有任何 Markdown 文件，右键打开会直接得到“目录中没有 Markdown 文件”。用户只能逐层右键到最深处，无法把一组规格文档当作一个整体浏览。

## What Changes

- 目录模式改为递归扫描子目录中的 `.md` / `.markdown` 文件，并把每个文件的相对路径带入页面。
- 侧栏文件列表改为可展开、折叠的文件树：
  - 文件节点以文件名为主标签，文档标题和完整相对路径放入悬停提示。
  - 目录优先于文件，层级内按自然序排列。
  - 只显示包含 Markdown 的目录，空目录和无 Markdown 目录不进入树。
- 默认只展开当前文档的祖先路径；跨文档链接跳转后，自动展开目标路径、滚动到节点并高亮。
- 从目录树之外通过链接载入的文档单独归入“链接文档”分组。
- 打开目录时优先显示根目录的 `README.md` / `index.md`；没有时显示层级最浅的第一篇文档。
- 递归扫描保留现有 500 篇 / 25 MiB 总量 / 单文件大小上限，并补充：
  - 跳过符号链接、junction 和已知重目录（如 `.git`、`node_modules`、`vendor`）；
  - 无权限或读取失败的子目录只产生警告，不终止整次扫描；
  - 接近上限时按“浅层文档优先”的确定性顺序截断，并明确提示。
- **BREAKING**：目录模式不再只列出当前层，现有“目录模式不递归”的文档描述和验收预期同步更新。

## Capabilities

### New Capabilities

- `directory-browsing`: 目录模式下的递归文档发现、文件树展示、树内定位与截断提示。

### Modified Capabilities

- `document-loading`: 目录扫描从“只扫描当前层”改为递归扫描，并新增忽略规则、遍历预算和确定性截断要求。

## Impact

- 扫描：`internal/scan/scan.go` 及测试。
- 渲染：`internal/render/directory.go`、`internal/render/page.go`、`internal/render/links.go`（相对路径与树外链接文档分组）。
- 前端：`internal/render/web/page.html`、`reader.css`、`reader.js`（文件树、展开折叠、活动节点定位）。
- 文档与验收：`README.md`、`cases/CHECKLIST.md` 中“目录不递归”的描述和用例。
- 不新增依赖，不引入 HTTP 服务或多页面输出；仍是一次性生成单个静态 HTML。
