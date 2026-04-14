package password

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
)

// 各字符集常量
const (
	CharNumbers     = "0123456789"
	CharLower       = "abcdefghijklmnopqrstuvwxyz"
	CharUpper       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharSpecial     = "`~!@#$%^&*()[{]}-_=+|;:'\",<.>/?"
	CharSpecialSafe = "-@#%^_+=.,"
)

// GenConfig 普通密码生成配置，只能选预定义字符集。
// Special 与 SpecialSafe 互斥，不可同时为 true。
type GenConfig struct {
	Number      bool
	Lower       bool
	Upper       bool
	Special     bool
	SpecialSafe bool
}

// StrongConfig 强密码生成配置。
// 数字 + 小写字母 + 大写字母固定启用，只需指定是否追加特殊字符集。
// Special 与 SpecialSafe 互斥，不可同时为 true。
type StrongConfig struct {
	Special     bool
	SpecialSafe bool
}

// Generate 使用 GenConfig 指定的预定义字符集生成一个随机密码。
func Generate(length int, cfg GenConfig) (string, error) {
	if cfg.Special && cfg.SpecialSafe {
		return "", fmt.Errorf("special 与 specialSafe 不能同时启用")
	}
	if length < 1 {
		return "", fmt.Errorf("密码长度必须大于 0")
	}
	charset := buildCharset(cfg.Number, cfg.Lower, cfg.Upper, cfg.Special, cfg.SpecialSafe)
	if len(charset) == 0 {
		return "", fmt.Errorf("至少需要启用一种字符集")
	}
	return generate(charset, length)
}

// GenerateN 使用 GenConfig 指定的预定义字符集生成 n 个随机密码。
func GenerateN(length, n int, cfg GenConfig) ([]string, error) {
	if n < 1 {
		return nil, fmt.Errorf("生成数量必须大于 0")
	}
	passwords := make([]string, 0, n)
	for i := 0; i < n; i++ {
		pwd, err := Generate(length, cfg)
		if err != nil {
			return nil, fmt.Errorf("生成第 %d 个密码失败：%w", i+1, err)
		}
		passwords = append(passwords, pwd)
	}
	return passwords, nil
}

// GenerateStrong 生成一个强密码。
// 固定使用数字 + 小写字母 + 大写字母三类基础字符集，
// 可通过 StrongConfig 追加特殊字符集。
// 生成的密码保证：每个字符集各出现至少一个字符、不在弱密码字典中、
// 无连续/重复字符序列、信息熵 ≥ 60 bits。
func GenerateStrong(length int, cfg StrongConfig) (string, error) {
	if cfg.Special && cfg.SpecialSafe {
		return "", fmt.Errorf("special 与 specialSafe 不能同时启用")
	}
	if length < 8 {
		return "", fmt.Errorf("强密码模式要求最小长度 8 位")
	}

	// 库层固定三类基础字符集，调用方只选是否追加特殊字符集
	charsets := []string{CharNumbers, CharLower, CharUpper}
	if cfg.Special {
		charsets = append(charsets, CharSpecial)
	}
	if cfg.SpecialSafe {
		charsets = append(charsets, CharSpecialSafe)
	}

	// 熵预校验：fail-fast，避免无意义重试
	totalSize := 0
	for _, cs := range charsets {
		totalSize += len([]rune(cs))
	}
	entropy := float64(length) * math.Log2(float64(totalSize))
	if entropy < 60 {
		minLen := int(math.Ceil(60 / math.Log2(float64(totalSize))))
		return "", fmt.Errorf("信息熵不足（%.1f bits），需 ≥ 60 bits，建议最小长度 %d 位", entropy, minLen)
	}

	const maxAttempts = 100
	for i := 0; i < maxAttempts; i++ {
		pwd, err := generateStrong(charsets, length)
		if err != nil {
			return "", fmt.Errorf("生成密码失败：%w", err)
		}
		if len(assessForStrong(pwd, charsets)) == 0 {
			return pwd, nil
		}
	}
	return "", fmt.Errorf("已重试 %d 次，无法生成满足强密码标准的密码，请检查参数", maxAttempts)
}

// GenerateStrongN 生成 n 个强密码，任意一次失败立即返回错误。
func GenerateStrongN(length, n int, cfg StrongConfig) ([]string, error) {
	if n < 1 {
		return nil, fmt.Errorf("生成数量必须大于 0")
	}
	passwords := make([]string, 0, n)
	for i := 0; i < n; i++ {
		pwd, err := GenerateStrong(length, cfg)
		if err != nil {
			return nil, fmt.Errorf("生成第 %d 个强密码失败：%w", i+1, err)
		}
		passwords = append(passwords, pwd)
	}
	return passwords, nil
}

// buildCharset 根据各字符集启用状态组合并返回字符集字符串（包私有）。
func buildCharset(number, lower, upper, special, specialSafe bool) string {
	var charset string
	if number {
		charset += CharNumbers
	}
	if lower {
		charset += CharLower
	}
	if upper {
		charset += CharUpper
	}
	if special {
		charset += CharSpecial
	}
	if specialSafe {
		charset += CharSpecialSafe
	}
	return charset
}

// generate 从字符集中随机生成指定长度的密码，使用 rune 安全处理 UTF-8（包私有）。
func generate(charset string, length int) (string, error) {
	runes := []rune(charset)
	charsetSize := big.NewInt(int64(len(runes)))
	result := make([]rune, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, charsetSize)
		if err != nil {
			return "", err
		}
		result[i] = runes[n.Int64()]
	}
	return string(result), nil
}
