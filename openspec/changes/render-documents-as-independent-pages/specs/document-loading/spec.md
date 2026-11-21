# Spec Delta

## MODIFIED Requirements

### Requirement: Single-file loading has a size guard

系统 SHALL 在读取单个 Markdown 文件前检查文件大小。超过 32 MiB 时 MUST 先询问用户是否仍然加载；用户确认后 MUST 读取并渲染，不再按产品上限拒绝。用户取消时 MUST NOT 读取文件。

#### Scenario: Oversized single file

- **WHEN** 用户打开一个超过单文件上限的 Markdown 文件
- **THEN** 系统先询问是否继续，而不是直接显示超限错误

#### Scenario: Oversized single file confirmed

- **WHEN** 用户打开一个超过 32 MiB 的 Markdown 文件并选择继续
- **THEN** 系统读取并渲染该文件，而不是显示超限错误

#### Scenario: Oversized single file cancelled

- **WHEN** 用户打开一个超过 32 MiB 的 Markdown 文件并选择取消
- **THEN** 系统不读取该文件并结束本次打开

#### Scenario: File at the limit

- **WHEN** 文件大小不超过 32 MiB
- **THEN** 文件按正常流程渲染

### Requirement: Recursive traversal is bounded

递归扫描 SHALL 先生成待加载工作量估计。待重建文档超过 3,000 篇或 Markdown 总大小超过 8 MiB 时，MUST 询问用户是否加载全部。用户确认加载全部后 MUST 取消数量、总量、深度和目录项限制；用户选择快速模式时 MAY 应用这些限制并给出截断提示。

#### Scenario: Depth limit reached

- **WHEN** 用户选择快速模式且目录层级超过允许的最大深度
- **THEN** 更深层的文件不被加载，页面提示扫描深度已截断

#### Scenario: Entry budget reached

- **WHEN** 用户选择快速模式且遍历的目录项数量达到预算上限
- **THEN** 扫描停止，已发现的文档仍然可用，页面提示扫描范围已截断

#### Scenario: Small directory

- **WHEN** 目录少于 3,000 篇且小于 8 MiB
- **THEN** 系统直接加载全部文档，不显示确认弹窗

#### Scenario: Large directory confirmed

- **WHEN** 目录超过软阈值且用户选择加载全部
- **THEN** 系统继续扫描并加载所有文档，不因数量或总量停止

#### Scenario: Large directory quick mode

- **WHEN** 用户选择只加载较快的一部分
- **THEN** 系统应用 3,000 篇 / 8 MiB 的快速模式限制，并在文件树底部显示截断提示

### Requirement: Directory truncation keeps a deterministic prefix

只有在用户选择快速模式或扫描触发保护条件时，目录扫描 SHALL 在达到文档数量上限或累计总字节上限时停止，并保留确定性前缀。候选文档 MUST 按相对扫描根目录的深度从浅到深、同深度按自然路径顺序选择。相同目录内容重复扫描 MUST 得到相同的文档集合。

#### Scenario: Document count limit

- **WHEN** 快速模式下目录中有 3 篇同层文档，数量上限为 2
- **THEN** 按自然顺序加载前 2 篇，并显示截断提示

#### Scenario: Shallow documents survive truncation

- **WHEN** 快速模式文档数量超过上限，且浅层和深层文档都可能被选中
- **THEN** 浅层文档优先保留，深层文档被截断

#### Scenario: Total size limit

- **WHEN** 快速模式下继续加入下一篇文档会超过总字节上限
- **THEN** 扫描停止，保留已加载的文档，并显示截断提示

#### Scenario: Quick mode count limit

- **WHEN** 用户选择快速模式且待加载文档超过 3,000 篇
- **THEN** 按浅层优先、自然序保留前 3,000 篇，并显示截断提示

#### Scenario: Quick mode size limit

- **WHEN** 继续加入下一篇文档会超过 8 MiB
- **THEN** 扫描停止，保留已加载的文档，并显示截断提示

#### Scenario: Deterministic rescan

- **WHEN** 同一目录在内容未变化时被扫描两次
- **THEN** 两次得到相同的文档集合和相同的加载顺序

### Requirement: Oversized directory entries are skipped

快速模式下，目录扫描 SHALL 跳过单个大小超过单文件预算的文档，并继续处理其余文件。加载全部模式 MUST NOT 因单文件大小跳过文档。警告 SHALL 使用相对于扫描根目录的路径标识被跳过的文件。

#### Scenario: Large file before small file

- **WHEN** 快速模式下第一个 Markdown 文件超过目录单文件预算，后面还有一个正常大小的 `b-small.md`
- **THEN** `b-small.md` 仍然出现在文件列表中

#### Scenario: Only oversized files

- **WHEN** 快速模式下所有 Markdown 文件都超过单文件预算
- **THEN** 系统返回可理解的空状态，说明存在 Markdown 但都因大小限制被跳过

#### Scenario: Skipped file warning

- **WHEN** 快速模式下跳过了至少一个超限文件，同时成功加载了其他文件
- **THEN** 页面显示警告，指出有文件因大小限制未加载

#### Scenario: Nested oversized file

- **WHEN** 快速模式下超限文件位于子目录中
- **THEN** 警告使用相对扫描根目录的路径标识该文件

#### Scenario: Oversized file in quick mode

- **WHEN** 快速模式下某个文档超过单文件预算
- **THEN** 该文件被跳过并显示相对路径警告，其他文档正常加载

#### Scenario: Oversized file in full mode

- **WHEN** 用户已确认加载全部，且某个文档超过单文件预算
- **THEN** 该文档仍然被读取和渲染
