package buildinfo

import "time"

var (
	Version   = "dev"
	BuildTime = ""
)

const (
	ReleaseURL = "https://github.com/iuroc/open-md/releases"
	AuthorName = "易舒"
	AuthorID   = "25768"
)

func IsReleaseBuild() bool {
	return Version != "" && Version != "dev" && BuildTime != ""
}

func VersionLabel() string {
	if !IsReleaseBuild() {
		return "开发版"
	}
	return "v" + Version
}

func BuildTimeDisplay() string {
	if BuildTime == "" {
		return "未注入"
	}
	parsed, err := time.Parse(time.RFC3339, BuildTime)
	if err != nil {
		return BuildTime
	}
	return parsed.Format("2006-01-02 15:04:05 -07:00")
}
