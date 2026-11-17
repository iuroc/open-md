package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"openmd/internal/browser"
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
	installedPath := a.Installer.InstalledExe()
	installed := fileExists(installedPath)
	choice := a.Dialog.Installer(installedPath, installed)
	switch choice {
	case ui.ChoiceInstall:
		if err := a.Installer.Install(currentExe); err != nil {
			return err
		}
		a.Dialog.Info("安装完成。\n\n现在可以右键 Markdown 文件，选择“用 OpenMD 阅读”。")
		return nil
	case ui.ChoiceUninstall:
		result, err := a.Installer.Uninstall(currentExe)
		if err != nil {
			return err
		}
		if result.NeedsHelper {
			a.Dialog.Info("卸载完成。程序文件将在退出后清理。")
			if err := a.Installer.SpawnUninstallHelper(currentExe); err != nil {
				return fmt.Errorf("启动卸载清理程序失败: %w", err)
			}
			return nil
		}
		a.Dialog.Info("卸载完成。")
		return nil
	default:
		return nil
	}
}

func (a *App) openPage(path string, isDir bool) error {
	var (
		page *render.Page
		err  error
	)
	if isDir {
		page, err = render.BuildDirPage(path)
	} else {
		page, err = render.BuildFilePage(path)
	}
	if err != nil {
		return err
	}
	htmlBytes, err := render.RenderPage(page)
	if err != nil {
		return err
	}
	if a.Temp == nil {
		return errors.New("临时目录管理器未初始化")
	}
	if a.Opener == nil {
		return errors.New("浏览器启动器未初始化")
	}

	_ = a.Temp.Cleanup()
	runDir, err := a.Temp.RunDir()
	if err != nil {
		fallback := tempdir.New(filepath.Join(os.TempDir(), "OpenMD"))
		_ = fallback.Cleanup()
		runDir, err = fallback.RunDir()
		if err != nil {
			return fmt.Errorf("创建临时目录失败: %w", err)
		}
	}

	htmlPath := filepath.Join(runDir, "index.html")
	tmpPath := htmlPath + ".tmp"
	if err := os.WriteFile(tmpPath, htmlBytes, 0644); err != nil {
		return fmt.Errorf("写入临时页面失败: %w", err)
	}
	if err := os.Rename(tmpPath, htmlPath); err != nil {
		return fmt.Errorf("保存临时页面失败: %w", err)
	}
	if err := a.Opener.Open(htmlPath); err != nil {
		return fmt.Errorf("打开默认浏览器失败: %w", err)
	}
	return nil
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
		dialog.Info("安装完成。")
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
			dialog.Info("卸载完成。程序文件将在退出后清理。")
			if err := installer.SpawnUninstallHelper(exe); err != nil {
				dialog.Error(err.Error())
				return true
			}
			return true
		}
		dialog.Info("卸载完成。")
		return true
	default:
		return false
	}
}
