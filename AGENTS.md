# OpenMD 开发说明

README 面向产品使用者，只放功能介绍和使用方式；构建、发布、测试和内部实现等开发向内容放在本文件。

## 项目约定

### 每次完成修改后

1. 先在当前分支提交所有已完成的改动，写清提交信息。除非明确要求，不要 push，也不要改写到已有提交。
2. 提交后再执行 `build.bat`，确保测试、`OpenMD.exe` 和 `source.git.bundle` 都是最新的。先提交再 build，bundle 才会包含最新提交。
3. `build.bat` 末尾有 `pause`，自动化调用时使用非交互方式：

   ```bash
   printf '\n' | cmd.exe //c build.bat
   ```

4. `OpenMD.exe` 和 `source.git.bundle` 已被 `.gitignore` 忽略，不要提交构建产物。
5. 如果测试或构建失败，先修复、重新提交，再报告完成；不要带着失败结果收尾。
6. 没有完成提交和构建时，不要把工作描述为已完成。

## 构建

需要 Go 1.26+，在项目根目录执行：

```bash
go test ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```

也可以直接双击 `build.bat`：自动运行测试、注入版本号和构建时间、构建 `OpenMD.exe`，并生成 `source.git.bundle`。

只想要源码包时，双击 `bundle.bat` 即可。

产物是单个 `OpenMD.exe`。手动执行上面的 `go build` 时页面会显示“开发版”；用 `build.bat` 构建才会写入正式版本号和构建时间。

## 持续集成与发布

GitHub Actions 工作流见 `.github/workflows/build.yml`：

- Pull Request：只运行 `go vet ./...` 和 `go test ./... -count=1`。
- push 到 `main`：只运行测试，不构建。
- push `v*` 标签：测试通过后构建并创建 GitHub Release，附加 `OpenMD.exe` 和 `OpenMD.exe.sha256`，同时上传 workflow artifact。
- 也可以在 Actions 页面手动触发构建（workflow_dispatch），产物作为 workflow artifact 保留 14 天。

版本号规则：tag 构建去掉前缀 `v` 作为版本号；手动触发的构建使用 `yyyy.MM.dd-<run_number>`。`build.bat` 会识别外部传入的 `BUILD_VERSION`、`BUILD_TIME`；设置 `CI=true` 时跳过 `source.git.bundle` 和 `pause`，设置 `OPENMD_SKIP_TESTS=true` 时跳过测试（CI 中测试已由前一个 job 完成）。

发布正式版本：

```bash
git tag v2026.11.22
git push origin v2026.11.22
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
