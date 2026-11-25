# Design

## Context

- 正文由 `github-markdown.css` 的 `.markdown-body` 设置 `font-size: 16px` 和 `line-height: 1.5`；阅读页在更具体的 `.doc-section.markdown-body` 上覆盖字体和字距。
- GitHub Markdown CSS 的标题、列表、表格和代码使用 `em`/`%`，只要改变 `.markdown-body` 的字号就会整体等比缩放；脚注和 CSV 表格是仅有的两处固定 `12px`。
- 源码视图的 `.source-code.chroma` 使用 `font-size: 1rem`，相对根元素而不是正文容器。
- 阅读偏好保存在 `localStorage`，`reader.js` 在首帧 reveal 前读取并应用；页面由 `site.go` 按 `cacheGeneration` 复用。
- 页面在系统默认浏览器中以 `file://` 打开，可能是 Chromium（Chrome/Edge）或 Firefox，不能假设只有一种滚动条实现。

## Goals / Non-Goals

**Goals:**

- 字号预设只缩放内容区域，默认档位与当前观感完全一致。
- 滚动条颜色由主题 token 推导，五套配色和明暗模式自动适配，无需逐主题维护。
- 字号和滚动条样式不改动滚动行为、锚点导航和滚动位置恢复。

**Non-Goals:**

- 不提供连续字号滑杆、自定义像素输入或行高设置。
- 不改变侧边栏、顶栏、目录和设置面板的排版。
- 不为了滚动条引入 JavaScript 或第三方库。

## Decisions

### 1. 字号通过正文容器上的 CSS 变量应用

在 `html` 上设置 `--reader-font-size`，正文规则使用 `font-size: var(--reader-font-size, 16px)`。

- 相比 `html { font-size: ... }`：根字号会同时改变所有 `rem` 单位，波及顶栏、面板和图标；正文容器方案天然只影响内容。
- 相比给每个 `.doc-section` 写 inline style：软导航替换 DOM 后无需重新套用。
- 相比 `zoom`/`transform`：不改变布局计算，滚动位置和锚点定位保持稳定。

字号写入 `.doc-section.markdown-body` 后：

- 标题（`2em` 等）、列表、表格、行内代码（`85%`）和代码块（`85%`）自动等比。
- `.markdown-body .footnotes` 与 `.csv-data` 的固定 `12px` 改为 `.75em`，默认仍是 `12px`。
- 源码视图的 `font-size: 1rem` 改为 `var(--reader-font-size, 1rem)`，默认仍是 `16px`。
- 行高为 unitless `1.5`，随字号等比，无需额外规则。

### 2. 字号档位和持久化

- 档位：`14` / `16` / `18` / `20`，对应“小 / 标准 / 大 / 特大”，默认 `16`。
- 键：`openmd-reader-size`；读取时校验档位，非法值回落 `16`。
- 控件沿用单行分段控件（4 段共享外框），放在“字体”和“字间距”之间。
- 初始化时与字体、字间距一起在首帧 reveal 前应用，避免字号跳变。

### 3. 滚动条颜色由主题 token 推导

```css
:root {
  --scrollbar-thumb: color-mix(in srgb, var(--fg-muted) 42%, transparent);
  --scrollbar-thumb-hover: color-mix(in srgb, var(--fg-muted) 62%, transparent);
  --scrollbar-track: transparent;
}
```

- 使用 `--fg-muted` 而不是 `--accent`：滑块保持中性，不抢正文和链接的视觉层级；同时冷暖随主题自然变化（海盐蓝偏冷灰，宣纸偏暖灰）。
- 使用 `color-mix` 而不是为十种明暗组合逐一写死色值：新增主题时自动继承，避免遗漏。
- 提供静态 `rgba(100, 116, 139, 0.42)` 作为 `color-mix` 不支持时的回退。

### 4. Chromium 用 WebKit 伪元素，Firefox 用标准属性

- Chromium/Edge/Safari 使用 `::-webkit-scrollbar`：宽 `10px`、透明轨道、透明边框加 `background-clip: padding-box` 形成圆角浮起滑块，`hover` 时加深。
- Firefox 不支持 WebKit 伪元素，在 `@supports (-moz-appearance: none)` 内使用 `scrollbar-color` 和 `scrollbar-width: thin`。
- 不采用“标准属性 + WebKit 伪元素同时生效”的写法：Chromium 121+ 只要设置了 `scrollbar-color`/`scrollbar-width`，就会忽略 `::-webkit-scrollbar`，hover 和圆角失效。按浏览器分派可以同时保住两边的效果。
- 不改变现有的 `body.split-mode .doc-section .source-view` 隐藏滚动条规则；其选择器更具体，会覆盖通用滑块样式。

### 5. 滚动条覆盖所有滚动容器

无选择器的 `::-webkit-scrollbar` 会匹配页面内所有滚动元素，包括窗口主滚动条、侧边栏、设置面板和分栏两栏，不需要逐个容器声明。

### 6. 缓存代次

`reader.css`、`reader.js` 和两个页面模板都参与缓存复用，提升 `cacheGeneration` 让已有页面重新生成并加载新样式。

## Risks / Trade-offs

- **`color-mix` 不可用** → 先用静态 `rgba` 作为 token 初始值，再用 `@supports (color: color-mix(in srgb, red, blue))` 覆盖为主题推导色。
- **WebKit 伪元素在 Firefox 无效** → Firefox 走标准属性分支，仍能获得主题颜色。
- **细滚动条更难用鼠标抓取** → 保留 10px 点击区域和 2px 透明内边距，滚轮、键盘和触控板不受影响。
- **字号放大后标题过大** → 标题保持与正文相同的缩放比例，和现有 GitHub Markdown 排版一致；不做单独限制。
- **打印时沿用阅读字号** → 打印样式不改字号，用户选择的阅读尺寸会带到打印结果；这是可接受的阅读偏好延续。
- **两个模板分叉** → `docpage.html` 与 `page.html` 同步增加字号分组，并用模板测试同时覆盖。

## Migration Plan

- 无需数据迁移；新增 `openmd-reader-size` 缺省即 `16px`。
- 提升 `cacheGeneration` 后，已有页面在下次打开时重新生成。
- 回滚时恢复旧版 `reader.css` / `reader.js` / 模板即可；旧代码会忽略新增的 `localStorage` 键。

## Open Questions

无。
