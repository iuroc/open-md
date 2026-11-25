# Spec Delta

## MODIFIED Requirements

### Requirement: Settings tab owns reading preferences

“设置”Tab SHALL 集中展示阅读相关的设置项，包括配色、字体、字号和字间距。任一设置 MUST 立即生效，并在后续页面加载中保留；未来的设置项 SHALL 作为该 Tab 内的分组继续扩展。

#### Scenario: Change theme

- **WHEN** 用户在“设置”Tab 中选择另一套配色
- **THEN** 页面立即应用该配色，重新打开页面后仍然使用该配色

#### Scenario: Change font

- **WHEN** 用户在“设置”Tab 中选择另一套正文字体
- **THEN** 正文立即使用该字体，重新打开页面后仍然使用该字体

#### Scenario: Change font size

- **WHEN** 用户在“设置”Tab 中选择另一档正文字号
- **THEN** 内容区域立即使用该字号，重新打开页面后仍然使用该字号

#### Scenario: Change letter spacing

- **WHEN** 用户在“设置”Tab 中选择另一档字符间距
- **THEN** 正文立即使用该字间距，重新打开页面后仍然使用该字间距

#### Scenario: Settings remain grouped

- **WHEN** 后续增加新的阅读设置
- **THEN** 新设置出现在“设置”Tab 内，而不是新增一个顶级 Tab
