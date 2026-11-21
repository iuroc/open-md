# Spec Delta

## ADDED Requirements

### Requirement: Wide tables scroll horizontally without compressed headers

系统 SHALL 防止表头单元格仅为适配阅读区宽度而折行。当表格自然宽度超过阅读区时，表格 MUST 在自身滚动容器内横向滚动；正文单元格 SHALL 继续按可用列宽换行。

#### Scenario: Many-column table

- **WHEN** 表格列数较多，自然宽度超过阅读区
- **THEN** 表头文字保持单行显示，表格出现横向滚动，而不是把表头压成多行

#### Scenario: Table fits the reading area

- **WHEN** 表格自然宽度不超过阅读区
- **THEN** 表格正常显示，不出现不必要的横向滚动

#### Scenario: Long header text

- **WHEN** 某个表头文字长于其内容列的自然宽度
- **THEN** 该列按表头文字撑开，整张表格横向滚动

#### Scenario: Long body content

- **WHEN** 正文单元格包含较长文本
- **THEN** 文本在列宽内换行，不把表头挤成多行
