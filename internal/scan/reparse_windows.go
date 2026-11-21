//go:build windows

package scan

import (
	"os"
	"syscall"
)

func isReparsePoint(path string) bool {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err == nil {
		if attributes, attrErr := syscall.GetFileAttributes(pointer); attrErr == nil {
			return attributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
		}
	}
	info, statErr := os.Lstat(path)
	return statErr == nil && info.Mode()&os.ModeSymlink != 0
}
