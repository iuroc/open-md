# Spec Delta

## Purpose

定义阅读页侧栏的视图切换与滚动方式，使文件列表和大纲在内容很长时仍能直接切换，让大纲当前条目始终出现在侧栏视野内，并让当前 Tab 的选择在当前会话内保留。

## ADDED Requirements

### Requirement: Sidebar view tabs stay reachable

侧栏的视图 Tab SHALL 固定在侧栏顶部，只有当前 Tab 的列表区域滚动。文件列表或大纲滚动任意距离后，Tab MUST 仍然可见并可直接点击切换。

#### Scenario: Long file list

- **WHEN** 用户在文件列表中滚动到中间或底部
- **THEN** 文件列表和大纲 Tab 仍固定在侧栏顶部
- **AND** 用户可以直接点击“大纲”切换视图

#### Scenario: Long outline

- **WHEN** 用户在大纲中滚动到中间或底部
- **THEN** 两个 Tab 仍固定在侧栏顶部
- **AND** 用户可以直接点击“文件列表”切换视图

### Requirement: Outline active item stays visible

内容滚动导致大纲高亮切换时，侧栏大纲列表 SHALL 以最小幅度滚动，使新的高亮项完整可见；高亮项已可见时 MUST NOT 移动列表。该滚动 MUST NOT 改变正文滚动位置或当前高亮标题。切换到“大纲”Tab 时，当前高亮项 SHALL 可见。

#### Scenario: Scroll reading content

- **WHEN** 用户滚动阅读区，高亮从一个标题切换到另一个标题
- **THEN** 大纲列表只在需要时滚动，使新的高亮项完整可见
- **AND** 阅读区的滚动位置和高亮标题保持不变

#### Scenario: Scroll source pane

- **WHEN** 双视图下用户滚动源码区，大纲高亮切换
- **THEN** 大纲列表以最小幅度滚动到新的高亮项
- **AND** 源码区的滚动位置不变

#### Scenario: Active item already visible

- **WHEN** 新的高亮项已经完整位于大纲列表的可见范围内
- **THEN** 大纲列表不滚动

#### Scenario: Switch to outline tab

- **WHEN** 用户在大纲列表被滚动到别处后切回“大纲”Tab
- **THEN** 当前高亮项立即出现在可见范围内

### Requirement: Sidebar tab choice survives refresh

侧栏当前视图 Tab 的选择 SHALL 在同一浏览器标签页的会话内保留：刷新页面或在该标签页内切换文档后，SHALL 恢复上次选择的 Tab。没有保存过选择、记录无效，或记录对应的 Tab 在当前页面不存在时，MUST 保持默认：有文件列表时显示“文件列表”，没有文件列表时显示“大纲”。程序自动切换 MUST NOT 覆盖已记录的选择。关闭浏览器后 MUST NOT 保留选择。

#### Scenario: Reload keeps the outline tab

- **WHEN** 用户切换到“大纲”Tab 后刷新页面
- **THEN** 页面仍停留在“大纲”Tab

#### Scenario: Switch documents in the same tab

- **WHEN** 用户在“大纲”Tab 下切换到另一篇文档
- **THEN** 新文档页面仍停留在“大纲”Tab

#### Scenario: No saved choice

- **WHEN** 用户在当前浏览器标签页第一次打开页面
- **THEN** 页面显示“文件列表”；当前页面没有文件列表时显示“大纲”

#### Scenario: Saved tab is unavailable

- **WHEN** 保存的选择是“文件列表”，但当前文档没有文件列表
- **THEN** 页面显示“大纲”
- **AND** 之后打开带有文件列表的文档时，仍恢复为“文件列表”

#### Scenario: New browser session

- **WHEN** 用户关闭浏览器后重新打开页面
- **THEN** 页面回到“文件列表”，不沿用上一次会话的选择
