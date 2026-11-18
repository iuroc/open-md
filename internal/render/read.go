package render

import (
	"fmt"
	"io"
	"os"
)

var errFileTooLarge = fmt.Errorf("文件超过大小限制")

func readFileLimited(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if limit <= 0 {
		return io.ReadAll(file)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errFileTooLarge
	}
	return data, nil
}
