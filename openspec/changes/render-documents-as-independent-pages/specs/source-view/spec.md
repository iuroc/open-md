# Spec Delta

## MODIFIED Requirements

### Requirement: Source view toggles from the topbar

系统 SHALL 在阅读页顶栏提供源码视图切换控件，使当前文档可以在渲染视图和源码视图之间切换。页面首次加载时 MUST 显示渲染视图；通过文件树或跨文档链接导航到另一篇文档时 MUST 保持当前视图模式；刷新页面时 MUST 回到渲染视图。

#### Scenario: Enter source view

- **WHEN** 用户点击顶栏的源码切换按钮
- **THEN** 当前文档显示带语法高亮的源码，而不是渲染后的正文

#### Scenario: Return to rendered view

- **WHEN** 用户再次点击源码切换按钮
- **THEN** 当前文档恢复显示渲染后的正文

#### Scenario: Switch documents in source mode

- **WHEN** 源码视图处于激活状态，用户导航到另一篇文档
- **THEN** 目标页面同样显示源码视图，而不是渲染视图

#### Scenario: Navigate in source mode

- **WHEN** 源码视图处于激活状态，用户通过文件树或跨文档链接打开另一篇文档
- **THEN** 新页面同样显示源码视图，而不是渲染视图

#### Scenario: Reload page

- **WHEN** 用户刷新当前文档页面
- **THEN** 页面显示渲染视图，而不是源码视图
