package password

import (
	"math"
	"strings"
	"testing"
)

// ---- CalcEntropy 单元测试 ----

// TestCalcEntropy_Empty 空密码熵值为 0
func TestCalcEntropy_Empty(t *testing.T) {
	if e := CalcEntropy(""); e != 0 {
		t.Errorf("期望 0，但得到 %.2f", e)
	}
}

// TestCalcEntropy_NumbersOnly 纯数字：字符集 10，16 位 → 16*log2(10) ≈ 53.15 bits
func TestCalcEntropy_NumbersOnly(t *testing.T) {
	e := CalcEntropy("1234567890123456")
	expected := 16.0 * 3.321928 // log2(10) ≈ 3.3219
	if e < expected*0.99 || e > expected*1.01 {
		t.Errorf("期望约 %.2f bits，但得到 %.2f bits", expected, e)
	}
}

// TestCalcEntropy_MixedCharsets 混合字符集（数字+小写+大写=62），20 位 → 约 119 bits
func TestCalcEntropy_MixedCharsets(t *testing.T) {
	e := CalcEntropy("aB3xY7mN2kQ9pL1rT5vW")
	// 62 个字符，log2(62) ≈ 5.954，20 位 → 119.07 bits
	if e < 100 || e > 140 {
		t.Errorf("混合字符集熵值超出预期范围 [100, 140]，实际：%.2f bits", e)
	}
}

// TestCalcEntropy_LevelThresholds 验证各等级分界正确
func TestCalcEntropy_LevelThresholds(t *testing.T) {
	// 纯数字短密码（低熵）
	lowEntropy := CalcEntropy("12345")
	if lowEntropy >= 40 {
		t.Errorf("期望低于 40 bits，但得到 %.2f bits", lowEntropy)
	}

	// 混合字符集长密码（高熵）
	highEntropy := CalcEntropy("aB3xY7mN2kQ9pL1rT5vW")
	if highEntropy < 80 {
		t.Errorf("期望大于 80 bits，但得到 %.2f bits", highEntropy)
	}
}

// TestCalcEntropy_SpecialSafeSubset CharSpecialSafe 的熵应按 10 个字符计算，而非 CharSpecial 的 33 个
func TestCalcEntropy_SpecialSafeSubset(t *testing.T) {
	// 密码仅含 CharSpecialSafe 中的字符（-@#），不含 CharSpecial 独有字符
	// 字符集：CharSpecialSafe(10) → entropy = 3 * log2(10) ≈ 9.97 bits
	e := CalcEntropy("-@#")
	expected := 3.0 * math.Log2(10)
	if e < expected*0.99 || e > expected*1.01 {
		t.Errorf("CharSpecialSafe 熵期望约 %.2f bits，但得到 %.2f bits（可能误用了 CharSpecial 大小）", expected, e)
	}
}

// TestCalcEntropy_FullSpecialLarger 含有 CharSpecial 独有字符（非 CharSpecialSafe）时应用完整字符集大小
func TestCalcEntropy_FullSpecialLarger(t *testing.T) {
	// '!' 在 CharSpecial 中但不在 CharSpecialSafe 中，应触发完整字符集（33 个）
	eFull := CalcEntropy("!")
	eSafe := CalcEntropy("-") // '-' 在 CharSpecialSafe 中
	if eFull <= eSafe {
		t.Errorf("完整特殊字符集的熵（%.2f）应大于安全子集熵（%.2f）", eFull, eSafe)
	}
}

// ---- hasSequential 单元测试 ----

// TestHasSequential_AscendingLetters 升序字母序列应返回 true
func TestHasSequential_AscendingLetters(t *testing.T) {
	cases := []string{"abc", "bcd", "xyz", "XYZ", "MNO"}
	for _, pwd := range cases {
		if !hasSequential(pwd) {
			t.Errorf("期望 %q 检测到连续字符，但未检测到", pwd)
		}
	}
}

// TestHasSequential_AscendingDigits 升序数字序列应返回 true
func TestHasSequential_AscendingDigits(t *testing.T) {
	cases := []string{"123", "234", "789"}
	for _, pwd := range cases {
		if !hasSequential(pwd) {
			t.Errorf("期望 %q 检测到连续字符，但未检测到", pwd)
		}
	}
}

// TestHasSequential_Descending 降序序列应返回 true
func TestHasSequential_Descending(t *testing.T) {
	cases := []string{"cba", "zyx", "321"}
	for _, pwd := range cases {
		if !hasSequential(pwd) {
			t.Errorf("期望 %q 检测到降序连续字符，但未检测到", pwd)
		}
	}
}

// TestHasSequential_NoSequential 无连续序列时返回 false
func TestHasSequential_NoSequential(t *testing.T) {
	cases := []string{"ace", "bdf", "acb", "a1b", "xB3", "xB3kQ"}
	for _, pwd := range cases {
		if hasSequential(pwd) {
			t.Errorf("期望 %q 未检测到连续字符，但检测到了", pwd)
		}
	}
}

// TestHasSequential_PunctuationNotSequential 标点符号的 ASCII 相邻序列不应被视为连续
// $%& (36,37,38)、:;< (58,59,60)、<=> (60,61,62) 不具备可猜测性，不应触发
func TestHasSequential_PunctuationNotSequential(t *testing.T) {
	cases := []string{"$%&", ":;<", "<=>", ">?@"}
	for _, pwd := range cases {
		if hasSequential(pwd) {
			t.Errorf("期望标点序列 %q 不被视为连续字符，但被误判", pwd)
		}
	}
}

// TestHasSequential_EmbeddedSequence 连续序列嵌入在较长密码中也应检测到
func TestHasSequential_EmbeddedSequence(t *testing.T) {
	if !hasSequential("x8abcY7") {
		t.Error("期望在 x8abcY7 中检测到 abc 序列")
	}
	if !hasSequential("p@123word") {
		t.Error("期望在 p@123word 中检测到 123 序列")
	}
}

// TestHasSequential_LengthLessThan3 长度小于 3 时不可能有连续序列
func TestHasSequential_LengthLessThan3(t *testing.T) {
	if hasSequential("ab") {
		t.Error("长度 2 不应检测到连续序列")
	}
	if hasSequential("a") {
		t.Error("长度 1 不应检测到连续序列")
	}
}

// ---- hasRepeated 单元测试 ----

// TestHasRepeated_ThreeSame 3 个相同字符应返回 true
func TestHasRepeated_ThreeSame(t *testing.T) {
	cases := []string{"aaa", "111", "ZZZ", "!!!"}
	for _, pwd := range cases {
		if !hasRepeated(pwd) {
			t.Errorf("期望 %q 检测到重复字符，但未检测到", pwd)
		}
	}
}

// TestHasRepeated_EmbeddedRepeat 重复字符嵌入较长密码中也应检测到
func TestHasRepeated_EmbeddedRepeat(t *testing.T) {
	if !hasRepeated("xaaaBcd") {
		t.Error("期望在 xaaaBcd 中检测到 aaa 序列")
	}
}

// TestHasRepeated_TwoSame 仅 2 个相同字符不应触发（不足 3 个）
func TestHasRepeated_TwoSame(t *testing.T) {
	cases := []string{"aa", "11", "aab", "aa1"}
	for _, pwd := range cases {
		if hasRepeated(pwd) {
			t.Errorf("期望 %q 未检测到重复字符，但检测到了", pwd)
		}
	}
}

// TestHasRepeated_NoRepeat 无重复时返回 false
func TestHasRepeated_NoRepeat(t *testing.T) {
	if hasRepeated("aAbBcC") {
		t.Error("期望 aAbBcC 未检测到重复字符")
	}
}

// ---- isCommonPassword 单元测试 ----

// TestIsCommonPassword_KnownWeak 已知弱密码应返回 true
func TestIsCommonPassword_KnownWeak(t *testing.T) {
	cases := []string{"123456", "password", "qwerty", "letmein", "admin"}
	for _, pwd := range cases {
		if !isCommonPassword(pwd) {
			t.Errorf("期望 %q 被识别为弱密码，但未识别", pwd)
		}
	}
}

// TestIsCommonPassword_CaseInsensitive 大小写不敏感匹配
func TestIsCommonPassword_CaseInsensitive(t *testing.T) {
	cases := []string{"PASSWORD", "Password", "QWERTY", "Admin"}
	for _, pwd := range cases {
		if !isCommonPassword(pwd) {
			t.Errorf("期望 %q（大写变体）被识别为弱密码，但未识别", pwd)
		}
	}
}

// TestIsCommonPassword_RandomStrong 随机强密码不应在字典中
func TestIsCommonPassword_RandomStrong(t *testing.T) {
	cases := []string{"xK3#mP9vQz2!", "Tr7@bW4nYs8!", "mN5!kQ2xV9p#"}
	for _, pwd := range cases {
		if isCommonPassword(pwd) {
			t.Errorf("期望 %q 不在弱密码字典中，但被误判", pwd)
		}
	}
}

// ---- Assess 综合测试 ----

// TestAssess_TooShort 长度不足应报告问题
func TestAssess_TooShort(t *testing.T) {
	result := Assess("Ab1#")
	if !containsIssue(result.Issues, "长度") {
		t.Errorf("期望报告长度问题，实际问题：%v", result.Issues)
	}
}

// TestAssess_NoDigit 缺少数字应报告问题
func TestAssess_NoDigit(t *testing.T) {
	result := Assess("abcdefGHIJKLMN")
	if !containsIssue(result.Issues, "数字") {
		t.Errorf("期望报告缺少数字，实际问题：%v", result.Issues)
	}
}

// TestAssess_NoLower 缺少小写字母应报告问题
func TestAssess_NoLower(t *testing.T) {
	result := Assess("ABCDEFGH12345678")
	if !containsIssue(result.Issues, "小写") {
		t.Errorf("期望报告缺少小写字母，实际问题：%v", result.Issues)
	}
}

// TestAssess_NoUpper 缺少大写字母应报告问题
func TestAssess_NoUpper(t *testing.T) {
	result := Assess("abcdefgh12345678")
	if !containsIssue(result.Issues, "大写") {
		t.Errorf("期望报告缺少大写字母，实际问题：%v", result.Issues)
	}
}

// TestAssess_Sequential 包含连续字符应报告问题
func TestAssess_Sequential(t *testing.T) {
	result := Assess("xYz1abcK9mP3nQ")
	if !containsIssue(result.Issues, "连续") {
		t.Errorf("期望报告连续字符问题，实际问题：%v", result.Issues)
	}
}

// TestAssess_Repeated 包含重复字符应报告问题
func TestAssess_Repeated(t *testing.T) {
	result := Assess("xYzaaaBcD1m2n3")
	if !containsIssue(result.Issues, "重复") {
		t.Errorf("期望报告重复字符问题，实际问题：%v", result.Issues)
	}
}

// TestAssess_CommonPassword 弱密码字典命中应报告问题
func TestAssess_CommonPassword(t *testing.T) {
	result := Assess("password")
	if !containsIssue(result.Issues, "弱密码") {
		t.Errorf("期望报告弱密码字典命中，实际问题：%v", result.Issues)
	}
}

// TestAssess_StrongPassword 真正的强密码不应有问题
func TestAssess_StrongPassword(t *testing.T) {
	result := Assess("xK3mP9vQz2bN7rL4wT8")
	if len(result.Issues) != 0 {
		t.Errorf("期望强密码无问题，实际问题：%v", result.Issues)
	}
	if result.Level < LevelStrong {
		t.Errorf("期望强度 ≥ LevelStrong，实际等级：%d", result.Level)
	}
}

// TestAssess_EntropyLevel 验证信息熵等级分配正确
func TestAssess_EntropyLevel(t *testing.T) {
	result := Assess("xK3mP9vQz2bN7rL4wT8j")
	if result.Level != LevelVeryStrong {
		t.Errorf("期望 LevelVeryStrong，实际：%d，熵值：%.2f", result.Level, result.Entropy)
	}
}

// ---- generateStrong 单元测试 ----

// TestGenerateStrong_EmptyCharsets 空字符集列表应返回错误
func TestGenerateStrong_EmptyCharsets(t *testing.T) {
	_, err := generateStrong([]string{}, 10)
	if err == nil {
		t.Error("期望空字符集返回错误，但未出错")
	}
}

// TestGenerateStrong_LengthSmallerThanCharsetCount 长度小于字符集数量应返回错误
func TestGenerateStrong_LengthSmallerThanCharsetCount(t *testing.T) {
	charsets := []string{CharNumbers, CharLower, CharUpper}
	_, err := generateStrong(charsets, 2)
	if err == nil {
		t.Error("期望长度不足时返回错误，但未出错")
	}
}

// TestGenerateStrong_ContainsAllCharsets 200 轮验证：每次生成的密码必须包含所有子集的字符
func TestGenerateStrong_ContainsAllCharsets(t *testing.T) {
	charsets := []string{CharNumbers, CharLower, CharUpper}
	const rounds = 200
	for round := 0; round < rounds; round++ {
		pwd, err := generateStrong(charsets, 20)
		if err != nil {
			t.Fatalf("第 %d 轮生成失败：%v", round, err)
		}
		for _, cs := range charsets {
			found := false
			for _, ch := range pwd {
				if strings.ContainsRune(cs, ch) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("第 %d 轮：密码 %q 未包含字符集 %q 中的任何字符", round, pwd, cs)
			}
		}
	}
}

// TestGenerateStrong_AllCharsInFullCharset 生成的密码中所有字符都属于合并字符集
func TestGenerateStrong_AllCharsInFullCharset(t *testing.T) {
	charsets := []string{CharNumbers, CharLower, CharUpper, CharSpecialSafe}
	fullCharset := strings.Join(charsets, "")
	const rounds = 100
	for round := 0; round < rounds; round++ {
		pwd, err := generateStrong(charsets, 20)
		if err != nil {
			t.Fatalf("第 %d 轮生成失败：%v", round, err)
		}
		for _, ch := range pwd {
			if !strings.ContainsRune(fullCharset, ch) {
				t.Errorf("第 %d 轮：密码 %q 包含不在合并字符集中的字符 %q", round, pwd, ch)
			}
		}
	}
}

// TestGenerateStrong_CorrectLength 生成密码长度必须与请求一致
func TestGenerateStrong_CorrectLength(t *testing.T) {
	charsets := []string{CharNumbers, CharLower, CharUpper}
	for _, length := range []int{8, 12, 16, 20, 32} {
		pwd, err := generateStrong(charsets, length)
		if err != nil {
			t.Fatalf("length=%d：生成失败：%v", length, err)
		}
		if len([]rune(pwd)) != length {
			t.Errorf("length=%d：期望密码长度 %d，实际 %d", length, length, len([]rune(pwd)))
		}
	}
}

// TestGenerateStrong_LengthEqualsCharsetCount 长度恰好等于字符集数量（边界情况）
func TestGenerateStrong_LengthEqualsCharsetCount(t *testing.T) {
	charsets := []string{CharNumbers, CharLower, CharUpper}
	const rounds = 200
	for round := 0; round < rounds; round++ {
		pwd, err := generateStrong(charsets, 3)
		if err != nil {
			t.Fatalf("第 %d 轮生成失败：%v", round, err)
		}
		if len([]rune(pwd)) != 3 {
			t.Errorf("期望长度 3，实际 %d", len([]rune(pwd)))
		}
		for _, cs := range charsets {
			found := false
			for _, ch := range pwd {
				if strings.ContainsRune(cs, ch) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("第 %d 轮：密码 %q 未包含字符集 %q 中的字符", round, pwd, cs)
			}
		}
	}
}

// ---- assessForStrong 单元测试 ----

// TestAssessForStrong_PassesWithTwoCharsets 只传两个子集时不应报告缺少其他字符集
func TestAssessForStrong_PassesWithTwoCharsets(t *testing.T) {
	// 生成一个纯数字+小写的密码，然后用匹配的 charsets 校验
	charsets := []string{CharNumbers, CharLower}
	// 构造一个足够长、无连续/重复、不在弱密码字典中的测试密码
	pwd := "a1b2c3d4e5f6g7h8"
	issues := assessForStrong(pwd, charsets)
	// 不应报告"缺少大写字母"，因为 charsets 中没有要求大写
	for _, issue := range issues {
		if strings.Contains(issue, "大写") {
			t.Errorf("不应报告缺少大写字母，实际问题：%v", issues)
		}
	}
}

// TestAssessForStrong_ReportsShortPassword 长度不足应报告问题
func TestAssessForStrong_ReportsShortPassword(t *testing.T) {
	charsets := []string{CharNumbers, CharLower, CharUpper}
	issues := assessForStrong("Ab1", charsets)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue, "长度") {
			found = true
		}
	}
	if !found {
		t.Errorf("期望报告长度问题，实际问题：%v", issues)
	}
}

// TestAssessForStrong_ReportsMissingCharset 密码缺少某个子集字符时应报告
func TestAssessForStrong_ReportsMissingCharset(t *testing.T) {
	// 密码仅含数字和小写，但 charsets 要求大写
	charsets := []string{CharNumbers, CharLower, CharUpper}
	pwd := "a1b2c3d4e5f6g7h8" // 无大写
	issues := assessForStrong(pwd, charsets)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue, "字符集") {
			found = true
		}
	}
	if !found {
		t.Errorf("期望报告缺少字符集，实际问题：%v", issues)
	}
}

// TestAssessForStrong_CleanPasswordPassesAllChecks 质量良好的密码应通过全部检查
func TestAssessForStrong_CleanPasswordPassesAllChecks(t *testing.T) {
	charsets := []string{CharNumbers, CharLower, CharUpper}
	pwd := "xK3mP9vQz2bN7rL4wT8j"
	issues := assessForStrong(pwd, charsets)
	if len(issues) != 0 {
		t.Errorf("期望无问题，实际问题：%v", issues)
	}
}

// containsIssue 检查 issues 列表中是否存在包含指定关键词的问题描述
func containsIssue(issues []string, keyword string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, keyword) {
			return true
		}
	}
	return false
}
