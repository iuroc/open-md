# Spec Delta

## Purpose

定义 OpenMD 加载单个 Markdown 文件和目录时的容量边界，确保超大输入不会拖垮进程，也不会让同一目录中本可阅读的其他文档消失。

## ADDED Requirements

### Requirement: Single-file loading has a size guard

系统 SHALL 在读取单个 Markdown 文件前检查文件大小。超过上限时 MUST 拒绝渲染，并给出包含文件名和上限的明确错误，而不是把整个文件读入内存。

#### Scenario: Oversized single file

- **WHEN** 用户打开一个超过单文件上限的 Markdown 文件
- **THEN** 系统显示明确的超限错误，进程保持响应，且不会尝试完整读取该文件

#### Scenario: File at the limit

- **WHEN** 文件大小不超过单文件上限
- **THEN** 文件按正常流程渲染

### Requirement: Oversized directory entries are skipped

目录扫描 SHALL 跳过单个大小超过目录单文件预算的文档，并继续处理其余文件。单个文件超限 MUST NOT 终止整个目录的扫描。

#### Scenario: Large file before small file

- **WHEN** 目录中第一个 Markdown 文件超过目录单文件预算，后面还有一个正常大小的 `b-small.md`
- **THEN** `b-small.md` 仍然出现在文件列表中，页面不报告“目录中没有 Markdown 文件”

#### Scenario: Only oversized files

- **WHEN** 目录中的所有 Markdown 文件都超过单文件预算
- **THEN** 系统返回可理解的空状态，说明存在 Markdown 但都因大小限制被跳过，而不是笼统地报告“目录中没有 Markdown 文件”

#### Scenario: Skipped file warning

- **WHEN** 目录扫描跳过了至少一个超限文件，同时成功加载了其他文件
- **THEN** 页面显示警告，指出有文件因大小限制未加载

### Requirement: Directory truncation keeps a deterministic prefix

目录扫描 SHALL 在达到文档数量上限或累计总字节上限时停止，并保留按现有排序规则确定的前缀。单个文件超过单文件预算时 MUST 跳过并继续；累计总量达到上限时 MUST 停止并显示提示。

#### Scenario: Document count limit

- **WHEN** 目录中有 3 篇文档，数量上限为 2
- **THEN** 按排序规则加载前 2 篇，并显示截断提示

#### Scenario: Total size limit

- **WHEN** 继续加入下一篇文档会超过总字节上限
- **THEN** 扫描停止，保留已加载的文档，并显示截断提示
