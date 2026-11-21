package ui

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type WindowsDialog struct{}

type installDialogCopy struct {
	instruction string
	content     string
	buttons     []taskDialogButtonSpec
	defaultID   int32
}

func installDialogCopyFor(mode InstallMode) installDialogCopy {
	const content = "右键 Markdown 文件或文件夹，选择 OpenMD 即可阅读。"
	if mode == InstallAfterInstall {
		return installDialogCopy{
			instruction: "OpenMD 已安装",
			content:     content,
			buttons: []taskDialogButtonSpec{
				{id: installButtonUninstall, text: "卸载"},
				{id: installButtonDone, text: "完成"},
			},
			defaultID: installButtonDone,
		}
	}
	return installDialogCopy{
		instruction: "OpenMD 已安装",
		content:     content,
		buttons: []taskDialogButtonSpec{
			{id: installButtonInstall, text: "覆盖安装"},
			{id: installButtonUninstall, text: "卸载"},
			{id: installButtonDone, text: "取消"},
		},
		defaultID: installButtonDone,
	}
}

func choiceFromButton(id int32) Choice {
	switch id {
	case installButtonInstall:
		return ChoiceInstall
	case installButtonUninstall:
		return ChoiceUninstall
	default:
		return ChoiceCancel
	}
}

func fallbackInstallerChoice(mode InstallMode, result uintptr) Choice {
	if mode == InstallAfterInstall {
		if result == 6 { // IDYES
			return ChoiceUninstall
		}
		return ChoiceCancel
	}
	switch result {
	case 6: // IDYES
		return ChoiceInstall
	case 7: // IDNO
		return ChoiceUninstall
	default:
		return ChoiceCancel
	}
}

func (d *WindowsDialog) Installer(mode InstallMode) Choice {
	spec := installDialogCopyFor(mode)
	icon, cleanup := loadAppIcon()
	defer cleanup()
	id, err := showTaskDialog(taskDialogOptions{
		title:        "OpenMD",
		instruction:  spec.instruction,
		content:      spec.content,
		buttons:      spec.buttons,
		icon:         icon,
		defaultID:    spec.defaultID,
		commandLinks: true,
	})
	if err != nil {
		if mode == InstallAfterInstall {
			text := spec.instruction + "\n\n" + spec.content + "\n\n是：卸载\n否：完成"
			return fallbackInstallerChoice(mode, messageBox(text, "OpenMD", mbYesNo))
		}
		text := spec.instruction + "\n\n" + spec.content + "\n\n是：覆盖安装\n否：卸载\n取消：退出"
		return fallbackInstallerChoice(mode, messageBox(text, "OpenMD", mbYesNoCancel))
	}
	return choiceFromButton(id)
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

func (d *WindowsDialog) Error(message string) {
	_, err := showTaskDialog(taskDialogOptions{
		title:         "OpenMD",
		instruction:   "出错了",
		content:       message,
		icon:          loadSystemIcon(idiError),
		commonButtons: mbOK,
	})
	if err == nil {
		return
	}
	messageBox("出错了\n\n"+message, "OpenMD", mbIconError)
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
