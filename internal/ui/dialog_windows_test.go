package ui

import (
	"strings"
	"testing"
)

func TestInstallDialogCopyAfterInstall(t *testing.T) {
	c := installDialogCopyFor(InstallAfterInstall)
	if c.instruction != "OpenMD 已安装" {
		t.Fatalf("instruction = %q", c.instruction)
	}
	if !strings.Contains(c.content, "右键") {
		t.Fatalf("content = %q", c.content)
	}
	if len(c.buttons) != 2 || c.buttons[0].id != 101 || c.buttons[0].text != "卸载" ||
		c.buttons[1].id != 102 || c.buttons[1].text != "完成" {
		t.Fatalf("buttons = %#v", c.buttons)
	}
	if c.defaultID != 102 {
		t.Fatalf("default = %d", c.defaultID)
	}
}

func TestInstallDialogCopyManage(t *testing.T) {
	c := installDialogCopyFor(InstallManage)
	if len(c.buttons) != 3 || c.buttons[0].id != 100 || c.buttons[0].text != "覆盖安装" ||
		c.buttons[1].id != 101 || c.buttons[1].text != "卸载" ||
		c.buttons[2].id != 102 || c.buttons[2].text != "取消" {
		t.Fatalf("buttons = %#v", c.buttons)
	}
	if c.defaultID != 102 {
		t.Fatalf("default = %d", c.defaultID)
	}
}

func TestChoiceFromButton(t *testing.T) {
	cases := []struct {
		id   int32
		want Choice
	}{
		{100, ChoiceInstall},
		{101, ChoiceUninstall},
		{102, ChoiceCancel},
		{2, ChoiceCancel},
	}
	for _, tc := range cases {
		if got := choiceFromButton(tc.id); got != tc.want {
			t.Fatalf("choiceFromButton(%d) = %v", tc.id, got)
		}
	}
}

func TestFallbackInstallerChoice(t *testing.T) {
	const yes, no, cancel = 6, 7, 2
	if got := fallbackInstallerChoice(InstallAfterInstall, yes); got != ChoiceUninstall {
		t.Fatalf("after yes = %v", got)
	}
	if got := fallbackInstallerChoice(InstallAfterInstall, no); got != ChoiceCancel {
		t.Fatalf("after no = %v", got)
	}
	if got := fallbackInstallerChoice(InstallAfterInstall, cancel); got != ChoiceCancel {
		t.Fatalf("after cancel = %v", got)
	}
	if got := fallbackInstallerChoice(InstallManage, yes); got != ChoiceInstall {
		t.Fatalf("manage yes = %v", got)
	}
	if got := fallbackInstallerChoice(InstallManage, no); got != ChoiceUninstall {
		t.Fatalf("manage no = %v", got)
	}
	if got := fallbackInstallerChoice(InstallManage, cancel); got != ChoiceCancel {
		t.Fatalf("manage cancel = %v", got)
	}
}
