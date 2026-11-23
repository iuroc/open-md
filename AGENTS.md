# OpenMD 开发说明

README 面向产品使用者，只放功能介绍和使用方式；构建、发布、测试和内部实现等开发向内容放在本文件。

## 项目约定

### 每次完成修改后

1. 先在当前分支提交所有已完成的改动，写清提交信息。除非明确要求，不要 push，也不要改写到已有提交。
2. 提交后执行 `go test ./...`，确保测试通过。
3. 正式构建和发布由 `.github/workflows/build.yml` 完成，本地不需要构建发布包。
4. 如果测试或构建失败，先修复、重新提交，再报告完成；不要带着失败结果收尾。
5. 没有完成提交和构建时，不要把工作描述为已完成。

## 构建

需要 Go 1.26+，在项目根目录执行：

```bash
go test ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```

产物是单个 `OpenMD.exe`。手动执行 `go build` 时页面会显示“开发版”；正式版本号和构建时间由 workflow 注入。

## 持续集成与发布

GitHub Actions 工作流见 `.github/workflows/build.yml`：

- Pull Request：只运行 `go vet ./...` 和 `go test ./... -count=1`。
- push 到 `main`：只运行测试，不构建。
- push `v*` 标签：测试通过后构建并创建 GitHub Release，附加 `OpenMD.exe` 和 `OpenMD.exe.sha256`，同时上传 workflow artifact。
- 也可以在 Actions 页面手动触发构建（workflow_dispatch），产物作为 workflow artifact 保留 14 天。

版本号规则：tag 构建去掉前缀 `v` 作为版本号；手动触发的构建使用 `yyyy.MM.dd-<run_number>`。后续正式版本统一使用 `v1.x.x` 语义化版本，tag 仍以 `v` 开头。workflow 通过 `-ldflags -X` 注入 `Version` 和 `BuildTime`。

发布正式版本：

```bash
git tag v1.0.0
git push origin v1.0.0
```

私有仓库的 Actions 和 Release 附件都需要有权限的账号访问；Windows runner 按 2 倍分钟计费。

## 开发

```bash
go test ./...
go vet ./...
```

测试不会修改真实的 Markdown 关联；注册表测试只使用 `HKCU\Software\OpenMDTest-<随机>` 临时键并自动清理。

生成人工验收案例：

```bash
go run ./tools/cases        # 轻量案例，输出到 ./cases
go run ./tools/cases -big   # 追加大小上限用例（约 100 MB）
```

`cases/` 不提交进仓库，可随时整目录删除。验收步骤见生成后的 `cases/CHECKLIST.md`。

## 临时文件

渲染缓存写入：

```text
%LOCALAPPDATA%\Programs\OpenMD\openmd-temp\rendered\<root-hash>\
  manifest.json
  tree.js
  reader.css
  reader.js
  doc-<hash>.html
```

同一源文件始终使用同一个输出路径，内容变化时原子覆盖。源文件、版本和文档集合未变化时直接复用已有页面。不会写到 Markdown 文件所在目录。旧缓存按 7 天 / 最近 50 个 root 清理。

如果安装目录不可写，会回退到 `%TEMP%\OpenMD`。

## 图标

- 图标源文件：`assets/openmd.ico`（白底、青色描边和 M 标记）
- 网页 favicon：`internal/render/web/favicon.png`，由同一个生成器输出并以内嵌 data URI 写入页面
- 重新生成：`go run ./tools/icon`
- 嵌入 exe：`rsrc -ico assets/openmd.ico -manifest assets/app.manifest -o rsrc_windows_amd64.syso`
- `assets/app.manifest` 提供 Common Controls v6、`asInvoker` 和 PerMonitorV2 DPI；修改后需重新生成 `.syso`
- 已提交的 `rsrc_windows_amd64.syso` 会被 `go build` 自动使用，日常构建不需要再安装 `rsrc`
