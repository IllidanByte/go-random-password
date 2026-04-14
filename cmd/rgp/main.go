package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/IllidanByte/go-random-password/password"
	"github.com/alecthomas/kong"
)

// version 由构建时通过 -ldflags "-X main.version=xxx" 注入
var version = "dev"

// boolValue 支持 --flag true / --flag false 写法（含空格）
// 同时保持 --flag（不传值）等同于 --flag true 的行为
type boolValue bool

// Decode 实现 kong 的自定义解码接口
func (b *boolValue) Decode(ctx *kong.DecodeContext) error {
	// 下一个 token 不是值类型（是另一个 flag 或 EOL），默认置为 true，不消耗任何 token
	if !ctx.Scan.Peek().IsValue() {
		*b = true
		return nil
	}
	token, err := ctx.Scan.PopValue("bool")
	if err != nil {
		return fmt.Errorf("读取 bool 值失败：%w", err)
	}
	switch strings.ToLower(fmt.Sprintf("%v", token.Value)) {
	case "true", "1":
		*b = true
	case "false", "0":
		*b = false
	default:
		return fmt.Errorf("无效的 bool 值 %q，期望 true 或 false", token.Value)
	}
	return nil
}

// GenCmd gen 子命令：普通随机密码生成
type GenCmd struct {
	Length      int       `short:"l" name:"length"       default:"20"   help:"密码长度（默认 20 位）"`
	Count       int       `short:"c" name:"count"        default:"1"    help:"生成密码数量（默认 1 个）"`
	Number      boolValue `name:"number"       default:"true"  help:"是否包含数字（默认启用）"`
	Lower       boolValue `name:"lower"        default:"true"  help:"是否包含小写字母（默认启用）"`
	Upper       boolValue `name:"upper"        default:"true"  help:"是否包含大写字母（默认启用）"`
	Special     boolValue `name:"special"      default:"false" help:"是否包含特殊符号，与 --special-safe 互斥（默认禁用）"`
	SpecialSafe boolValue `name:"special-safe" default:"false" help:"是否包含安全特殊符号，与 --special 互斥（默认禁用）"`
}

// Run 执行 gen 子命令
func (cmd *GenCmd) Run() error {
	if cmd.Length < 1 {
		return fmt.Errorf("密码长度必须大于 0")
	}
	if cmd.Count < 1 {
		return fmt.Errorf("生成数量必须大于 0")
	}

	passwords, err := password.GenerateN(cmd.Length, cmd.Count, password.GenConfig{
		Number:      bool(cmd.Number),
		Lower:       bool(cmd.Lower),
		Upper:       bool(cmd.Upper),
		Special:     bool(cmd.Special),
		SpecialSafe: bool(cmd.SpecialSafe),
	})
	if err != nil {
		return fmt.Errorf("生成密码失败：%w", err)
	}
	for _, pwd := range passwords {
		fmt.Println(pwd)
	}
	return nil
}

// StrongCmd strong 子命令：强密码生成
// 库层固定启用数字 + 小写字母 + 大写字母，可通过 --special / --special-safe 追加特殊字符集
type StrongCmd struct {
	Length      int       `short:"l" name:"length"       default:"20"   help:"密码长度（默认 20 位，最小 8 位）"`
	Count       int       `short:"c" name:"count"        default:"1"    help:"生成密码数量（默认 1 个）"`
	Special     boolValue `name:"special"      default:"false" help:"追加特殊符号字符集，与 --special-safe 互斥（默认禁用）"`
	SpecialSafe boolValue `name:"special-safe" default:"false" help:"追加安全特殊符号字符集，与 --special 互斥（默认禁用）"`
}

// Run 执行 strong 子命令
func (cmd *StrongCmd) Run() error {
	if cmd.Special && cmd.SpecialSafe {
		return fmt.Errorf("--special 与 --special-safe 不能同时启用")
	}
	if cmd.Length < 8 {
		return fmt.Errorf("强密码模式要求最小长度 8 位")
	}
	if cmd.Count < 1 {
		return fmt.Errorf("生成数量必须大于 0")
	}

	passwords, err := password.GenerateStrongN(cmd.Length, cmd.Count, password.StrongConfig{
		Special:     bool(cmd.Special),
		SpecialSafe: bool(cmd.SpecialSafe),
	})
	if err != nil {
		return err
	}
	for _, pwd := range passwords {
		fmt.Println(pwd)
	}
	return nil
}

// CLI 顶层命令行结构
var cli struct {
	Version kong.VersionFlag `name:"version" short:"v" help:"打印版本号并退出"`
	Gen     GenCmd           `cmd:"gen" help:"生成随机密码（普通模式）"`
	Strong  StrongCmd        `cmd:"strong" help:"生成强密码（强密码模式，固定包含数字 + 大小写字母）"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("rgp"),
		kong.Description("随机密码生成工具"),
		kong.Vars{"version": version},
	)
	if err := ctx.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "错误："+err.Error())
		os.Exit(1)
	}
}
