package app

import (
	"openmd/internal/browser"
	"openmd/internal/install"
	"openmd/internal/tempdir"
	"openmd/internal/ui"
)

type Installer interface {
	Install(sourceExe string) error
	Uninstall(sourceExe string) (install.UninstallResult, error)
	SpawnUninstallHelper(sourceExe string) error
	InstalledExe() string
}

type App struct {
	Dialog     ui.Dialog
	Installer  Installer
	Temp       *tempdir.Manager
	Opener     browser.Opener
	Executable func() (string, error)
}
