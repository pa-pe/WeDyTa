package utils

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"regexp"
	"strings"
	"unicode"
)

// CamelToSnake convert from camelCase to snake_case
func CamelToSnake(input string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(input, "${1}_${2}")
	return strings.ToLower(snake)
}

// SnakeToCamel convert from snake_case to camelCase
func SnakeToCamel(input string) string {
	parts := strings.Split(input, "_")
	for i, part := range parts {
		parts[i] = cases.Title(language.Und).String(part)
	}
	return strings.Join(parts, "")
}

// InvertCaseStyle invert from snake_case to camelCase and vice versa
func InvertCaseStyle(input string) string {
	if IsCamelCase(input) {
		return CamelToSnake(input)
	}
	return SnakeToCamel(input)
}

// IsCamelCase check is camelCase
func IsCamelCase(input string) bool {
	for _, r := range input {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func CleanPrefixes(s string, prefixes []string) string {
	s = strings.TrimSpace(s)
	upper := strings.ToUpper(s)
	for _, prefix := range prefixes {
		prefixUpper := strings.ToUpper(prefix)
		if strings.HasPrefix(upper, prefixUpper) {
			return strings.TrimSpace(s[len(prefix):])
		}
	}
	return s
}
