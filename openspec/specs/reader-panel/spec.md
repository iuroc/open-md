# reader-panel Specification

## Purpose

让阅读页的信息与设置面板按稳定的分类展示文件信息、阅读设置和关于信息，并随设置项增长保持可导航。

## Requirements

### Requirement: Panel uses categorized tabs

信息与设置面板 SHALL 提供“文件 / 设置 / 关于”三个 Tab，同一时间 MUST 只显示一个 Tab 的内容。首次打开 MUST 显示“文件”；同一次页面会话内关闭再打开 SHALL 保留上次停留的 Tab；页面重新加载后 MUST 回到“文件”。

#### Scenario: Open the panel for the first time

- **WHEN** 用户首次打开信息与设置面板
- **THEN** “文件”Tab 处于选中状态，其他 Tab 内容不可见

#### Scenario: Switch tabs

- **WHEN** 用户点击“设置”或“关于”
- **THEN** 只有对应 Tab 的内容显示，其他内容隐藏

#### Scenario: Reopen during the same session

- **WHEN** 用户停留在“关于”Tab，关闭面板后再次打开
- **THEN** 面板仍停留在“关于”Tab

#### Scenario: Reload page

- **WHEN** 用户刷新页面后重新打开面板
- **THEN** 面板回到“文件”Tab

### Requirement: File tab shows active document properties

“文件”Tab SHALL 显示当前文档的绝对路径和所在目录，并在切换文档时更新。该 Tab SHALL 提供“复制路径”操作，并反馈复制成功或失败。

#### Scenario: Switch documents

- **WHEN** 用户在目录模式或链接文档模式下切换到另一篇文档
- **THEN** “文件”Tab 中的路径和目录更新为新文档的信息

#### Scenario: Copy path

- **WHEN** 用户点击“复制路径”
- **THEN** 当前文档的绝对路径被写入剪贴板，并显示“已复制”反馈

#### Scenario: No path available

- **WHEN** 当前没有可复制的文档路径
- **THEN** 页面显示“没有可复制的路径”，不写入空内容

### Requirement: Settings tab owns reading preferences

“设置”Tab SHALL 集中展示阅读相关的设置项，当前至少包含配色选择。选择配色 MUST 立即生效，并在后续页面加载中保留；未来的设置项 SHALL 作为该 Tab 内的分组继续扩展。

#### Scenario: Change theme

- **WHEN** 用户在“设置”Tab 中选择另一套配色
- **THEN** 页面立即应用该配色，重新打开页面后仍然使用该配色

#### Scenario: Settings remain grouped

- **WHEN** 后续增加新的阅读设置
- **THEN** 新设置出现在“设置”Tab 内，而不是新增一个顶级 Tab

### Requirement: About tab shows identity and build information

“关于”Tab SHALL 显示阅读器简介、作者、版本和构建时间。正式构建 SHALL 提供发布页链接；开发构建 MUST 显示未提供发布页的说明。

#### Scenario: Release build

- **WHEN** 用户打开正式构建的“关于”Tab
- **THEN** 页面显示发布页链接和更新提示

#### Scenario: Development build

- **WHEN** 用户打开开发构建的“关于”Tab
- **THEN** 页面显示开发构建未提供发布页的说明，不显示发布页链接

### Requirement: Panel keyboard navigation

面板 Tab SHALL 支持鼠标和键盘操作。左右方向键 SHALL 在 Tab 间移动，Home/End SHALL 跳到首尾；当前 Tab MUST 通过 `aria-selected` 和可见焦点反馈。

#### Scenario: Move with arrow keys

- **WHEN** 焦点位于某个 Tab，用户按下左方向键或右方向键
- **THEN** 焦点和选中状态移动到相邻 Tab，并显示对应内容

#### Scenario: Focus is visible

- **WHEN** 用户用键盘聚焦某个 Tab
- **THEN** 该 Tab 显示清晰的焦点轮廓

### Requirement: Panel layout keeps navigation visible

面板的标题栏和 Tab 栏 MUST 保持固定，只有 Tab 内容区滚动；内容高度变化 MUST NOT 把 Tab 栏移出视野。

#### Scenario: Long settings content

- **WHEN** “设置”Tab 的内容超过面板可用高度
- **THEN** 内容区独立滚动，标题栏和三个 Tab 仍然可见

### Requirement: Panel labels stay concise

面板标题 SHALL 使用品牌名 `OpenMD`；各 Tab 内容 MUST NOT 以“文件”“设置”“关于”作为重复的分组标题。

#### Scenario: Inspect each tab

- **WHEN** 用户依次打开三个 Tab
- **THEN** 内容区不出现与当前 Tab 名称重复的分组标题
