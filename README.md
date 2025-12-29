# xeger-go

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

根据正则表达式反向生成匹配的随机字符串。

## 功能特性

- 🎲 根据正则表达式生成匹配的随机字符串
- 🔄 支持常见正则语法：字符类、量词、分组、选择等
- 📁 支持从文件批量读取正则表达式
- 🔒 并发安全
- 🛠️ 提供命令行工具

## 安装

```bash
go get github.com/gyyyy/xeger-go
```

## 快速开始

### 作为库使用

```go
package main

import (
    "fmt"
    "github.com/gyyyy/xeger-go"
)

func main() {
    // 创建 Xeger 实例，参数为量词最大展开次数增量（0 使用默认值 10）
    x := xeger.New(0)

    // 从正则表达式字符串生成
    result, err := x.GenerateFromString(`[a-z]+@[a-z]+\.(com|org|net)`)
    if err != nil {
        panic(err)
    }
    fmt.Println(result) // 例如：abc@example.com
}
```

### 使用编译后的正则表达式

```go
import "regexp"

reg := regexp.MustCompile(`\d{3}-\d{4}-\d{4}`)
result, err := x.Generate(reg)
// 例如：123-4567-8901
```

### 从 Reader 批量生成

```go
import "strings"

reader := strings.NewReader("[a-z]+\n\\d{4}\n(foo|bar)")
results := x.GenerateFromReader(reader)
// 按行读取正则，返回生成结果切片
```

## 命令行工具

### 安装

```bash
go install github.com/gyyyy/xeger-go/cmd@latest

# 重命名为 xeger（可选）
mv $(go env GOPATH)/bin/cmd $(go env GOPATH)/bin/xeger
```

### 使用方法

```bash
# 基本用法：生成匹配正则的字符串
xeger -r "[a-z]{5}"

# 指定生成次数（1-100）
xeger -r "\d{3}-\d{4}" -n 10

# 对输出进行编码（html/uri/query）
xeger -r "<script>" -e html

# 从文件读取正则表达式
xeger -f patterns.txt

# 输出到文件
xeger -r "[A-Z]{10}" -n 20 -o output.txt
```

### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-r` | 正则表达式 | - |
| `-n` | 输出次数（1-100） | 1 |
| `-e` | 编码模式：`html`/`uri`/`query` | - |
| `-f` | 输入文件路径（按行读取正则） | - |
| `-o` | 输出文件路径 | 标准输出 |

## 支持的正则语法

| 语法 | 说明 | 示例 |
|------|------|------|
| 字面量 | 直接匹配字符 | `hello` |
| `.` | 任意字符（可打印 ASCII） | `.+` |
| `[abc]` | 字符类 | `[a-zA-Z0-9]` |
| `[^abc]` | 否定字符类 | `[^0-9]` |
| `\d` `\D` | 数字 / 非数字 | `\d{4}` |
| `\w` `\W` | 单词字符 / 非单词字符 | `\w+` |
| `\s` `\S` | 空白 / 非空白 | `\s*` |
| `*` | 零次或多次 | `a*` |
| `+` | 一次或多次 | `a+` |
| `?` | 零次或一次 | `a?` |
| `{n}` | 恰好 n 次 | `a{3}` |
| `{n,}` | 至少 n 次 | `a{2,}` |
| `{n,m}` | n 到 m 次 | `a{2,5}` |
| `(...)` | 捕获组 | `(abc)+` |
| `\|` | 选择 | `cat\|dog` |
| `\b` | 词边界 | `\bword\b` |

## 示例

```go
x := xeger.New(0)

// 邮箱格式
x.GenerateFromString(`[a-z]+@[a-z]+\.(com|org|net)`)
// => "abc@xyz.com"

// IP 地址格式
x.GenerateFromString(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
// => "192.168.1.100"

// 手机号格式
x.GenerateFromString(`1[3-9]\d{9}`)
// => "13812345678"

// UUID 格式
x.GenerateFromString(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
// => "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

## 许可证

[MIT License](LICENSE)