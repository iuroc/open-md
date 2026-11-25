# Spec Delta

## Purpose

定义阅读页可选的配色主题与正文排版偏好，使长文阅读可以按舒适度调整，同时保持默认观感、代码可读性和跨文档一致性。

## ADDED Requirements

### Requirement: Paper theme is available

阅读页 SHALL 提供名为“宣纸”的 `paper` 配色，包含明暗两版。选择该配色 MUST 立即应用到整个阅读页，并与现有配色一样在后续页面加载中保留。

#### Scenario: Select paper theme

- **WHEN** 用户在“设置”Tab 中选择“宣纸”
- **THEN** 页面立即使用宣纸配色的纸底、墨色文字和黛青强调色
- **AND** 重新打开页面后仍使用宣纸配色

#### Scenario: Paper theme follows system dark mode

- **WHEN** 系统处于深色模式且用户选择“宣纸”
- **THEN** 页面使用宣纸配色的深色版本，文字与背景保持可读对比度

### Requirement: Font presets change rendered prose

“设置”Tab SHALL 提供正文字体预设：默认、宋体、楷体、仿宋。选择任一预设 MUST 立即应用到渲染后的正文，并保留到后续页面加载。

#### Scenario: Switch body font

- **WHEN** 用户选择“宋体”
- **THEN** 渲染后的正文使用宋体字体栈
- **AND** 刷新页面后仍使用该预设

#### Scenario: Missing font falls back gracefully

- **WHEN** 所选字体未安装在当前系统
- **THEN** 正文回退到该预设的通用 `serif` 字体栈，仍然可读

### Requirement: Letter spacing presets change rendered prose

“设置”Tab SHALL 提供字符间距预设：标准、疏朗、舒展、宽松，分别对应 `0`、`0.03em`、`0.06em`、`0.10em`。选择任一预设 MUST 立即应用到渲染后的正文，并保留到后续页面加载。

#### Scenario: Increase letter spacing

- **WHEN** 用户选择“舒展”
- **THEN** 渲染后的正文使用 `0.06em` 字符间距
- **AND** 刷新页面后仍使用该预设

### Requirement: Code remains monospaced

字体和字符间距预设 MUST NOT 改变代码块、行内代码和源码视图的等宽字体；这些区域的字符间距 MUST 保持正常。

#### Scenario: Code block is unaffected

- **WHEN** 用户选择“楷体”和“宽松”字间距
- **THEN** 正文使用楷体和宽松字距
- **AND** 代码块、行内代码和源码视图仍使用等宽字体与正常字距

### Requirement: Appearance preferences persist and keep defaults

阅读页 SHALL 把配色、字体和字间距偏好保存在 `localStorage`。用户没有选择新偏好时，页面 MUST 保持现有的默认配色和默认排版。

#### Scenario: Preferences survive soft navigation

- **WHEN** 用户设置字体和字间距后在文件树中切换到另一篇文档
- **THEN** 新文档继续使用相同的字体和字间距

#### Scenario: First visit keeps defaults

- **WHEN** 用户在没有任何已保存偏好的情况下打开页面
- **THEN** 页面使用现有的默认配色、默认字体和标准字间距
