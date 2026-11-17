package install

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

type RegistryAssociator struct {
	ClassesRoot string
}

func NewRegistryAssociator() *RegistryAssociator {
	return &RegistryAssociator{ClassesRoot: `Software\Classes`}
}

func (a *RegistryAssociator) Register(exePath string) error {
	root := a.root()
	progID := root + `\OpenMD.Markdown`
	command := fmt.Sprintf(`"%s" "%%1"`, exePath)
	dirCommand := fmt.Sprintf(`"%s" --dir "%%1"`, exePath)
	backgroundCommand := fmt.Sprintf(`"%s" --dir "%%V"`, exePath)
	icon := fmt.Sprintf(`"%s",0`, exePath)

	values := []struct {
		path  string
		name  string
		value string
	}{
		{progID, "", "Markdown 文档"},
		{progID + `\DefaultIcon`, "", icon},
		{progID + `\shell\open\command`, "", command},

		{root + `\Applications\OpenMD.exe`, "FriendlyAppName", "OpenMD"},
		{root + `\Applications\OpenMD.exe\shell\open\command`, "", command},
		{root + `\Applications\OpenMD.exe\SupportedTypes`, ".md", ""},
		{root + `\Applications\OpenMD.exe\SupportedTypes`, ".markdown", ""},

		{root + `\.md\OpenWithProgids`, "OpenMD.Markdown", ""},
		{root + `\.markdown\OpenWithProgids`, "OpenMD.Markdown", ""},

		{root + `\SystemFileAssociations\.md\shell\OpenMD`, "MUIVerb", "用 OpenMD 阅读"},
		{root + `\SystemFileAssociations\.md\shell\OpenMD`, "Icon", icon},
		{root + `\SystemFileAssociations\.md\shell\OpenMD\command`, "", command},
		{root + `\SystemFileAssociations\.markdown\shell\OpenMD`, "MUIVerb", "用 OpenMD 阅读"},
		{root + `\SystemFileAssociations\.markdown\shell\OpenMD`, "Icon", icon},
		{root + `\SystemFileAssociations\.markdown\shell\OpenMD\command`, "", command},

		{root + `\Directory\shell\OpenMD`, "MUIVerb", "用 OpenMD 浏览 Markdown"},
		{root + `\Directory\shell\OpenMD`, "Icon", icon},
		{root + `\Directory\shell\OpenMD\command`, "", dirCommand},
		{root + `\Directory\Background\shell\OpenMD`, "MUIVerb", "用 OpenMD 浏览 Markdown"},
		{root + `\Directory\Background\shell\OpenMD`, "Icon", icon},
		{root + `\Directory\Background\shell\OpenMD\command`, "", backgroundCommand},
	}
	for _, item := range values {
		if err := setStringValue(item.path, item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (a *RegistryAssociator) Unregister() error {
	root := a.root()
	for _, path := range []string{
		root + `\.md\OpenWithProgids`,
		root + `\.markdown\OpenWithProgids`,
	} {
		if err := deleteValue(path, "OpenMD.Markdown"); err != nil {
			return err
		}
	}
	for _, path := range []string{
		root + `\OpenMD.Markdown`,
		root + `\Applications\OpenMD.exe`,
		root + `\SystemFileAssociations\.md\shell\OpenMD`,
		root + `\SystemFileAssociations\.markdown\shell\OpenMD`,
		root + `\Directory\shell\OpenMD`,
		root + `\Directory\Background\shell\OpenMD`,
	} {
		if err := deleteTree(registry.CURRENT_USER, path); err != nil {
			return err
		}
	}
	return nil
}

func (a *RegistryAssociator) root() string {
	if a.ClassesRoot == "" {
		return `Software\Classes`
	}
	return a.ClassesRoot
}

func setStringValue(path, name, value string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(name, value)
}

func deleteValue(path, name string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err == registry.ErrNotExist {
		return nil
	}
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.DeleteValue(name); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}

func deleteTree(parent registry.Key, path string) error {
	key, err := registry.OpenKey(parent, path, registry.READ)
	if err == registry.ErrNotExist {
		return nil
	}
	if err != nil {
		return err
	}
	names, err := key.ReadSubKeyNames(-1)
	_ = key.Close()
	if err != nil {
		return err
	}
	for _, name := range names {
		if err := deleteTree(parent, path+`\`+name); err != nil {
			return err
		}
	}
	if err := registry.DeleteKey(parent, path); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}
