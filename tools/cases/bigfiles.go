package main

import (
	"io"
	"os"
)

const (
	bigHeader = "# 大文件用例\n\n"
	bigBlock  = "## 小节\n\n这是一段用于触发大小上限的占位正文。\n\n"
)

func BigFixtures(sizes BigSizes) []Fixture {
	return []Fixture{
		{Path: "limits/skip/big-33mib.md", Write: func(w io.Writer) error {
			return writeRepeatedMarkdown(w, sizes.Oversize)
		}},
		fixture("limits/skip/small.md", "# 跳过超限文件后仍可阅读\n\n这个文件应正常显示。\n"),
		{Path: "limits/total/01.md", Write: func(w io.Writer) error {
			return writeRepeatedMarkdown(w, sizes.TotalFile)
		}},
		{Path: "limits/total/02.md", Write: func(w io.Writer) error {
			return writeRepeatedMarkdown(w, sizes.TotalFile)
		}},
		{Path: "limits/total/03.md", Write: func(w io.Writer) error {
			return writeRepeatedMarkdown(w, sizes.TotalFile)
		}},
		{Path: "limits/total/04.md", Write: func(w io.Writer) error {
			return writeRepeatedMarkdown(w, sizes.TotalFile)
		}},
		{Path: "limits/total/05.md", Write: func(w io.Writer) error {
			return writeRepeatedMarkdown(w, sizes.TotalFile)
		}},
		{
			Path: "limits/all-big/big.md",
			Link: "limits/skip/big-33mib.md",
		},
		{Path: "limits/at-limit/big-32mib.md", Write: func(w io.Writer) error {
			return writeRepeatedMarkdown(w, sizes.AtLimit)
		}},
	}
}

// writeRepeatedMarkdown streams exactly size bytes of valid Markdown without
// allocating the whole file in memory.
func writeRepeatedMarkdown(w io.Writer, size int64) error {
	if size <= 0 {
		return nil
	}
	if size < int64(len(bigHeader)) {
		return writeRepeatedByte(w, '\n', size)
	}
	written, err := io.WriteString(w, bigHeader)
	if err != nil {
		return err
	}
	remaining := size - int64(written)

	for remaining >= int64(len(bigBlock)) {
		written, err = io.WriteString(w, bigBlock)
		if err != nil {
			return err
		}
		remaining -= int64(written)
	}
	return writeRepeatedByte(w, '\n', remaining)
}

func writeRepeatedByte(w io.Writer, b byte, n int64) error {
	buf := make([]byte, 4096)
	for i := range buf {
		buf[i] = b
	}
	for n > 0 {
		chunk := int64(len(buf))
		if n < chunk {
			chunk = n
		}
		written, err := w.Write(buf[:chunk])
		if err != nil {
			return err
		}
		n -= int64(written)
	}
	return nil
}

func linkOrCopy(src, dst string) error {
	if err := os.Link(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
