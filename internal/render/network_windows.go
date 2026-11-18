package render

import (
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var getDriveType = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetDriveTypeW")

func isNetworkPath(path string) bool {
	if strings.HasPrefix(path, `\\`) {
		return true
	}
	volume := filepath.VolumeName(path)
	if len(volume) != 2 || volume[1] != ':' {
		return false
	}
	rootPtr, err := windows.UTF16PtrFromString(volume + `\`)
	if err != nil {
		return true
	}
	driveType, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(rootPtr)))
	// 4 = DRIVE_REMOTE. 1 = DRIVE_NO_ROOT_DIR, which can be a disconnected
	// mapped drive; skip it rather than risk a blocking network stat.
	return driveType == 4 || driveType == 1
}
