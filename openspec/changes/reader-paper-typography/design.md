# Design

## Context

- 当前阅读页有四套配色（`blue` / `teal` / `violet` / `graphite`），token 定义在 `reader.css` 的 `:root` 和 `html[data-theme="..."]`，选择结果存 `localStorage` 的 `openmd-theme`。
- 正文由 `github-markdown.css` 的 `.markdown-body` 设置 sans 字体栈和 `line-height: 1.5`，覆盖了 `body` 的 `--font-body`；调整正文排版需要在更具体的选择器上覆盖。
- “设置”Tab 在 `docpage.html` 和 `page.html` 中，现有配色使用两列胶囊按钮；`reader.js` 负责按钮状态和持久化。
- 页面运行在 `file://` 且受 CSP 限制，不能可靠加载网络字体；`reader.js` 随页面缓存复用，改动后需要提升 `cacheGeneration`。
- `reader-panel` 的现有要求已经预留下设分组，但只明确包含配色。

## Goals / Non-Goals

**Goals:**

- 新增 `paper`（宣纸）主题，明暗两版与现有主题同构。
- 字体和字符间距预设只作用于渲染正文，代码块和源码视图保持等宽、正常字距。
- 偏好持久化，刷新和软导航后保持；未设置时默认观感不变。

**Non-Goals:**

- 不增加字号、行高、页边距设置。
- 不支持自定义字体输入，不打包 webfont，不做纸张纹理或噪点。
- 不修改现有四套配色。

## Decisions

### 1. 排版通过 `html` 上的 CSS 变量应用

在 `html` 上设置 `--reader-font` 和 `--reader-letter-spacing`，正文选择器使用它们；设置按钮只切换变量值。

- 相比给每个 `.doc-section` 写 inline style：软导航替换文档内容后不需要重新套用。
- 相比 `html[data-reader-font="song"]` 之类的属性规则：新增预设不需要复制整套 CSS。
- 默认字体通过移除自定义属性让 `var(--reader-font, var(--font-body))` 回退，避免空值产生无效声明。

### 2. 宣纸配色 token

| token | 明 | 暗 |
|---|---|---|
| `--bg` | `#f6f0e3` | `#1a1814` |
| `--bg-sidebar` | `#efe7d6` | `#211e18` |
| `--bg-subtle` | `#f1ebdd` | `#26221c` |
| `--bg-hover` | `#e7ddc9` | `#2e2922` |
| `--fg` | `#2f2a23` | `#e6ddcb` |
| `--fg-muted` | `#7c705e` | `#a89b85` |
| `--border` | `#ddd2bd` | `#3a342a` |
| `--accent` | `#3f5d6b` | `#93b3c0` |
| `--accent-fg` | `#ffffff` | `#10171b` |
| `--accent-soft` | `#e4ecef` | `#24343b` |

- 强调色用黛青而不是赭石，避免“暖米色 + 陶土色”的通用搭配。
- 在头部内联样式中补宣纸的初始背景，并在主题白名单中加入 `paper`，避免样式加载前露白或被忽略。
- 宣纸按钮在配色网格中横跨整行，避免第五项单独落在半格。

### 3. 控件使用胶囊按钮预设

字体和字间距各用 2×2 胶囊按钮，与现有配色按钮共享样式。

- `<select>` 更紧凑，但不如按钮可扫视；range 适合连续值，而字间距只需要几个明确档位。
- 按钮使用 `data-reader-font` / `data-reader-spacing` 和 `aria-pressed`，点击后立即应用并持久化。

### 4. 字体栈和回退

- 默认：沿用 `--font-body`。
- 宋体：`"SimSun", "宋体", "Songti SC", "Noto Serif CJK SC", serif`
- 楷体：`"KaiTi", "楷体", "Kaiti SC", STKaiti, serif`
- 仿宋：`"FangSong", "仿宋", "FangSong_GB2312", STFangsong, serif`

字体未安装时回退到通用 `serif`，不报错、不阻断阅读。

### 5. 字符间距和代码隔离

- 正文容器 `letter-spacing` 使用 `--reader-letter-spacing`；预设为 `0` / `0.03em` / `0.06em` / `0.10em`。
- `pre`、`code`、`kbd`、`samp` 显式重置为等宽字体和 `letter-spacing: normal`。
- 源码视图不在 `.markdown-body` 内，保持现有等宽字体和正常字距。

### 6. 持久化与加载时机

- 键：`openmd-reader-font`（`default|song|kai|fang`）、`openmd-reader-spacing`（`0|0.03|0.06|0.1`）。
- 读取时校验预设值，非法值回退默认；沿用 `readPref` / `writePref` 的异常容错。
- 在页面 reveal 之前应用，避免闪动；软导航只替换正文 DOM，`html` 上的变量不受影响。

### 7. 缓存代次

`reader.js`、`reader.css` 和页面模板都参与缓存复用，提升 `cacheGeneration` 让已有页面重新生成。

## Risks / Trade-offs

- **字体不可用** → 每个预设都以通用 `serif` 收尾；必要时用户可切回默认。
- **字距过宽影响窄屏** → 预设上限 `0.10em`，并保持代码区正常字距。
- **宣纸暗色对比不足** → 使用表中已验证的墨色/纸色组合，浏览器验收时检查正文、弱化文字和链接。
- **`localStorage` 不可用** → 偏好不持久化，但页面功能和默认观感不受影响。
- **两套模板分叉** → `docpage.html` 与 `page.html` 同步更新，并用模板测试同时覆盖。
- **旧偏好兼容** → 新增键独立于 `openmd-theme`，现有用户不会被迫切换主题。

## Migration Plan

- 无需数据迁移；新偏好键缺省即默认。
- 提升缓存代次后，已有页面会在下次打开时重新生成。
- 回滚时恢复旧版 `reader.js` / `reader.css` / 模板即可；旧代码会忽略新增的 `localStorage` 键。

## Open Questions

无。行高与字号明确留到后续变更。
