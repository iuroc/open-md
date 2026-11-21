# Spec Delta

## Purpose

让用户把多层嵌套的 Markdown 目录当作一棵可浏览的文件树，在同一阅读页内展开、切换和定位文档，而不必逐层右键打开子目录。

## ADDED Requirements

### Requirement: Documents are presented as a file tree

目录模式 SHALL 把递归发现的所有 Markdown 文档按相对路径组织成文件树。目录节点 SHALL 能展开和折叠，文件节点 SHALL 能切换到对应文档。树 MUST 只包含至少有一个 Markdown 后代的目录。

#### Scenario: Nested documents form a tree

- **WHEN** 被打开的目录包含 `specs/source-view/spec.md` 和 `changes/archive/foo/proposal.md`
- **THEN** 侧栏显示 `specs`、`source-view`、`changes`、`archive`、`foo` 等目录节点，并在对应层级显示文件节点

#### Scenario: Directory without Markdown is absent

- **WHEN** 目录中存在只包含图片或文本文件的子目录
- **THEN** 文件树不显示该子目录

#### Scenario: File selection

- **WHEN** 用户点击文件树中的 `proposal.md`
- **THEN** 阅读区切换到该文档，文件节点显示为当前选中状态

### Requirement: File nodes use filenames as labels

文件节点 SHALL 以文件名作为主标签，例如 `proposal.md`、`design.md`、`spec.md`。文档标题和完整相对路径 MUST 通过悬停提示或等效方式提供，MUST NOT 依赖标题区分同一目录下的同名文件。

#### Scenario: Repeated spec filenames

- **WHEN** 不同目录下存在多个 `spec.md`
- **THEN** 每个节点都显示 `spec.md`，用户通过所在目录层级区分它们

#### Scenario: Repeated generic titles

- **WHEN** 多个文档的 H1 都是 `# Spec Delta`
- **THEN** 文件树仍然以文件名展示这些节点，不出现一排重复的“Spec Delta”

#### Scenario: Long filename

- **WHEN** 文件名或可用侧栏宽度不足以完整显示标签
- **THEN** 标签按可用宽度截断，完整标题或路径仍可通过悬停提示查看

### Requirement: Tree order is predictable

同一层级内 SHALL 先显示目录，再显示文件，并分别按自然序排列。日期前缀 SHALL 按数值顺序排列，而不是纯字符顺序。

#### Scenario: Directories before files

- **WHEN** 同一目录同时包含 `specs/` 和 `tasks.md`
- **THEN** `specs/` 显示在 `tasks.md` 之前

#### Scenario: Natural file order

- **WHEN** 同一目录包含 `2.md`、`10.md`、`a.md`
- **THEN** 显示顺序为 `2.md`、`10.md`、`a.md`

#### Scenario: Date-prefixed directories

- **WHEN** 同一层包含 `2026-11-2-change` 和 `2026-11-10-change`
- **THEN** `2026-11-2-change` 排在 `2026-11-10-change` 之前

### Requirement: Active document path is revealed

页面加载时 SHALL 只展开当前文档的祖先目录，其他目录保持折叠。切换文档或点击跨文档链接后，树 MUST 展开目标文档的祖先路径、把目标节点滚动到可见区域并显示选中状态。

#### Scenario: Initial active document

- **WHEN** 用户打开一个多层目录
- **THEN** 只有当前文档所在的目录路径被展开，其余目录折叠

#### Scenario: Cross-document link target

- **WHEN** 当前文档中的链接指向一个位于折叠目录中的文档，用户点击该链接
- **THEN** 阅读区切换到目标文档，树展开目标路径并高亮对应文件节点

#### Scenario: Switch through tree

- **WHEN** 用户在树中打开另一个目录下的文档
- **THEN** 该路径保持展开，当前节点获得选中状态

### Requirement: Initial document selection is deterministic

打开目录时，系统 SHALL 优先选择根目录的 `README.md` 或 `index.md`；不存在时 MUST 选择层级最浅的第一个文档，同层按自然序取第一个。

#### Scenario: Root README exists

- **WHEN** 被打开目录的根层存在 `README.md`
- **THEN** 页面初始显示 `README.md`

#### Scenario: No root entry document

- **WHEN** 根层没有 `README.md` 或 `index.md`
- **THEN** 页面初始显示层级最浅的第一个文档，而不是归档深处的一篇

### Requirement: Linked documents outside the scan root are separated

通过文档链接载入、但不在扫描根目录内的 Markdown 文档 SHALL 单独归入“链接文档”分组，MUST NOT 混入按真实路径生成的文件树。

#### Scenario: Link points outside the opened directory

- **WHEN** 根目录文档链接到扫描根之外的一个 Markdown 文件
- **THEN** 该文档出现在“链接文档”分组，文件树结构保持不变

### Requirement: File tree is keyboard operable

文件树的目录和文件控件 SHALL 可通过键盘聚焦和触发。目录控件 MUST 暴露 `aria-expanded` 状态；文件控件 MUST 暴露当前选中状态。

#### Scenario: Toggle a directory with keyboard

- **WHEN** 键盘焦点位于一个折叠的目录控件上，用户按下 Enter 或空格
- **THEN** 该目录展开，并更新 `aria-expanded`

#### Scenario: Open a file with keyboard

- **WHEN** 键盘焦点位于文件控件上，用户按下 Enter 或空格
- **THEN** 阅读区切换到该文档
