package main

import (
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

// newDecodeCtx 使用给定的未类型化 token 构造 DecodeContext，Value 字段在此处可为 nil
func newDecodeCtx(args ...string) *kong.DecodeContext {
	return &kong.DecodeContext{Scan: kong.Scan(args...)}
}

// testCLI 是集成测试使用的参数结构体
type testCLI struct {
	Number      boolValue `name:"number"       default:"true"`
	Lower       boolValue `name:"lower"        default:"true"`
	Upper       boolValue `name:"upper"        default:"true"`
	Special     boolValue `name:"special"      default:"false"`
	SpecialSafe boolValue `name:"special-safe" default:"false"`
	Count       int       `name:"count"        default:"1"`
	Length      int       `name:"length"       default:"20"`
}

// parseArgs 通过 kong 解析参数，返回填充后的 testCLI
func parseArgs(t *testing.T, args []string) testCLI {
	t.Helper()
	var c testCLI
	p, err := kong.New(&c)
	if err != nil {
		t.Fatalf("构建 parser 失败：%v", err)
	}
	_, err = p.Parse(args)
	if err != nil {
		t.Fatalf("解析参数失败：%v", err)
	}
	return c
}

// withCLI 备份全局 cli，执行 f，结束后还原
func withCLI(t *testing.T, f func()) {
	t.Helper()
	orig := cli
	defer func() { cli = orig }()
	f()
}

// ---- 单元测试：直接构造 DecodeContext ----

// TestBoolDecode_NoValue 测试不传值时默认为 true，且不消耗任何 token
func TestBoolDecode_NoValue(t *testing.T) {
	var b boolValue
	ctx := newDecodeCtx()
	if err := b.Decode(ctx); err != nil {
		t.Fatalf("不期望出错，但得到：%v", err)
	}
	if !bool(b) {
		t.Error("期望默认为 true，但得到 false")
	}
}

// TestBoolDecode_AdjacentFlag 测试下一个 token 是 flag 时不消耗它，且值默认为 true
func TestBoolDecode_AdjacentFlag(t *testing.T) {
	var b boolValue
	ctx := &kong.DecodeContext{
		Scan: kong.ScanFromTokens(kong.Token{Value: "count", Type: kong.FlagToken}),
	}
	if err := b.Decode(ctx); err != nil {
		t.Fatalf("不期望出错，但得到：%v", err)
	}
	if !bool(b) {
		t.Error("期望默认为 true，但得到 false")
	}
	if ctx.Scan.Peek().IsEOL() {
		t.Error("--count token 被意外消耗")
	}
}

// TestBoolDecode_ExplicitTrue 测试显式传 "true" 和 "1"
func TestBoolDecode_ExplicitTrue(t *testing.T) {
	for _, val := range []string{"true", "True", "TRUE", "1"} {
		var b boolValue
		ctx := newDecodeCtx(val)
		if err := b.Decode(ctx); err != nil {
			t.Fatalf("val=%q：不期望出错，但得到：%v", val, err)
		}
		if !bool(b) {
			t.Errorf("val=%q：期望 true，但得到 false", val)
		}
	}
}

// TestBoolDecode_ExplicitFalse 测试显式传 "false" 和 "0"
func TestBoolDecode_ExplicitFalse(t *testing.T) {
	for _, val := range []string{"false", "False", "FALSE", "0"} {
		b := boolValue(true)
		ctx := newDecodeCtx(val)
		if err := b.Decode(ctx); err != nil {
			t.Fatalf("val=%q：不期望出错，但得到：%v", val, err)
		}
		if bool(b) {
			t.Errorf("val=%q：期望 false，但得到 true", val)
		}
	}
}

// TestBoolDecode_InvalidValue 测试非法值返回错误
func TestBoolDecode_InvalidValue(t *testing.T) {
	var b boolValue
	ctx := newDecodeCtx("abc")
	if err := b.Decode(ctx); err == nil {
		t.Error("期望非法值返回错误，但未出错")
	}
}

// ---- 集成测试：通过 kong.Parse 解析完整参数 ----

// TestIntegration_AdjacentFlag 核心回归测试：--special 后紧跟 --count 2 时两个参数均正确解析
func TestIntegration_AdjacentFlag(t *testing.T) {
	c := parseArgs(t, []string{"--special", "--count", "2"})
	if !bool(c.Special) {
		t.Error("期望 Special=true，但得到 false")
	}
	if c.Count != 2 {
		t.Errorf("期望 Count=2，但得到 %d（--count 被意外丢弃）", c.Count)
	}
}

// TestIntegration_ExplicitBool 测试显式赋值写法
func TestIntegration_ExplicitBool(t *testing.T) {
	c := parseArgs(t, []string{"--lower", "false", "--upper", "true"})
	if bool(c.Lower) {
		t.Error("期望 Lower=false，但得到 true")
	}
	if !bool(c.Upper) {
		t.Error("期望 Upper=true，但得到 false")
	}
}

// TestIntegration_DefaultValues 测试不传任何 bool flag 时默认值正确
func TestIntegration_DefaultValues(t *testing.T) {
	c := parseArgs(t, []string{})
	if !bool(c.Number) {
		t.Error("期望 Number 默认为 true")
	}
	if !bool(c.Lower) {
		t.Error("期望 Lower 默认为 true")
	}
	if !bool(c.Upper) {
		t.Error("期望 Upper 默认为 true")
	}
	if bool(c.Special) {
		t.Error("期望 Special 默认为 false")
	}
	if bool(c.SpecialSafe) {
		t.Error("期望 SpecialSafe 默认为 false")
	}
}

// TestIntegration_SpecialSafe 测试 --special-safe 单独启用时字符集包含安全特殊符号
func TestIntegration_SpecialSafe(t *testing.T) {
	c := parseArgs(t, []string{"--special-safe", "--count", "3"})
	if !bool(c.SpecialSafe) {
		t.Error("期望 SpecialSafe=true，但得到 false")
	}
	if bool(c.Special) {
		t.Error("期望 Special=false，但得到 true")
	}
	if c.Count != 3 {
		t.Errorf("期望 Count=3，但得到 %d（--count 被意外丢弃）", c.Count)
	}
}

// TestIntegration_SpecialSafeFalse 测试 --special-safe false 显式关闭
func TestIntegration_SpecialSafeFalse(t *testing.T) {
	c := parseArgs(t, []string{"--special-safe", "false"})
	if bool(c.SpecialSafe) {
		t.Error("期望 SpecialSafe=false，但得到 true")
	}
}

// ---- buildCharset 单元测试 ----

// TestBuildCharset_SpecialSafe 测试 buildCharset 在 SpecialSafe 启用时输出正确字符集
func TestBuildCharset_SpecialSafe(t *testing.T) {
	withCLI(t, func() {
		cli.Number = false
		cli.Lower = false
		cli.Upper = false
		cli.Special = false
		cli.SpecialSafe = true

		got, err := buildCharset()
		if err != nil {
			t.Fatalf("不期望出错：%v", err)
		}
		if got != charSpecialSafe {
			t.Errorf("期望字符集 %q，但得到 %q", charSpecialSafe, got)
		}
	})
}

// TestBuildCharset_SpecialSafeNotIncludedByDefault 测试 SpecialSafe=false 时字符集为空
func TestBuildCharset_SpecialSafeNotIncludedByDefault(t *testing.T) {
	withCLI(t, func() {
		cli.Number = false
		cli.Lower = false
		cli.Upper = false
		cli.Special = false
		cli.SpecialSafe = false

		got, err := buildCharset()
		if err != nil {
			t.Fatalf("不期望出错：%v", err)
		}
		if got != "" {
			t.Errorf("期望空字符集，但得到 %q", got)
		}
	})
}

// TestBuildCharset_MutualExclusion 验证 --special 与 --special-safe 同时启用时 buildCharset 返回错误
func TestBuildCharset_MutualExclusion(t *testing.T) {
	withCLI(t, func() {
		cli.Special = true
		cli.SpecialSafe = true

		_, err := buildCharset()
		if err == nil {
			t.Error("期望返回互斥错误，但未出错")
		}
	})
}

// TestBuildCharset_CharSpecialSafeDashFirst 验证 charSpecialSafe 中 '-' 位于首位
func TestBuildCharset_CharSpecialSafeDashFirst(t *testing.T) {
	if len(charSpecialSafe) == 0 || charSpecialSafe[0] != '-' {
		t.Errorf("期望 charSpecialSafe 首字符为 '-'，实际为 %q", string(charSpecialSafe[0]))
	}
}

// ---- generatePassword 单元测试 ----

// TestGeneratePassword_EmptyCharset 验证空字符集时 generatePassword 返回错误
func TestGeneratePassword_EmptyCharset(t *testing.T) {
	_, err := generatePassword("", 10)
	if err == nil {
		t.Error("期望空字符集返回错误，但未出错")
	}
}

// TestGeneratePassword_CharsFromSpecialSafeOnly 验证启用 --special-safe 时生成密码的所有字符均来自 charSpecialSafe
func TestGeneratePassword_CharsFromSpecialSafeOnly(t *testing.T) {
	const rounds = 200
	for i := 0; i < rounds; i++ {
		pwd, err := generatePassword(charSpecialSafe, 20)
		if err != nil {
			t.Fatalf("生成密码失败：%v", err)
		}
		for _, ch := range pwd {
			if !strings.ContainsRune(charSpecialSafe, ch) {
				t.Errorf("密码 %q 包含不在 charSpecialSafe 中的字符 %q", pwd, ch)
			}
		}
	}
}
