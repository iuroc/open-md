# 内部打包品牌隔离 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 内部 exe 的用户可见内容不含 `iuroc` / GitHub 身份；公开构建完全由 workflow 完成；本地打包脚本不再提交。

**Architecture:** 用 `!internal` / `internal` build tag 把品牌变量拆成两个文件。公开构建不带 tag，只编译 `brand_public.go`；内部构建带 `-tags internal`，只编译 `brand_internal.go`，因此公开身份字符串不会进入内部二进制。模板按 `AuthorID` 条件渲染作者；可见文案与样式资源名中性化。`build.bat` / `bundle.bat` 移出 Git，本地 `build.bat` 用 `-tags internal` 打包。

**Tech Stack:** Go 1.26、html/template、GitHub Actions (windows-latest)、Windows batch。

**Spec:** `docs/superpowers/specs/2026-11-23-internal-build-branding-design.md`

## Global Constraints

- 公开验证命令：`go vet ./...`、`go test ./...`（不带 tag）。
- 内部验证命令：`go test -tags internal ./...`、`go build -tags internal`。
- 内部用户可见输出不得包含 `iuroc`、`github.com/iuroc`、`GitHub Releases`、`github-markdown.css`。
- 永远不要 `git add`：`build.bat`、`bundle.bat`、`internal/buildinfo/brand_internal.go`、`internal/render/internal_brand_test.go`、`OpenMD.exe`。
- 模块名保持 `openmd`，包路径保持 `openmd/internal/buildinfo`。
- 中文界面文案和全角括号保持原样。

## Review Focus

1. 没有 `brand_internal.go` 的公开构建必须能编译并通过测试。
2. 带 `-tags internal` 的构建必须排除 `brand_public.go`，且 `iuroc` 不出现在二进制中。
3. 作者渲染三种情况：有工号（无链接）、有链接无工号、两者都为空。
4. 站点资源：页面引用并写出 `markdown.css`，不能再引用 `github-markdown.css`。
5. 开发构建（`Version=dev`）仍显示“开发版”、隐藏发布链接；中性文案仍包含“晚于当前构建时间”。

---

### Task 0: 提交设计与计划

**Files:**
- Add: `docs/superpowers/specs/2026-11-23-internal-build-branding-design.md`
- Add: `docs/superpowers/plans/2026-11-23-internal-build-branding.md`

- [ ] **Step 1: Commit**

```bash
git add docs/superpowers/specs/2026-11-23-internal-build-branding-design.md \
        docs/superpowers/plans/2026-11-23-internal-build-branding.md
git commit -m "docs: plan internal build branding isolation" \
  -m "Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 1: Build tag 品牌拆分与作者条件渲染

**Files:**
- Modify: `internal/buildinfo/buildinfo.go`
- Create: `internal/buildinfo/brand_public.go`
- Modify: `internal/render/docpage.go:28-50`
- Modify: `internal/render/page.go:59-61,149-152`
- Modify: `internal/render/web/page.html:9,192`
- Modify: `internal/render/web/docpage.html:27,244`
- Modify: `internal/render/page_test.go:1010-1071`
- Create: `internal/render/docpage_test.go`

**Interfaces:**
- Produces: `buildinfo.AuthorName string`、`buildinfo.AuthorID string`、`buildinfo.AuthorURL string`、`buildinfo.ReleaseURL string`（均为包级变量，由 brand 文件提供）。
- Produces: `render.DocPageData.AuthorID string`、`render.pageData.AuthorID string`（包内）。

- [ ] **Step 1: 写失败测试**

把 `internal/render/page_test.go` 的 `TestRenderPageShowsBuildInfoAndReleaseLink` 替换为：

```go
func TestRenderPageShowsBuildInfoAndReleaseLink(t *testing.T) {
	oldVersion, oldBuildTime := buildinfo.Version, buildinfo.BuildTime
	defer func() {
		buildinfo.Version, buildinfo.BuildTime = oldVersion, oldBuildTime
	}()
	buildinfo.Version = "2026.11.18-0959"
	buildinfo.BuildTime = "2026-11-18T09:59:12+08:00"

	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(out))

	authorMeta := `name="author" content="` + buildinfo.AuthorName
	if buildinfo.AuthorID != "" {
		authorMeta += "（" + buildinfo.AuthorID + "）"
	}
	authorMeta += `"`

	want := []string{
		`id="settings-panel"`,
		authorMeta,
		`data-theme="blue"`,
		`data-theme="teal"`,
		"v2026.11.18-0959",
		"2026-11-18 09:59:12 +08:00",
		`class="release-link panel-action"`,
		"晚于当前构建时间",
	}
	if buildinfo.ReleaseURL != "" {
		want = append(want, buildinfo.ReleaseURL)
	}
	if buildinfo.AuthorID != "" {
		want = append(want, buildinfo.AuthorName+"（"+buildinfo.AuthorID+"）")
	} else if buildinfo.AuthorURL != "" {
		want = append(want, `href="`+buildinfo.AuthorURL+`"`)
	}
	for _, w := range want {
		if !strings.Contains(s, w) {
			t.Fatalf("missing %q", w)
		}
	}
}
```

把 `TestRenderPageHidesReleaseLinkForDevBuild` 中硬编码的 URL 检查替换为：

```go
	if buildinfo.ReleaseURL != "" && strings.Contains(s, buildinfo.ReleaseURL) {
		t.Fatal("dev build must not expose the release URL")
	}
```

在 `page_test.go` 追加：

```go
func TestRenderPageShowsAuthorIDInsteadOfLink(t *testing.T) {
	oldName, oldID, oldURL := buildinfo.AuthorName, buildinfo.AuthorID, buildinfo.AuthorURL
	defer func() {
		buildinfo.AuthorName, buildinfo.AuthorID, buildinfo.AuthorURL = oldName, oldID, oldURL
	}()
	buildinfo.AuthorName = "内部作者"
	buildinfo.AuthorID = "12345"
	buildinfo.AuthorURL = "https://example.invalid/profile"

	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(out))
	if !strings.Contains(s, `name="author" content="内部作者（12345）"`) {
		t.Fatal("author meta must include the employee ID")
	}
	if !strings.Contains(s, "内部作者（12345）") {
		t.Fatal("author line must include the employee ID")
	}
	if strings.Contains(s, `href="https://example.invalid/profile"`) {
		t.Fatal("author link must not be rendered when an employee ID is set")
	}
}

func TestRenderPageShowsPlainAuthorWhenNoIDOrURL(t *testing.T) {
	oldName, oldID, oldURL := buildinfo.AuthorName, buildinfo.AuthorID, buildinfo.AuthorURL
	defer func() {
		buildinfo.AuthorName, buildinfo.AuthorID, buildinfo.AuthorURL = oldName, oldID, oldURL
	}()
	buildinfo.AuthorName = "匿名作者"
	buildinfo.AuthorID = ""
	buildinfo.AuthorURL = ""

	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(out))
	if !strings.Contains(s, `name="author" content="匿名作者"`) {
		t.Fatal("author meta must contain the plain author name")
	}
	if !strings.Contains(s, "匿名作者") {
		t.Fatal("author line must contain the plain author name")
	}
}
```

创建 `internal/render/docpage_test.go`：

```go
package render

import (
	"html"
	"html/template"
	"strings"
	"testing"
)

func TestRenderDocPageShowsAuthorIDInsteadOfLink(t *testing.T) {
	out, err := RenderDocPage(DocPageData{
		Title:      "A",
		Doc:        Document{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")},
		AuthorName: "内部作者",
		AuthorID:   "12345",
		AuthorURL:  "https://example.invalid/profile",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(out))
	if !strings.Contains(s, `name="author" content="内部作者（12345）"`) {
		t.Fatal("author meta must include the employee ID")
	}
	if !strings.Contains(s, "内部作者（12345）") {
		t.Fatal("author line must include the employee ID")
	}
	if strings.Contains(s, `href="https://example.invalid/profile"`) {
		t.Fatal("author link must not be rendered when an employee ID is set")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/render -run 'TestRenderPageShows|TestRenderPageHides|TestRenderDocPageShows' -count=1`

Expected: 编译失败，提示 `unknown field AuthorID` 或 `buildinfo.AuthorID undefined`。

- [ ] **Step 3: 拆分 buildinfo**

`internal/buildinfo/buildinfo.go` 删除 `const` 块，最终只保留：

```go
package buildinfo

import "time"

var (
	Version   = "dev"
	BuildTime = ""
)

func IsReleaseBuild() bool { /* 不变 */ }
func VersionLabel() string { /* 不变 */ }
func BuildTimeDisplay() string { /* 不变 */ }
```

创建 `internal/buildinfo/brand_public.go`：

```go
//go:build !internal

package buildinfo

var (
	AuthorName = "github.com/iuroc"
	AuthorID   = ""
	AuthorURL  = "https://github.com/iuroc"
	ReleaseURL = "https://github.com/iuroc/open-md/releases"
)
```

- [ ] **Step 4: 增加 AuthorID 数据字段**

`internal/render/docpage.go` 的 `DocPageData` 在 `AuthorName` 后加 `AuthorID string`；`prepareDocPageData` 在 AuthorName 默认值后加：

```go
	if data.AuthorID == "" {
		data.AuthorID = buildinfo.AuthorID
	}
```

`internal/render/page.go` 的 `pageData` 在 `AuthorName` 后加 `AuthorID string`；`RenderPage` 的数据初始化在 `AuthorName` 后加：

```go
		AuthorID:         buildinfo.AuthorID,
```

- [ ] **Step 5: 改模板**

`internal/render/web/page.html` 和 `internal/render/web/docpage.html` 的 `<meta name="author">` 都改成：

```html
<meta name="author" content="{{.AuthorName}}{{if .AuthorID}}（{{.AuthorID}}）{{end}}">
```

两处作者行都改成：

```html
<div><dt>作者</dt><dd class="author-line">{{if .AuthorID}}{{.AuthorName}}（{{.AuthorID}}）{{else if .AuthorURL}}<a href="{{.AuthorURL}}" target="_blank" rel="noopener noreferrer">{{.AuthorName}}</a>{{else}}{{.AuthorName}}{{end}}</dd></div>
```

- [ ] **Step 6: 运行测试，确认通过**

Run: `go test ./internal/render ./internal/buildinfo -count=1`

Expected: PASS。

- [ ] **Step 7: 全量验证并提交**

```bash
go vet ./...
go test ./... -count=1
git add internal/buildinfo internal/render
git commit -m "refactor: make branding configurable by build tag" \
  -m "Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 2: 中性化可见文案与样式资源名

**Files:**
- Modify: `internal/render/web/page.html:2,197`
- Modify: `internal/render/web/docpage.html:19,88,249`
- Modify: `internal/render/site.go:672`
- Modify: `internal/render/page_test.go:911-930`
- Modify: `internal/render/site_test.go:89-91`

- [ ] **Step 1: 写失败测试**

`page_test.go` 把 `TestRenderPageIncludesGithubMarkdownCSS` 改名为 `TestRenderPageIncludesMarkdownCSS`，失败信息改为 `"markdown CSS not embedded"`，并追加：

```go
func TestRenderPageHasNoVendorBrandingText(t *testing.T) {
	page := &Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	}
	out, err := RenderPage(page)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, bad := range []string{"github-markdown.css", "GitHub Releases"} {
		if strings.Contains(s, bad) {
			t.Fatalf("rendered page still contains %q", bad)
		}
	}
}
```

`site_test.go` 的 `TestBuildDirSiteCreatesIndependentPages` 在 `reader.css` 检查后加：

```go
	if !strings.Contains(string(page), `href="markdown.css"`) {
		t.Fatal("shared markdown.css not referenced")
	}
	if !fileExists(filepath.Join(filepath.Dir(site.InitialPage), "markdown.css")) {
		t.Fatal("markdown.css was not written")
	}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/render -run 'TestRenderPageIncludesMarkdownCSS|TestRenderPageHasNoVendorBrandingText|TestBuildDirSiteCreatesIndependentPages' -count=1`

Expected: FAIL，提示仍引用 `github-markdown.css` 或页面包含 `GitHub Releases`。

- [ ] **Step 3: 修改文案和资源名**

两个模板的 CSS 注释改为：

```html
<!-- Body styles based on an MIT-licensed stylesheet, Copyright (c) Sindre Sorhus. -->
```

两处提示改为：

```html
<p class="about-hint">最新版本的构建时间晚于当前构建时间时，说明有更新。</p>
```

`docpage.html` 的样式表链接改为：

```html
<link rel="stylesheet" href="markdown.css">
```

`internal/render/site.go` 的 `writeAssets` 映射改为：

```go
		"markdown.css":        githubMarkdownCSS,
```

- [ ] **Step 4: 运行测试，确认通过**

Run: `go test ./internal/render -count=1`

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
go vet ./...
go test ./... -count=1
git add internal/render
git commit -m "refactor: neutralize visible copy and stylesheet asset name" \
  -m "Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 3: 构建脚本移出仓库，workflow 自包含

**Files:**
- Modify: `.github/workflows/build.yml:58-63`
- Modify: `.gitignore`
- Modify: `AGENTS.md`
- Remove from index: `build.bat`、`bundle.bat`（工作区文件保留）

- [ ] **Step 1: 修改 workflow**

把 `Build` 步骤替换为：

```yaml
      - name: Build
        shell: pwsh
        run: |
          $ld = "-s -w -H=windowsgui -X openmd/internal/buildinfo.Version=$env:BUILD_VERSION -X openmd/internal/buildinfo.BuildTime=$env:BUILD_TIME"
          go build -trimpath -ldflags $ld -o OpenMD.exe .
```

- [ ] **Step 2: 忽略本地脚本**

`.gitignore` 追加：

```gitignore
/build.bat
/bundle.bat
```

- [ ] **Step 3: 更新 AGENTS.md**

把“每次完成修改后”一节替换为：

~~~markdown
### 每次完成修改后

1. 先在当前分支提交所有已完成的改动，写清提交信息。除非明确要求，不要 push，也不要改写到已有提交。
2. 提交后执行 `go test ./...`，确保测试通过。
3. 正式构建和发布由 `.github/workflows/build.yml` 完成，本地不需要构建发布包。
4. 如果测试或构建失败，先修复、重新提交，再报告完成；不要带着失败结果收尾。
5. 没有完成提交和构建时，不要把工作描述为已完成。
~~~

把“构建”一节替换为：

~~~markdown
## 构建

需要 Go 1.26+，在项目根目录执行：

```bash
go test ./...
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```

手动执行 `go build` 时页面会显示“开发版”；正式版本号和构建时间由 workflow 注入。
~~~

删除 `build.bat` / `bundle.bat` 相关段落。CI 一节末尾关于 `build.bat` 识别环境变量的句子替换为：

~~~markdown
版本号规则：tag 构建去掉前缀 `v` 作为版本号；手动触发的构建使用 `yyyy.MM.dd-<run_number>`。后续正式版本统一使用 `v1.x.x` 语义化版本，tag 仍以 `v` 开头。workflow 通过 `-ldflags -X` 注入 `Version` 和 `BuildTime`。
~~~

- [ ] **Step 4: 移出 Git 跟踪并提交**

```bash
git rm --cached build.bat bundle.bat
git add .gitignore .github/workflows/build.yml AGENTS.md
git commit -m "chore: keep build scripts out of the repository" \
  -m "Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

- [ ] **Step 5: 验证脚本仍在本地且被忽略**

Run: `ls build.bat bundle.bat && git check-ignore -v build.bat bundle.bat`

Expected: 两个文件都存在；输出两条 `.gitignore` 规则。

- [ ] **Step 6: 验证公开构建**

Run: `go vet ./... && go test ./... -count=1`

Expected: PASS。

---

### Task 4: 本地内部打包与端到端验证

**Files (全部不提交):**
- Create: `internal/buildinfo/brand_internal.go`
- Create: `internal/render/internal_brand_test.go`
- Rewrite: `build.bat`
- Keep: `bundle.bat`
- Modify: `.git/info/exclude`

- [ ] **Step 1: 写内部品牌配置**

创建 `internal/buildinfo/brand_internal.go`：

```go
//go:build internal

package buildinfo

var (
	AuthorName = "易舒"
	AuthorID   = "25768"
	AuthorURL  = ""
	ReleaseURL = "https://t.2980.com/510Qkm"
)
```

- [ ] **Step 2: 写内部构建验收测试**

创建 `internal/render/internal_brand_test.go`：

```go
//go:build internal

package render

import (
	"html/template"
	"strings"
	"testing"

	"openmd/internal/buildinfo"
)

func TestInternalBuildHasNoPublicBranding(t *testing.T) {
	page, err := RenderPage(&Page{
		Mode:      "file",
		Title:     "A",
		Documents: []Document{{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	docPage, err := RenderDocPage(DocPageData{
		Title: "A",
		Doc:   Document{ID: "doc-0", Title: "A", HTML: template.HTML("<p>a</p>")},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{string(page), string(docPage)} {
		if !strings.Contains(s, buildinfo.AuthorName) {
			t.Fatalf("internal author name %q missing", buildinfo.AuthorName)
		}
		if buildinfo.AuthorID != "" && !strings.Contains(s, buildinfo.AuthorID) {
			t.Fatalf("internal author ID %q missing", buildinfo.AuthorID)
		}
		for _, bad := range []string{"iuroc", "github.com/", "GitHub Releases", "github-markdown.css"} {
			if strings.Contains(s, bad) {
				t.Fatalf("internal output still contains %q", bad)
			}
		}
	}
}
```

- [ ] **Step 3: 加入本地忽略**

追加到 `.git/info/exclude`：

```text
/internal/buildinfo/brand_internal.go
/internal/render/internal_brand_test.go
```

- [ ] **Step 4: 重写本地 build.bat**

```bat
@echo off
setlocal
cd /d "%~dp0"

where go >nul 2>nul
if errorlevel 1 (
    echo Go was not found in PATH.
    pause
    exit /b 1
)

if not exist "internal\buildinfo\brand_internal.go" (
    echo Missing internal\buildinfo\brand_internal.go.
    pause
    exit /b 1
)

if not defined BUILD_VERSION (
    for /f "delims=" %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy.MM.dd-HHmm"') do set "BUILD_VERSION=%%i"
)
if not defined BUILD_TIME (
    for /f "delims=" %%i in ('powershell -NoProfile -Command "Get-Date -Format yyyy-MM-ddTHH:mm:sszzz"') do set "BUILD_TIME=%%i"
)

echo Running tests...
go test -tags internal ./... -count=1
if errorlevel 1 (
    echo.
    echo Tests failed. Build aborted.
    pause
    exit /b 1
)

echo.
echo Building OpenMD.exe...
echo   Version:    %BUILD_VERSION%
echo   Build time: %BUILD_TIME%
go build -tags internal -trimpath -buildvcs=false -ldflags "-s -w -H=windowsgui -X openmd/internal/buildinfo.Version=%BUILD_VERSION% -X openmd/internal/buildinfo.BuildTime=%BUILD_TIME%" -o OpenMD.exe .
if errorlevel 1 (
    echo.
    echo Build failed.
    pause
    exit /b 1
)

echo.
echo Build complete: %CD%\OpenMD.exe
pause
```

`bundle.bat` 保持现有内容不变。

- [ ] **Step 5: 运行内部测试**

Run: `go test -tags internal ./... -count=1`

Expected: PASS，包括 `TestInternalBuildHasNoPublicBranding`。

- [ ] **Step 6: 构建内部 exe**

Run: `printf '\n' | cmd.exe //c build.bat`

Expected: 测试通过，输出 `Build complete: ...\OpenMD.exe`。

- [ ] **Step 7: 验证二进制与工作区**

```bash
grep -a -i iuroc OpenMD.exe
grep -a "GitHub Releases" OpenMD.exe
grep -a "github-markdown.css" OpenMD.exe
git status --short
```

Expected: 前三条 `grep` 无输出；`git status --short` 为空。

```bash
grep -a -i github OpenMD.exe | head
```

Expected: 可能仍有 Chroma 主题键或 Go 模块路径中的 `github`，属于第三方依赖，界面不可见；不要求为空。

- [ ] **Step 8: 确认公开构建仍正常**

Run: `go vet ./... && go test ./... -count=1 && go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD-public.exe .`

Expected: PASS，`OpenMD-public.exe` 生成成功。随后删除该文件：

```bash
rm -f OpenMD-public.exe
```
