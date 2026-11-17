package install

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

func notifyShellChanged() {
	const (
		shcneAssocChanged = 0x08000000
		shcnfIDList       = 0x0000
	)
	shell32 := windows.NewLazySystemDLL("shell32.dll")
	_, _, _ = shell32.NewProc("SHChangeNotify").Call(shcneAssocChanged, shcnfIDList, 0, 0)
}

func spawnUninstallHelper(sourceExe, installDir string) error {
	helper, err := os.CreateTemp("", "openmd-uninstall-*.exe")
	if err != nil {
		return err
	}
	helperPath := helper.Name()
	if err := helper.Close(); err != nil {
		_ = os.Remove(helperPath)
		return err
	}
	if err := copyFile(sourceExe, helperPath); err != nil {
		_ = os.Remove(helperPath)
		return err
	}
	cmd := exec.Command(helperPath, "--uninstall-helper", installDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Start(); err != nil {
		_ = os.Remove(helperPath)
		return err
	}
	return cmd.Process.Release()
}

func RunUninstallHelper(installDir string) error {
	var lastErr error
	for attempt := 0; attempt < 20; attempt++ {
		lastErr = os.RemoveAll(installDir)
		if lastErr == nil || os.IsNotExist(lastErr) {
			lastErr = nil
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	exe, err := os.Executable()
	if err == nil {
		cmd := exec.Command("cmd.exe", "/c", `ping -n 3 127.0.0.1 >nul & del /f /q "`+exe+`"`)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = cmd.Start()
	}
	return lastErr
}
