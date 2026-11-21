# Spec Delta

## ADDED Requirements

### Requirement: Tree visually expresses hierarchy

文件树 SHALL 通过固定宽度的引导槽、缩进和竖向引导线表达层级。同一层级的文件夹和文件标签 MUST 从相同水平位置开始。文件夹和文件 MUST 通过字重、颜色或图标至少一种方式区分。

#### Scenario: Folder and file alignment

- **WHEN** 同一层同时存在文件夹和文件
- **THEN** 文件夹名称和文件名称从同一水平位置开始

#### Scenario: Deep nesting

- **WHEN** 文件树展开到多层目录
- **THEN** 每一层都有明确的缩进和引导线，用户可以看出父子关系

#### Scenario: Active document

- **WHEN** 某个文件是当前文档
- **THEN** 该节点使用强调色、浅色背景或左侧标记显示选中状态，而不是整块高饱和填充

## MODIFIED Requirements

### Requirement: Documents are presented as a file tree

目录模式 SHALL 把递归发现的所有 Markdown 文档按相对路径组织成文件树。目录节点 SHALL 能展开和折叠，文件节点 SHALL 是指向独立文档页面的链接。树 MUST 只包含至少有一个 Markdown 后代的目录。树子节点 SHALL 在文件夹首次展开时创建，MUST NOT 一次性把所有节点加入 DOM。

#### Scenario: Nested documents form a tree

- **WHEN** 被打开的目录包含 `specs/source-view/spec.md` 和 `changes/archive/foo/proposal.md`
- **THEN** 侧栏显示 `specs`、`source-view`、`changes`、`archive`、`foo` 等目录节点，并在对应层级显示文件节点

#### Scenario: Directory without Markdown is absent

- **WHEN** 目录中存在只包含图片或文本文件的子目录
- **THEN** 文件树不显示该子目录

#### Scenario: File selection

- **WHEN** 用户点击文件树中的 `proposal.md`
- **THEN** 浏览器导航到该文档的独立页面，文件节点显示为当前选中状态

#### Scenario: Large tree

- **WHEN** 目录包含大量文档且多数文件夹处于折叠状态
- **THEN** 页面只为根节点和当前展开路径创建 DOM

### Requirement: Active document path is revealed

页面加载时 SHALL 展开当前文档的祖先目录，并恢复同一 root 上次保存的展开状态。导航到另一篇文档后，目标页面 MUST 展开并高亮该文档路径；用户新展开或折叠的目录 MUST 在后续页面导航中保持。

#### Scenario: Initial active document

- **WHEN** 用户直接打开一个深层文档页面
- **THEN** 该文档的祖先目录被展开并高亮

#### Scenario: Cross-document link target

- **WHEN** 当前文档中的链接指向另一篇文档
- **THEN** 目标页面加载后展开并高亮其文件节点

#### Scenario: Switch through tree

- **WHEN** 用户在树中打开另一个目录下的文档
- **THEN** 新页面的文件树展开该路径，当前节点获得选中状态

#### Scenario: Expansion survives navigation

- **WHEN** 用户展开 `archive/` 后打开其中一篇文档
- **THEN** 新页面中的 `archive/` 仍保持展开

### Requirement: Linked documents outside the scan root are separated

通过文档链接载入、但不在扫描根目录内的 Markdown 文档 SHALL 单独归入“链接文档”分组，MUST NOT 混入按真实路径生成的文件树。该分组中的每个文档 MUST 是指向其独立页面的链接。

#### Scenario: Link points outside the opened directory

- **WHEN** 根目录文档链接到扫描根之外的一个 Markdown 文件
- **THEN** 该文档出现在“链接文档”分组，文件树结构保持不变

#### Scenario: Open a linked document

- **WHEN** 用户点击“链接文档”中的条目
- **THEN** 浏览器导航到该文档的独立页面
