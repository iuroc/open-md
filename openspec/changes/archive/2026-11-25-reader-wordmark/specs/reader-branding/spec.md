# Spec Delta

## Purpose

定义 OpenMD 字标在阅读页中的组成、出现位置、主题适配和小尺寸可读性，让品牌名与现有的 M 图标形成统一标识。

## ADDED Requirements

### Requirement: Wordmark combines the icon mark and the brand text

文档页侧栏标题和信息与设置面板标题 SHALL 使用同一个 OpenMD 字标，字标 MUST 由 M 图标标记和可见文本 `OpenMD` 组成。

#### Scenario: Sidebar header shows the wordmark

- **WHEN** 用户打开任一文档页
- **THEN** 侧栏标题显示 M 标记和 `OpenMD` 文本，而不是纯文字标题

#### Scenario: Settings panel shows the same wordmark

- **WHEN** 用户打开信息与设置面板
- **THEN** 面板标题显示与侧栏一致的 M 标记和 `OpenMD` 文本

### Requirement: Wordmark adapts to the active theme

字标的 M 标记 SHALL 使用当前主题的强调色，文字 SHALL 使用主题前景色，并在明暗模式和所有可选配色下保持可读。

#### Scenario: Switching theme updates the wordmark

- **WHEN** 用户切换配色
- **THEN** M 标记跟随新的强调色，文字跟随新的前景色

#### Scenario: Dark mode keeps contrast

- **WHEN** 系统处于深色模式
- **THEN** 字标在深色背景上保持可读，M 和文字不丢失对比

### Requirement: Wordmark stays legible at small sizes

字标 SHALL 在侧栏的 18px 和设置面板的 14px 字号下保持 M 标记与文字可区分，MUST NOT 相互重叠、裁切或挤出容器。

#### Scenario: Settings panel at 14px

- **WHEN** 用户查看信息与设置面板标题
- **THEN** M 标记和 `OpenMD` 文本都完整可见且可辨认

#### Scenario: Sidebar at 18px

- **WHEN** 用户查看侧栏标题
- **THEN** M 标记与 `OpenMD` 文本保持清晰间距，不挤压侧栏按钮

### Requirement: Wordmark uses system serif fallbacks

字标文本 SHALL 使用系统衬线字体栈，并在命名字体不可用时 MUST 回退到通用衬线字体，不依赖网络字体或额外字体文件。

#### Scenario: Named fonts are unavailable

- **WHEN** 页面运行在没有安装命名衬线字体的系统上
- **THEN** 字标使用通用衬线字体渲染，文字仍然完整显示

### Requirement: Wordmark text remains accessible

字标中的 M 图形 MUST 对辅助技术隐藏，可见文本 MUST 暴露为品牌名 `OpenMD`，且不重复朗读。

#### Scenario: Screen reader reads the brand once

- **WHEN** 辅助技术读取侧栏或设置面板标题
- **THEN** 品牌名只朗读一次为 `OpenMD`，M 图形不被单独朗读
