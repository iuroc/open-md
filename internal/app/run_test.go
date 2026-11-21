package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openmd/internal/install"
	"openmd/internal/render"
	"openmd/internal/tempdir"
	"openmd/internal/ui"
)

type fakeOpener struct {
	path string
}

func (f *fakeOpener) Open(path string) error {
	f.path = path
	return nil
}

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

func (f *fakeDialog) ConfirmLoad(message string) ui.LoadDecision {
	f.confirmSeen = append(f.confirmSeen, message)
	return f.load
}

func (f *fakeDialog) Error(message string) {
	f.errors = append(f.errors, message)
}

type fakeInstaller struct {
	installedExe    string
	installSource   string
	installCalled   bool
	uninstallCalled bool
	helperCalled    bool
	uninstallResult install.UninstallResult
}

func (f *fakeInstaller) Install(sourceExe string) error {
	f.installCalled = true
	f.installSource = sourceExe
	return nil
}

func (f *fakeInstaller) Uninstall(sourceExe string) (install.UninstallResult, error) {
	f.uninstallCalled = true
	return f.uninstallResult, nil
}

func (f *fakeInstaller) SpawnUninstallHelper(sourceExe string) error {
	f.helperCalled = true
	return nil
}

func (f *fakeInstaller) InstalledExe() string {
	return f.installedExe
}

func TestRunReadFileRendersTempHTMLAndOpensBrowser(t *testing.T) {
	md := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(md, []byte("# Hello\n\n![x](pic.png)\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opener := &fakeOpener{}
	a := &App{
		Dialog:     &fakeDialog{},
		Installer:  &fakeInstaller{},
		Temp:       tempdir.New(t.TempDir()),
		Opener:     opener,
		Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil },
	}
	if err := a.Run(Invocation{Mode: ModeReadFile, Path: md}); err != nil {
		t.Fatal(err)
	}
	if opener.path == "" {
		t.Fatal("browser not opened")
	}
	if !strings.HasPrefix(filepath.Base(opener.path), "doc-") {
		t.Fatalf("opened %q", opener.path)
	}
	data, err := os.ReadFile(opener.path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Hello") {
		t.Fatal("rendered content missing")
	}
	if !strings.Contains(string(data), `href="reader.css"`) {
		t.Fatal("shared stylesheet missing")
	}
	entries, err := os.ReadDir(filepath.Dir(md))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "note.md" {
		t.Fatalf("markdown directory was modified: %#v", entries)
	}
}

func TestRunReadDirBuildsDocumentList(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("# B\n"), 0644); err != nil {
		t.Fatal(err)
	}
	opener := &fakeOpener{}
	a := &App{
		Dialog:     &fakeDialog{},
		Installer:  &fakeInstaller{},
		Temp:       tempdir.New(t.TempDir()),
		Opener:     opener,
		Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil },
	}
	if err := a.Run(Invocation{Mode: ModeReadDir, Path: dir}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(opener.path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `id="doc-list"`) || !strings.Contains(string(data), "A") {
		t.Fatalf("document page missing: %s", data)
	}
	tree, err := os.ReadFile(filepath.Join(filepath.Dir(opener.path), "tree.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tree), "data-tree-path") || !strings.Contains(string(tree), "a.md") || !strings.Contains(string(tree), "b.md") {
		t.Fatalf("document tree missing: %s", tree)
	}
}

func TestConfirmLoadMapsDialogChoice(t *testing.T) {
	d := &fakeDialog{load: ui.LoadQuick}
	a := &App{Dialog: d}
	got := a.confirmLoad(render.Estimate{PendingDocs: 3001, PendingBytes: 9 << 20})
	if got != render.LoadQuick {
		t.Fatalf("choice = %v", got)
	}
	if len(d.confirmSeen) != 1 || !strings.Contains(d.confirmSeen[0], "3001") {
		t.Fatalf("confirm message = %#v", d.confirmSeen)
	}
}

func TestRunFreshInstallInstallsThenAsksAfterInstall(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "OpenMD.exe")
	fi := &fakeInstaller{installedExe: installed}
	d := &fakeDialog{choice: ui.ChoiceCancel}
	a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil }}
	if err := a.Run(Invocation{Mode: ModeInstall}); err != nil {
		t.Fatal(err)
	}
	if !fi.installCalled || fi.installSource != `C:\Downloads\OpenMD.exe` {
		t.Fatalf("install = %v %q", fi.installCalled, fi.installSource)
	}
	if len(d.installModes) != 1 || d.installModes[0] != ui.InstallAfterInstall {
		t.Fatalf("modes = %#v", d.installModes)
	}
	if fi.uninstallCalled {
		t.Fatal("uninstall called")
	}
}

func TestRunFreshInstallCanUninstallAfterInstall(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "OpenMD.exe")
	fi := &fakeInstaller{installedExe: installed}
	d := &fakeDialog{choice: ui.ChoiceUninstall}
	a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil }}
	if err := a.Run(Invocation{Mode: ModeInstall}); err != nil {
		t.Fatal(err)
	}
	if !fi.installCalled || !fi.uninstallCalled {
		t.Fatalf("install=%v uninstall=%v", fi.installCalled, fi.uninstallCalled)
	}
}

func TestRunInstalledAsksManageBeforeAction(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "OpenMD.exe")
	if err := os.WriteFile(installed, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	fi := &fakeInstaller{installedExe: installed}
	d := &fakeDialog{choice: ui.ChoiceCancel}
	a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return installed, nil }}
	if err := a.Run(Invocation{Mode: ModeInstall}); err != nil {
		t.Fatal(err)
	}
	if len(d.installModes) != 1 || d.installModes[0] != ui.InstallManage {
		t.Fatalf("modes = %#v", d.installModes)
	}
	if fi.installCalled || fi.uninstallCalled {
		t.Fatalf("install=%v uninstall=%v", fi.installCalled, fi.uninstallCalled)
	}
}

func TestRunInstalledOverwriteInstalls(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "OpenMD.exe")
	if err := os.WriteFile(installed, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	fi := &fakeInstaller{installedExe: installed}
	d := &fakeDialog{choice: ui.ChoiceInstall}
	a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return `C:\Downloads\OpenMD.exe`, nil }}
	if err := a.Run(Invocation{Mode: ModeInstall}); err != nil {
		t.Fatal(err)
	}
	if !fi.installCalled {
		t.Fatal("install not called")
	}
}

func TestRunInstalledUninstallSpawnsHelperWhenNeeded(t *testing.T) {
	installed := filepath.Join(t.TempDir(), "OpenMD.exe")
	if err := os.WriteFile(installed, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	fi := &fakeInstaller{installedExe: installed, uninstallResult: install.UninstallResult{NeedsHelper: true}}
	d := &fakeDialog{choice: ui.ChoiceUninstall}
	a := &App{Dialog: d, Installer: fi, Executable: func() (string, error) { return installed, nil }}
	if err := a.Run(Invocation{Mode: ModeInstall}); err != nil {
		t.Fatal(err)
	}
	if !fi.uninstallCalled || !fi.helperCalled {
		t.Fatalf("uninstall=%v helper=%v", fi.uninstallCalled, fi.helperCalled)
	}
}
