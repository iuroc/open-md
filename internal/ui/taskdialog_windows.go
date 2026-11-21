package ui

import (
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	tdfUseHIconMain            = 0x0002
	tdfAllowDialogCancellation = 0x0008
	tdfUseCommandLinks         = 0x0010

	// TASKDIALOGCONFIG is declared with pack(1); on 64-bit it is 160 bytes.
	taskDialogConfigSize = 160
	taskDialogButtonSize = 12

	installButtonInstall   int32 = 100
	installButtonUninstall int32 = 101
	installButtonDone      int32 = 102

	idiError       = 32513
	idiInformation = 32516

	mbOK          = 0x00000001
	mbYesNoCancel = 0x00000003
	mbYesNo       = 0x00000004
	mbIconError   = 0x00000010
)

var (
	kernel32               = windows.NewLazySystemDLL("kernel32.dll")
	user32                 = windows.NewLazySystemDLL("user32.dll")
	shell32                = windows.NewLazySystemDLL("shell32.dll")
	comctl32               = windows.NewLazySystemDLL("comctl32.dll")
	procGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
	procLoadIconW          = user32.NewProc("LoadIconW")
	procExtractIconExW     = shell32.NewProc("ExtractIconExW")
	procDestroyIcon        = user32.NewProc("DestroyIcon")
	procTaskDialogIndirect = comctl32.NewProc("TaskDialogIndirect")
)

type taskDialogButtonSpec struct {
	id   int32
	text string
}

type taskDialogOptions struct {
	title         string
	instruction   string
	content       string
	buttons       []taskDialogButtonSpec
	icon          uintptr
	defaultID     int32
	commonButtons uint32
	commandLinks  bool
}

func buildTaskDialogConfig(opts taskDialogOptions, hInstance uintptr) ([]byte, []byte, []*uint16) {
	var refs []*uint16
	strPtr := func(s string) uint64 {
		if s == "" {
			return 0
		}
		p, err := windows.UTF16PtrFromString(s)
		if err != nil {
			panic(err)
		}
		refs = append(refs, p)
		return uint64(uintptr(unsafe.Pointer(p)))
	}

	title := strPtr(opts.title)
	instruction := strPtr(opts.instruction)
	content := strPtr(opts.content)

	btns := make([]byte, taskDialogButtonSize*len(opts.buttons))
	for i, b := range opts.buttons {
		off := i * taskDialogButtonSize
		binary.LittleEndian.PutUint32(btns[off:], uint32(b.id))
		binary.LittleEndian.PutUint64(btns[off+4:], strPtr(b.text))
	}

	flags := uint32(tdfAllowDialogCancellation)
	if len(opts.buttons) > 0 && opts.commandLinks {
		flags |= tdfUseCommandLinks
	}
	if opts.icon != 0 {
		flags |= tdfUseHIconMain
	}

	cfg := make([]byte, taskDialogConfigSize)
	binary.LittleEndian.PutUint32(cfg[0:], taskDialogConfigSize) // cbSize
	binary.LittleEndian.PutUint64(cfg[4:], 0)                    // hwndParent
	binary.LittleEndian.PutUint64(cfg[12:], uint64(hInstance))   // hInstance
	binary.LittleEndian.PutUint32(cfg[20:], flags)               // dwFlags
	binary.LittleEndian.PutUint32(cfg[24:], opts.commonButtons)  // dwCommonButtons
	binary.LittleEndian.PutUint64(cfg[28:], title)               // pszWindowTitle
	binary.LittleEndian.PutUint64(cfg[36:], uint64(opts.icon))   // hMainIcon
	binary.LittleEndian.PutUint64(cfg[44:], instruction)         // pszMainInstruction
	binary.LittleEndian.PutUint64(cfg[52:], content)             // pszContent
	binary.LittleEndian.PutUint32(cfg[60:], uint32(len(opts.buttons)))
	if len(btns) > 0 {
		binary.LittleEndian.PutUint64(cfg[64:], uint64(uintptr(unsafe.Pointer(&btns[0])))) // pButtons
	}
	binary.LittleEndian.PutUint32(cfg[72:], uint32(opts.defaultID)) // nDefaultButton
	return cfg, btns, refs
}

func showTaskDialog(opts taskDialogOptions) (int32, error) {
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	cfg, btns, refs := buildTaskDialogConfig(opts, hInstance)

	if err := procTaskDialogIndirect.Find(); err != nil {
		return 0, err
	}
	var button int32
	hr, _, _ := procTaskDialogIndirect.Call(
		uintptr(unsafe.Pointer(&cfg[0])),
		uintptr(unsafe.Pointer(&button)),
		0,
		0,
	)
	runtime.KeepAlive(refs)
	runtime.KeepAlive(btns)
	runtime.KeepAlive(cfg)

	if int32(hr) < 0 {
		return 0, fmt.Errorf("TaskDialogIndirect 失败：0x%08X", uint32(hr))
	}
	return button, nil
}

func loadAppIcon() (uintptr, func()) {
	exe, err := os.Executable()
	if err != nil {
		return loadSystemIcon(idiInformation), func() {}
	}
	p, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return loadSystemIcon(idiInformation), func() {}
	}
	var large, small uintptr
	n, _, _ := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(p)),
		0,
		uintptr(unsafe.Pointer(&large)),
		uintptr(unsafe.Pointer(&small)),
		1,
	)
	runtime.KeepAlive(p)
	if small != 0 {
		procDestroyIcon.Call(small)
	}
	if n == 0 || large == 0 {
		return loadSystemIcon(idiInformation), func() {}
	}
	return large, func() { procDestroyIcon.Call(large) }
}

func loadSystemIcon(id uintptr) uintptr {
	icon, _, _ := procLoadIconW.Call(0, id)
	return icon
}
