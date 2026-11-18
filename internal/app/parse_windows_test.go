package app

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestParseArgsFromQuotedRootCommandLine(t *testing.T) {
	const commandLine = `"C:\Tools\OpenMD.exe" --dir "G:\"`
	commandLinePtr, err := windows.UTF16PtrFromString(commandLine)
	if err != nil {
		t.Fatal(err)
	}
	var argc int32
	argv, err := windows.CommandLineToArgv(commandLinePtr, &argc)
	if err != nil {
		t.Fatal(err)
	}
	args := make([]string, 0, argc-1)
	for i := int32(1); i < argc; i++ {
		args = append(args, windows.UTF16ToString(argv[i][:]))
	}

	inv, err := ParseArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Mode != ModeReadDir || inv.Path != `G:\` {
		t.Fatalf("invocation = %#v, want ModeReadDir G:\\", inv)
	}
}
