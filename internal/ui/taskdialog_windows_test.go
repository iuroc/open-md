package ui

import (
	"encoding/binary"
	"testing"
)

func TestBuildTaskDialogConfigPacked(t *testing.T) {
	opts := taskDialogOptions{
		title:        "OpenMD",
		instruction:  "OpenMD 已安装",
		content:      "正文",
		buttons:      []taskDialogButtonSpec{{id: 100, text: "覆盖安装"}, {id: 101, text: "卸载"}},
		icon:         0x1234,
		defaultID:    100,
		commandLinks: true,
	}
	cfg, btns, refs := buildTaskDialogConfig(opts, 0x5678)
	if len(cfg) != 160 {
		t.Fatalf("len(cfg) = %d", len(cfg))
	}
	if got := binary.LittleEndian.Uint32(cfg[0:]); got != 160 {
		t.Fatalf("cbSize = %d", got)
	}
	if got := binary.LittleEndian.Uint64(cfg[4:]); got != 0 {
		t.Fatalf("hwndParent = %#x", got)
	}
	if got := binary.LittleEndian.Uint64(cfg[12:]); got != 0x5678 {
		t.Fatalf("hInstance = %#x", got)
	}
	if got := binary.LittleEndian.Uint32(cfg[20:]); got != 0x1a {
		t.Fatalf("flags = %#x", got)
	}
	if got := binary.LittleEndian.Uint64(cfg[36:]); got != 0x1234 {
		t.Fatalf("mainIcon = %#x", got)
	}
	if got := binary.LittleEndian.Uint32(cfg[60:]); got != 2 {
		t.Fatalf("cButtons = %d", got)
	}
	if got := binary.LittleEndian.Uint64(cfg[64:]); got == 0 {
		t.Fatal("pButtons is NULL")
	}
	if got := binary.LittleEndian.Uint32(cfg[72:]); got != 100 {
		t.Fatalf("nDefaultButton = %d", got)
	}
	if len(btns) != 24 {
		t.Fatalf("len(btns) = %d", len(btns))
	}
	if got := binary.LittleEndian.Uint32(btns[0:]); got != 100 {
		t.Fatalf("button 0 id = %d", got)
	}
	if got := binary.LittleEndian.Uint64(btns[4:]); got == 0 {
		t.Fatal("button 0 text is NULL")
	}
	if got := binary.LittleEndian.Uint32(btns[12:]); got != 101 {
		t.Fatalf("button 1 id = %d", got)
	}
	if got := binary.LittleEndian.Uint64(btns[16:]); got == 0 {
		t.Fatal("button 1 text is NULL")
	}
	if len(refs) != 5 {
		t.Fatalf("len(refs) = %d", len(refs))
	}
}

func TestBuildTaskDialogConfigNoCommandLinksWithoutButtons(t *testing.T) {
	cfg, _, _ := buildTaskDialogConfig(taskDialogOptions{
		title:         "OpenMD",
		instruction:   "出错了",
		content:       "正文",
		icon:          0x1234,
		commonButtons: 1,
	}, 0x5678)
	if got := binary.LittleEndian.Uint32(cfg[20:]); got&0x10 != 0 {
		t.Fatalf("command-link flag set without buttons: %#x", got)
	}
	if got := binary.LittleEndian.Uint32(cfg[24:]); got != 1 {
		t.Fatalf("common buttons = %#x", got)
	}
}
