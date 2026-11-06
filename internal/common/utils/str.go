package utils

import (
	"strings"
	"unicode"
)

// KebabCase 驼峰或下划线命名转短横线小写，如 MenuDemo → menu-demo
func KebabCase(s string) string {
	var b strings.Builder
	pendingSep := false
	for _, r := range strings.TrimSpace(s) {
		if r == '_' || r == '-' || r == ' ' {
			pendingSep = true
			continue
		}
		if r >= 'A' && r <= 'Z' {
			if b.Len() > 0 && !pendingSep {
				b.WriteByte('-')
			}
			r = unicode.ToLower(r)
		}
		if pendingSep && b.Len() > 0 {
			b.WriteByte('-')
		}
		b.WriteRune(r)
		pendingSep = false
	}
	return b.String()
}

// PluralKebab 短横线命名转英文复数，如 dict-item → dict-items、category → categories
// 接口路径与前端目录共用同一套复数规则，避免生成结果与现有路由风格不一致
func PluralKebab(s string) string {
	word := KebabCase(s)
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") ||
		strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") {
		return word + "es"
	}
	if len(word) > 1 && strings.HasSuffix(word, "y") && !strings.ContainsRune("aeiou", rune(word[len(word)-2])) {
		return word[:len(word)-1] + "ies"
	}
	return word + "s"
}
