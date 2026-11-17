package install

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeAssoc struct {
	registeredPath   string
	unregisterCalled bool
	err              error
}

func (f *fakeAssoc) Register(exePath string) error {
	f.registeredPath = exePath
	return f.err
}

func (f *fakeAssoc) Unregister() error {
	f.unregisterCalled = true
	return f.err
}

func TestInstallCopiesExeAndRegisters(t *testing.T) {
	src := filepath.Join(t.TempDir(), "OpenMD.exe")
	if err := os.WriteFile(src, []byte("binary"), 0644); err != nil {
		t.Fatal(err)
	}
	installDir := filepath.Join(t.TempDir(), "OpenMD")
	fake := &fakeAssoc{}
	notified := 0
	in := &Installer{InstallDir: installDir, Assoc: fake, Notify: func() { notified++ }}
	if err := in.Install(src); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(in.InstalledExe()); err != nil {
		t.Fatal(err)
	}
	if fake.registeredPath != in.InstalledExe() {
		t.Fatalf("registered %q", fake.registeredPath)
	}
	if notified != 1 {
		t.Fatalf("notify = %d", notified)
	}
	if _, err := os.Stat(filepath.Join(installDir, "openmd-temp")); err != nil {
		t.Fatalf("temp directory not created: %v", err)
	}
}

func TestInstallFromInstalledPathDoesNotOverwriteItself(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), "OpenMD")
	in := &Installer{InstallDir: installDir, Assoc: &fakeAssoc{}, Notify: func() {}}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(in.InstalledExe(), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := in.Install(in.InstalledExe()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(in.InstalledExe())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "keep" {
		t.Fatalf("installed exe was overwritten: %q", data)
	}
}

func TestInstallAssociationFailureRollsBackFreshInstall(t *testing.T) {
	src := filepath.Join(t.TempDir(), "OpenMD.exe")
	if err := os.WriteFile(src, []byte("binary"), 0644); err != nil {
		t.Fatal(err)
	}
	installDir := filepath.Join(t.TempDir(), "OpenMD")
	in := &Installer{
		InstallDir: installDir,
		Assoc:      &fakeAssoc{err: errors.New("access denied")},
		Notify:     func() {},
	}
	if err := in.Install(src); err == nil {
		t.Fatal("want error")
	}
	if _, err := os.Stat(in.InstalledExe()); !os.IsNotExist(err) {
		t.Fatalf("failed install left an exe behind: %v", err)
	}
}

func TestInstallUpdateRestoresOldExeOnFailure(t *testing.T) {
	src := filepath.Join(t.TempDir(), "OpenMD.exe")
	if err := os.WriteFile(src, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	installDir := filepath.Join(t.TempDir(), "OpenMD")
	in := &Installer{
		InstallDir: installDir,
		Assoc:      &fakeAssoc{err: errors.New("access denied")},
		Notify:     func() {},
	}
	if err := os.MkdirAll(installDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(in.InstalledExe(), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := in.Install(src); err == nil {
		t.Fatal("want error")
	}
	data, err := os.ReadFile(in.InstalledExe())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "old" {
		t.Fatalf("previous exe not restored: %q", data)
	}
}

func TestSpawnUninstallHelperUsesInjectedSpawn(t *testing.T) {
	var gotExe, gotDir string
	in := &Installer{
		InstallDir: filepath.Join(t.TempDir(), "OpenMD"),
		Assoc:      &fakeAssoc{},
		Notify:     func() {},
		Spawn: func(exe, installDir string) error {
			gotExe, gotDir = exe, installDir
			return nil
		},
	}
	if err := in.SpawnUninstallHelper(`C:\Local\OpenMD.exe`); err != nil {
		t.Fatal(err)
	}
	if gotExe != `C:\Local\OpenMD.exe` || gotDir != in.InstallDir {
		t.Fatalf("spawn args = %q %q", gotExe, gotDir)
	}
}

func TestUninstallFromOtherLocationRemovesInstallDir(t *testing.T) {
	installDir := t.TempDir()
	fake := &fakeAssoc{}
	in := &Installer{InstallDir: installDir, Assoc: fake, Notify: func() {}}
	if err := os.WriteFile(in.InstalledExe(), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	tempDir := filepath.Join(installDir, "openmd-temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatal(err)
	}
	res, err := in.Uninstall(filepath.Join(t.TempDir(), "OpenMD.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if res.NeedsHelper {
		t.Fatal("unexpected helper")
	}
	if !fake.unregisterCalled {
		t.Fatal("unregister not called")
	}
	if _, err := os.Stat(installDir); !os.IsNotExist(err) {
		t.Fatal("install dir still exists")
	}
}

func TestUninstallFromInstalledPathNeedsHelper(t *testing.T) {
	installDir := t.TempDir()
	fake := &fakeAssoc{}
	in := &Installer{InstallDir: installDir, Assoc: fake, Notify: func() {}}
	if err := os.WriteFile(in.InstalledExe(), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	res, err := in.Uninstall(in.InstalledExe())
	if err != nil {
		t.Fatal(err)
	}
	if !res.NeedsHelper {
		t.Fatal("want helper")
	}
	if !fake.unregisterCalled {
		t.Fatal("unregister not called")
	}
	if _, err := os.Stat(in.InstalledExe()); err != nil {
		t.Fatal("installed exe removed too early")
	}
}
