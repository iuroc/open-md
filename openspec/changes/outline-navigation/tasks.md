# Tasks

## 1. 大纲行号数据

- [ ] 1.1 在 `Heading` 上记录 0 起算的原始源码行号：处理 BOM 和 front matter 的前缀偏移，并补充含 front matter / BOM 的单元测试，运行 `go test ./internal/render` 通过
- [ ] 1.2 在 `docpage.html` 和 `page.html` 的大纲模板输出 `data-toc-line`，并更新渲染测试断言两个模板都包含行号

## 2. 前端定位与高亮

- [ ] 2.1 在 `reader.js` 实现源码行到滚动位置的换算：用 DOM Range 处理自动换行，按 section 和宽度缓存，验证源码区高位行和折行标题都能得到合理位置
- [ ] 2.2 重写大纲点击处理：阅读模式定位阅读区，双视图同时定位阅读区和源码行，源码模式保持源码模式并定位源码行；验证三种模式下点击同一大纲项后高亮项都是被点击的标题
- [ ] 2.3 调整 `updateActiveTOC()`：双视图以当前滚动面板顶部为原点，源码模式不再提前返回，源码区滚动时按源码行位置更新高亮；验证滚动阅读区、源码区和源码模式页面时高亮都跟随当前可见标题
- [ ] 2.4 在 `reader.css` 只为新的定位偏移做必要调整，并用 290px 侧栏、双视图和源码模式截图确认大纲与内容可见

## 3. 缓存与集成验证

- [ ] 3.1 将 `cacheGeneration` 提升到新的唯一值，运行 `go vet ./...` 和 `go test ./... -count=1` 通过
- [ ] 3.2 用真实 Markdown 渲染页面，在阅读、双视图、源码三种模式下点击和滚动大纲，确认定位、对齐和高亮一致；记录检查结果

## Workflow follow-up

- 在项目的 review 要求满足后归档该变更。
- 归档后确认 `openspec/specs/outline-navigation/spec.md` 已生成，并确认 `source-view` 的旧要求已移除。
