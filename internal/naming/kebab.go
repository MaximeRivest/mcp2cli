package naming

import (
	"strings"
	"unicode"
)

// ToKebabCase converts common identifier styles like camelCase, snake_case,
// PascalCase, and space-separated labels to kebab-case.
func ToKebabCase(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	var out []rune
	var prev rune
	for i, r := range []rune(input) {
		switch {
		case r == '_' || r == ' ' || r == '/' || r == '.':
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		case unicode.IsUpper(r):
			nextLower := false
			if i+1 < len([]rune(input)) {
				nextLower = unicode.IsLower([]rune(input)[i+1])
			}
			if len(out) > 0 && out[len(out)-1] != '-' && (unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower)) {
				out = append(out, '-')
			}
			out = append(out, unicode.ToLower(r))
		default:
			out = append(out, unicode.ToLower(r))
		}
		prev = r
	}

	result := strings.Trim(string(out), "-")
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return result
}
