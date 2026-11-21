# Tasks

## 1. Scan limits and confirmation

- [x] 1.1 先写扫描测试：无限制模式不按 500 篇/总量截断，快速模式仍保留确定性截断；实现 `Limits` 的无限制语义并运行测试通过
- [x] 1.2 先写单文件确认测试：超过 32 MiB 时调用确认回调，确认后读取，取消时不读取；实现后测试通过
- [x] 1.3 实现待重建工作量估计（文档数、字节数）与 3,000 篇 / 8 MiB 软阈值判断，补单元测试

## 2. Site cache and page generation

- [x] 2.1 先写稳定路径测试：同一源路径重复生成使用同一 root 目录和 HTML 文件名；实现 root key、文档 hash 和原子覆盖
- [x] 2.2 先写 manifest 复用测试：mtime/size/version/known-set 未变时跳过渲染，单篇变化只重建该篇；实现后测试通过
- [x] 2.3 先写清理测试：删除源文档后旧页面被清理，旧 root 按 TTL/数量清理；实现后测试通过
- [x] 2.4 实现并行渲染 worker pool，单篇失败生成错误页且不影响其他文档；补测试

## 3. Single-document page and shared assets

- [x] 3.1 重构页面模板为单文档页面，使用共享 CSS/JS 和 `tree.js`，补充模板测试
- [x] 3.2 生成 `tree.js` 数据，包含目录树、当前文档、链接文档和展开状态；补单元测试
- [x] 3.3 跨文档链接重写为 `doc-<hash>.html#anchor`，页内锚点和脚注保持本地；补链接重写测试
- [x] 3.4 生成链接文档页面，并确保单篇渲染失败只影响该页

## 4. File tree and browser interactions

- [x] 4.1 实现方案 1 的文件树 CSS：固定引导槽、缩进引导线、文件夹/文件字重和轻量选中态；补 CSS 测试
- [x] 4.2 实现 `reader.js` 的懒渲染树、展开状态 sessionStorage、当前路径高亮和跨页面恢复；补 JS/浏览器验证
- [x] 4.3 实现源码模式跨页面保持、刷新重置；补浏览器验证
- [x] 4.4 验证浏览器前进/后退、跨页锚点、TOC、主题、全宽、侧栏和信息面板
- [x] 4.5 保持文件树滚动位置：在树链接点击、滚动和 pagehide 时保存，跨页面恢复；用浏览器验证

## 5. Application integration and confirmation dialog

- [x] 5.1 扩展 `ui.Dialog`：增加加载全部/快速模式/取消确认；实现 Windows MessageBox 并补测试
- [x] 5.2 接入 `app`：扫描、确认、生成站点、打开初始页面；删除旧的单页 run 目录写入逻辑并保持安装/卸载流程
- [x] 5.3 更新 `tempdir.Manager`：新增稳定的 rendered root 目录和清理策略；补测试

## 6. Documentation and verification

- [x] 6.1 更新 README：独立页面、缓存、软阈值、快速模式、源码模式和文件树说明
- [x] 6.2 更新验收清单和案例说明，移除旧“单页/500 篇”预期
- [x] 6.3 运行 `go test ./... -count=1`、`go vet ./...`，确认全部通过
- [x] 6.4 用真实 `openspec/` 和 3,000+ 篇合成目录验证首次生成、缓存复用、确认弹窗、树导航和源码模式
- [x] 6.5 提交所有改动并执行 `build.bat`，确认 `OpenMD.exe` 和 `source.git.bundle` 更新
