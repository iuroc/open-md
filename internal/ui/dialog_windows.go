package ui

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type Choice int

const (
	ChoiceCancel Choice = iota
	ChoiceInstall
	ChoiceUninstall
)

type LoadDecision int

const (
	LoadAll LoadDecision = iota
	LoadQuick
	LoadCancel
)

type Dialog interface {
	Installer(installedPath string, installed bool) Choice
	ConfirmLoad(message string) LoadDecision
	Info(message string)
	Error(message string)
}

type WindowsDialog struct{}

func (d *WindowsDialog) Installer(installedPath string, installed bool) Choice {
	var text string
	if installed {
		text = "OpenMD 已安装：\n" + installedPath + "\n\n是：重新安装 / 更新\n否：卸载\n取消：退出"
	} else {
		text = "OpenMD 尚未安装。\n\n是：安装并注册右键菜单\n否：卸载（当前未安装）\n取消：退出"
	}
	result := messageBox(text, "OpenMD", 0x00000003|0x00000020)
	switch result {
	case 6: // IDYES
		return ChoiceInstall
	case 7: // IDNO
		return ChoiceUninstall
	default:
		return ChoiceCancel
	}
}

func (d *WindowsDialog) ConfirmLoad(message string) LoadDecision {
	result := messageBox(message, "OpenMD", 0x00000003|0x00000020)
	switch result {
	case 6: // IDYES
		return LoadAll
	case 7: // IDNO
		return LoadQuick
	default:
		return LoadCancel
	}
}

func (d *WindowsDialog) Info(message string) {
	messageBox(message, "OpenMD", 0x00000000|0x00000040)
}

func (d *WindowsDialog) Error(message string) {
	messageBox(message, "OpenMD", 0x00000000|0x00000010)
}

func messageBox(text, caption string, flags uint32) uintptr {
	user32 := windows.NewLazySystemDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	textPtr, _ := windows.UTF16PtrFromString(text)
	captionPtr, _ := windows.UTF16PtrFromString(caption)
	result, _, _ := proc.Call(
		0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(flags),
	)
	return result
}
