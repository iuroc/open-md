# Tasks

详细步骤见 `docs/superpowers/plans/2026-11-25-builtin-manual.md`。

## 1. 手册包

- [ ] 1.1 新增 `render.FileURL`，以及 `internal/manual` 的内容指纹、原子提取和 `.complete` 标记；用 `fstest.MapFS` 覆盖指纹变化、完整提取和残缺目录恢复。
- [ ] 1.2 实现 `manual.Ensure`：内容指纹、锁文件串行化、`current.json` 复用、`BuildDirSite` 生成、只清理超过 24 小时的非当前版本；覆盖首次生成、复用、锁超时、陈旧锁和清理。

## 2. 阅读页入口

- [ ] 2.1 `SiteOptions` / `DocPageData` / `Page` 增加 `ManualURL`，`siteManifest` 参与缓存失效；`docpage.html` 和 `page.html` 的「关于」Tab 增加「打开使用手册」，新标签页打开。
- [ ] 2.2 提升 `cacheGeneration`，补模板和站点缓存失效测试。

## 3. 应用集成

- [ ] 3.1 `App` 增加可注入的 `Manual`，`openPage` 在每个尝试的缓存基目录内先生成/复用手册，失败时隐藏入口；`Main` 接入 `manual.Ensure`。
- [ ] 3.2 补应用测试：正常显示入口、生成失败隐藏入口、主缓存不可写时在回退目录生成。

## 4. 手册内容

- [ ] 4.1 编写 `index.md` 和 13 个章节，统一「你会学到 / 基本写法 / 在 OpenMD 里怎么看 / 常见错误 / 延伸阅读」结构；正文原创，延伸阅读只放外部链接。
- [ ] 4.2 补内容契约测试：文件齐全、单 H1、五个小节、相对链接可达、无远程图片、首页链接全部章节。

## 5. 文档与集成验证

- [ ] 5.1 更新 README 的能力说明和 AGENTS 的手册缓存布局。
- [ ] 5.2 运行 `go vet ./...`、`go test ./... -count=1`，用真实页面验证「关于 → 打开使用手册」、章节跳转、源码/双视图和入口隐藏。

## 验证记录

待实现完成后补充。
