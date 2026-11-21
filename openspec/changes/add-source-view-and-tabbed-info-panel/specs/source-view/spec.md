# Spec Delta

## Purpose

让用户在阅读页内直接查看当前 Markdown 的原始内容，并在不离开阅读上下文的情况下复制源码。

## ADDED Requirements

### Requirement: Source view toggles from the topbar

系统 SHALL 在阅读页顶栏提供源码视图切换控件，使当前文档可以在渲染视图和源码视图之间切换。页面首次加载时 MUST 显示渲染视图；切换文档时 MUST 保持当前视图模式。

#### Scenario: Enter source view

- **WHEN** 用户点击顶栏的源码切换按钮
- **THEN** 当前文档显示带语法高亮的源码，而不是渲染后的正文

#### Scenario: Return to rendered view

- **WHEN** 用户再次点击源码切换按钮
- **THEN** 当前文档恢复显示渲染后的正文

#### Scenario: Switch documents in source mode

- **WHEN** 源码视图处于激活状态，用户切换到另一篇文档
- **THEN** 新文档同样显示源码视图，而不是渲染视图

#### Scenario: Reload page

- **WHEN** 用户刷新阅读页或重新打开文档
- **THEN** 页面显示渲染视图，而不是上次的源码视图

### Requirement: Source view shows the original highlighted Markdown

源码视图 SHALL 显示当前文档的原始 Markdown 文本，包括文档开头的 front matter，并使用语法高亮呈现。源码 MUST 只读，页面 MUST NOT 提供修改或保存源码的控件。

#### Scenario: Front matter is visible

- **WHEN** 文档以 `---` 或 `+++` 包围的 front matter 开头
- **THEN** 源码视图显示该 front matter 块，而渲染视图仍然不显示它

#### Scenario: Syntax highlighting

- **WHEN** 文档包含标题、强调、链接、列表或围栏代码块
- **THEN** 源码视图中这些结构按 Markdown 语法高亮显示

#### Scenario: Read-only source

- **WHEN** 用户查看源码视图
- **THEN** 页面不提供编辑、保存或格式化源码的控件

### Requirement: Source lines wrap

源码视图 SHALL 自动换行显示超长行，源码内容区 MUST NOT 产生横向滚动。

#### Scenario: Long paragraph wraps

- **WHEN** 文档包含一行超过阅读区宽度的长段落
- **THEN** 该行在源码视图中自动折行，不发生横向滚动

#### Scenario: Long URL wraps

- **WHEN** 文档包含一行超过阅读区宽度的长链接或连续字符
- **THEN** 该行在源码视图中自动折行，不发生横向滚动

### Requirement: Source can be copied

源码视图 SHALL 提供复制当前文档完整源码的操作，并在操作后给出成功或失败反馈。复制内容 MUST 与源码视图显示的内容一致，包括 front matter、缩进和空行。

#### Scenario: Copy full source

- **WHEN** 用户点击“复制源码”
- **THEN** 当前文档的完整源码被写入剪贴板，并显示“已复制”反馈

#### Scenario: Copy empty source

- **WHEN** 当前文档没有可复制的源码内容
- **THEN** 页面显示“没有可复制的源码”，不写入空内容

#### Scenario: Clipboard failure

- **WHEN** 浏览器拒绝写入剪贴板且回退复制也失败
- **THEN** 页面显示“复制失败”反馈

### Requirement: Source selection is scoped

源码视图中的文本 SHALL 可被选中。当键盘焦点位于源码区域内时，`Ctrl/Cmd + A` MUST 只选中该文档的源码，而不是整个页面。

#### Scenario: Select all inside source

- **WHEN** 焦点位于源码区域，用户按下 `Ctrl/Cmd + A`
- **THEN** 选中范围只包含当前源码内容

#### Scenario: Select all outside source

- **WHEN** 焦点不在源码区域，用户按下 `Ctrl/Cmd + A`
- **THEN** 保持浏览器的默认全选行为

### Requirement: Missing source has an explicit state

当文档没有可用源码时，源码视图 SHALL 显示明确的空状态，并 MUST NOT 提供复制操作。

#### Scenario: Read failure in source mode

- **WHEN** 文档读取失败，用户进入源码视图
- **THEN** 页面显示“该文档没有可显示的源码”，且不显示复制源码按钮

### Requirement: Switching views preserves position

在渲染视图和源码视图之间切换时，页面 SHALL 保持切换前的滚动位置，并按新内容高度做合理截断。

#### Scenario: Toggle mid-document

- **WHEN** 用户在长文档中部切换到源码视图
- **THEN** 页面停留在相近的滚动位置，不会跳回顶部或跳到文档末尾

#### Scenario: Return to rendered view

- **WHEN** 用户从源码视图切回渲染视图
- **THEN** 页面同样停留在相近的滚动位置

### Requirement: Outline navigation leaves source mode

当源码视图处于激活状态时，点击大纲项 SHALL 先回到渲染视图，再滚动到对应标题。

#### Scenario: Click outline in source mode

- **WHEN** 源码视图处于激活状态，用户点击侧栏大纲中的标题
- **THEN** 页面切换回渲染视图，并滚动到该标题
