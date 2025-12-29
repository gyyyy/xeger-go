package xeger

import (
	"regexp"
	"strings"
	"testing"
)

// 测试用例结构
type testCase struct {
	name    string
	pattern string
}

// 基础功能测试用例
var basicTestCases = []testCase{
	// 字面量
	{"literal", `hello`},
	{"literal_with_special", `hello\.world`},

	// 字符类
	{"char_class_simple", `[abc]`},
	{"char_class_range", `[a-z]`},
	{"char_class_multi_range", `[a-zA-Z0-9]`},
	{"char_class_negated", `[^abc]`},
	{"char_class_negated_range", `[^a-z]`},

	// 预定义字符类
	{"digit", `\d`},
	{"non_digit", `\D`},
	{"word", `\w`},
	{"non_word", `\W`},
	{"whitespace", `\s`},
	{"non_whitespace", `\S`},

	// 任意字符
	{"any_char", `.`},
	{"any_char_multiple", `...`},

	// 量词
	{"star", `a*`},
	{"plus", `a+`},
	{"question", `a?`},
	{"repeat_exact", `a{3}`},
	{"repeat_range", `a{2,5}`},
	{"repeat_min", `a{2,}`},
	{"repeat_zero_or_more", `[a-z]*`},

	// 分组
	{"group_simple", `(abc)`},
	{"group_nested", `((a)(b))`},
	{"group_with_quantifier", `(abc)+`},

	// 选择
	{"alternation", `a|b|c`},
	{"alternation_group", `(cat|dog|bird)`},
	{"alternation_complex", `(foo|bar)(baz|qux)`},

	// 锚点（应该不影响生成）
	{"anchor_start", `^hello`},
	{"anchor_end", `world$`},
	{"anchor_both", `^hello$`},

	// 复杂模式
	{"email_like", `[a-z]+@[a-z]+\.[a-z]{2,4}`},
	{"ip_like", `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`},
	{"phone_like", `\d{3}-\d{3}-\d{4}`},
	{"url_like", `https?://[a-z]+\.[a-z]+`},
	{"uuid_like", `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`},

	// 边界情况
	{"empty_alternation", `(|a)`},
	{"optional_group", `(abc)?`},
	{"nested_quantifiers", `(a+)+`},
	{"complex_char_class", `[a-zA-Z_][a-zA-Z0-9_]*`},
}

// 测试生成的字符串是否能匹配原正则表达式
func TestGenerateMatchesPattern(t *testing.T) {
	x := New(0)
	for _, tc := range basicTestCases {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := regexp.Compile(tc.pattern)
			if err != nil {
				t.Fatalf("无法编译正则表达式 %q: %v", tc.pattern, err)
			}

			// 生成多次以增加覆盖率
			for i := 0; i < 100; i++ {
				result, err := x.Generate(reg)
				if err != nil {
					t.Errorf("生成失败: %v", err)
					continue
				}

				str := string(result)
				if !reg.MatchString(str) {
					t.Errorf("生成的字符串 %q 不匹配正则 %q", str, tc.pattern)
				}
			}
		})
	}
}

// 测试 GenerateFromString
func TestGenerateFromString(t *testing.T) {
	x := New(0)
	testCases := []string{
		`[a-z]+`,
		`\d{3}-\d{4}`,
		`(hello|world)`,
	}

	for _, pattern := range testCases {
		t.Run(pattern, func(t *testing.T) {
			reg := regexp.MustCompile(pattern)
			for i := 0; i < 50; i++ {
				result, err := x.GenerateFromString(pattern)
				if err != nil {
					t.Errorf("生成失败: %v", err)
					continue
				}
				if !reg.MatchString(result) {
					t.Errorf("生成的字符串 %q 不匹配正则 %q", result, pattern)
				}
			}
		})
	}
}

// 测试 GenerateFromReader
func TestGenerateFromReader(t *testing.T) {
	x := New(0)
	input := "[a-z]+\n\\d{3}\n(foo|bar)"
	reader := strings.NewReader(input)
	results := x.GenerateFromReader(reader)

	if len(results) != 3 {
		t.Errorf("期望 3 个结果，得到 %d 个", len(results))
	}

	patterns := []string{`[a-z]+`, `\d{3}`, `(foo|bar)`}
	for i, result := range results {
		if i < len(patterns) {
			reg := regexp.MustCompile(patterns[i])
			if !reg.MatchString(result) {
				t.Errorf("结果 %d: %q 不匹配 %q", i, result, patterns[i])
			}
		}
	}
}

// 测试单个选择项（OpAlternate 修复验证）
func TestSingleAlternation(t *testing.T) {
	x := New(0)

	// 测试只有一个选项的选择
	patterns := []string{
		`(a)`,
		`(abc)`,
		`(hello|)`, // 空选项
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			reg := regexp.MustCompile(pattern)
			for i := 0; i < 20; i++ {
				result, err := x.GenerateFromString(pattern)
				if err != nil {
					t.Errorf("生成失败: %v", err)
					continue
				}
				if !reg.MatchString(result) {
					t.Errorf("生成的字符串 %q 不匹配 %q", result, pattern)
				}
			}
		})
	}
}

// 测试空字符类边界情况
func TestEdgeCases(t *testing.T) {
	x := New(0)

	// 测试单个选择项
	t.Run("single_alternation", func(t *testing.T) {
		result, err := x.GenerateFromString(`(a)`)
		if err != nil {
			t.Errorf("生成失败: %v", err)
			return
		}
		if result != "a" {
			t.Errorf("期望 'a'，得到 %q", result)
		}
	})

	// 测试空模式
	t.Run("empty_pattern", func(t *testing.T) {
		result, err := x.GenerateFromString(``)
		if err != nil {
			t.Errorf("生成失败: %v", err)
			return
		}
		if result != "" {
			t.Errorf("期望空字符串，得到 %q", result)
		}
	})

	// 测试只有量词的模式
	t.Run("only_star", func(t *testing.T) {
		reg := regexp.MustCompile(`a*`)
		for i := 0; i < 50; i++ {
			result, err := x.GenerateFromString(`a*`)
			if err != nil {
				t.Errorf("生成失败: %v", err)
				continue
			}
			if !reg.MatchString(result) {
				t.Errorf("生成的字符串 %q 不匹配", result)
			}
		}
	})
}

// 测试特殊字符处理
func TestSpecialCharacters(t *testing.T) {
	x := New(0)
	specialPatterns := []string{
		`\+`,
		`\*`,
		`\?`,
		`\[`,
		`\]`,
		`\(`,
		`\)`,
		`\{`,
		`\}`,
		`\\`,
		`\.`,
		`\^`,
		`\$`,
		`\|`,
	}

	for _, pattern := range specialPatterns {
		t.Run(pattern, func(t *testing.T) {
			reg := regexp.MustCompile(pattern)
			result, err := x.GenerateFromString(pattern)
			if err != nil {
				t.Errorf("生成失败: %v", err)
				return
			}
			if !reg.MatchString(result) {
				t.Errorf("生成的字符串 %q 不匹配 %q", result, pattern)
			}
		})
	}
}

// 测试 Unicode 字符
func TestUnicode(t *testing.T) {
	x := New(0)
	patterns := []testCase{
		{"unicode_range", `[一-龥]`},
		{"unicode_literal", `你好`},
		{"mixed_unicode", `[a-z]+[一-龥]+`},
	}

	for _, tc := range patterns {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := regexp.Compile(tc.pattern)
			if err != nil {
				t.Skipf("跳过无法编译的模式: %v", err)
				return
			}
			for i := 0; i < 20; i++ {
				result, err := x.GenerateFromString(tc.pattern)
				if err != nil {
					t.Errorf("生成失败: %v", err)
					continue
				}
				if !reg.MatchString(result) {
					t.Errorf("生成的字符串 %q 不匹配 %q", result, tc.pattern)
				}
			}
		})
	}
}

// 测试词边界
func TestWordBoundary(t *testing.T) {
	x := New(0)
	// \b 在生成时会产生空格，这可能导致不匹配
	pattern := `\bword\b`
	reg := regexp.MustCompile(pattern)

	matchCount := 0
	for i := 0; i < 100; i++ {
		result, err := x.GenerateFromString(pattern)
		if err != nil {
			t.Errorf("生成失败: %v", err)
			continue
		}
		if reg.MatchString(result) {
			matchCount++
		}
	}
	t.Logf("词边界测试: %d/100 匹配", matchCount)
}

// 测试非贪婪量词
func TestNonGreedy(t *testing.T) {
	x := New(0)
	patterns := []string{
		`a+?`,
		`a*?`,
		`a??`,
		`a{2,5}?`,
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			reg := regexp.MustCompile(pattern)
			for i := 0; i < 50; i++ {
				result, err := x.GenerateFromString(pattern)
				if err != nil {
					t.Errorf("生成失败: %v", err)
					continue
				}
				if !reg.MatchString(result) {
					t.Errorf("生成的字符串 %q 不匹配 %q", result, pattern)
				}
			}
		})
	}
}

// 性能测试 - 简单模式
func BenchmarkGenerateSimple(b *testing.B) {
	x := New(0)
	reg := regexp.MustCompile(`[a-z]+`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Generate(reg)
	}
}

// 性能测试 - 复杂模式
func BenchmarkGenerateComplex(b *testing.B) {
	x := New(0)
	reg := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,4}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Generate(reg)
	}
}

// 性能测试 - 带选择的模式
func BenchmarkGenerateAlternation(b *testing.B) {
	x := New(0)
	reg := regexp.MustCompile(`(SELECT|DELETE|UPDATE|INSERT).+(FROM|INTO|SET).+`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Generate(reg)
	}
}

// 性能测试 - 嵌套量词
func BenchmarkGenerateNestedQuantifiers(b *testing.B) {
	x := New(0)
	reg := regexp.MustCompile(`((a+b)+c)+`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Generate(reg)
	}
}

// 性能测试 - UUID 模式
func BenchmarkGenerateUUID(b *testing.B) {
	x := New(0)
	reg := regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Generate(reg)
	}
}

// 性能测试 - GenerateFromString
func BenchmarkGenerateFromString(b *testing.B) {
	x := New(0)
	pattern := `[a-zA-Z0-9]{10,20}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.GenerateFromString(pattern)
	}
}

// 内存分配测试
func BenchmarkGenerateAllocs(b *testing.B) {
	x := New(0)
	reg := regexp.MustCompile(`[a-z]{10}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x.Generate(reg)
	}
}

// 测试自定义最大展开次数
func TestCustomMaxExpand(t *testing.T) {
	// 使用较大的展开次数
	x := New(10)

	// 测试 * 量词
	t.Run("star_with_max", func(t *testing.T) {
		reg := regexp.MustCompile(`^a*$`)
		maxLen := 0
		for i := 0; i < 100; i++ {
			result, err := x.GenerateFromString(`a*`)
			if err != nil {
				t.Errorf("生成失败: %v", err)
				continue
			}
			if len(result) > maxLen {
				maxLen = len(result)
			}
			if !reg.MatchString(result) {
				t.Errorf("生成的字符串 %q 不匹配", result)
			}
		}
		t.Logf("* 量词最大长度: %d", maxLen)
	})

	// 测试 + 量词
	t.Run("plus_with_max", func(t *testing.T) {
		reg := regexp.MustCompile(`^a+$`)
		maxLen := 0
		for i := 0; i < 100; i++ {
			result, err := x.GenerateFromString(`a+`)
			if err != nil {
				t.Errorf("生成失败: %v", err)
				continue
			}
			if len(result) > maxLen {
				maxLen = len(result)
			}
			if !reg.MatchString(result) {
				t.Errorf("生成的字符串 %q 不匹配", result)
			}
		}
		t.Logf("+ 量词最大长度: %d", maxLen)
	})
}

// 并发安全性测试
func TestConcurrentGenerate(t *testing.T) {
	x := New(0)
	reg := regexp.MustCompile(`[a-z]+`)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, err := x.Generate(reg)
				if err != nil {
					t.Errorf("并发生成失败: %v", err)
				}
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// SQL 注入模式测试（与原测试类似）
func TestSQLInjectionPatterns(t *testing.T) {
	x := New(0)
	pattern := `^\+\/v(8|9)|\b(and|or)\b.{1,6}?(=|>|<|\bin\b|\blike\b)|\/\*.+?\*\/|<\s*script\b|\bEXEC\b|UNION.+?SELECT|UPDATE.+?SET|INSERT\s+INTO.+?VALUES|(SELECT|DELETE).+?FROM|(CREATE|ALTER|DROP|TRUNCATE)\s+(TABLE|DATABASE)`

	reg := regexp.MustCompile(pattern)
	matchCount := 0
	totalTests := 100

	for i := 0; i < totalTests; i++ {
		result, err := x.GenerateFromString(pattern)
		if err != nil {
			t.Errorf("生成失败: %v", err)
			continue
		}
		if reg.MatchString(result) {
			matchCount++
		}
	}

	t.Logf("SQL 注入模式匹配率: %d/%d (%.1f%%)", matchCount, totalTests, float64(matchCount)/float64(totalTests)*100)

	// 期望至少有一定比例能匹配
	if matchCount < totalTests/2 {
		t.Errorf("匹配率过低: %d/%d", matchCount, totalTests)
	}
}
