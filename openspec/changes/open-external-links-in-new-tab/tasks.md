# Tasks

## 1. 链接打开方式

- [x] 1.1 在 `safeLinkRenderer` 中按最终 `href` 前缀识别外部链接，为 `http://`、`https://`、`file://`、`//` 输出 `target="_blank" rel="noopener noreferrer"`；验证：新增外部 HTTP 和本地 PDF 测试，断言两个属性存在，运行 `go test ./internal/render -run TestExternalLink`
- [x] 1.2 确认 `#doc=` 内部链接、页内锚点、`mailto:` 和 `tel:` 不输出 `target`；验证：新增内部锚点和邮件链接测试，运行 `go test ./internal/render -run TestInternalLink`
- [x] 1.3 更新 README，说明外部链接和本地非 Markdown 文件在新标签页打开，页内锚点和已预渲染文档保持在当前页；验证：README 中的描述与测试断言一致

## 2. 集成验证

- [x] 2.1 运行完整测试和静态检查：`go test ./... -count=1`、`go vet ./...`，两者均无输出错误
- [ ] 2.2 运行 `printf '\n' | cmd.exe //c build.bat`，确认测试通过、版本号和构建时间已注入、`OpenMD.exe` 生成成功

## Workflow follow-up

- 完成人工验收后归档变更。
