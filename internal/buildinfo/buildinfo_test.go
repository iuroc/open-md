package buildinfo

import "testing"

func TestIsReleaseBuild(t *testing.T) {
	oldVersion, oldBuildTime := Version, BuildTime
	defer func() {
		Version, BuildTime = oldVersion, oldBuildTime
	}()

	Version, BuildTime = "dev", ""
	if IsReleaseBuild() {
		t.Fatal("dev build must not be a release build")
	}
	Version, BuildTime = "2026.11.18-0959", ""
	if IsReleaseBuild() {
		t.Fatal("missing build time must not be a release build")
	}
	Version, BuildTime = "2026.11.18-0959", "2026-11-18T09:59:12+08:00"
	if !IsReleaseBuild() {
		t.Fatal("version + build time must be a release build")
	}
}

func TestBuildTimeDisplay(t *testing.T) {
	oldBuildTime := BuildTime
	defer func() { BuildTime = oldBuildTime }()

	BuildTime = "2026-11-18T09:59:12+08:00"
	if got := BuildTimeDisplay(); got != "2026-11-18 09:59:12 +08:00" {
		t.Fatalf("got %q", got)
	}
	BuildTime = ""
	if got := BuildTimeDisplay(); got != "未注入" {
		t.Fatalf("got %q", got)
	}
}

func TestVersionLabel(t *testing.T) {
	oldVersion, oldBuildTime := Version, BuildTime
	defer func() { Version, BuildTime = oldVersion, oldBuildTime }()

	Version, BuildTime = "dev", ""
	if got := VersionLabel(); got != "开发版" {
		t.Fatalf("got %q", got)
	}
	Version, BuildTime = "2026.11.18-0959", "2026-11-18T09:59:12+08:00"
	if got := VersionLabel(); got != "v2026.11.18-0959" {
		t.Fatalf("got %q", got)
	}
}
