# 内部打包品牌隔离设计

日期：2026-11-23
状态：已确认

## 目标

内部共享的 `OpenMD.exe` 在界面和页面源码中不出现 `iuroc` 和项目源自 GitHub 的痕迹。公开版行为不变。

## 非目标

- 不新建分支，不重写 Git 历史。
- 不对外分享源码，因此不处理 LICENSE、README、docs 等源码级内容。
- 不清理二进制里第三方依赖的技术字样（Chroma 的 `github` 主题键、Go 模块信息中的 `github.com/...` 模块路径）。

## 约束

- 公开仓库只保留源码和 `.github/workflows/build.yml`。
- `build.bat`、`bundle.bat` 不再提交，加入 `.gitignore`，本地保留用于内部打包。
- 内部品牌配置放在 `internal/buildinfo/brand_internal.go`，不提交，使用 `.git/info/exclude` 忽略。
- 公开构建与内部构建共用同一份源码，用 Go build tag 选择品牌文件。

## 方案

- `internal/buildinfo/buildinfo.go` 只保留 `Version`、`BuildTime` 和展示逻辑。
- `internal/buildinfo/brand_public.go`（`//go:build !internal`）声明公开品牌变量。
- `internal/buildinfo/brand_internal.go`（`//go:build internal`，本地未跟踪）声明内部品牌变量。
- 模板按 `AuthorID` / `AuthorURL` 条件渲染作者：有工号显示“姓名（工号）”，否则显示作者链接。
- “GitHub Releases 里最新版本……”提示改为中性文案。
- 产出 HTML 里的 `github-markdown.css` 改名为 `markdown.css`，注释去掉 github 字样，保留 Sindre Sorhus 的 MIT 署名。
- 公开构建由 workflow 直接 `go build`；本地 `build.bat` 用 `-tags internal` 构建内部 exe。

## 为什么不用 init 覆盖

如果用 `init()` 覆盖公开默认值，公开身份字符串仍会被编译进内部二进制。build tag 让 `brand_public.go` 在内部构建中完全不参与编译，才能保证 `iuroc` 不进入内部 exe。

## 兼容性

- 公开构建不带 tag，只编译 `brand_public.go`。
- 内部构建带 `-tags internal`，只编译 `brand_internal.go`。
- 测试读取 `buildinfo` 变量而不是硬编码身份，两种构建都能通过。

## 验收

- `go test ./...`、`go vet ./...` 通过。
- `go test -tags internal ./...` 通过。
- 内部 `build.bat` 产出的 `OpenMD.exe` 中 `grep -a -i iuroc` 无结果，页面输出不含 `GitHub Releases`、`github-markdown.css`。
