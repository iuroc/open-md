# Tasks

## 1. 宣纸主题

- [x] 1.1 在 `internal/render/web/reader.css` 增加 `html[data-theme="paper"]` 明暗 token、`.theme-dot-paper` 和 `.theme-option[data-theme="paper"]` 整行规则（色值见 design.md）；在 `internal/render/page_test.go` 增加 `TestPaperThemeStyles`，断言浅色 `--bg: #f6f0e3`、暗色 `#1a1814`、色点和 `grid-column: 1 / -1`，运行 `go test ./internal/render -run TestPaperThemeStyles -count=1` 通过。
- [x] 1.2 在 `docpage.html` 和 `page.html` 的头部主题白名单加入 `paper`，在 `docpage.html` 头部内联样式补宣纸明暗初始背景，并在两个模板的配色网格加入“宣纸”按钮；在 `reader.js` 的 `validThemes` 加入 `"paper"`；增加 `TestPaperThemeRegistration` 断言模板、白名单和 `validThemes`，运行 `go test ./internal/render -run TestPaperThemeRegistration -count=1` 通过。
- [x] 1.3 更新 `README.md` 的配色说明为五套并写明“宣纸”，用 `grep -n "宣纸" README.md` 确认。

## 2. 正文排版

- [x] 2.1 在 `reader.css` 的 `.doc-section.markdown-body` 增加 `font-family: var(--reader-font, var(--font-body));` 和 `letter-spacing: var(--reader-letter-spacing, 0);`，并为 `pre`、`code`、`kbd`、`samp` 重置等宽字体与 `letter-spacing: normal`；增加 `TestReaderTypographyCSS` 断言这些声明，运行通过。
- [x] 2.2 在 `reader.js` 增加字体预设（默认/宋体/楷体/仿宋）、字间距预设（`0`/`0.03`/`0.06`/`0.1`）、合法值校验和 `applyReaderFont` / `applyReaderSpacing`（设置 `--reader-font` / `--reader-letter-spacing`、更新按钮 `aria-pressed`、按需写入 `localStorage`），初始化时读取 `openmd-reader-font` / `openmd-reader-spacing` 后应用；增加 `TestReaderJSTypographyPresets` 断言键名、预设值和变量名，运行通过。

## 3. 设置面板

- [x] 3.1 在 `docpage.html` 和 `page.html` 的“设置”Tab 加入“字体”和“字间距”两组 2×2 胶囊按钮（`data-reader-font` / `data-reader-spacing`，`aria-pressed`），在 `reader.css` 增加 `.setting-grid` / `.setting-option` 样式（或与 `.theme-option` 共享规则）；增加 `TestReaderSettingsGroups` 断言两组控件，运行通过。
- [x] 3.2 在 `reader.js` 查询并在初始化时绑定 `[data-reader-font]` / `[data-reader-spacing]` 按钮，点击后立即应用并持久化；增加 `TestReaderJSTypographyWiring` 断言查询和事件绑定，运行通过。
- [x] 3.3 更新 `README.md` 的设置说明为“配色、字体、字间距”，用 `grep -n "字体" README.md` 确认。

## 4. 集成与缓存

- [x] 4.1 提升 `internal/render/site.go` 的 `cacheGeneration` 值；运行 `go vet ./...` 和 `go test ./... -count=1` 全部通过。
- [x] 4.2 用生成页面在浏览器验证：宣纸明暗两版、四种字体、四档字距、刷新与软导航后保持、代码块和源码视图不受影响、默认观感不变；记录每项结果。
- [ ] 4.3 提交全部改动，提交信息说明宣纸主题与正文排版设置；确认 `git status` 干净。
