package ui

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

type InstallMode int

const (
	InstallAfterInstall InstallMode = iota
	InstallManage
)

type Dialog interface {
	Installer(mode InstallMode) Choice
	ConfirmLoad(message string) LoadDecision
	Error(message string)
}
