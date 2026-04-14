package password

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
)

// StrengthLevel 密码强度等级
type StrengthLevel int

const (
	LevelWeak       StrengthLevel = iota // 弱：< 40 bits
	LevelFair                            // 一般：40–60 bits
	LevelStrong                          // 强：60–80 bits
	LevelVeryStrong                      // 极强：≥ 80 bits
)

// StrengthResult 密码强度评估结果
type StrengthResult struct {
	Entropy float64
	Level   StrengthLevel
	Issues  []string // 未通过的检查项，为空表示通过所有检查
}

// Assess 综合评估密码强度，返回完整结果。
func Assess(pwd string) StrengthResult {
	result := StrengthResult{}

	// 1. 检查长度（按字符数计算，与熵计算和连续/重复检测保持一致）
	if len([]rune(pwd)) < 8 {
		result.Issues = append(result.Issues, "密码长度不足 8 位")
	}

	// 2. 检查字符集覆盖（不强制要求特殊字符）
	if !strings.ContainsAny(pwd, CharNumbers) {
		result.Issues = append(result.Issues, "缺少数字")
	}
	if !strings.ContainsAny(pwd, CharLower) {
		result.Issues = append(result.Issues, "缺少小写字母")
	}
	if !strings.ContainsAny(pwd, CharUpper) {
		result.Issues = append(result.Issues, "缺少大写字母")
	}

	// 3. 检查弱密码字典
	if isCommonPassword(pwd) {
		result.Issues = append(result.Issues, "密码在常见弱密码列表中")
	}

	// 4. 检测连续字符序列（升序或降序，3+ 个）
	if hasSequential(pwd) {
		result.Issues = append(result.Issues, "包含连续字符序列")
	}

	// 5. 检测重复字符（3+ 个相同连续字符）
	if hasRepeated(pwd) {
		result.Issues = append(result.Issues, "包含重复字符模式")
	}

	// 6. 计算信息熵并给出评级
	result.Entropy = CalcEntropy(pwd)
	switch {
	case result.Entropy >= 80:
		result.Level = LevelVeryStrong
	case result.Entropy >= 60:
		result.Level = LevelStrong
	case result.Entropy >= 40:
		result.Level = LevelFair
	default:
		result.Level = LevelWeak
	}
	if result.Entropy < 60 {
		result.Issues = append(result.Issues, fmt.Sprintf("信息熵不足（%.1f bits），需 ≥ 60 bits", result.Entropy))
	}

	return result
}

// CalcEntropy 根据密码中出现的字符类别推算信息熵（bits）。
// 算法：统计实际使用的字符集总大小，entropy = len(pwd) * log2(charsetSize)。
// 特殊字符区分两种情况：
//   - 密码中含有 CharSpecial 里不属于 CharSpecialSafe 的字符 → 字符集大小用完整 CharSpecial
//   - 密码中特殊字符全部属于 CharSpecialSafe → 字符集大小仅用 CharSpecialSafe
func CalcEntropy(pwd string) float64 {
	if len(pwd) == 0 {
		return 0
	}

	charsetSize := 0
	if strings.ContainsAny(pwd, CharNumbers) {
		charsetSize += 10
	}
	if strings.ContainsAny(pwd, CharLower) {
		charsetSize += 26
	}
	if strings.ContainsAny(pwd, CharUpper) {
		charsetSize += 26
	}

	// 判断密码中是否含有超出 CharSpecialSafe 范围的特殊字符
	hasFullSpecial := false
	hasSafeSpecial := false
	for _, ch := range pwd {
		if strings.ContainsRune(CharSpecial, ch) {
			if strings.ContainsRune(CharSpecialSafe, ch) {
				hasSafeSpecial = true
			} else {
				hasFullSpecial = true
				break // 发现完整字符集中的字符，无需继续
			}
		}
	}
	if hasFullSpecial {
		charsetSize += len(CharSpecial) // 完整特殊字符集：33 个
	} else if hasSafeSpecial {
		charsetSize += len(CharSpecialSafe) // 安全子集：10 个
	}

	if charsetSize == 0 {
		return 0
	}
	return float64(len([]rune(pwd))) * math.Log2(float64(charsetSize))
}

// hasSequential 检测密码中是否存在 3+ 个字母或数字的连续序列（升序或降序）。
// 仅检测有实际安全语义的字符类：a–z、A–Z、0–9。
// 标点符号的 ASCII 相邻（如 $%&、:;<）不具备可猜测性，不纳入检测范围。
func hasSequential(pwd string) bool {
	isAlphanumeric := func(r rune) bool {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
	}

	runes := []rune(pwd)
	for i := 0; i+2 < len(runes); i++ {
		a, b, c := runes[i], runes[i+1], runes[i+2]
		// 三个字符都必须是字母或数字，才做连续性判断
		if !isAlphanumeric(a) || !isAlphanumeric(b) || !isAlphanumeric(c) {
			continue
		}
		// 升序连续
		if b == a+1 && c == a+2 {
			return true
		}
		// 降序连续
		if b == a-1 && c == a-2 {
			return true
		}
	}
	return false
}

// hasRepeated 检测密码中是否存在 3+ 个相同的连续字符。
// 例如：aaa、111、ZZZ 等。
func hasRepeated(pwd string) bool {
	runes := []rune(pwd)
	for i := 0; i+2 < len(runes); i++ {
		if runes[i] == runes[i+1] && runes[i+1] == runes[i+2] {
			return true
		}
	}
	return false
}

// isCommonPassword 检查密码是否在弱密码字典中（大小写不敏感）。
func isCommonPassword(pwd string) bool {
	_, found := weakPasswords[strings.ToLower(pwd)]
	return found
}

// assessForStrong 校验密码是否满足强密码生成要求，检查规则与 charsets 动态绑定（包私有）。
// 与 Assess 不同：不硬编码"必须含数字/小写/大写"，而是检查 charsets 中每个子集都有字符出现；
// 也不将熵阈值作为失败条件（熵预校验在 GenerateStrong 入口处已做）。
func assessForStrong(pwd string, charsets []string) []string {
	var issues []string
	if len([]rune(pwd)) < 8 {
		issues = append(issues, "密码长度不足 8 位")
	}
	// 检查每个子集在密码中至少出现一个字符
	for _, cs := range charsets {
		if !strings.ContainsAny(pwd, cs) {
			issues = append(issues, "缺少指定字符集中的字符")
		}
	}
	if isCommonPassword(pwd) {
		issues = append(issues, "密码在常见弱密码列表中")
	}
	if hasSequential(pwd) {
		issues = append(issues, "包含连续字符序列")
	}
	if hasRepeated(pwd) {
		issues = append(issues, "包含重复字符模式")
	}
	return issues
}

// generateStrong 从指定子集列表生成密码，使用 rune 安全处理 UTF-8（包私有）。
// 保证每个子集各出现至少一个字符，其余位置从合并字符集随机填充，最终 Fisher-Yates shuffle。
func generateStrong(charsets []string, length int) (string, error) {
	if len(charsets) == 0 {
		return "", fmt.Errorf("字符集不能为空")
	}
	if length < len(charsets) {
		return "", fmt.Errorf("密码长度（%d）不能小于启用的字符集数量（%d）", length, len(charsets))
	}

	// 将各字符集预转为 rune 切片，支持多字节 UTF-8 字符
	charsetRunes := make([][]rune, len(charsets))
	for i, cs := range charsets {
		charsetRunes[i] = []rune(cs)
	}

	result := make([]rune, length)

	// 从每个子集各随机取 1 个字符，放入前 N 个位置
	for i, runes := range charsetRunes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(runes))))
		if err != nil {
			return "", err
		}
		result[i] = runes[n.Int64()]
	}

	// 构建合并字符集（rune 切片），填充剩余位置
	var fullRunes []rune
	for _, runes := range charsetRunes {
		fullRunes = append(fullRunes, runes...)
	}
	charsetSize := big.NewInt(int64(len(fullRunes)))
	for i := len(charsets); i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetSize)
		if err != nil {
			return "", err
		}
		result[i] = fullRunes[n.Int64()]
	}

	// Fisher-Yates shuffle（全程使用 crypto/rand 保证密码学安全）
	for i := length - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		result[i], result[j.Int64()] = result[j.Int64()], result[i]
	}

	return string(result), nil
}
