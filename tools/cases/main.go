package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
)

func main() {
	out := flag.String("out", "cases", "输出目录")
	big := flag.Bool("big", false, "追加大小上限用例（约 100 MB）")
	force := flag.Bool("force", false, "覆盖不是本工具生成的非空目录")
	flag.Parse()

	if err := Generate(Options{Out: *out, Big: *big, Force: *force}); err != nil {
		log.Fatal(err)
	}

	abs, err := filepath.Abs(*out)
	if err != nil {
		abs = *out
	}
	if *big {
		fmt.Printf("已生成含大小上限用例的验收案例：%s\n", abs)
	} else {
		fmt.Printf("已生成轻量验收案例：%s（大小上限用例请加 -big）\n", abs)
	}
}
