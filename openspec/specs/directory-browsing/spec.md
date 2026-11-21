# directory-browsing Specification

## Purpose

让用户把多层嵌套的 Markdown 目录当作一棵可浏览的文件树，在同一阅读页内展开、切换和定位文档，而不必逐层右键打开子目录。

## Requirements

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

### Requirement: Sidebar scroll survives navigation

目录模式的文件树滚动位置 SHALL 在同一 root 的文档页面之间保持。通过文件树、跨文档链接或浏览器前进/后退导航后，MUST 恢复离开前的位置。首次打开某个 root、没有保存位置时 MAY 将当前文档滚动到可见区域。

#### Scenario: Switch documents from the tree

- **WHEN** 用户在文件树滚动到中间位置，然后点击另一篇文档
- **THEN** 新页面的侧栏保持相同的滚动位置

#### Scenario: Cross-document link

- **WHEN** 用户阅读文档 A 时通过正文链接打开文档 B
- **THEN** 文档 B 页面的侧栏保持文档 A 页面离开时的滚动位置

#### Scenario: First open

- **WHEN** 用户首次打开某个 root 的文档页面
- **THEN** 文件树展开并显示当前文档节点，不需要恢复旧位置

### Requirement: Initial document selection is deterministic

打开目录时，系统 SHALL 优先选择根目录的 `README.md` 或 `index.md`；不存在时 MUST 选择层级最浅的第一个文档，同层按自然序取第一个。

#### Scenario: Root README exists

- **WHEN** 被打开目录的根层存在 `README.md`
- **THEN** 页面初始显示 `README.md`

#### Scenario: No root entry document

- **WHEN** 根层没有 `README.md` 或 `index.md`
- **THEN** 页面初始显示层级最浅的第一个文档，而不是归档深处的一篇

### Requirement: Linked documents outside the scan root are separated

通过文档链接载入、但不在扫描根目录内的 Markdown 文档 SHALL 单独归入“链接文档”分组，MUST NOT 混入按真实路径生成的文件树。该分组中的每个文档 MUST 是指向其独立页面的链接。

#### Scenario: Link points outside the opened directory

- **WHEN** 根目录文档链接到扫描根之外的一个 Markdown 文件
- **THEN** 该文档出现在“链接文档”分组，文件树结构保持不变

#### Scenario: Open a linked document

- **WHEN** 用户点击“链接文档”中的条目
- **THEN** 浏览器导航到该文档的独立页面

### Requirement: File tree is keyboard operable

文件树的目录和文件控件 SHALL 可通过键盘聚焦和触发。目录控件 MUST 暴露 `aria-expanded` 状态；文件控件 MUST 暴露当前选中状态。

#### Scenario: Toggle a directory with keyboard

- **WHEN** 键盘焦点位于一个折叠的目录控件上，用户按下 Enter 或空格
- **THEN** 该目录展开，并更新 `aria-expanded`

#### Scenario: Open a file with keyboard

- **WHEN** 键盘焦点位于文件控件上，用户按下 Enter 或空格
- **THEN** 阅读区切换到该文档
