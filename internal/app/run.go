package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"openmd/internal/browser"
	"openmd/internal/buildinfo"
	"openmd/internal/install"
	"openmd/internal/render"
	"openmd/internal/tempdir"
	"openmd/internal/ui"
)

func (a *App) Run(inv Invocation) error {
	switch inv.Mode {
	case ModeInstall:
		return a.runInstaller()
	case ModeReadFile:
		return a.openPage(inv.Path, false)
	case ModeReadDir:
		return a.openPage(inv.Path, true)
	default:
		return errors.New("未知启动模式")
	}
}

func (a *App) runInstaller() error {
	if a.Dialog == nil || a.Installer == nil || a.Executable == nil {
		return errors.New("安装程序未正确初始化")
	}
	currentExe, err := a.Executable()
	if err != nil {
		return fmt.Errorf("无法确定当前程序路径: %w", err)
	}
	if !fileExists(a.Installer.InstalledExe()) {
		if err := a.Installer.Install(currentExe); err != nil {
			return err
		}
		if a.Dialog.Installer(ui.InstallAfterInstall) == ui.ChoiceUninstall {
			return a.uninstall(currentExe)
		}
		return nil
	}
	switch a.Dialog.Installer(ui.InstallManage) {
	case ui.ChoiceInstall:
		return a.Installer.Install(currentExe)
	case ui.ChoiceUninstall:
		return a.uninstall(currentExe)
	default:
		return nil
	}
}

func (a *App) uninstall(currentExe string) error {
	result, err := a.Installer.Uninstall(currentExe)
	if err != nil {
		return err
	}
	if result.NeedsHelper {
		if err := a.Installer.SpawnUninstallHelper(currentExe); err != nil {
			return fmt.Errorf("启动卸载清理程序失败: %w", err)
		}
	}
	return nil
}

func (a *App) openPage(path string, isDir bool) error {
	if a.Temp == nil {
		return errors.New("临时目录管理器未初始化")
	}
	if a.Opener == nil {
		return errors.New("浏览器启动器未初始化")
	}
	_ = a.Temp.Cleanup()

	build := func(base string) (*render.Site, error) {
		opts := render.SiteOptions{
			CacheBase:  filepath.Join(base, "rendered"),
			Version:    buildinfo.VersionLabel(),
			AllowCache: buildinfo.IsReleaseBuild(),
			Confirm:    a.confirmLoad,
		}
		if isDir {
			return render.BuildDirSite(path, opts)
		}
		return render.BuildFileSite(path, opts)
	}

	site, err := build(a.Temp.BaseDir)
	if err != nil {
		if errors.Is(err, render.ErrLoadCancelled) {
			return nil
		}
		fallback := tempdir.New(filepath.Join(os.TempDir(), "OpenMD"))
		_ = fallback.Cleanup()
		site, err = build(fallback.BaseDir)
		if err != nil {
			if errors.Is(err, render.ErrLoadCancelled) {
				return nil
			}
			return err
		}
	}
	if err := a.Opener.Open(site.InitialPage); err != nil {
		return fmt.Errorf("打开默认浏览器失败: %w", err)
	}
	return nil
}

func (a *App) confirmLoad(estimate render.Estimate) render.LoadChoice {
	if a.Dialog == nil {
		return render.LoadAll
	}
	message := fmt.Sprintf(
		"本次需要生成 %d 篇文档，合计 %s。\n\n生成期间需要等待一段时间。\n\n是：加载全部（不再限制数量、总量、深度和目录项）\n否：只加载较快的一部分（3,000 篇 / 8 MiB）\n取消：退出",
		estimate.PendingDocs, formatSize(estimate.PendingBytes),
	)
	switch a.Dialog.ConfirmLoad(message) {
	case ui.LoadAll:
		return render.LoadAll
	case ui.LoadQuick:
		return render.LoadQuick
	default:
		return render.LoadCancel
	}
}

func formatSize(size int64) string {
	if size >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1<<20))
	}
	if size >= 1<<10 {
		return fmt.Sprintf("%.1f KB", float64(size)/(1<<10))
	}
	return fmt.Sprintf("%d B", size)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func Main() {
	dialog := &ui.WindowsDialog{}
	installDir, err := install.DefaultInstallDir()
	if err != nil {
		dialog.Error("无法确定安装目录:\n" + err.Error())
		return
	}
	installer := &install.Installer{
		InstallDir: installDir,
		Assoc:      install.NewRegistryAssociator(),
	}
	if handleInternal(os.Args[1:], installer, dialog) {
		return
	}
	application := &App{
		Dialog:     dialog,
		Installer:  installer,
		Temp:       tempdir.New(filepath.Join(installDir, "openmd-temp")),
		Opener:     browser.NewWindowsOpener(),
		Executable: os.Executable,
	}
	invocation, err := ParseArgs(os.Args[1:])
	if err != nil {
		dialog.Error(err.Error())
		return
	}
	if err := application.Run(invocation); err != nil {
		dialog.Error(err.Error())
	}
}

func handleInternal(args []string, installer *install.Installer, dialog ui.Dialog) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "--uninstall-helper":
		if len(args) == 2 {
			_ = install.RunUninstallHelper(args[1])
		}
		return true
	case "--install":
		if len(args) != 1 {
			dialog.Error("--install 不接受额外参数")
			return true
		}
		exe, err := os.Executable()
		if err != nil {
			dialog.Error("无法确定当前程序路径:\n" + err.Error())
			return true
		}
		if err := installer.Install(exe); err != nil {
			dialog.Error(err.Error())
			return true
		}
		return true
	case "--uninstall":
		if len(args) != 1 {
			dialog.Error("--uninstall 不接受额外参数")
			return true
		}
		exe, err := os.Executable()
		if err != nil {
			dialog.Error("无法确定当前程序路径:\n" + err.Error())
			return true
		}
		result, err := installer.Uninstall(exe)
		if err != nil {
			dialog.Error(err.Error())
			return true
		}
		if result.NeedsHelper {
			if err := installer.SpawnUninstallHelper(exe); err != nil {
				dialog.Error(err.Error())
			}
		}
		return true
	default:
		return false
	}
}
