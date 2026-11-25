# outline-navigation Specification

## Purpose

定义点击侧栏大纲和滚动阅读区、源码区时，页面如何定位到对应标题，并让大纲高亮始终与被点击或当前可见的标题一致。

## Requirements

### Requirement: Outline click follows the active view

点击大纲项时，系统 SHALL 按当前视图选择定位目标：阅读模式滚动阅读内容；双视图优先滚动阅读区，并把源码区对齐到同一标题；源码模式保持源码视图并滚动到对应标题行。定位完成后，大纲高亮 MUST 是被点击的标题。

#### Scenario: Click outline in render mode

- **WHEN** 阅读模式下用户点击大纲中的某个标题
- **THEN** 页面滚动到该标题，大纲高亮该标题

#### Scenario: Click outline in split mode

- **WHEN** 双视图下用户点击大纲中的某个标题
- **THEN** 阅读区显示该标题，源码区对齐到该标题所在的源码行，大纲高亮该标题

#### Scenario: Click outline in source mode

- **WHEN** 源码模式下用户点击大纲中的某个标题
- **THEN** 页面保持源码模式，源码滚动到该标题所在行，大纲高亮该标题

### Requirement: Outline highlight tracks the scrolling view

大纲高亮 SHALL 跟随用户当前滚动的一侧内容更新。双视图下滚动阅读区或源码区、源码模式下滚动页面、阅读模式下滚动页面时，当前标题都必须与可见位置一致。高亮判定 MUST 以滚动区域的可见顶部为基准，而不是窗口顶部。

#### Scenario: Scroll rendered pane in split mode

- **WHEN** 双视图下用户滚动阅读区
- **THEN** 大纲高亮对应当前阅读区顶部的标题

#### Scenario: Scroll source pane in split mode

- **WHEN** 双视图下用户滚动源码区
- **THEN** 大纲高亮对应当前源码区顶部的标题

#### Scenario: Scroll in source mode

- **WHEN** 源码模式下用户滚动页面
- **THEN** 大纲高亮对应当前源码顶部的标题

#### Scenario: Pane starts below the topbar

- **WHEN** 双视图的阅读区从顶栏下方开始，用户点击或滚动到某个标题
- **THEN** 该标题位于阅读区顶部时即被高亮，不会停留在一节之前

### Requirement: Outline navigation keeps the current document

大纲定位 SHALL 只改变当前文档内部的滚动位置，MUST NOT 切换当前文档或离开源码视图。

#### Scenario: Click outline after switching documents

- **WHEN** 用户切换到另一篇文档后点击新文档的大纲项
- **THEN** 页面仍停留在该文档，并滚动到对应标题

#### Scenario: Outline click does not change view mode

- **WHEN** 用户在任何视图模式下点击大纲项
- **THEN** 当前阅读模式保持不变
