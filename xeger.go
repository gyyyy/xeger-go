package xeger

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"regexp"
	"regexp/syntax"
	"sync"
	"unicode"
)

// DefaultMaxExpand 量词（*、+、{n,}）的默认最大展开次数增量
const DefaultMaxExpand = 10

// Xeger 根据正则表达式反向生成匹配的随机字符串
type Xeger struct {
	mu        sync.Mutex    // 互斥锁，保证并发安全
	maxExpand int           // 量词的最大展开次数增量
	rand      *rand.Rand    // 随机数生成器，用于随机选择字符和重复次数
	buf       *bytes.Buffer // 字符串构建缓冲区
}

// intn 返回 [0, n) 范围内的随机整数
func (x *Xeger) intn(n int) int {
	return x.rand.IntN(n)
}

// generate 递归处理正则语法树，生成匹配的字符串
func (x *Xeger) generate(regSyntax *syntax.Regexp) {
	switch regSyntax.Op {
	// 字面量：直接写入字符，如 "hello"
	case syntax.OpLiteral:
		for _, r := range regSyntax.Rune {
			x.buf.WriteRune(r)
		}
	// 字符类：如 [a-z]、[^0-9]，按字符总数均匀随机选一个
	case syntax.OpCharClass:
		if len(regSyntax.Rune)%2 != 0 || len(regSyntax.Rune) == 0 {
			break
		}
		// 处理否定字符类 [^...]，将范围限制在可打印 ASCII 字符内
		runes := regSyntax.Rune
		if runes[0] == 0 && runes[len(runes)-1] == unicode.MaxRune {
			// 复制一份避免修改原始数据
			runes = make([]rune, len(regSyntax.Rune))
			copy(runes, regSyntax.Rune)
			if runes[1] >= 0x20 {
				runes[0] = 0x20 // 空格
			}
			if runes[len(runes)-2] <= 0x7e {
				runes[len(runes)-1] = 0x7e // ~
			}
		}
		// 计算所有范围的字符总数
		var total int
		for i := 0; i < len(runes); i += 2 {
			total += int(runes[i+1]-runes[i]) + 1
		}
		if total <= 0 {
			break
		}
		// 按字符总数均匀随机选择
		n := x.intn(total)
		for i := 0; i < len(runes); i += 2 {
			var (
				low  = runes[i]
				high = runes[i+1]
				size = int(high-low) + 1
			)
			if n < size {
				x.buf.WriteRune(low + rune(n))
				break
			}
			n -= size
		}
	// 任意字符：. 匹配任意字符，这里限制为可打印 ASCII 字符
	case syntax.OpAnyCharNotNL, syntax.OpAnyChar:
		x.buf.WriteRune(rune(x.intn(0x5f) + 0x20)) // 0x20-0x7e 可打印字符
	// 词边界：\b，生成空格作为边界
	case syntax.OpWordBoundary:
		x.buf.WriteRune(0x20)
	// 捕获组：(...)，递归处理子表达式
	case syntax.OpCapture:
		x.generate(regSyntax.Sub[0])
	// 零或多次：*，生成 [0, maxExpand) 次
	case syntax.OpStar:
		for i := 0; i < x.intn(x.maxExpand); i++ {
			x.generate(regSyntax.Sub[0])
		}
	// 一或多次：+，生成 [1, maxExpand+1) 次
	case syntax.OpPlus:
		for i := 0; i < x.intn(x.maxExpand)+1; i++ {
			x.generate(regSyntax.Sub[0])
		}
	// 可选：?，50% 概率生成
	case syntax.OpQuest:
		if x.intn(2) == 1 {
			x.generate(regSyntax.Sub[0])
		}
	// 重复次数：{n}、{n,}、{n,m}
	case syntax.OpRepeat:
		var n int
		if regSyntax.Max < 0 { // {n,} 无上限，生成 [min, min+maxExpand) 次
			n = x.intn(x.maxExpand)
		} else if regSyntax.Min < regSyntax.Max { // {n,m} 在范围内随机
			n = x.intn(regSyntax.Max - regSyntax.Min + 1)
		}
		for i := 0; i < regSyntax.Min+n; i++ {
			x.generate(regSyntax.Sub[0])
		}
	// 连接：依次处理所有子表达式
	case syntax.OpConcat:
		for _, sub := range regSyntax.Sub {
			x.generate(sub)
		}
	// 选择：a|b|c，随机选一个分支
	case syntax.OpAlternate:
		if len(regSyntax.Sub) == 0 {
			break
		}
		if len(regSyntax.Sub) == 1 {
			x.generate(regSyntax.Sub[0])
		} else {
			x.generate(regSyntax.Sub[x.intn(len(regSyntax.Sub))])
		}
	}
}

// Generate 根据编译后的正则表达式生成匹配的随机字节序列
//
// 参数：
//   - reg: 编译后的正则表达式
//
// 返回：
//   - []byte: 生成的匹配字节序列
//   - error: 可能发生的错误
func (x *Xeger) Generate(reg *regexp.Regexp) (result []byte, err error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	// 捕获可能的 panic，转换为错误返回
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("generate panic: %v", p)
			result = nil
		}
	}()
	// 解析正则表达式为语法树
	regSyntax, err := syntax.Parse(reg.String(), syntax.Perl)
	if err != nil {
		return nil, err
	}
	x.buf.Reset()
	x.generate(regSyntax)
	// 返回副本以避免后续调用覆盖数据
	result = make([]byte, x.buf.Len())
	copy(result, x.buf.Bytes())
	return result, nil
}

// GenerateFromString 根据正则表达式字符串生成匹配的随机字符串
//
// 参数：
//   - s: 正则表达式字符串
//
// 返回：
//   - string: 生成的匹配字符串
//   - error: 可能发生的错误
func (x *Xeger) GenerateFromString(s string) (string, error) {
	reg, err := regexp.Compile(s)
	if err != nil {
		return "", err
	}
	b, err := x.Generate(reg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// GenerateFromReader 从 Reader 按行读取正则表达式，逐行生成匹配的字符串
//
// 参数：
//   - reader: io.Reader 接口，用于读取正则表达式字符串
//
// 返回：
//   - []string: 生成的匹配字符串切片
func (x *Xeger) GenerateFromReader(reader io.Reader) []string {
	var results []string
	// 按行读取正则表达式并生成匹配字符串
	for scanner := bufio.NewScanner(reader); scanner.Scan(); {
		if line := scanner.Text(); line != "" {
			result, _ := x.GenerateFromString(line)
			results = append(results, result)
		}
	}
	return results
}

// New 创建一个使用默认配置的 Xeger 实例
//
// 参数：
//   - maxExpand: 量词的最大展开次数增量
//
// 返回：
//   - *Xeger: 新创建的 Xeger 实例
func New(maxExpand int) *Xeger {
	if maxExpand < 1 {
		maxExpand = DefaultMaxExpand
	}
	if maxExpand > 100 {
		maxExpand = 100
	}
	return &Xeger{
		maxExpand: maxExpand,
		rand:      rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		buf:       bytes.NewBuffer(make([]byte, 0, 1024)),
	}
}
