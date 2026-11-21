package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	markerName    = ".openmd-cases"
	markerContent = "OpenMD 人工验收案例，由 tools/cases 生成，可安全删除。\n"
)

type Options struct {
	Out   string
	Big   bool
	Force bool
	Sizes BigSizes
}

type BigSizes struct {
	Oversize  int64
	TotalFile int64
	AtLimit   int64
}

func DefaultBigSizes() BigSizes {
	return BigSizes{
		Oversize:  33 << 20,
		TotalFile: 11 << 19,
		AtLimit:   32 << 20,
	}
}

type Fixture struct {
	Path  string
	Data  []byte
	Write func(io.Writer) error
	Dir   bool
	// Link, when set, creates a hard link (or a copy) from this relative
	// source path instead of writing new content.
	Link string
}

func Generate(opts Options) error {
	if opts.Out == "" {
		opts.Out = "cases"
	}
	if opts.Sizes == (BigSizes{}) {
		opts.Sizes = DefaultBigSizes()
	}

	if err := prepareOutputDir(opts.Out, opts.Force); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(opts.Out, markerName), []byte(markerContent), 0o644); err != nil {
		return fmt.Errorf("写入标记文件: %w", err)
	}
	for _, fixture := range fixtures(opts) {
		if err := writeFixture(opts.Out, fixture); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(opts.Out, "CHECKLIST.md"), Checklist(opts.Big), 0o644); err != nil {
		return fmt.Errorf("写入验收清单: %w", err)
	}
	return nil
}

func writeFixture(root string, fixture Fixture) error {
	if fixture.Path == "" {
		return errors.New("案例路径为空")
	}
	path := filepath.Join(root, filepath.FromSlash(fixture.Path))
	if fixture.Dir {
		return os.MkdirAll(path, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if fixture.Link != "" {
		return linkOrCopy(filepath.Join(root, filepath.FromSlash(fixture.Link)), path)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if fixture.Write != nil {
		if err := fixture.Write(file); err != nil {
			return fmt.Errorf("写入 %s: %w", fixture.Path, err)
		}
		return nil
	}
	if _, err := file.Write(fixture.Data); err != nil {
		return fmt.Errorf("写入 %s: %w", fixture.Path, err)
	}
	return nil
}

func prepareOutputDir(out string, force bool) error {
	info, err := os.Stat(out)
	if errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(out, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("输出路径不是目录：%s", out)
	}

	entries, err := os.ReadDir(out)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}

	marked := false
	for _, entry := range entries {
		if entry.Name() == markerName {
			marked = true
			break
		}
	}
	if !marked && !force {
		return fmt.Errorf("输出目录已存在且不是本工具生成的：%s；确认后可加 -force 覆盖", out)
	}
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	return os.MkdirAll(out, 0o755)
}
