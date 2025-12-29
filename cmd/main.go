package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gyyyy/xeger-go"
)

// encode 编码转换
func encode(s, encoding string) string {
	switch strings.ToLower(encoding) {
	case "html":
		return html.EscapeString(s)
	case "uri":
		return url.PathEscape(s)
	case "query":
		return url.QueryEscape(s)
	}
	return s
}

func main() {
	var (
		reg      = flag.String("r", "", "正则表达式")
		times    = flag.Int("n", 1, "输出次数：范围 1 到 100（默认 1）")
		encoding = flag.String("e", "", "编码模式：html/uri/query")
		input    = flag.String("f", "", "输入文件路径")
		output   = flag.String("o", "", "输出文件路径")
	)
	flag.Parse()
	var (
		in  *os.File
		out *os.File
		err error
	)
	// 打开输入文件
	if *input != "" {
		in, err = os.Open(*input)
		if err != nil {
			log.Fatalln(err)
		}
		defer in.Close()
	}
	// 打开输出文件
	if *output != "" {
		if path, err := filepath.Abs(*output); err == nil {
			*output = path
		}
		if out, err = os.Create(*output); err != nil {
			log.Fatalln(err)
		}
		defer out.Close()
	}
	if out == nil {
		out = os.Stdout
	}
	if in != nil {
		x := xeger.New(0)
		// 从文件逐行读取正则表达式并生成
		for i, result := range x.GenerateFromReader(in) {
			// 编码并输出
			fmt.Fprintf(out, "%d. %s\n", i+1, encode(result, *encoding))
		}
	} else {
		// 根据命令行参数生成
		if *reg == "" {
			log.Fatalln("正则表达式不能为空，请使用 -r 参数指定")
		}
		if *times < 1 {
			*times = 1
		}
		if *times > 100 {
			*times = 100
		}
		x := xeger.New(max(*times, xeger.DefaultMaxExpand))
		// 生成指定次数的字符串
		for i := 0; i < *times; {
			// 生成字符串
			result, err := x.GenerateFromString(*reg)
			if err != nil {
				log.Printf("生成失败：%v\n", err)
				continue
			}
			// 编码并输出
			fmt.Fprintf(out, "%d. %s\n", i+1, encode(result, *encoding))
			i++
		}
	}
	// 输出结果路径
	if *output != "" {
		log.Println("结果输出至文件：", *output)
	}
}
