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
		if len(args) != 2 || args[1] == "" {
			return Invocation{}, errors.New("--dir 需要一个目录路径")
		}
		return Invocation{Mode: ModeReadDir, Path: args[1]}, nil
	}

	if len(args) != 1 {
		return Invocation{}, errors.New("一次只能打开一个文件")
	}
	if strings.HasPrefix(args[0], "--") {
		return Invocation{}, errors.New("未知参数：" + args[0])
	}
	if args[0] == "" {
		return Invocation{}, errors.New("文件路径不能为空")
	}
	return Invocation{Mode: ModeReadFile, Path: args[0]}, nil
}
