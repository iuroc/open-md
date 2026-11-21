# Spec Delta

## MODIFIED Requirements

### Requirement: Source lines wrap

源码视图 SHALL 默认自动换行显示超长行。用户 SHALL 能在“设置”Tab 中关闭自动换行；关闭后源码 SHALL 保留原始行宽，并允许通过横向滚动查看超长行；开启时 MUST NOT 产生横向滚动。设置 MUST 在页面重新加载后保留，并 SHALL 同时作用于单独源码视图和双视图中的源码栏。打印源码时 MUST 始终自动换行。

#### Scenario: Long paragraph wraps

- **WHEN** 文档包含一行超过阅读区宽度的长段落，且用户未修改自动换行设置
- **THEN** 该行在源码视图中自动折行，不发生横向滚动

#### Scenario: Long URL wraps

- **WHEN** 文档包含一行超过阅读区宽度的长链接或连续字符，且用户未修改自动换行设置
- **THEN** 该行在源码视图中自动折行，不发生横向滚动

#### Scenario: Disable wrapping

- **WHEN** 用户在“设置”Tab 中关闭源码自动换行
- **THEN** 源码保留原始行宽，超长行可通过横向滚动查看

#### Scenario: Split source pane follows the setting

- **WHEN** 自动换行已关闭，用户进入双视图并横向滚动左侧源码栏
- **THEN** 源码栏保持原始行宽，并显示横向滚动条

#### Scenario: Wrapping setting persists

- **WHEN** 用户关闭源码自动换行后刷新页面
- **THEN** 源码视图仍保持不换行状态

#### Scenario: Print always wraps

- **WHEN** 自动换行已关闭，用户打印源码视图
- **THEN** 打印结果中的源码自动折行，且不出现横向滚动条
