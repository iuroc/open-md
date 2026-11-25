# Spec Delta

## Purpose

定义随 OpenMD 发布的内嵌 Markdown 使用手册：它包含什么、如何生成、从哪里进入，以及在生成不可用时如何降级。

## ADDED Requirements

### Requirement: Manual ships with the application

应用 SHALL 内嵌一本面向零基础读者的中文 Markdown 手册，至少包含入门章节、场景章节、速查表和常见问题。手册 MUST 在无网络环境下完整可用，章节之间 MUST 能通过相对链接互相跳转。

#### Scenario: Read the manual offline

- **WHEN** 用户在断网状态下打开使用手册
- **THEN** 手册首页和所有章节都能正常显示
- **AND** 章节之间的跳转可以正常工作

#### Scenario: Navigate chapters

- **WHEN** 用户从手册首页点击任意章节链接
- **THEN** 浏览器打开对应章节，侧栏文件列表包含全部章节

### Requirement: About tab exposes the manual

阅读页的「信息与设置 → 关于」Tab SHALL 在手册可用时显示「打开使用手册」入口，并在新标签页打开手册首页。手册不可用时 MUST NOT 显示该入口。

#### Scenario: Open manual from a document

- **WHEN** 用户打开任意 Markdown 文件并进入「关于」Tab
- **THEN** 页面显示「打开使用手册」
- **AND** 点击后在新标签页打开手册首页，当前文档页面保持原状

#### Scenario: Manual unavailable

- **WHEN** 手册生成失败或不可用
- **THEN**「关于」Tab 不显示手册入口
- **AND** 文档本身仍能正常打开和阅读

### Requirement: Manual uses normal reader pages

手册页面 SHALL 复用普通文档的阅读界面：文件树、大纲、阅读 / 源码 / 双视图、打印、配色和排版设置 MUST 与普通文档一致。手册页面 MUST NOT 依赖网络资源加载正文或样式。

#### Scenario: Inspect chapter source

- **WHEN** 用户打开手册的某一章并切换到源码或双视图
- **THEN** 页面显示该章的 Markdown 原文，并与正文保持同步

#### Scenario: Print a chapter

- **WHEN** 用户在手册页面执行打印
- **THEN** 打印内容遵循与普通文档相同的白底、代码换行和表格适配规则

### Requirement: Manual generation does not block reading

打开文档时，应用 SHALL 按内嵌内容指纹生成或复用手册站点。生成失败 MUST NOT 阻止文档打开、弹出错误对话框或改变文档内容。

#### Scenario: First document open generates the manual

- **WHEN** 用户首次打开文档
- **THEN** 应用生成手册站点后正常打开文档
- **AND**「关于」Tab 随后可以打开手册

#### Scenario: Reuse generated manual

- **WHEN** 用户再次打开文档且手册内容未变化
- **THEN** 应用复用手册站点，不重复渲染所有章节

#### Scenario: Generation failure

- **WHEN** 手册目录不可写或渲染失败
- **THEN** 文档仍按正常流程打开
- **AND** 页面不显示手册入口，也不出现额外错误提示

#### Scenario: Open several documents at once

- **WHEN** 用户同时打开多个文档，且手册尚未生成
- **THEN** 每个文档都正常打开
- **AND** 手册已完整生成时显示入口，未完成时隐藏入口，页面不会指向缺章的半成品手册

### Requirement: Manual content is original with attributed further reading

手册正文 SHALL 为原创内容。引用外部教程时 MUST 只使用链接并标注出处，MUST NOT 直接摘录受许可证约束的第三方原文。

#### Scenario: Further reading

- **WHEN** 用户阅读任意一章
- **THEN** 该章末尾可以提供指向 Markdown Guide、CommonMark 或 GitHub Docs 等站点的延伸阅读链接
- **AND** 手册正文中不包含第三方教程的原文摘录
