# Spec Delta

## ADDED Requirements

### Requirement: Directory scan is recursive

目录扫描 SHALL 递归发现扫描根目录及其所有子目录中的 `.md` 和 `.markdown` 文件。根层没有 Markdown、但子目录中存在 Markdown 时，系统 MUST 正常打开这些文档，而不是报告目录为空。

#### Scenario: Root has no direct Markdown

- **WHEN** 打开的目录根层只有子目录，Markdown 文件全部位于子目录中
- **THEN** 页面正常加载这些文档，不显示“目录中没有 Markdown 文件”

#### Scenario: Deeply nested document

- **WHEN** Markdown 文件位于扫描根下多层子目录中
- **THEN** 该文件出现在文件树中，并可被打开阅读

#### Scenario: Mixed extensions

- **WHEN** 子目录中同时存在 `.md`、`.MD`、`.markdown` 和非 Markdown 文件
- **THEN** 三种 Markdown 扩展名都被扫描，非 Markdown 文件被忽略

### Requirement: Directory scan skips links and heavy directories

递归扫描 MUST NOT 跟随符号链接、junction 或其他目录重解析点。系统 SHALL 跳过已知的重目录，例如 `.git`、`node_modules`、`vendor`。单个子目录不可读时 MUST 记录警告并继续扫描其余目录。

#### Scenario: Symlink or junction cycle

- **WHEN** 子目录中存在指回祖先目录的符号链接或 junction
- **THEN** 扫描不进入该链接，不产生重复文档，也不会无限循环

#### Scenario: Heavy directory is ignored

- **WHEN** 扫描根包含 `node_modules` 或 `.git` 目录，其中存在 Markdown 文件
- **THEN** 这些目录中的文档不进入扫描结果

#### Scenario: Unreadable subdirectory

- **WHEN** 扫描过程中某个子目录因权限不足无法读取
- **THEN** 扫描继续处理其他目录，并在页面显示一条可理解的警告

### Requirement: Recursive traversal is bounded

递归扫描 SHALL 设置最大目录深度和最大访问条目预算。达到任一预算时 MUST 停止继续遍历、保留已发现的文档，并显示明确的截断提示。

#### Scenario: Depth limit reached

- **WHEN** 目录层级超过允许的最大深度
- **THEN** 更深层的文件不被加载，页面提示扫描深度已截断

#### Scenario: Entry budget reached

- **WHEN** 遍历的目录项数量达到预算上限
- **THEN** 扫描停止，已发现的文档仍然可用，页面提示扫描范围已截断

## MODIFIED Requirements

### Requirement: Oversized directory entries are skipped

目录扫描 SHALL 在任意层级跳过单个大小超过目录单文件预算的文档，并继续处理其余文件。单个文件超限 MUST NOT 终止整次扫描。警告 SHALL 使用相对于扫描根目录的路径标识被跳过的文件。

#### Scenario: Large file before small file

- **WHEN** 目录中第一个 Markdown 文件超过目录单文件预算，后面还有一个正常大小的 `b-small.md`
- **THEN** `b-small.md` 仍然出现在文件列表中，页面不报告“目录中没有 Markdown 文件”

#### Scenario: Only oversized files

- **WHEN** 目录中的所有 Markdown 文件都超过单文件预算
- **THEN** 系统返回可理解的空状态，说明存在 Markdown 但都因大小限制被跳过，而不是笼统地报告“目录中没有 Markdown 文件”

#### Scenario: Skipped file warning

- **WHEN** 目录扫描跳过了至少一个超限文件，同时成功加载了其他文件
- **THEN** 页面显示警告，指出有文件因大小限制未加载

#### Scenario: Nested oversized file

- **WHEN** 超限文件位于子目录中，而其他子目录存在正常文档
- **THEN** 警告使用相对扫描根目录的路径标识该文件，其他文档正常加载

### Requirement: Directory truncation keeps a deterministic prefix

目录扫描 SHALL 在达到文档数量上限或累计总字节上限时停止，并保留确定性前缀。候选文档 MUST 按相对扫描根目录的深度从浅到深、同深度按自然路径顺序选择；单个文件超过单文件预算时 MUST 跳过并继续。累计总量达到上限时 MUST 停止并显示提示。相同目录内容重复扫描 MUST 得到相同的文档集合。

#### Scenario: Document count limit

- **WHEN** 目录中有 3 篇同层文档，数量上限为 2
- **THEN** 按自然顺序加载前 2 篇，并显示截断提示

#### Scenario: Shallow documents survive truncation

- **WHEN** 文档数量超过上限，且浅层和深层文档都可能被选中
- **THEN** 浅层文档优先保留，深层文档被截断

#### Scenario: Total size limit

- **WHEN** 继续加入下一篇文档会超过总字节上限
- **THEN** 扫描停止，保留已加载的文档，并显示截断提示

#### Scenario: Deterministic rescan

- **WHEN** 同一目录在内容未变化时被扫描两次
- **THEN** 两次得到相同的文档集合和相同的加载顺序
