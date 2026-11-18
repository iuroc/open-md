package app

import (
	"errors"
	"strings"
)

type Mode int

const (
	ModeInstall Mode = iota
	ModeReadFile
	ModeReadDir
)

type Invocation struct {
	Mode Mode
	Path string
}

func ParseArgs(args []string) (Invocation, error) {
	if len(args) == 0 {
		return Invocation{Mode: ModeInstall}, nil
	}

	if args[0] == "--dir" {
		if len(args) != 2 {
			return Invocation{}, errors.New("--dir 需要一个目录路径")
		}
		path := normalizePathArg(args[1])
		if path == "" {
			return Invocation{}, errors.New("--dir 需要一个目录路径")
		}
		return Invocation{Mode: ModeReadDir, Path: path}, nil
	}

	if len(args) != 1 {
		return Invocation{}, errors.New("一次只能打开一个文件")
	}
	if strings.HasPrefix(args[0], "--") {
		return Invocation{}, errors.New("未知参数：" + args[0])
	}
	path := normalizePathArg(args[0])
	if path == "" {
		return Invocation{}, errors.New("文件路径不能为空")
	}
	return Invocation{Mode: ModeReadFile, Path: path}, nil
}

// Windows command lines quote an argument as "%1". If the path ends with a
// backslash, such as the root "G:\", the backslash escapes the closing quote.
// CommandLineToArgvW then returns a path ending with a literal quote. Windows
// file names cannot contain quotes, so strip those artifacts here.
func normalizePathArg(path string) string {
	hadTrailingQuote := strings.HasSuffix(path, `"`)
	path = strings.TrimRight(path, `"`)
	// CommandLineToArgvW consumes the backslash in a quoted drive root
	// ("G:\") and returns `G:"`. Windows file names cannot contain quotes, so
	// restore the root separator when the quote was the only trailing artifact.
	if hadTrailingQuote && len(path) == 2 && path[1] == ':' {
		path += `\`
	}
	return path
}
