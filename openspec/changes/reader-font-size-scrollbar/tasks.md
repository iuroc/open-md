# Tasks

## 1. 字号样式与控件

- [x] 1.1 在 `internal/render/web/reader.css` 的 `.doc-section.markdown-body` 增加 `font-size: var(--reader-font-size, 16px)`，把 `.footnotes` 和 `.csv-data` 的固定 `12px` 改为 `.75em`，把 `.source-code.chroma` 的 `1rem` 改为 `var(--reader-font-size, 1rem)`；在 `internal/render/page_test.go` 增加 `TestReaderFontSizeCSS` 断言这些规则，运行 `go test ./internal/render -run TestReaderFontSizeCSS -count=1` 通过。
- [x] 1.2 在 `docpage.html` 和 `page.html` 的“设置”Tab 中，在“字体”和“字间距”之间加入“字号”单行分段控件（`data-reader-size` 为 `14`/`16`/`18`/`20`，文本“小/标准/大/特大”，默认 `16` 为 `aria-pressed="true"`）；增加 `TestReaderSettingsGroups` 的字号断言，运行 `go test ./internal/render -run TestReaderSettingsGroups -count=1` 通过。

## 2. 字号逻辑

- [x] 2.1 在 `reader.js` 增加字号预设映射（`14`/`16`/`18`/`20` → `14px`/默认值不写变量/`18px`/`20px`）、合法值校验和 `applyReaderSize`（设置或移除 `--reader-font-size`、更新 `aria-pressed`、按需写入 `openmd-reader-size`），初始化时读取后应用；增加 `TestReaderJSFontSizePresets` 断言键名、变量名和存储键，运行通过。
- [x] 2.2 在 `reader.js` 查询 `[data-reader-size]` 按钮并在初始化时绑定点击，点击后立即应用并持久化；增加 `TestReaderJSFontSizeWiring` 断言查询、绑定和初始化调用，运行通过。

## 3. 主题滚动条

- [x] 3.1 在 `reader.css` 增加 `--scrollbar-thumb` / `--scrollbar-thumb-hover` / `--scrollbar-track` token（静态 `rgba` 回退 + `@supports` 内的 `color-mix` 主题推导），以及 `::-webkit-scrollbar` 系列规则（10px、透明轨道、圆角滑块、hover 加深）；在 `@supports (-moz-appearance: none)` 内为 Firefox 设置 `scrollbar-color` 和 `scrollbar-width: thin`；增加 `TestThemedScrollbars` 断言 token、伪元素和 Firefox 分支，运行通过。
- [x] 3.2 确认 `body.split-mode .doc-section .source-view` 的隐藏滚动条规则仍然生效，且通用滚动条规则不影响它；在 `TestThemedScrollbars` 中断言该规则仍存在，运行通过。

## 4. 集成与验证

- [x] 4.1 更新 `README.md` 的设置说明，加入“字号”和随主题变化的滚动条；用 `grep -n "字号" README.md` 确认。
- [x] 4.2 提升 `internal/render/site.go` 的 `cacheGeneration`；运行 `go vet ./...` 和 `go test ./... -count=1` 全部通过。
- [x] 4.3 生成页面做浏览器验证：四档字号只影响内容区域、默认 16px 观感不变、标题/代码/脚注等比、刷新和软导航后保持；五套配色明暗下滚动条颜色、hover、分栏两栏和源码栏隐藏正常；记录每项结果。
- [x] 4.4 提交全部改动，提交信息说明字号预设与主题滚动条；确认 `git status` 干净。

## Workflow follow-up

- 实现完成并验证后归档 change，同步 `reader-appearance` 和 `reader-panel` 主 spec。
