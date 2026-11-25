# Spec Delta

## ADDED Requirements

### Requirement: Font size presets change content text only

“设置”Tab SHALL 提供正文字号预设：小 `14px`、标准 `16px`、大 `18px`、特大 `20px`，默认标准。选择任一预设 MUST 立即应用到内容区域，并保留到后续页面加载；侧边栏、顶栏、设置面板和目录的字号 MUST NOT 改变。

#### Scenario: Switch content font size

- **WHEN** 用户在“设置”Tab 中选择“大”
- **THEN** 渲染后的正文使用 `18px` 字号
- **AND** 侧边栏、顶栏、设置面板和目录的字号保持不变

#### Scenario: Content elements scale together

- **WHEN** 用户选择“特大”
- **THEN** 标题、列表、表格、行内代码和代码块随正文等比放大
- **AND** 脚注和 CSV 表格不再固定为 `12px`

#### Scenario: Font size survives navigation

- **WHEN** 用户设置字号后在文件树中切换到另一篇文档，或刷新页面
- **THEN** 新页面继续使用该字号

#### Scenario: First visit keeps the current size

- **WHEN** 用户在没有任何已保存字号偏好的情况下打开页面
- **THEN** 内容区域使用现有的 `16px` 字号

### Requirement: Scrollbars follow the active theme

阅读页的滚动条 SHALL 使用当前配色的弱化文字色作为滑块颜色，并随配色和明暗模式变化。滚动条 MUST 覆盖窗口主滚动区域、侧边栏、设置面板和分栏阅读的两栏；分栏源码栏 MUST 保持隐藏滚动条。滚动条外观 MUST NOT 改变滚动位置恢复或锚点导航行为。

#### Scenario: Scrollbar changes with theme

- **WHEN** 用户在海盐蓝和宣纸之间切换配色
- **THEN** 滚动条滑块分别呈现对应的冷灰和暖灰色调
- **AND** 滚动条轨道保持透明或与页面底色融合

#### Scenario: Scrollbar follows dark mode

- **WHEN** 系统处于深色模式
- **THEN** 滚动条滑块在深色背景上保持可见，并与当前配色协调

#### Scenario: Split view panes are themed

- **WHEN** 用户进入分栏阅读并滚动渲染正文或源码栏
- **THEN** 两栏的滚动条使用主题颜色
- **AND** 源码栏仍不显示滚动条

#### Scenario: Scroll behavior is unchanged

- **WHEN** 用户刷新页面或使用前进/后退
- **THEN** 滚动位置恢复与锚点导航的结果和改动前一致
