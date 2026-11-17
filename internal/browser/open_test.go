package browser

import (
	"errors"
	"reflect"
	"testing"
)

func TestWindowsOpenerUsesRundll32(t *testing.T) {
	var calls [][]string
	o := &WindowsOpener{Run: func(name string, args ...string) error {
		calls = append(calls, append([]string{name}, args...))
		return nil
	}}
	if err := o.Open(`C:\temp\openmd\index.html`); err != nil {
		t.Fatal(err)
	}
	want := []string{"rundll32.exe", "url.dll,FileProtocolHandler", `C:\temp\openmd\index.html`}
	if !reflect.DeepEqual(calls[0], want) {
		t.Fatalf("calls = %#v", calls)
	}
}

func TestWindowsOpenerFallsBackToExplorer(t *testing.T) {
	calls := 0
	o := &WindowsOpener{Run: func(name string, args ...string) error {
		calls++
		if calls == 1 {
			return errors.New("rundll32 failed")
		}
		return nil
	}}
	if err := o.Open("x.html"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestWindowsOpenerReturnsErrorWhenBothFail(t *testing.T) {
	o := &WindowsOpener{Run: func(string, ...string) error { return errors.New("no browser") }}
	if err := o.Open("x.html"); err == nil {
		t.Fatal("want error")
	}
}
