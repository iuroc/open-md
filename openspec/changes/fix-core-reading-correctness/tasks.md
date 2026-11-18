# Tasks

## 1. 标题 ID 与锚点

- [x] 1.1 在 `RenderOptions` 中增加当前文档 ID，并让 `RewriteURL` 把 `#fragment` 重写为 `#doc=<id>&anchor=<fragment>`；验证：新增 `TestRenderMarkdownRewritesSameDocumentAnchor`，断言输出包含 `#doc=doc-0&anchor=install`，运行 `go test ./internal/render -run TestRenderMarkdownRewritesSameDocumentAnchor`
- [x] 1.2 在 `RenderMarkdown` 的 AST walk 中按标题可见文本生成 Unicode slug，并在文档内去重；验证：新增中文标题、重复标题、英文兼容用例，运行 `go test ./internal/render -run 'TestHeading'`，英文 `# Hello World` 仍得到 `hello-world`
- [x] 1.3 确保同文档和已预渲染文档的锚点都能解析到 `doc-N-<slug>`；验证：新增同文档、英文跨文档、中文跨文档测试，运行 `go test ./internal/render -run 'Test.*Anchor'`
- [x] 1.4 在 README 的“文档间链接”中补充锚点行为说明；验证：README 中出现“页内锚点”和“跨文档锚点”的描述，且与实际行为一致

## 2. 脚注 ID 隔离

- [x] 2.1 把 `newMarkdown` 改为接收文档 ID，并使用 `extension.NewFootnote(extension.WithFootnoteIDPrefix(docID+"-"))`；验证：新增测试断言脚注输出包含 `id="doc-0-fn:1"`、`id="doc-0-fnref:1"`
- [x] 2.2 增加多文档脚注测试，确认两篇文档各自使用 `[^1]` 时 ID 不重复，回链指向本文档；验证：`go test ./internal/render -run TestFootnote` 通过，输出中 `id="fn:1"` 出现次数为 0
- [x] 2.3 检查目录模式和链接文档模式共用该逻辑；验证：`go test ./internal/render -run TestBuildDirPage` 和 `TestBuildFilePagePreRendersLinkedLocalDocument` 不回归

## 3. Front matter

- [x] 3.1 实现 `stripFrontMatter`，识别文档开头的 `---` / `+++` 块，限制扫描范围并做 key/value 形状校验；验证：新增 YAML、TOML、无结束分隔符、正文中间水平线四组测试，运行 `go test ./internal/render -run TestFrontMatter`
- [x] 3.2 确认剥离后标题和大纲从正文 H1 开始，不包含元数据；验证：测试断言 HTML 不含 `title: Hi`，且 TOC 不含 front matter 字段
- [x] 3.3 在 README 的特性或已知限制中说明 front matter 会被剥离；验证：README 描述与测试行为一致

## 4. 本地路径解析与异常隔离

- [x] 4.1 抽出共享的本地目标解析函数：绝对路径优先、相对路径字面优先、URL 解码回退、网络路径在 `Stat` 前排除；验证：新增 `100%.md`、`foo%bar.md`、`a%20b.md` 三组测试，运行 `go test ./internal/render -run TestResolveLocalDestination`
- [x] 4.2 改造 `RewriteURL`：解析失败时返回降级目标而不是 error，确保 AST walk 不再中止整篇文档；验证：新增非法链接和非法图片测试，断言正文其余内容仍渲染，运行 `go test ./internal/render -run TestMalformed`
- [x] 4.3 让 `resolveLinkedMarkdown` 复用 4.1 的解析函数，使 `100%.md` 这类链接文档也能被预渲染；验证：新增链接文档含 `100%.md` 的测试，断言文档列表包含该目标
- [x] 4.4 确保 `knownDocs` 查找使用解析后的绝对路径，字面 `%` 文件名和 URL 编码空格都能正确命中；验证：新增 `TestKnownDocumentWithPercentName` 和 `TestKnownDocumentWithEncodedSpace`，运行对应测试
- [x] 4.5 更新 README 的“文档间链接”和“已知限制”，说明本地路径优先按文件名解析；验证：README 描述与 4.1 的行为一致

## 5. 加载上限与目录扫描

- [x] 5.1 在 `scan.Limits` 中增加单文件预算并维护 skipped 信息；单个文件超限时记录 warning 后继续，累计总量超限时才停止；验证：新增“大文件在小文件之前”“大文件在小文件之后”“全部超大”三组测试，运行 `go test ./internal/scan -run TestDirectory`
- [x] 5.2 增加“存在 Markdown 但全部被跳过”的独立错误，和 `ErrNoMarkdown` 区分；验证：新增仅含超大文件的测试，断言返回新的错误类型和可理解的消息
- [x] 5.3 在 `BuildFilePage` 读取前 `Stat` 并检查单文件上限，超限返回包含文件名和上限的错误；验证：新增 `TestBuildFilePageRejectsOversizedFile`，使用可注入的小上限，运行 `go test ./internal/render -run TestBuildFilePage`
- [x] 5.4 增加带 limit 的读取 helper，并在目录文档读取路径使用，防止 `Stat` 与读取之间文件变大；验证：新增 helper 单元测试，断言超过 limit 时返回错误而不是完整读入
- [x] 5.5 更新 README 的目录限制和已知限制，写明单文件上限、跳过警告和截断行为；验证：README 中的数字和行为与新实现一致

## 6. 集成验证

- [x] 6.1 运行完整测试和静态检查：`go test ./... -count=1`、`go vet ./...`，两者均无输出错误
- [x] 6.2 生成一组覆盖锚点、中文标题、脚注、front matter、`100%.md`、超大文件的样例文档，用 OpenMD 打开并检查页面：锚点可跳转、脚注不串篇、front matter 不显示、特殊文件名可读、超大文件有明确提示
- [ ] 6.3 运行 `printf '\n' | cmd.exe //c build.bat`，确认测试通过、版本号和构建时间已注入、`OpenMD.exe` 生成成功

## Workflow follow-up

- 完成人工验收后归档变更。
