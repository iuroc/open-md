package browser

import "os/exec"

type Opener interface {
	Open(path string) error
}

type CommandRunner func(name string, args ...string) error

type WindowsOpener struct {
	Run CommandRunner
}

func NewWindowsOpener() *WindowsOpener {
	return &WindowsOpener{Run: startProcess}
}

func (o *WindowsOpener) Open(path string) error {
	run := o.Run
	if run == nil {
		run = startProcess
	}
	if err := run("rundll32.exe", "url.dll,FileProtocolHandler", path); err == nil {
		return nil
	}
	return run("explorer.exe", path)
}

func startProcess(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
