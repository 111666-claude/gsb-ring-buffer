// Command ring-buffer 按脚本驱动环形缓冲。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"example.com/ringbuffer"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ring-buffer", flag.ContinueOnError)
	flags.SetOutput(stderr)
	capacity := flags.Int("cap", 8, "容量")
	ops := flags.Int("ops", 0, "写入这么多个递增数字")
	write := flags.String("write", "", "写入这段文本")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	buffer := ringbuffer.New(*capacity)
	if *ops > 0 {
		for index := 0; index < *ops; index++ {
			token := []byte(strconv.Itoa(index % 10))
			buffer.Write(token)
		}
	} else if *write != "" {
		buffer.Write([]byte(*write))
	} else {
		fmt.Fprintln(stderr, "用法：ring-buffer --cap 8 --write abcdefghij | --ops 100000")
		return 2
	}
	content := buffer.Bytes()
	if len(content) <= 40 {
		fmt.Fprintf(stdout, "content=%s\n", string(content))
	}
	fmt.Fprintf(stdout, "len=%d cap=%d copied=%d\n", buffer.Len(), buffer.Cap(), buffer.Copied())
	_ = strings.TrimSpace
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
