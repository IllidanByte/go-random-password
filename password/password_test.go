package password

import (
	"strings"
	"testing"
)

// ---- buildCharset 单元测试（包私有白盒测试）----

// TestBuildCharset_SpecialSafe 测试 SpecialSafe 启用时输出正确字符集
func TestBuildCharset_SpecialSafe(t *testing.T) {
	got := buildCharset(false, false, false, false, true)
	if got != CharSpecialSafe {
		t.Errorf("期望字符集 %q，但得到 %q", CharSpecialSafe, got)
	}
}

// TestBuildCharset_AllDisabled 测试全部禁用时字符集为空
func TestBuildCharset_AllDisabled(t *testing.T) {
	got := buildCharset(false, false, false, false, false)
	if got != "" {
		t.Errorf("期望空字符集，但得到 %q", got)
	}
}

// TestBuildCharset_CharSpecialSafeDashFirst 验证 CharSpecialSafe 中 '-' 位于首位
func TestBuildCharset_CharSpecialSafeDashFirst(t *testing.T) {
	if len(CharSpecialSafe) == 0 || CharSpecialSafe[0] != '-' {
		t.Errorf("期望 CharSpecialSafe 首字符为 '-'，实际为 %q", string(CharSpecialSafe[0]))
	}
}

// ---- Generate 单元测试 ----

// TestGenerate_MutualExclusion 验证 Special 与 SpecialSafe 同时启用时返回错误
func TestGenerate_MutualExclusion(t *testing.T) {
	_, err := Generate(10, GenConfig{Special: true, SpecialSafe: true})
	if err == nil {
		t.Error("期望返回互斥错误，但未出错")
	}
}

// TestGenerate_NoCharset 验证未启用任何字符集时返回错误
func TestGenerate_NoCharset(t *testing.T) {
	_, err := Generate(10, GenConfig{})
	if err == nil {
		t.Error("期望返回空字符集错误，但未出错")
	}
}

// TestGenerate_InvalidLength 验证长度 < 1 时返回错误
func TestGenerate_InvalidLength(t *testing.T) {
	_, err := Generate(0, GenConfig{Number: true})
	if err == nil {
		t.Error("期望长度 0 返回错误，但未出错")
	}
}

// TestGenerate_CharsFromSpecialSafeOnly 验证启用 SpecialSafe 时生成的字符均来自 CharSpecialSafe
func TestGenerate_CharsFromSpecialSafeOnly(t *testing.T) {
	const rounds = 200
	for i := 0; i < rounds; i++ {
		pwd, err := Generate(20, GenConfig{SpecialSafe: true})
		if err != nil {
			t.Fatalf("生成密码失败：%v", err)
		}
		for _, ch := range pwd {
			if !strings.ContainsRune(CharSpecialSafe, ch) {
				t.Errorf("密码 %q 包含不在 CharSpecialSafe 中的字符 %q", pwd, ch)
			}
		}
	}
}

// TestGenerate_CorrectLength 验证生成密码长度与请求一致
func TestGenerate_CorrectLength(t *testing.T) {
	for _, length := range []int{1, 8, 16, 20, 32} {
		pwd, err := Generate(length, GenConfig{Number: true, Lower: true})
		if err != nil {
			t.Fatalf("length=%d：生成失败：%v", length, err)
		}
		if len([]rune(pwd)) != length {
			t.Errorf("length=%d：期望密码长度 %d，实际 %d", length, length, len([]rune(pwd)))
		}
	}
}

// TestGenerate_MultibyteCharset 验证多字节 UTF-8 字符集可正确生成完整字符（rune 安全）
func TestGenerate_MultibyteCharset(t *testing.T) {
	// 直接调用私有 generate 函数测试 rune 安全性
	charset := "中文日本語한국어"
	const length = 10
	pwd, err := generate(charset, length)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	// 验证字符数（rune 数）等于 length，而非字节数
	runes := []rune(pwd)
	if len(runes) != length {
		t.Errorf("期望 %d 个字符（rune），得到 %d", length, len(runes))
	}
	// 验证每个字符都来自字符集
	charsetRunes := []rune(charset)
	charsetSet := make(map[rune]struct{}, len(charsetRunes))
	for _, r := range charsetRunes {
		charsetSet[r] = struct{}{}
	}
	for _, r := range runes {
		if _, ok := charsetSet[r]; !ok {
			t.Errorf("密码包含不在字符集中的字符 %q", string(r))
		}
	}
}

// ---- GenerateN 单元测试 ----

// TestGenerateN_InvalidCount 验证 n < 1 时返回错误
func TestGenerateN_InvalidCount(t *testing.T) {
	_, err := GenerateN(10, 0, GenConfig{Number: true})
	if err == nil {
		t.Error("期望 n=0 返回错误，但未出错")
	}
}

// TestGenerateN_ReturnsCorrectCount 验证返回数量与 n 一致
func TestGenerateN_ReturnsCorrectCount(t *testing.T) {
	passwords, err := GenerateN(20, 5, GenConfig{Number: true, Lower: true, Upper: true})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(passwords) != 5 {
		t.Errorf("期望 5 个密码，得到 %d", len(passwords))
	}
}

// ---- GenerateStrong 单元测试 ----

// TestGenerateStrong_DefaultConfig 验证默认配置（无特殊字符集）能成功生成强密码
func TestGenerateStrong_DefaultConfig(t *testing.T) {
	pwd, err := GenerateStrong(20, StrongConfig{})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len([]rune(pwd)) != 20 {
		t.Errorf("期望长度 20，得到 %d", len([]rune(pwd)))
	}
}

// TestGenerateStrong_WithSpecialSafe 验证追加安全特殊字符集能成功生成强密码
func TestGenerateStrong_WithSpecialSafe(t *testing.T) {
	pwd, err := GenerateStrong(20, StrongConfig{SpecialSafe: true})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len([]rune(pwd)) != 20 {
		t.Errorf("期望长度 20，得到 %d", len([]rune(pwd)))
	}
}

// TestGenerateStrong_LengthTooShort 验证长度 < 8 返回错误
func TestGenerateStrong_LengthTooShort(t *testing.T) {
	_, err := GenerateStrong(7, StrongConfig{})
	if err == nil {
		t.Error("期望长度不足返回错误，但未出错")
	}
}

// TestGenerateStrong_MutualExclusion 验证 Special 与 SpecialSafe 同时启用返回错误
func TestGenerateStrong_MutualExclusion(t *testing.T) {
	_, err := GenerateStrong(20, StrongConfig{Special: true, SpecialSafe: true})
	if err == nil {
		t.Error("期望互斥错误，但未出错")
	}
}

// TestGenerateStrong_AlwaysContainsThreeCharsets 200 轮：强密码始终包含数字、小写、大写三类字符
func TestGenerateStrong_AlwaysContainsThreeCharsets(t *testing.T) {
	const rounds = 200
	for i := 0; i < rounds; i++ {
		pwd, err := GenerateStrong(20, StrongConfig{})
		if err != nil {
			t.Fatalf("第 %d 轮生成失败：%v", i, err)
		}
		if !strings.ContainsAny(pwd, CharNumbers) {
			t.Errorf("第 %d 轮：强密码缺少数字：%q", i, pwd)
		}
		if !strings.ContainsAny(pwd, CharLower) {
			t.Errorf("第 %d 轮：强密码缺少小写字母：%q", i, pwd)
		}
		if !strings.ContainsAny(pwd, CharUpper) {
			t.Errorf("第 %d 轮：强密码缺少大写字母：%q", i, pwd)
		}
	}
}
