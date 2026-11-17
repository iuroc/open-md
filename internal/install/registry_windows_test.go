package install

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows/registry"
)

func TestRegistryAssociatorPreservesOtherAssociations(t *testing.T) {
	root := fmt.Sprintf(`Software\OpenMDTest-%d`, time.Now().UnixNano())
	defer deleteTree(registry.CURRENT_USER, root)

	key, _, err := registry.CreateKey(registry.CURRENT_USER, root+`\.md`, registry.ALL_ACCESS)
	if err != nil {
		t.Fatal(err)
	}
	if err := key.SetStringValue("", "OtherApp.md"); err != nil {
		t.Fatal(err)
	}
	openWith, _, err := registry.CreateKey(key, "OpenWithProgids", registry.ALL_ACCESS)
	if err != nil {
		t.Fatal(err)
	}
	if err := openWith.SetStringValue("Other.Prog", ""); err != nil {
		t.Fatal(err)
	}
	_ = openWith.Close()
	_ = key.Close()

	a := &RegistryAssociator{ClassesRoot: root}
	if err := a.Register(`C:\Tools\OpenMD.exe`); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		root + `\OpenMD.Markdown\shell\open\command`,
		root + `\.md\OpenWithProgids`,
		root + `\SystemFileAssociations\.md\shell\OpenMD\command`,
		root + `\Directory\shell\OpenMD\command`,
		root + `\Directory\Background\shell\OpenMD\command`,
		root + `\Applications\OpenMD.exe\shell\open\command`,
	} {
		k, err := registry.OpenKey(registry.CURRENT_USER, path, registry.READ)
		if err != nil {
			t.Fatalf("missing key %q: %v", path, err)
		}
		_ = k.Close()
	}

	commandKey, err := registry.OpenKey(registry.CURRENT_USER, root+`\OpenMD.Markdown\shell\open\command`, registry.READ)
	if err != nil {
		t.Fatal(err)
	}
	command, _, err := commandKey.GetStringValue("")
	_ = commandKey.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(command, `C:\Tools\OpenMD.exe`) || !strings.Contains(command, `%1`) {
		t.Fatalf("command = %q", command)
	}

	if err := a.Unregister(); err != nil {
		t.Fatal(err)
	}

	mdKey, err := registry.OpenKey(registry.CURRENT_USER, root+`\.md`, registry.READ)
	if err != nil {
		t.Fatal(err)
	}
	defaultValue, _, err := mdKey.GetStringValue("")
	_ = mdKey.Close()
	if err != nil || defaultValue != "OtherApp.md" {
		t.Fatalf("default association changed: %q %v", defaultValue, err)
	}

	knownOpenWith, err := registry.OpenKey(registry.CURRENT_USER, root+`\.md\OpenWithProgids`, registry.READ)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := knownOpenWith.GetStringValue("Other.Prog"); err != nil {
		t.Fatal("other app's OpenWithProgids value removed")
	}
	if _, _, err := knownOpenWith.GetStringValue("OpenMD.Markdown"); err == nil {
		t.Fatal("OpenMD value still present")
	}
	_ = knownOpenWith.Close()

	if keyExists(root + `\OpenMD.Markdown`) {
		t.Fatal("OpenMD ProgID key still present")
	}
	if keyExists(root + `\Directory\shell\OpenMD`) {
		t.Fatal("Directory shell key still present")
	}
}

func keyExists(path string) bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.READ)
	if err != nil {
		return false
	}
	_ = key.Close()
	return true
}
