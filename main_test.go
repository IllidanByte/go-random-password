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

// testGenCmd 是集成测试使用的参数结构体，镜像 GenCmd 的 flag 定义
type testGenCmd struct {
	Number      boolValue `name:"number"       default:"true"`
	Lower       boolValue `name:"lower"        default:"true"`
	Upper       boolValue `name:"upper"        default:"true"`
	Special     boolValue `name:"special"      default:"false"`
	SpecialSafe boolValue `name:"special-safe" default:"false"`
	Count       int       `name:"count"        default:"1"`
	Length      int       `name:"length"       default:"20"`
}

// testStrongCmd 是 StrongCmd 集成测试使用的参数结构体
type testStrongCmd struct {
	Special     boolValue `name:"special"      default:"false"`
	SpecialSafe boolValue `name:"special-safe" default:"false"`
	Count       int       `name:"count"        default:"1"`
	Length      int       `name:"length"       default:"20"`
}

// parseGenArgs 通过 kong 解析 gen 子命令参数，返回填充后的 testGenCmd
func parseGenArgs(t *testing.T, args []string) testGenCmd {
	t.Helper()
	var c testGenCmd
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

// parseStrongArgs 通过 kong 解析 strong 子命令参数，返回填充后的 testStrongCmd
func parseStrongArgs(t *testing.T, args []string) testStrongCmd {
	t.Helper()
	var c testStrongCmd
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

// parseCLI 使用与真实入口完全相同的顶层结构体（GenCmd + StrongCmd 子命令）解析参数
// 返回路由到的子命令名称（"gen" 或 "strong"）以及填充后的各子命令结构体
// 注意：此处故意省略 VersionFlag，以避免 --version 触发 os.Exit
func parseCLI(t *testing.T, args []string) (cmd string, gen GenCmd, strong StrongCmd) {
	t.Helper()
	var c struct {
		Gen    GenCmd    `cmd:"gen"`
		Strong StrongCmd `cmd:"strong"`
	}
	p, err := kong.New(&c, kong.Name("rgp"), kong.Vars{"version": "test"})
	if err != nil {
		t.Fatalf("构建顶层 parser 失败：%v", err)
	}
	ctx, err := p.Parse(args)
	if err != nil {
		t.Fatalf("解析参数失败（args=%v）：%v", args, err)
	}
	return ctx.Command(), c.Gen, c.Strong
}

// parseCLIExpectError 解析参数并期望返回错误，用于验证非法参数组合
func parseCLIExpectError(t *testing.T, args []string) error {
	t.Helper()
	var c struct {
		Gen    GenCmd    `cmd:"gen"`
		Strong StrongCmd `cmd:"strong"`
	}
	p, err := kong.New(&c, kong.Name("rgp"), kong.Vars{"version": "test"})
	if err != nil {
		t.Fatalf("构建顶层 parser 失败：%v", err)
	}
	_, err = p.Parse(args)
	return err
}

// ---- 顶层 CLI 路由集成测试 ----

// TestCLI_GenRouting 验证 rgp gen 正确路由到 gen 子命令
func TestCLI_GenRouting(t *testing.T) {
	cmd, _, _ := parseCLI(t, []string{"gen"})
	if cmd != "gen" {
		t.Errorf("期望路由到 gen，实际：%q", cmd)
	}
}

// TestCLI_GenDefaults 验证 gen 子命令默认值正确
func TestCLI_GenDefaults(t *testing.T) {
	_, gen, _ := parseCLI(t, []string{"gen"})
	if !bool(gen.Number) {
		t.Error("gen: 期望 Number 默认为 true")
	}
	if !bool(gen.Lower) {
		t.Error("gen: 期望 Lower 默认为 true")
	}
	if !bool(gen.Upper) {
		t.Error("gen: 期望 Upper 默认为 true")
	}
	if bool(gen.Special) {
		t.Error("gen: 期望 Special 默认为 false")
	}
	if bool(gen.SpecialSafe) {
		t.Error("gen: 期望 SpecialSafe 默认为 false")
	}
	if gen.Length != 20 {
		t.Errorf("gen: 期望 Length 默认为 20，实际：%d", gen.Length)
	}
	if gen.Count != 1 {
		t.Errorf("gen: 期望 Count 默认为 1，实际：%d", gen.Count)
	}
}

// TestCLI_GenWithFlags 验证 gen 子命令的 flag 能正确解析
func TestCLI_GenWithFlags(t *testing.T) {
	_, gen, _ := parseCLI(t, []string{"gen", "--special", "--count", "3", "--length", "16"})
	if !bool(gen.Special) {
		t.Error("gen: 期望 Special=true")
	}
	if gen.Count != 3 {
		t.Errorf("gen: 期望 Count=3，实际：%d", gen.Count)
	}
	if gen.Length != 16 {
		t.Errorf("gen: 期望 Length=16，实际：%d", gen.Length)
	}
}

// TestCLI_GenLowerFalse 验证 gen --lower false 能正确禁用小写字母
func TestCLI_GenLowerFalse(t *testing.T) {
	_, gen, _ := parseCLI(t, []string{"gen", "--lower", "false"})
	if bool(gen.Lower) {
		t.Error("gen: 期望 Lower=false，但得到 true")
	}
}

// TestCLI_StrongRouting 验证 rgp strong 正确路由到 strong 子命令
func TestCLI_StrongRouting(t *testing.T) {
	cmd, _, _ := parseCLI(t, []string{"strong"})
	if cmd != "strong" {
		t.Errorf("期望路由到 strong，实际：%q", cmd)
	}
}

// TestCLI_StrongDefaults 验证 strong 子命令默认值正确
func TestCLI_StrongDefaults(t *testing.T) {
	_, _, strong := parseCLI(t, []string{"strong"})
	if bool(strong.Special) {
		t.Error("strong: 期望 Special 默认为 false")
	}
	if bool(strong.SpecialSafe) {
		t.Error("strong: 期望 SpecialSafe 默认为 false")
	}
	if strong.Length != 20 {
		t.Errorf("strong: 期望 Length 默认为 20，实际：%d", strong.Length)
	}
	if strong.Count != 1 {
		t.Errorf("strong: 期望 Count 默认为 1，实际：%d", strong.Count)
	}
}

// TestCLI_StrongWithSpecial 验证 strong --special 能正确解析
func TestCLI_StrongWithSpecial(t *testing.T) {
	_, _, strong := parseCLI(t, []string{"strong", "--special", "--length", "16"})
	if !bool(strong.Special) {
		t.Error("strong: 期望 Special=true")
	}
	if strong.Length != 16 {
		t.Errorf("strong: 期望 Length=16，实际：%d", strong.Length)
	}
}

// TestCLI_StrongWithSpecialSafe 验证 strong --special-safe 能正确解析
func TestCLI_StrongWithSpecialSafe(t *testing.T) {
	_, _, strong := parseCLI(t, []string{"strong", "--special-safe"})
	if !bool(strong.SpecialSafe) {
		t.Error("strong: 期望 SpecialSafe=true")
	}
	if bool(strong.Special) {
		t.Error("strong: 期望 Special=false（互斥）")
	}
}

// TestCLI_InvalidSubcommand 验证非法子命令会返回解析错误
func TestCLI_InvalidSubcommand(t *testing.T) {
	err := parseCLIExpectError(t, []string{"unknown"})
	if err == nil {
		t.Error("期望非法子命令返回错误，但未出错")
	}
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

// ---- 集成测试：通过 kong.Parse 解析完整参数（gen 子命令）----

// TestGenIntegration_AdjacentFlag 核心回归测试：--special 后紧跟 --count 2 时两个参数均正确解析
func TestGenIntegration_AdjacentFlag(t *testing.T) {
	c := parseGenArgs(t, []string{"--special", "--count", "2"})
	if !bool(c.Special) {
		t.Error("期望 Special=true，但得到 false")
	}
	if c.Count != 2 {
		t.Errorf("期望 Count=2，但得到 %d（--count 被意外丢弃）", c.Count)
	}
}

// TestGenIntegration_ExplicitBool 测试显式赋值写法
func TestGenIntegration_ExplicitBool(t *testing.T) {
	c := parseGenArgs(t, []string{"--lower", "false", "--upper", "true"})
	if bool(c.Lower) {
		t.Error("期望 Lower=false，但得到 true")
	}
	if !bool(c.Upper) {
		t.Error("期望 Upper=true，但得到 false")
	}
}

// TestGenIntegration_DefaultValues 测试不传任何 bool flag 时默认值正确
func TestGenIntegration_DefaultValues(t *testing.T) {
	c := parseGenArgs(t, []string{})
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

// TestGenIntegration_SpecialSafe 测试 --special-safe 单独启用时字符集包含安全特殊符号
func TestGenIntegration_SpecialSafe(t *testing.T) {
	c := parseGenArgs(t, []string{"--special-safe", "--count", "3"})
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

// TestGenIntegration_SpecialSafeFalse 测试 --special-safe false 显式关闭
func TestGenIntegration_SpecialSafeFalse(t *testing.T) {
	c := parseGenArgs(t, []string{"--special-safe", "false"})
	if bool(c.SpecialSafe) {
		t.Error("期望 SpecialSafe=false，但得到 true")
	}
}

// ---- 集成测试：strong 子命令 ----

// TestStrongIntegration_DefaultValues 测试 strong 子命令默认值
func TestStrongIntegration_DefaultValues(t *testing.T) {
	c := parseStrongArgs(t, []string{})
	if bool(c.Special) {
		t.Error("期望 Special 默认为 false")
	}
	if bool(c.SpecialSafe) {
		t.Error("期望 SpecialSafe 默认为 false")
	}
	if c.Count != 1 {
		t.Errorf("期望 Count 默认为 1，但得到 %d", c.Count)
	}
	if c.Length != 20 {
		t.Errorf("期望 Length 默认为 20，但得到 %d", c.Length)
	}
}

// TestStrongIntegration_WithSpecial 测试 --special 参数解析
func TestStrongIntegration_WithSpecial(t *testing.T) {
	c := parseStrongArgs(t, []string{"--special", "--count", "3"})
	if !bool(c.Special) {
		t.Error("期望 Special=true，但得到 false")
	}
	if c.Count != 3 {
		t.Errorf("期望 Count=3，但得到 %d", c.Count)
	}
}

// TestStrongIntegration_WithSpecialSafe 测试 --special-safe 参数解析
func TestStrongIntegration_WithSpecialSafe(t *testing.T) {
	c := parseStrongArgs(t, []string{"--special-safe"})
	if !bool(c.SpecialSafe) {
		t.Error("期望 SpecialSafe=true，但得到 false")
	}
}

// ---- buildCharset 单元测试 ----

// TestBuildCharset_SpecialSafe 测试 buildCharset 在 SpecialSafe 启用时输出正确字符集
func TestBuildCharset_SpecialSafe(t *testing.T) {
	got, err := buildCharset(false, false, false, false, true)
	if err != nil {
		t.Fatalf("不期望出错：%v", err)
	}
	if got != charSpecialSafe {
		t.Errorf("期望字符集 %q，但得到 %q", charSpecialSafe, got)
	}
}

// TestBuildCharset_SpecialSafeNotIncludedByDefault 测试全部禁用时字符集为空
func TestBuildCharset_SpecialSafeNotIncludedByDefault(t *testing.T) {
	got, err := buildCharset(false, false, false, false, false)
	if err != nil {
		t.Fatalf("不期望出错：%v", err)
	}
	if got != "" {
		t.Errorf("期望空字符集，但得到 %q", got)
	}
}

// TestBuildCharset_MutualExclusion 验证 --special 与 --special-safe 同时启用时 buildCharset 返回错误
func TestBuildCharset_MutualExclusion(t *testing.T) {
	_, err := buildCharset(false, false, false, true, true)
	if err == nil {
		t.Error("期望返回互斥错误，但未出错")
	}
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
