# Tasks

## 1. Recursive scanning core

- [x] 1.1 在 `internal/scan/scan_test.go` 先写递归扫描测试：根层无 Markdown 但子目录有、深层文件、`.md` / `.MD` / `.markdown` 混合；运行测试确认失败
- [x] 1.2 改写 `scan.Directory` 为深度优先递归，返回相对路径；运行 1.1 的测试通过，并确认 `go test ./internal/scan` 全绿
- [x] 1.3 先写忽略规则测试：`.git`、`node_modules`、`vendor`、符号链接/junction 不进入结果；实现忽略与重解析点跳过，运行测试通过
- [x] 1.4 先写遍历预算测试：最大深度、最大访问条目数触顶时保留已发现文档并产生警告；实现预算并运行测试通过
- [x] 1.5 先写截断顺序测试：500 篇/总字节上限下浅层文档优先、同层自然序、重复扫描结果一致；实现确定性截断并运行测试通过
- [x] 1.6 先写超限嵌套文件测试：警告使用相对路径且不中断扫描；调整警告文案并运行测试通过

## 2. Directory page and tree model

- [x] 2.1 先写 `BuildDirPage` 集成测试：递归文档带正确相对路径、根 README 优先为初始文档、无 README 时选最浅第一篇；实现后测试通过
- [x] 2.2 在 `internal/render` 新增树节点模型与构建函数，先写单元测试覆盖目录优先、自然序、只保留含 Markdown 的目录；实现后测试通过
- [x] 2.3 先写页面数据测试：树只包含扫描根内文档，根外链接文档进入独立分组；实现后测试通过
- [x] 2.4 更新 `internal/render/directory.go` 和相关测试，确保原有“跳过超限文件、继续加载”的行为不回归

## 3. Sidebar file tree UI

- [x] 3.1 先写模板测试：目录模式渲染嵌套树、文件节点显示 basename、目录节点带 `aria-expanded`、文件节点带选中状态；实现 `page.html` 后测试通过
- [x] 3.2 先写样式测试：树缩进、标签省略、活动节点高亮、链接文档分组存在；补充 `reader.css` 并运行测试通过
- [x] 3.3 实现 `reader.js` 的目录展开/折叠、切换文档后展开祖先并滚动定位；用页面测试或 DOM 断言验证目标节点变为可见
- [x] 3.4 验证键盘操作：目录和文件控件可聚焦，Enter/空格可切换或打开，`aria-expanded` 同步更新；补充相应测试

## 4. Documentation and acceptance cases

- [x] 4.1 更新 `README.md` 的目录模式、已知限制和文档间链接说明，确认不再出现“不递归”的旧描述
- [x] 4.2 更新 `cases` 生成器与 `CHECKLIST.md` 中“目录不递归”的验收项，改为递归文件树、同名文件、深层路径和空目录预期；运行相关测试通过

## 5. Integration verification

- [x] 5.1 运行 `go test ./...` 和 `go vet ./...`，确认全部通过
- [x] 5.2 用真实 `openspec/` 目录生成页面，检查文件树、初始文档、跨文档链接定位和 500 篇截断提示；记录验证结果
- [ ] 5.3 提交所有改动并执行 `build.bat`，确认 `OpenMD.exe` 和 `source.git.bundle` 更新成功

## Workflow follow-up

- 完成审阅后归档该 change，并同步 `document-loading` 与 `directory-browsing` 主 spec。
