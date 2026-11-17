package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type UninstallResult struct {
	NeedsHelper bool
}

type Installer struct {
	InstallDir string
	Assoc      Associator
	Notify     func()
	Spawn      func(exe, installDir string) error
}

func DefaultInstallDir() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		var err error
		base, err = os.UserCacheDir()
		if err != nil {
			return "", err
		}
	}
	if base == "" {
		return "", errors.New("无法确定用户本地应用目录")
	}
	return filepath.Join(base, "Programs", "OpenMD"), nil
}

func (i *Installer) InstalledExe() string {
	return filepath.Join(i.InstallDir, "OpenMD.exe")
}

func (i *Installer) Install(sourceExe string) error {
	if i.Assoc == nil {
		return errors.New("安装程序缺少关联器")
	}
	if i.InstallDir == "" {
		return errors.New("安装目录为空")
	}
	if err := os.MkdirAll(i.InstallDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}
	target := i.InstalledExe()
	if samePath(sourceExe, target) {
		if err := i.Assoc.Register(target); err != nil {
			return fmt.Errorf("写入右键菜单失败: %w", err)
		}
	} else if err := i.replaceWithBackup(sourceExe, target); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(i.InstallDir, "openmd-temp"), 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	i.notify()
	return nil
}

func (i *Installer) replaceWithBackup(sourceExe, target string) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".OpenMD-*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时安装文件失败: %w", err)
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := copyFile(sourceExe, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("复制程序失败: %w", err)
	}

	backupPath := target + ".bak"
	hadOld := fileExists(target)
	if hadOld {
		_ = os.Remove(backupPath)
		if err := copyFile(target, backupPath); err != nil {
			_ = os.Remove(tmpPath)
			_ = os.Remove(backupPath)
			return fmt.Errorf("备份旧程序失败: %w", err)
		}
	}

	if err := os.Rename(tmpPath, target); err != nil {
		// Windows does not always replace an existing file via Rename.
		if removeErr := os.Remove(target); removeErr != nil {
			_ = os.Remove(tmpPath)
			_ = os.Remove(backupPath)
			return fmt.Errorf("替换程序失败: %w", err)
		}
		if err := os.Rename(tmpPath, target); err != nil {
			_ = os.Remove(tmpPath)
			i.restoreBackup(backupPath, target, hadOld)
			return fmt.Errorf("替换程序失败: %w", err)
		}
	}

	if err := i.Assoc.Register(target); err != nil {
		_ = os.Remove(target)
		if hadOld {
			if restoreErr := copyFile(backupPath, target); restoreErr != nil {
				return fmt.Errorf("写入右键菜单失败: %w；恢复旧程序也失败: %v", err, restoreErr)
			}
		}
		_ = os.Remove(backupPath)
		return fmt.Errorf("写入右键菜单失败: %w", err)
	}
	_ = os.Remove(backupPath)
	return nil
}

func (i *Installer) restoreBackup(backupPath, target string, hadOld bool) {
	if hadOld {
		_ = os.Remove(target)
		_ = copyFile(backupPath, target)
	}
	_ = os.Remove(backupPath)
}

func (i *Installer) Uninstall(sourceExe string) (UninstallResult, error) {
	if i.Assoc == nil {
		return UninstallResult{}, errors.New("安装程序缺少关联器")
	}
	if err := i.Assoc.Unregister(); err != nil {
		return UninstallResult{}, fmt.Errorf("移除右键菜单失败: %w", err)
	}
	i.notify()

	_ = os.RemoveAll(filepath.Join(i.InstallDir, "openmd-temp"))
	if samePath(sourceExe, i.InstalledExe()) {
		return UninstallResult{NeedsHelper: true}, nil
	}
	if err := os.RemoveAll(i.InstallDir); err != nil {
		return UninstallResult{}, fmt.Errorf("删除安装目录失败: %w", err)
	}
	return UninstallResult{}, nil
}

func (i *Installer) SpawnUninstallHelper(sourceExe string) error {
	spawn := i.Spawn
	if spawn == nil {
		spawn = spawnUninstallHelper
	}
	return spawn(sourceExe, i.InstallDir)
}

func (i *Installer) notify() {
	if i.Notify != nil {
		i.Notify()
		return
	}
	notifyShellChanged()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}
