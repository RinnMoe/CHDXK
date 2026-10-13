package bkjw

import (
	"regexp"
	"strings"
)

var constructorPattern = regexp.MustCompile(`\bnew\s+(?:window\s*\.\s*)?(TaskActivity|CourseTable)\s*\(`)

// Mask comments and optionally strings while preserving offsets. This locates
// data constructors without executing JavaScript or matching quoted examples.
func scriptCode(input string, maskStrings bool) (string, error) {
	out := []byte(input)
	for i := 0; i < len(input); {
		if input[i] == '\'' || input[i] == '"' || input[i] == '`' {
			start, quote := i, input[i]
			i++
			for i < len(input) && input[i] != quote {
				if input[i] == '\\' {
					i++
				}
				i++
			}
			if i >= len(input) {
				return "", codeError("PERSONAL_PARSE_FAILED")
			}
			i++
			if maskStrings {
				for j := start; j < i; j++ {
					out[j] = ' '
				}
			}
			continue
		}
		if strings.HasPrefix(input[i:], "//") {
			for i < len(input) && input[i] != '\n' && input[i] != '\r' {
				out[i] = ' '
				i++
			}
			continue
		}
		if strings.HasPrefix(input[i:], "/*") {
			end := strings.Index(input[i+2:], "*/")
			if end < 0 {
				return "", codeError("PERSONAL_PARSE_FAILED")
			}
			end += i + 4
			for i < end {
				if out[i] != '\n' && out[i] != '\r' {
					out[i] = ' '
				}
				i++
			}
			continue
		}
		i++
	}
	return string(out), nil
}
