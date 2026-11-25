# Proposal

## Why

侧栏和设置面板里的「OpenMD」目前只是普通加粗 sans，和应用图标的 M 标记没有联系，品牌识别度弱。应用已经有稳定的 M 图标，阅读页也需要一个与图标一致、在小尺寸下仍清晰的字标。

## What Changes

- 新增统一的 OpenMD 字标：内联 SVG 的图标 M（复用现有图标折线）+ 衬线「OpenMD」（`Open` 字重 400、`MD` 字重 700）。
- 字标用于文档页侧栏标题和信息与设置面板标题；`docpage.html` 与 `page.html` 同步更新。
- M 使用主题的 `--accent`，文字使用 `--fg`，随现有配色和明暗模式自动适配。
- 字标在 14px 与 18px 下保持可读；SVG 仅作装饰并对辅助技术隐藏，可见文字仍是真实文本。
- 模板和共享 CSS 变化后提升 `cacheGeneration`，让已有渲染缓存重新生成。
- 不引入自定义字体文件，不改 favicon、exe 图标或浏览器标签标题，不添加动画、纹理或渐变。

## Capabilities

### New Capabilities

- `reader-branding`: OpenMD 字标在阅读页中的组成、出现位置、主题适配和小尺寸可读性。

### Modified Capabilities

无。字标只改变品牌名的呈现方式，不改变既有 spec 中「面板标题使用品牌名 OpenMD」等要求。

## Impact

- `internal/render/web/docpage.html`
- `internal/render/web/page.html`
- `internal/render/web/reader.css`
- `internal/render/page_test.go` 及模板相关测试
- `internal/render/site.go` 的 `cacheGeneration`
- 不影响文档内容渲染、导航交互、主题 token 或设置项行为。
