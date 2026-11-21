# OpenMD 人工验收案例生成器 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 OpenMD 提供一个按需生成 `cases/` 人工验收案例集的 Go 工具，覆盖现有全部用户可见功能和未来的递归扫描 + 文件树验收素材。

**Architecture:** 在 `tools/cases/` 新增独立 Go 命令，用固定的 `Fixture` 清单描述案例文件，`Generate` 负责安全清理输出目录、写标记、写文件和流式生成大文件，同时输出 `CHECKLIST.md`。生成物通过 `.gitignore` 排除。

**Tech Stack:** Go 1.26 标准库；无新增依赖。

**Spec:** `docs/superpowers/specs/2026-11-21-test-cases-generator-design.md`

## Global Constraints

- 生成的输出目录默认 `./cases`，生成物绝不提交进 Git。
- 默认尺寸：`limits/skip/big-33mib.md` = 33 MiB；`limits/total/01.md` … `05.md` 各 5.5 MiB（5.5 MiB = `11 << 19` 字节，合计 27.5 MiB > 25 MiB）；`limits/at-limit/big-32mib.md` = 恰好 32 MiB；`limits-count/` = 501 篇。
- 大小上限常量必须与 OpenMD 一致：目录 500 篇 / 25 MiB 总量 / 单文件 32 MiB，单文件模式 32 MiB。
- 大文件必须流式写入，不得整块放进内存。
- 生成结果必须确定：无时间戳、无随机数；同一输入重复生成内容一致。
- 输出目录安全规则：不存在或为空 → 生成；含 `.openmd-cases` 标记 → 清空重生成；非空且无标记 → 拒绝，除非 `-force`。
- 中文、空格、`%`、`#` 文件名必须原样落盘。
- 案例内容只求触发格式与功能，不追求语义合理。
- Windows 环境；提交前所有测试通过。

## Review Focus

- 指向包含用户文件的现有目录时必须拒绝生成，除非 `-force`；Task 1 的 `TestPrepareRefusesUnmarkedDir` 固定该行为。
- 上次生成中断或重复运行时，残留文件必须被清掉且不丢标记；Task 1 的 `TestGenerateCleansMarkedDir` 固定该行为。
- 33 MiB / 32 MiB / 5.5 MiB 这些边界值必须精确到字节，否则大小上限验收结论会错；Task 3 的 `TestBigFixtureSizes` 固定该行为。
- 特殊文件名（`100%`、`foo%bar`、`a b`、`中文`、`hash#name`）必须能落盘并出现在预期路径；Task 2 的 `TestLightFixtureStructure` 固定该行为。
- 生成必须可重复且无随机差异；Task 1 的 `TestGenerateIsDeterministic` 固定该行为。

---

### Task 1: 生成器骨架、安全输出目录、确定性

**Files:**
- Create: `tools/cases/main.go`
- Create: `tools/cases/generate.go`
- Test: `tools/cases/generate_test.go`

**Interfaces:**
- Produces:
  - `type Options struct { Out string; Big bool; Force bool; Sizes BigSizes }`
  - `type BigSizes struct { Oversize, TotalFile, AtLimit int64 }`
  - `func DefaultBigSizes() BigSizes`
  - `type Fixture struct { Path string; Data []byte; Write func(io.Writer) error }`
  - `func Generate(opts Options) error`
  - `func prepareOutputDir(out string, force bool) error`
  - `const markerName = ".openmd-cases"`

- [ ] **Step 1: Write the failing tests**

在 `tools/cases/generate_test.go` 中写：

```go
func TestPrepareRefusesUnmarkedDir(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("user data"), 0o644)
    if err := prepareOutputDir(dir, false); err == nil {
        t.Fatal("expected refusal for unmarked non-empty dir")
    }
    if _, err := os.Stat(filepath.Join(dir, "keep.txt")); err != nil {
        t.Fatal("user file must survive refusal")
    }
}

func TestPrepareForceClearsUnmarkedDir(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("user data"), 0o644)
    if err := prepareOutputDir(dir, true); err != nil {
        t.Fatal(err)
    }
    if _, err := os.Stat(filepath.Join(dir, "keep.txt")); !os.IsNotExist(err) {
        t.Fatal("force should clear the dir")
    }
}

func TestGenerateCleansMarkedDir(t *testing.T) {
    dir := t.TempDir()
    if err := Generate(Options{Out: dir, Sizes: SmallTestSizes()}); err != nil {
        t.Fatal(err)
    }
    stray := filepath.Join(dir, "stray.md")
    os.WriteFile(stray, []byte("stray"), 0o644)
    if err := Generate(Options{Out: dir, Sizes: SmallTestSizes()}); err != nil {
        t.Fatal(err)
    }
    if _, err := os.Stat(stray); !os.IsNotExist(err) {
        t.Fatal("marked dir should be cleared before regeneration")
    }
    if _, err := os.Stat(filepath.Join(dir, markerName)); err != nil {
        t.Fatal("marker must exist after generation")
    }
}

func TestGenerateIsDeterministic(t *testing.T) {
    first, second := t.TempDir(), t.TempDir()
    if err := Generate(Options{Out: first, Sizes: SmallTestSizes()}); err != nil {
        t.Fatal(err)
    }
    if err := Generate(Options{Out: second, Sizes: SmallTestSizes()}); err != nil {
        t.Fatal(err)
    }
    compareTrees(t, first, second)
}
```

同时提供测试辅助 `SmallTestSizes()`（返回适合测试的小尺寸：`Oversize: 64 << 10`, `TotalFile: 16 << 10`, `AtLimit: 64 << 10`）和 `compareTrees(t, a, b string)`（递归比较相对路径与文件内容）。

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tools/cases -run 'TestPrepare|TestGenerate' -v`
Expected: FAIL，提示 `undefined: prepareOutputDir` / `undefined: Generate`。

- [ ] **Step 3: Implement CLI and output safety**

`tools/cases/main.go`：解析 `-out`（默认 `cases`）、`-big`、`-force`，调用 `Generate`；出错时 `log.Fatal`，成功时打印输出目录和提示（轻量版/含大文件）。

`tools/cases/generate.go`：
- `DefaultBigSizes()` 返回 `Oversize: 33 << 20`, `TotalFile: 11 << 19`, `AtLimit: 32 << 20`。
- `prepareOutputDir`：目录不存在 → `os.MkdirAll`；存在且含 `markerName` → `os.RemoveAll` 后重建；存在但为空 → 直接使用；存在、非空、无标记且 `force=false` → 返回错误，错误文本包含输出路径和“-force”；`force=true` → 清空重建。
- `Generate`：`opts.Sizes` 为零值时替换为 `DefaultBigSizes()`；`prepareOutputDir` 后先写标记文件，再按切片顺序遍历 `fixtures(opts)` 写文件（本任务返回空列表 + 一个最小 `CHECKLIST.md` 占位，Task 2 替换），最后写 `CHECKLIST.md`。
- 本任务 `fixtures` 可返回空切片，以便测试安全规则与确定性（确定性测试比较两棵树，空目录同样适用）。

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tools/cases -run 'TestPrepare|TestGenerate' -v`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add tools/cases docs/superpowers/specs/2026-11-21-test-cases-generator-design.md
git commit -m "feat: add safe generator skeleton for manual test cases

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 2: 轻量案例文件与验收清单

**Files:**
- Create: `tools/cases/fixtures.go`
- Create: `tools/cases/checklist.go`
- Modify: `tools/cases/generate.go`（接入 `LightFixtures()` 与 `Checklist(opts.Big)`）
- Test: `tools/cases/generate_test.go`（追加）

**Interfaces:**
- Consumes: Task 1 的 `Fixture`、`Generate`、`Options`。
- Produces:
  - `func LightFixtures() []Fixture`
  - `func Checklist(big bool) []byte`
  - `func fixtures(opts Options) []Fixture`（Task 3 在 `opts.Big` 时追加 `BigFixtures`）

- [ ] **Step 1: Write the failing structure test**

```go
func TestLightFixtureStructure(t *testing.T) {
    dir := t.TempDir()
    if err := Generate(Options{Out: dir}); err != nil {
        t.Fatal(err)
    }
    for _, want := range []string{
        "CHECKLIST.md", markerName,
        "01-headings.md", "05-code.md", "06-footnotes.md",
        "07-footnotes-peer.md", "08-anchors.md", "09-links-target.md",
        "10-frontmatter-yaml.md", "11-frontmatter-toml.md", "12-hr-mid.md",
        "13-links.md", "14-images.md", "15-broken.md", "16-empty.md",
        "17-bom.md", "18-unicode.md", "19-html.md", "20-long-line.md",
        "100%.md", "foo%bar.md", "a b.md", "中文文件名.md",
        "UPPER.MARKDOWN", "hash#name.md",
        "fallback/link.md", "fallback/a b.md",
        "literal/link.md", "literal/a%20b.md", "literal/a b.md",
        "sub/nested.md",
        "tree/root.md", "tree/guides/deep/deeper/leaf.md",
        "tree/中文目录/子目录/更多.md",
        "tree/no-md/note.txt", "tree/mixed/logo.png",
        "tree/duplicates/a.md", "tree/duplicates/b/a.md",
        "tree/sort-order/10.md",
        "limits-count/doc-001.md", "limits-count/doc-501.md",
        "common/pixel.png", "common/手册.pdf",
    } {
        if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(want))); err != nil {
            t.Errorf("missing %s: %v", want, err)
        }
    }
    for _, absent := range []string{
        "limits/skip/big-33mib.md", "limits/total/01.md",
        "limits/all-big/big.md", "limits/at-limit/big-32mib.md",
    } {
        if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(absent))); !os.IsNotExist(err) {
            t.Errorf("%s must not exist without -big", absent)
        }
    }
    info, err := os.Stat(filepath.Join(dir, "tree", "empty"))
    if err != nil || !info.IsDir() {
        t.Fatal("tree/empty must be an empty directory")
    }
    if _, err := os.Stat(filepath.Join(dir, "16-empty.md")); err != nil {
        t.Fatal(err)
    }
    if size := mustSize(t, filepath.Join(dir, "16-empty.md")); size != 0 {
        t.Fatalf("16-empty.md must be empty, got %d", size)
    }
    if !bytes.HasPrefix(mustRead(t, filepath.Join(dir, "17-bom.md")), []byte{0xEF, 0xBB, 0xBF}) {
        t.Fatal("17-bom.md must start with UTF-8 BOM")
    }
}
```

测试辅助 `mustRead(t, path) []byte` 和 `mustSize(t, path) int64` 也放在 `generate_test.go`。再追加内容抽查测试 `TestLightFixtureContent`：`10-frontmatter-yaml.md` 以 `---\n` 开头且包含 `# 正文标题`；`fallback/link.md` 含 `a%20b.md`；`15-broken.md` 含 `101%.md`；`14-images.md` 含 `common/pixel.png`；`09-links-target.md` 同时含 `## Usage` 与 `# 中文标题`；`tree/duplicates/a.md` 与 `tree/duplicates/b/a.md` 内容不同。

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./tools/cases -run TestLightFixture -v`
Expected: FAIL，缺文件。

- [ ] **Step 3: Implement light fixtures**

按设计文档布局生成，每篇只需最小内容：

- `01`–`20`：分别覆盖设计文档列出的 Markdown 特征；`06-footnotes.md` 与 `07-footnotes-peer.md` 各用 `[^1]`；`09-links-target.md` 提供 `## Usage` 与 `# 中文标题` 供锚点用例链接；`13-links.md` 含 `https://example.com`、`mailto:`、`tel:`、`common/手册.pdf`、`09-links-target.md`、`\\server\share\doc.md`；`14-images.md` 引用 `common/pixel.png`、`common/pixel one.png`、`common/不存在.png` 和一张 `https://example.com/x.png`；`15-broken.md` 含 `[bad](101%.md)` 与坏图片；`19-html.md` 含 `<script>alert(1)</script>` 与 `<b>inline</b>`；`20-long-line.md` 用约 100 KB 的 `"abcdefghij"` 重复行。
- 根层文件名用例按设计落盘；`fallback/link.md` 链接 `a%20b.md`，`fallback/` 只放 `a b.md`；`literal/` 同时放 `a%20b.md` 与 `a b.md`，`literal/link.md` 链接 `a%20b.md`，两个字面文件内容不同。
- `tree/` 按设计生成嵌套目录、`empty/` 空目录、`no-md/note.txt`、`mixed/`、`duplicates/`、`sort-order/`、`retry/`（无需额外文件）。
- `limits-count/` 生成 `doc-001.md` 到 `doc-501.md`，每篇内容为 `# doc-N\n\n占位内容。\n`。
- `common/pixel.png` 用最小合法 PNG 字节（1×1 像素即可）；`common/pixel one.png` 复制同一图片；`common/手册.pdf` 写 `%PDF-1.4\n` 起头的占位字节。
- `checklist.go` 的 `Checklist(big bool)` 输出设计文档第 1–12 节：标题、生成/安装准备、逐项“操作 → 期望结果”；大小上限与未来文件树小节固定存在；`big=false` 时在大小上限小节注明“需使用 -big 重新生成”。

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tools/cases -run 'TestLightFixture|TestGenerate|TestPrepare' -v`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add tools/cases
git commit -m "feat: generate manual acceptance fixtures and checklist

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 3: 大小上限用例、流式大文件与硬链接

**Files:**
- Create: `tools/cases/bigfiles.go`
- Modify: `tools/cases/fixtures.go`（`fixtures` 在 `opts.Big` 时追加 `BigFixtures(opts.Sizes)`）
- Test: `tools/cases/generate_test.go`（追加）

**Interfaces:**
- Consumes: Task 1 的 `BigSizes`、`Fixture`；Task 2 的 `fixtures`。
- Produces:
  - `func BigFixtures(sizes BigSizes) []Fixture`
  - `func writeRepeatedMarkdown(w io.Writer, size int64) error`
  - `func linkOrCopy(src, dst string) error`

- [ ] **Step 1: Write the failing tests**

```go
func TestBigSizesDefaults(t *testing.T) {
    got := DefaultBigSizes()
    want := BigSizes{Oversize: 33 << 20, TotalFile: 11 << 19, AtLimit: 32 << 20}
    if got != want {
        t.Fatalf("got %+v want %+v", got, want)
    }
}

func TestBigFixtureSizes(t *testing.T) {
    dir := t.TempDir()
    sizes := SmallTestSizes()
    if err := Generate(Options{Out: dir, Big: true, Sizes: sizes}); err != nil {
        t.Fatal(err)
    }
    checks := map[string]int64{
        "limits/skip/big-33mib.md":      sizes.Oversize,
        "limits/total/01.md":            sizes.TotalFile,
        "limits/total/05.md":            sizes.TotalFile,
        "limits/at-limit/big-32mib.md":  sizes.AtLimit,
        "limits/all-big/big.md":         sizes.Oversize,
    }
    for rel, want := range checks {
        if got := mustSize(t, filepath.Join(dir, filepath.FromSlash(rel))); got != want {
            t.Errorf("%s: got %d want %d", rel, got, want)
        }
    }
    if !bytes.Equal(mustRead(t, filepath.Join(dir, "limits", "skip", "big-33mib.md")),
        mustRead(t, filepath.Join(dir, "limits", "all-big", "big.md"))) {
        t.Fatal("all-big must reuse the 33 MiB content")
    }
    if _, err := os.Stat(filepath.Join(dir, "limits", "skip", "small.md")); err != nil {
        t.Fatal("skip case needs a small neighbour document")
    }
}

func TestWriteRepeatedMarkdownExactSize(t *testing.T) {
    for _, size := range []int64{0, 1, 1023, 1 << 20} {
        var buf bytes.Buffer
        if err := writeRepeatedMarkdown(&buf, size); err != nil {
            t.Fatal(err)
        }
        if int64(buf.Len()) != size {
            t.Fatalf("size %d: wrote %d", size, buf.Len())
        }
        if !utf8.Valid(buf.Bytes()) {
            t.Fatalf("size %d: output is not valid UTF-8", size)
        }
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./tools/cases -run 'TestBig|TestWriteRepeated' -v`
Expected: FAIL，`undefined: BigFixtures` / `writeRepeatedMarkdown`。

- [ ] **Step 3: Implement big fixtures**

- `writeRepeatedMarkdown`：先写 `# 大文件用例\n\n`，再以固定块（如 `## 小节\n\n这是一段用于触发大小上限的占位正文。\n\n`）循环写入，直到达到目标字节数；剩余字节用换行补齐，保证输出恰好 `size` 字节且是合法 UTF-8 / Markdown。整个函数只用一个小缓冲区循环 `w.Write`，不构造 `size` 大的切片。
- `BigFixtures`：
  - `limits/skip/big-33mib.md`（`Write` 流式写入 `sizes.Oversize`）+ `limits/skip/small.md`（普通小文档）。
  - `limits/total/01.md`–`05.md`，每个 `sizes.TotalFile`。
  - `limits/all-big/big.md`：用 `Write` 调用 `linkOrCopy` 指向已写好的 skip 大文件。`BigFixtures` 必须把 `limits/skip/big-33mib.md` 排在 `limits/all-big/big.md` 之前；`Generate` 按切片顺序写盘，因此源文件先存在。
  - `limits/at-limit/big-32mib.md`，流式写入 `sizes.AtLimit`。
- `linkOrCopy`：`os.Link(src, dst)`；失败则流式 `io.Copy` 并返回 nil（不因文件系统不支持硬链接而失败）。
- `Checklist(true)` 的大小上限小节保留操作步骤：打开 `limits/skip/` 应看到跳过警告且 `small.md` 可用；打开 `limits/total/` 应截断；打开 `limits/all-big/` 应显示“均超过大小限制”的空状态；直接打开 `limits/skip/big-33mib.md` 应报超限；直接打开 `limits/at-limit/big-32mib.md` 应能渲染（很慢）。

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./tools/cases -v`
Expected: PASS，且测试不会写出超过 100 MB 的文件（使用 `SmallTestSizes`）。

- [ ] **Step 5: Commit**

```bash
git add tools/cases
git commit -m "feat: add size-limit fixtures with streaming and hard links

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 4: 仓库说明、真实生成验证与构建

**Files:**
- Modify: `.gitignore`
- Modify: `README.md`（「开发」小节）
- Test: 运行真实生成器并核对产物

**Interfaces:**
- Consumes: 前三个任务的 `tools/cases`。
- Produces: 可用的 `go run ./tools/cases` / `-big` 工作流。

- [ ] **Step 1: Update `.gitignore` and README**

`.gitignore` 追加一行 `/cases/`。README「开发」小节增加：

```markdown
生成人工验收案例：

```bash
go run ./tools/cases        # 轻量案例，输出到 ./cases
go run ./tools/cases -big   # 追加大小上限用例（约 100 MB）
```

`cases/` 不提交进仓库，可随时整目录删除。验收步骤见生成后的 `cases/CHECKLIST.md`。
```

- [ ] **Step 2: Run all tests and vet**

Run: `go test ./... && go vet ./...`
Expected: 全部 PASS，vet 无输出。

- [ ] **Step 3: Generate the real light suite and inspect**

Run: `go run ./tools/cases -out cases -force`
Expected: 打印生成成功；随后确认：

```bash
test -f cases/CHECKLIST.md && test -f "cases/100%.md" && test -f cases/tree/guides/deep/deeper/leaf.md
find cases -type f | wc -l
```

预期文件数 ≥ 560（含 501 篇计数用例），且不存在 `cases/limits/skip/big-33mib.md`。

- [ ] **Step 4: Generate the real big suite and verify sizes**

Run: `go run ./tools/cases -out cases -big -force`
Expected: 生成成功；确认大文件尺寸：

```bash
stat -c '%s %n' cases/limits/skip/big-33mib.md cases/limits/at-limit/big-32mib.md cases/limits/total/01.md
```
预期分别为 34603008、33554432、5767168 字节（`stat` 在 Git Bash 下可用；不可用则用 `ls -l` 核对）。

- [ ] **Step 5: Commit docs and build**

```bash
git add .gitignore README.md docs/superpowers/specs docs/superpowers/plans
git commit -m "docs: document generated manual acceptance cases

Co-Authored-By: Claude Code <noreply@anthropic.com>"
printf '\n' | cmd.exe //c build.bat
```

Expected: 测试通过，`OpenMD.exe` 与 `source.git.bundle` 更新成功。
