# Spec Delta

## Purpose

定义阅读页在打印时如何跟随当前阅读模式，如何把双视图收敛为单栏正文，以及如何处理代码块和表格的横向溢出，确保打印结果完整、可读且不出现横向滚动条。

## ADDED Requirements

### Requirement: Print follows the reading mode

系统 SHALL 按当前阅读模式决定打印内容：阅读模式打印渲染正文；源码模式打印源码；双视图 MUST 只打印渲染正文，MUST NOT 打印左侧源码栏。打印内容 MUST 从文档顶部开始完整输出，MUST NOT 受屏幕滚动位置影响。

#### Scenario: Print in read mode

- **WHEN** 用户在阅读模式执行打印
- **THEN** 打印结果包含完整的渲染正文

#### Scenario: Print in source mode

- **WHEN** 用户在源码模式执行打印
- **THEN** 打印结果包含当前文档的完整源码

#### Scenario: Print in split view

- **WHEN** 用户在双视图执行打印
- **THEN** 打印结果只包含渲染正文，不包含左侧源码栏

#### Scenario: Print ignores screen scroll

- **WHEN** 用户在文档中部执行打印
- **THEN** 打印结果仍从文档顶部开始并包含全部内容

### Requirement: Split view prints as a single rendered column

双视图打印时，页面 SHALL 临时取消分栏、固定高度和内部滚动容器，按普通文档流输出完整渲染正文。

#### Scenario: Split layout is flattened for print

- **WHEN** 双视图页面进入打印
- **THEN** 渲染正文以单栏普通文档流输出，不再出现左右分栏或内部滚动区域

### Requirement: Printed code blocks wrap

打印时代码块 SHALL 临时启用自动换行，长行 MUST NOT 产生横向滚动条或被裁切。

#### Scenario: Long code line wraps in print

- **WHEN** 打印包含超长代码行的文档
- **THEN** 该行在打印结果中自动折行，不出现横向滚动条

#### Scenario: Long URL in code wraps in print

- **WHEN** 打印包含超长连续字符串或 URL 的代码块
- **THEN** 该内容在打印结果中折行显示，不被裁切

### Requirement: Printed tables fit the page width

打印时表格 SHALL 适配页面宽度，表头和单元格 SHALL 允许换行，表格 MUST NOT 出现横向滚动条。

#### Scenario: Wide table fits the page

- **WHEN** 打印包含宽表格的文档
- **THEN** 表格收缩到页面宽度内，不出现横向滚动条

#### Scenario: Header wraps in print

- **WHEN** 打印的表格表头单元格内容较长
- **THEN** 表头文本允许折行，而不是保持单行溢出

### Requirement: Source always wraps in print

即使屏幕上的源码自动换行已关闭，打印源码模式时源码 SHALL 自动换行，MUST NOT 出现横向滚动条。

#### Scenario: Print source with wrapping disabled

- **WHEN** 用户关闭源码自动换行后打印源码视图
- **THEN** 打印结果中的源码仍然自动换行，且不出现横向滚动条

### Requirement: Printed pages use a white background

打印页面 SHALL 使用纯白色背景，MUST NOT 保留当前配色或深色模式的页面底色。打印文本、代码和表格 SHALL 使用适合白底的可读颜色。

#### Scenario: Print from a light theme

- **WHEN** 用户使用浅色配色执行打印
- **THEN** 打印结果背景为纯白色

#### Scenario: Print from a dark theme

- **WHEN** 用户使用深色模式执行打印
- **THEN** 打印结果背景仍为纯白色，正文和代码保持深色可读

### Requirement: Source toolbar is not printed

源码模式打印时，源码工具栏 MUST NOT 出现在打印结果中，文件名和复制控件都不打印，只打印源码正文。

#### Scenario: Print source view without toolbar

- **WHEN** 用户在源码模式执行打印
- **THEN** 打印结果不包含源码工具栏、文件名或复制按钮
