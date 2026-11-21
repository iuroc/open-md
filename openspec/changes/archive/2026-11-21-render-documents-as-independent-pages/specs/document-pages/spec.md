# Spec Delta

## Purpose

把每篇 Markdown 渲染为独立的静态 HTML 页面，并通过稳定缓存、共享资源和增量生成，让大目录在不依赖 HTTP 服务的前提下仍能快速打开和重复阅读。

## ADDED Requirements

### Requirement: Each document has its own page

系统 SHALL 为每篇 Markdown 生成一个独立 HTML 页面。目录模式的文件树节点和跨文档链接 MUST 指向对应页面，并在同一标签页打开。

#### Scenario: Open a document from the tree

- **WHEN** 用户在文件树中点击 `proposal.md`
- **THEN** 浏览器导航到该文档的独立页面，阅读区只包含这一篇文档

#### Scenario: Cross-document link

- **WHEN** 当前文档包含指向 `other.md#usage` 的链接
- **THEN** 点击后导航到 `other.md` 的页面并滚动到 `usage` 标题

#### Scenario: Browser back

- **WHEN** 用户从文档 A 跳到文档 B 后点击浏览器后退
- **THEN** 回到文档 A 的页面

### Requirement: Page paths are stable

同一源文件 MUST 在重复打开时复用同一个输出 HTML 路径。系统 MUST NOT 因运行次数增加而生成重复的随机页面。

#### Scenario: Reopen the same directory

- **WHEN** 用户两次打开同一个目录，文档内容未变化
- **THEN** 每篇文档使用相同的输出路径，旧页面被复用或覆盖，而不是新增副本

#### Scenario: Source file changes

- **WHEN** 某篇 Markdown 的内容发生变化后再次打开
- **THEN** 对应的 HTML 被覆盖为最新内容，其他未变化页面保持不变

### Requirement: Unchanged documents are reused

系统 SHALL 使用源文件 mtime、size、OpenMD 版本和文档集合 hash 判断缓存是否有效。全部条件匹配且输出存在时，MUST 跳过读取和渲染。

#### Scenario: Cache hit

- **WHEN** 文档集合、源文件元信息和 OpenMD 版本都未变化
- **THEN** 打开目录时直接复用已有页面，不重新读取或渲染这些文档

#### Scenario: Cache miss for one file

- **WHEN** 只有一篇文档的 mtime 或 size 变化
- **THEN** 只重新生成这一篇页面，其余页面继续复用

#### Scenario: Document set changes

- **WHEN** 目录中新增、删除或重命名了文档
- **THEN** 首版重新生成所有页面，避免旧链接映射失效

#### Scenario: Version changes

- **WHEN** OpenMD 版本或页面资源版本变化
- **THEN** 所有页面重新生成

### Requirement: Stale pages are removed

系统 SHALL 删除当前扫描根下已经不对应任何源文档的旧页面，并按 TTL 或最近使用数量清理不再打开的 root。

#### Scenario: Deleted document

- **WHEN** 用户删除一篇 Markdown 后再次打开该目录
- **THEN** 对应的旧 HTML 被删除，文件树中不再出现该文档

#### Scenario: Old root cleanup

- **WHEN** 某个 root 的缓存超过保留期限或超出最近使用的 root 数量
- **THEN** 该 root 的页面和 manifest 被清理

### Requirement: Pages share navigation assets

同一 root 的所有页面 SHALL 共享 CSS、JavaScript 和文件树数据，MUST NOT 在每个页面重复内联完整文件树。

#### Scenario: Shared tree data

- **WHEN** 打开同一目录下的任意文档页面
- **THEN** 页面加载同一份 `tree.js`、`reader.css` 和 `reader.js`

#### Scenario: Large tree

- **WHEN** 目录包含大量文档
- **THEN** 侧栏只渲染根节点和当前展开路径，展开文件夹时才创建其子节点

### Requirement: Pages are generated before opening

系统 SHALL 在打开浏览器前生成所有需要更新的页面。单篇渲染失败 MUST NOT 阻止其他页面生成。

#### Scenario: One broken document

- **WHEN** 某篇文档读取或渲染失败
- **THEN** 该文档生成错误页面，其他文档仍正常生成并可阅读
