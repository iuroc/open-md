# Proposal

## Why

当前阅读器在几个常见 Markdown 场景下会产生错误结果：页内锚点失效、多文档脚注串篇、一个特殊链接导致整篇文档不可读、中文标题锚点不稳定，以及目录扫描被单个超大文件拖死。这些问题直接影响“打开就能读”的核心承诺，应作为第一批正确性修复处理。

## What Changes

- 修正同文档锚点：`[跳转](#install)` 应跳到被加上文档前缀后的真实标题 ID。
- 修正跨文档锚点：中文标题、带标点标题生成稳定 ID，`other.md#中文标题` 能正确跳转。
- 按文档隔离脚注 ID（`fn:` / `fnref:`），多文档页面不再出现重复 ID。
- 本地路径解析改为“文件系统优先”，支持文件名中的 `%` 等字符；URL 解析失败只影响单个链接，不再让整篇文档失败。
- 剥离 YAML/TOML front matter，不把元数据渲染成正文或大纲标题。
- 目录扫描遇到单个超大文件时跳过并继续，而不是终止整个目录；达到文档数量上限时才停止。
- 为单文件模式增加大小上限，超限时给出明确错误，避免一次性读入超大文件。
- 为上述行为补充回归测试，并更新 README 中受影响的已知限制。

明确不在本变更范围内：远程/UNC 图片拦截、CSP nonce、`openmd://` 按需加载、临时文件保留策略、卸载 `UserChoice`、GBK/UTF-16 编码支持。

## Capabilities

### New Capabilities

- `markdown-rendering`: 负责 Markdown 到阅读页的渲染契约，包括稳定标题 ID、同文档与跨文档锚点、脚注 ID 隔离、front matter 处理和异常链接隔离。
- `document-loading`: 负责单文件与目录文档的加载契约，包括文件大小上限、目录扫描截断策略和超大文件跳过行为。

### Modified Capabilities

无。

## Impact

- 代码：`internal/render/document.go`、`internal/render/url.go`、`internal/render/links.go`、`internal/render/page.go`、`internal/scan/scan.go` 及相关测试。
- 行为：现有 Markdown 的锚点、脚注、front matter 和超大目录打开结果会发生变化；不涉及注册表、安装流程或浏览器启动方式。
- 文档：README 的目录限制、链接文档说明和已知限制需要同步。
- 依赖：不新增第三方依赖。
