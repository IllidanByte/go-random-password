package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/alecthomas/kong"
)

// 各字符集常量
const (
	charNumbers     = "0123456789"
	charLower       = "abcdefghijklmnopqrstuvwxyz"
	charUpper       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charSpecial     = "`~!@#$%^&*()[{]}-_=+|;:'\",<.>/?"
	charSpecialSafe = "-@#%^_+=.,"
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

// CLI 命令行参数定义
var cli struct {
	Version     kong.VersionFlag `name:"version" short:"v" help:"打印版本号并退出"`
	Length      int              `short:"l" name:"length"  default:"20"   help:"密码长度（默认 20 位）"`
	Count       int              `short:"c" name:"count"   default:"1"    help:"生成密码数量（默认 1 个）"`
	Number      boolValue        `name:"number"  default:"true"  help:"是否包含数字（默认启用）"`
	Lower       boolValue        `name:"lower"   default:"true"  help:"是否包含小写字母（默认启用）"`
	Upper       boolValue        `name:"upper"   default:"true"  help:"是否包含大写字母（默认启用）"`
	Special     boolValue        `name:"special"      default:"false" help:"是否包含特殊符号，与 --special-safe 互斥（默认禁用）"`
	SpecialSafe boolValue        `name:"special-safe" default:"false" help:"是否包含安全特殊符号，与 --special 互斥（默认禁用）"`
}

func main() {
	kong.Parse(&cli,
		kong.Name("rgp"),
		kong.Description("随机密码生成工具"),
		kong.Vars{"version": version},
	)

	// 校验参数合法性
	if cli.Length < 1 {
		_, _ = fmt.Fprintln(os.Stderr, "错误：密码长度必须大于 0")
		os.Exit(1)
	}
	if cli.Count < 1 {
		_, _ = fmt.Fprintln(os.Stderr, "错误：生成数量必须大于 0")
		os.Exit(1)
	}

	// --special 与 --special-safe 互斥
	if cli.Special && cli.SpecialSafe {
		_, _ = fmt.Fprintln(os.Stderr, "错误：--special 与 --special-safe 不能同时启用")
		os.Exit(1)
	}

	// 组合启用的字符集
	charset, err := buildCharset()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(1)
	}
	if len(charset) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "错误：至少需要启用一种字符集")
		os.Exit(1)
	}

	// 生成并输出密码
	for i := 0; i < cli.Count; i++ {
		pwd, err := generatePassword(charset, cli.Length)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "生成密码失败：%v\n", err)
			os.Exit(1)
		}
		fmt.Println(pwd)
	}
}

// buildCharset 根据启用的选项组合字符集，--special 与 --special-safe 不能同时启用
func buildCharset() (string, error) {
	if cli.Special && cli.SpecialSafe {
		return "", fmt.Errorf("--special 与 --special-safe 不能同时启用")
	}
	var charset string
	if cli.Number {
		charset += charNumbers
	}
	if cli.Lower {
		charset += charLower
	}
	if cli.Upper {
		charset += charUpper
	}
	if cli.Special {
		charset += charSpecial
	}
	if cli.SpecialSafe {
		charset += charSpecialSafe
	}
	return charset, nil
}

// generatePassword 从指定字符集中随机生成指定长度的密码
func generatePassword(charset string, length int) (string, error) {
	if len(charset) == 0 {
		return "", fmt.Errorf("字符集不能为空")
	}
	result := make([]byte, length)
	charsetSize := big.NewInt(int64(len(charset)))
	for i := range result {
		n, err := rand.Int(rand.Reader, charsetSize)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
