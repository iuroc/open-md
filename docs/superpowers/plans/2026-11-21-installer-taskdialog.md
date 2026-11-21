# OpenMD Installer TaskDialog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the installer MessageBox with a native TaskDialog, install directly on first run, and show at most one management dialog.

**Architecture:** Add a comctl32 v6 manifest, a packed `TASKDIALOGCONFIG` serializer, and a TaskDialog wrapper in `internal/ui`. Change the Dialog interface to an `InstallMode` enum; `app.runInstaller` installs first when absent and asks manage actions when present.

**Tech Stack:** Go 1.26, `golang.org/x/sys/windows`, Win32 comctl32 `TaskDialogIndirect`, `rsrc` for resources.

**Spec:** `docs/superpowers/specs/2026-11-21-installer-taskdialog-design.md`

## Global Constraints

- Windows-only; no cgo; `go build` must remain a pure Go build.
- No WebView2, Electron, or bundled runtime.
- Button IDs: `100` install/overwrite, `101` uninstall, `102` done/cancel.
- `TASKDIALOGCONFIG` is packed: 64-bit `cbSize` is **160**; `TASKDIALOG_BUTTON` is **12** bytes.
- Never set `TDF_USE_COMMAND_LINKS` when `cButtons == 0`.
- Success path shows at most one dialog; no `Info` success popups.
- Commit before running `build.bat`; build with `printf '\n' | cmd.exe //c build.bat`.

## Review Focus

- Packed layout mistakes produce `E_INVALIDARG`; the byte-offset test must pin every field the dialog reads.
- Command-link flags with no custom buttons produce `E_INVALIDARG`; the error dialog must not set them.
- Fresh install must finish before the post-install dialog; a failed install shows only the error dialog.
- Uninstalling while running the installed copy must still spawn the helper.
- The MessageBox fallback must map “No” to cancel/keep-installed, never to install.

---

### Task 1: Manifest and resource file

**Files:**
- Create: `assets/app.manifest`
- Create: `docs/superpowers/specs/2026-11-21-installer-taskdialog-design.md`
- Create: `docs/superpowers/plans/2026-11-21-installer-taskdialog.md`
- Modify: `rsrc_windows_amd64.syso` (regenerated)

**Interfaces:**
- Produces: an exe manifest that activates comctl32 v6 for TaskDialogInDirect.

- [ ] **Step 1: Write `assets/app.manifest`**

Use this exact content:

```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity type="win32" name="OpenMD" version="1.0.0.0" processorArchitecture="*"/>
  <description>OpenMD</description>
  <dependency>
    <dependentAssembly>
      <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls" version="6.0.0.0" processorArchitecture="*" publicKeyToken="6595b64144ccf1df" language="*"/>
    </dependentAssembly>
  </dependency>
  <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3">
    <security>
      <requestedPrivileges>
        <requestedExecutionLevel level="asInvoker" uiAccess="false"/>
      </requestedPrivileges>
    </security>
  </trustInfo>
  <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1">
    <application>
      <supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
    </application>
  </compatibility>
  <application xmlns="urn:schemas-microsoft-com:asm.v3">
    <windowsSettings>
      <dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">true/pm</dpiAware>
      <dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">PerMonitorV2</dpiAwareness>
    </windowsSettings>
  </application>
</assembly>
```

- [ ] **Step 2: Regenerate the resource file**

Run from the repo root:

```bash
"$(go env GOPATH)/bin/rsrc.exe" -ico assets/openmd.ico -manifest assets/app.manifest -o rsrc_windows_amd64.syso
```

Expected: `rsrc_windows_amd64.syso` is rewritten, no error.

- [ ] **Step 3: Build the current app**

Run:

```bash
go build -trimpath -ldflags "-s -w -H=windowsgui" -o OpenMD.exe .
```

Expected: build succeeds. The dialog is still the old MessageBox until Task 3.

- [ ] **Step 4: Commit**

```bash
git add assets/app.manifest rsrc_windows_amd64.syso docs/superpowers/specs/2026-11-21-installer-taskdialog-design.md docs/superpowers/plans/2026-11-21-installer-taskdialog.md
git commit -m "feat: embed comctl32 v6 manifest for native task dialogs"
```

---

### Task 2: Packed TaskDialog wrapper

**Files:**
- Create: `internal/ui/taskdialog_windows.go`
- Create: `internal/ui/taskdialog_windows_test.go`

**Interfaces:**
- Produces:
  - `type taskDialogButtonSpec struct { id int32; text string }`
  - `type taskDialogOptions struct { title, instruction, content string; buttons []taskDialogButtonSpec; icon uintptr; defaultID int32; commonButtons uint32; commandLinks bool }`
  - `func buildTaskDialogConfig(opts taskDialogOptions, hInstance uintptr) (cfg []byte, btnBuf []byte, refs []*uint16)`
  - `func showTaskDialog(opts taskDialogOptions) (int32, error)`
  - `func loadAppIcon() (uintptr, func())`
  - `func loadSystemIcon(id uintptr) uintptr`
- Constants: `taskDialogConfigSize = 160`, `taskDialogButtonSize = 12`, flags `0x0002/0x0008/0x0010`, button IDs `100/101/102`, `idiError = 32513`, `idiInformation = 32516`, `mbOK = 1`, `mbIconError = 0x10`.

- [ ] **Step 1: Write the failing packed-layout test**

Create `internal/ui/taskdialog_windows_test.go` with:

```go
func TestBuildTaskDialogConfigPacked(t *testing.T) {
    opts := taskDialogOptions{
        title:         "OpenMD",
        instruction:   "OpenMD 已安装",
        content:       "正文",
        buttons:       []taskDialogButtonSpec{{id: 100, text: "覆盖安装"}, {id: 101, text: "卸载"}},
        icon:          0x1234,
        defaultID:     100,
        commandLinks:  true,
    }
    cfg, btns, refs := buildTaskDialogConfig(opts, 0x5678)
    if len(cfg) != 160 { t.Fatalf("len(cfg) = %d", len(cfg)) }
    if got := binary.LittleEndian.Uint32(cfg[0:]); got != 160 { t.Fatalf("cbSize = %d", got) }
    if got := binary.LittleEndian.Uint64(cfg[4:]); got != 0 { t.Fatalf("hwndParent = %#x", got) }
    if got := binary.LittleEndian.Uint64(cfg[12:]); got != 0x5678 { t.Fatalf("hInstance = %#x", got) }
    if got := binary.LittleEndian.Uint32(cfg[20:]); got != 0x1a { t.Fatalf("flags = %#x", got) }
    if got := binary.LittleEndian.Uint64(cfg[36:]); got != 0x1234 { t.Fatalf("mainIcon = %#x", got) }
    if got := binary.LittleEndian.Uint32(cfg[60:]); got != 2 { t.Fatalf("cButtons = %d", got) }
    if got := binary.LittleEndian.Uint64(cfg[64:]); got == 0 { t.Fatal("pButtons is NULL") }
    if got := binary.LittleEndian.Uint32(cfg[72:]); got != 100 { t.Fatalf("nDefaultButton = %d", got) }
    if len(btns) != 24 { t.Fatalf("len(btns) = %d", len(btns)) }
    if got := binary.LittleEndian.Uint32(btns[0:]); got != 100 { t.Fatalf("button 0 id = %d", got) }
    if got := binary.LittleEndian.Uint64(btns[4:]); got == 0 { t.Fatal("button 0 text is NULL") }
    if got := binary.LittleEndian.Uint32(btns[12:]); got != 101 { t.Fatalf("button 1 id = %d", got) }
    if got := binary.LittleEndian.Uint64(btns[16:]); got == 0 { t.Fatal("button 1 text is NULL") }
    if len(refs) != 5 { t.Fatalf("len(refs) = %d", len(refs)) }
}

func TestBuildTaskDialogConfigNoCommandLinksWithoutButtons(t *testing.T) {
    cfg, _, _ := buildTaskDialogConfig(taskDialogOptions{
        title: "OpenMD", instruction: "出错了", content: "正文",
        icon: 0x1234, commonButtons: 1,
    }, 0x5678)
    if got := binary.LittleEndian.Uint32(cfg[20:]); got&0x10 != 0 {
        t.Fatalf("command-link flag set without buttons: %#x", got)
    }
    if got := binary.LittleEndian.Uint32(cfg[24:]); got != 1 { t.Fatalf("common buttons = %#x", got) }
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui -run 'TestBuildTaskDialogConfig' -v`

Expected: compile failure because `buildTaskDialogConfig` and the types do not exist.

- [ ] **Step 3: Implement the packed serializer**

Create `internal/ui/taskdialog_windows.go`:

- Encode `TASKDIALOGCONFIG` into a `[]byte` of length 160 at these little-endian offsets:
  - `0` cbSize (4), `4` hwndParent (8), `12` hInstance (8), `20` dwFlags (4), `24` dwCommonButtons (4),
  - `28` pszWindowTitle (8), `36` mainIcon (8), `44` pszMainInstruction (8), `52` pszContent (8),
  - `60` cButtons (4), `64` pButtons (8), `72` nDefaultButton (4),
  - `132` pszFooter (8; leave zero).
- Encode each button into `btnBuf` at `12*i`: `id` at `0` (4), text pointer at `4` (8).
- `flags = tdfAllowDialogCancellation`; add `tdfUseCommandLinks` only when `len(buttons) > 0 && commandLinks`; add `tdfUseHIconMain` when `icon != 0`.
- Create UTF-16 pointers with `windows.UTF16PtrFromString`; append each pointer to `refs` so the caller can `runtime.KeepAlive(refs)`.
- `showTaskDialog`:
  - Find `comctl32.NewLazySystemDLL("comctl32.dll").NewProc("TaskDialogIndirect")`; return an error if not found.
  - Build the config with `hInstance` from `kernel32!GetModuleHandleW(0)`.
  - Call with `(cfgPtr, &button, 0, 0)`; return an error when `int32(hr) < 0`.
  - `runtime.KeepAlive` the refs, button buffer, and config after the call.
- `loadAppIcon`:
  - `os.Executable()` → `ExtractIconExW(exe, 0, &large, &small, 1)`.
  - Destroy `small` immediately; return `large` and a cleanup that destroys it.
  - If `large == 0`, return `loadSystemIcon(idiInformation)` with a no-op cleanup.
- `loadSystemIcon` uses `user32!LoadIconW(0, id)`; shared, no destroy.
- Reuse the existing `messageBox` helper in `dialog_windows.go`; do not define a second one.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/ui -run 'TestBuildTaskDialogConfig' -v`

Expected: both tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/taskdialog_windows.go internal/ui/taskdialog_windows_test.go
git commit -m "feat: add packed TaskDialog wrapper"
```

---

### Task 3: InstallMode interface and app flow

**Files:**
- Create: `internal/ui/dialog.go`
- Modify: `internal/ui/dialog_windows.go`
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`
- Create: `internal/ui/dialog_windows_test.go`

**Interfaces:**
- Produces:
  - `type InstallMode int` with `InstallAfterInstall` and `InstallManage`.
  - `Dialog.Installer(mode InstallMode) Choice`; `Info` removed from the interface.
  - `func installDialogCopyFor(mode InstallMode) installDialogCopy`
  - `func choiceFromButton(id int32) Choice`
  - `func fallbackInstallerChoice(mode InstallMode, result uintptr) Choice`
- Consumes: Task 2's `taskDialogOptions`, `taskDialogButtonSpec`, `showTaskDialog`, `loadAppIcon`, `loadSystemIcon`, `messageBox`.

- [ ] **Step 1: Write the failing UI tests**

Create `internal/ui/dialog_windows_test.go`:

```go
func TestInstallDialogCopyAfterInstall(t *testing.T) {
    c := installDialogCopyFor(InstallAfterInstall)
    if c.instruction != "OpenMD 已安装" { t.Fatalf("instruction = %q", c.instruction) }
    if !strings.Contains(c.content, "右键") { t.Fatalf("content = %q", c.content) }
    if len(c.buttons) != 2 || c.buttons[0].id != 101 || c.buttons[0].text != "卸载" ||
        c.buttons[1].id != 102 || c.buttons[1].text != "完成" {
        t.Fatalf("buttons = %#v", c.buttons)
    }
    if c.defaultID != 102 { t.Fatalf("default = %d", c.defaultID) }
}

func TestInstallDialogCopyManage(t *testing.T) {
    c := installDialogCopyFor(InstallManage)
    if len(c.buttons) != 3 || c.buttons[0].id != 100 || c.buttons[0].text != "覆盖安装" ||
        c.buttons[1].id != 101 || c.buttons[1].text != "卸载" ||
        c.buttons[2].id != 102 || c.buttons[2].text != "取消" {
        t.Fatalf("buttons = %#v", c.buttons)
    }
    if c.defaultID != 102 { t.Fatalf("default = %d", c.defaultID) }
}

func TestChoiceFromButton(t *testing.T) {
    cases := []struct{ id int32; want Choice }{
        {100, ChoiceInstall}, {101, ChoiceUninstall}, {102, ChoiceCancel}, {2, ChoiceCancel},
    }
    for _, tc := range cases {
        if got := choiceFromButton(tc.id); got != tc.want {
            t.Fatalf("choiceFromButton(%d) = %v", tc.id, got)
        }
    }
}

func TestFallbackInstallerChoice(t *testing.T) {
    const yes, no, cancel = 6, 7, 2
    if got := fallbackInstallerChoice(InstallAfterInstall, yes); got != ChoiceUninstall { t.Fatalf("after yes = %v", got) }
    if got := fallbackInstallerChoice(InstallAfterInstall, no); got != ChoiceCancel { t.Fatalf("after no = %v", got) }
    if got := fallbackInstallerChoice(InstallAfterInstall, cancel); got != ChoiceCancel { t.Fatalf("after cancel = %v", got) }
    if got := fallbackInstallerChoice(InstallManage, yes); got != ChoiceInstall { t.Fatalf("manage yes = %v", got) }
    if got := fallbackInstallerChoice(InstallManage, no); got != ChoiceUninstall { t.Fatalf("manage no = %v", got) }
    if got := fallbackInstallerChoice(InstallManage, cancel); got != ChoiceCancel { t.Fatalf("manage cancel = %v", got) }
}
```

- [ ] **Step 2: Write the failing app-flow tests**

Replace the two installer tests in `internal/app/run_test.go` with:

```go
func TestRunFreshInstallInstallsThenAsksAfterInstall(t *testing.T) {
    installed := filepath.Join(t.TempDir(), "OpenMD.exe") // deliberately absent
    fi := &fakeInstaller{installedExe: installed}
    d := &fakeDialog{choice: ui.ChoiceCancel}
    a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil }}
    if err := a.Run(Invocation{Mode: ModeInstall}); err != nil { t.Fatal(err) }
    if !fi.installCalled || fi.installSource != `C:\Downloads\OpenMD.exe` { t.Fatalf("install = %v %q", fi.installCalled, fi.installSource) }
    if len(d.installModes) != 1 || d.installModes[0] != ui.InstallAfterInstall { t.Fatalf("modes = %#v", d.installModes) }
    if fi.uninstallCalled { t.Fatal("uninstall called") }
}

func TestRunFreshInstallCanUninstallAfterInstall(t *testing.T) {
    installed := filepath.Join(t.TempDir(), "OpenMD.exe")
    fi := &fakeInstaller{installedExe: installed}
    d := &fakeDialog{choice: ui.ChoiceUninstall}
    a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil }}
    if err := a.Run(Invocation{Mode: ModeInstall}); err != nil { t.Fatal(err) }
    if !fi.installCalled || !fi.uninstallCalled { t.Fatalf("install=%v uninstall=%v", fi.installCalled, fi.uninstallCalled) }
}

func TestRunInstalledAsksManageBeforeAction(t *testing.T) {
    installed := filepath.Join(t.TempDir(), "OpenMD.exe")
    if err := os.WriteFile(installed, []byte("x"), 0644); err != nil { t.Fatal(err) }
    fi := &fakeInstaller{installedExe: installed}
    d := &fakeDialog{choice: ui.ChoiceCancel}
    a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return installed, nil }}
    if err := a.Run(Invocation{Mode: ModeInstall}); err != nil { t.Fatal(err) }
    if len(d.installModes) != 1 || d.installModes[0] != ui.InstallManage { t.Fatalf("modes = %#v", d.installModes) }
    if fi.installCalled || fi.uninstallCalled { t.Fatalf("install=%v uninstall=%v", fi.installCalled, fi.uninstallCalled) }
}

func TestRunInstalledOverwriteInstalls(t *testing.T) {
    installed := filepath.Join(t.TempDir(), "OpenMD.exe")
    if err := os.WriteFile(installed, []byte("x"), 0644); err != nil { t.Fatal(err) }
    fi := &fakeInstaller{installedExe: installed}
    d := &fakeDialog{choice: ui.ChoiceInstall}
    a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil }}
    if err := a.Run(Invocation{Mode: ModeInstall}); err != nil { t.Fatal(err) }
    if !fi.installCalled { t.Fatal("install not called") }
}

func TestRunInstalledUninstallSpawnsHelperWhenNeeded(t *testing.T) {
    installed := filepath.Join(t.TempDir(), "OpenMD.exe")
    if err := os.WriteFile(installed, []byte("x"), 0644); err != nil { t.Fatal(err) }
    fi := &fakeInstaller{installedExe: installed, uninstallResult: install.UninstallResult{NeedsHelper: true}}
    d := &fakeDialog{choice: ui.ChoiceUninstall}
    a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return installed, nil }}
    if err := a.Run(Invocation{Mode: ModeInstall}); err != nil { t.Fatal(err) }
    if !fi.uninstallCalled || !fi.helperCalled { t.Fatalf("uninstall=%v helper=%v", fi.uninstallCalled, fi.helperCalled) }
}
```

Update `fakeDialog`:

```go
type fakeDialog struct {
    choice       ui.Choice
    load         ui.LoadDecision
    confirmSeen  []string
    errors       []string
    installModes []ui.InstallMode
}

func (f *fakeDialog) Installer(mode ui.InstallMode) ui.Choice {
    f.installModes = append(f.installModes, mode)
    return f.choice
}
```

Remove `Info` and the `events`/`infos` fields; remove `events` from `fakeInstaller` if it is now unused.

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./internal/ui ./internal/app -run 'TestInstallDialog|TestChoiceFromButton|TestFallback|TestRunFresh|TestRunInstalled' -v`

Expected: compile failure because `InstallMode`, `installDialogCopyFor`, and the new `Dialog` signature do not exist.

- [ ] **Step 4: Create `internal/ui/dialog.go`**

Move the platform-independent types and interface here, and delete the old copies from `dialog_windows.go`:

```go
type Choice int
const ( ChoiceCancel Choice = iota; ChoiceInstall; ChoiceUninstall )

type LoadDecision int
const ( LoadAll LoadDecision = iota; LoadQuick; LoadCancel )

type InstallMode int
const ( InstallAfterInstall InstallMode = iota; InstallManage )

type Dialog interface {
    Installer(mode InstallMode) Choice
    ConfirmLoad(message string) LoadDecision
    Error(message string)
}
```

- [ ] **Step 5: Implement the TaskDialog installer and error prompts**

In `internal/ui/dialog_windows.go`:

- Define `installDialogCopy` and `installDialogCopyFor`:
  - `InstallAfterInstall`: instruction `"OpenMD 已安装"`, content `"右键 Markdown 文件或文件夹，选择 OpenMD 即可阅读。"`, buttons `{101,"卸载"},{102,"完成"}`, default `102`.
  - `InstallManage`: same instruction/content, buttons `{100,"覆盖安装"},{101,"卸载"},{102,"取消"}`, default `102`.
- `choiceFromButton`: `100→ChoiceInstall`, `101→ChoiceUninstall`, default `ChoiceCancel`.
- `fallbackInstallerChoice`: after-install `6→Uninstall`, `7→Cancel`; manage `6→Install`, `7→Uninstall`, `2→Cancel`.
- `WindowsDialog.Installer(mode)`:
  - Get the copy, load the app icon, call `showTaskDialog` with title `"OpenMD"`, `commandLinks: true`.
  - On error, call `messageBox` with `MB_YESNOCANCEL` (manage) or `MB_YESNO` (after install) and the fallback text; map with `fallbackInstallerChoice`.
  - Otherwise return `choiceFromButton(id)`.
- `WindowsDialog.Error(message)`:
  - Call `showTaskDialog` with instruction `"出错了"`, content `message`, error icon, `commonButtons: mbOK`.
  - On error, `messageBox("出错了\n\n"+message, "OpenMD", mbIconError)`.
- Remove `Info` from `WindowsDialog`.
- Keep `ConfirmLoad` unchanged (MessageBox).
- Reuse the existing `messageBox` helper; do not add a duplicate.
- Remove the old `Installer(installedPath string, installed bool)` and old `Info` implementation.

- [ ] **Step 6: Implement the app flow**

In `internal/app/run.go`:

- `runInstaller`:
  - Resolve `currentExe`.
  - `installed := fileExists(a.Installer.InstalledExe())`.
  - If not installed: `Install(currentExe)`; then `if a.Dialog.Installer(ui.InstallAfterInstall) == ui.ChoiceUninstall { return a.uninstall(currentExe) }`; otherwise `nil`.
  - If installed: switch on `a.Dialog.Installer(ui.InstallManage)`: `ChoiceInstall → Install(currentExe)`, `ChoiceUninstall → a.uninstall(currentExe)`, default `nil`.
- Add `func (a *App) uninstall(currentExe string) error`:
  - Call `Uninstall`; on `NeedsHelper`, call `SpawnUninstallHelper`; wrap the helper error as `启动卸载清理程序失败: %w`.
- Remove all `a.Dialog.Info(...)` calls in `runInstaller`.
- In `handleInternal`, remove the success `dialog.Info(...)` calls for `--install` and `--uninstall`; keep `dialog.Error` on failure.

- [ ] **Step 7: Run all tests**

Run: `go test ./... -count=1`

Expected: all tests pass.

- [ ] **Step 8: Commit**

```bash
git add internal/ui internal/app
git commit -m "feat: install directly and show one TaskDialog management prompt"
```

---

### Task 4: Documentation, cleanup, and build

**Files:**
- Modify: `README.md`
- Delete: `tools/taskdialogdemo/` (untracked throwaway spike)

**Interfaces:**
- Consumes: the final behavior from Task 3.

- [ ] **Step 1: Update README install/uninstall sections**

Replace the current 安装 steps with:

```markdown
1. 双击 `OpenMD.exe`（放在 Downloads 等任意可执行目录都可以）。
2. 程序会直接安装到 `%LOCALAPPDATA%\Programs\OpenMD\OpenMD.exe`，并注册右键菜单。
3. 安装完成后弹出一个对话框；选择“完成”即可，需要卸载时在同一对话框选择“卸载”。
```

Replace the 卸载 section with:

```markdown
双击安装目录里的 `OpenMD.exe`，在对话框里选择“卸载”。

也可以执行：

```bash
OpenMD.exe --uninstall
```
```

- [ ] **Step 2: Update the icon/resource section**

Change the embedding command to:

```bash
rsrc -ico assets/openmd.ico -manifest assets/app.manifest -o rsrc_windows_amd64.syso
```

Add a line: `assets/app.manifest` 提供 Common Controls v6、`asInvoker` 和 PerMonitorV2 DPI；修改后需重新生成 `.syso`。

- [ ] **Step 3: Note the native dialog limitation**

Add to 已知限制:

```markdown
- 安装与错误对话框使用 Windows 原生 TaskDialog，不跟随阅读页配色和深色模式。
```

- [ ] **Step 4: Remove the throwaway demo**

Run:

```bash
rm -rf tools/taskdialogdemo
```

Expected: the directory no longer exists and is not part of the commit.

- [ ] **Step 5: Run tests**

Run: `go test ./... -count=1`

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add README.md
git commit -m "docs: describe direct install and task dialog prompts"
```

- [ ] **Step 7: Build**

Run:

```bash
printf '\n' | cmd.exe //c build.bat
```

Expected: tests pass, `OpenMD.exe` and `source.git.bundle` are regenerated.

- [ ] **Step 8: Manual visual check**

Run `OpenMD.exe` with no arguments and confirm:

- Fresh state installs first, then shows the post-install dialog with `卸载 / 完成`.
- Second launch shows the manage dialog with `覆盖安装 / 卸载 / 取消`.
- `OpenMD.exe --uninstall` uninstalls silently.

To re-test the fresh path, run `OpenMD.exe --uninstall`, then double-click `OpenMD.exe` again.

If anything fails, fix, commit, and re-run `build.bat` before reporting completion.
