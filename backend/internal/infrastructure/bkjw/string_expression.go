package bkjw

import (
	"log/slog"
	"regexp"
	"strings"
)

// Decode only string literals joined by +, optionally grouped by parentheses.
// No identifiers, function calls, properties or scripts are evaluated.
func jsStringExpression(raw string) (string, error) {
	return boundStringExpression(raw, nil)
}

func boundStringExpression(raw string, bindings map[string]string) (string, error) {
	if len(raw) > 65536 {
		return "", codeError("PERSONAL_PARSE_FAILED")
	}
	cursor, terms := 0, 0
	space := func() {
		for cursor < len(raw) && strings.ContainsRune(" \t\r\n", rune(raw[cursor])) {
			cursor++
		}
	}
	var expression func(int) (string, error)
	var term func(int) (string, error)
	term = func(depth int) (string, error) {
		space()
		terms++
		if depth > 8 || terms > 64 || cursor >= len(raw) {
			return "", codeError("PERSONAL_PARSE_FAILED")
		}
		if raw[cursor] == '(' {
			cursor++
			value, err := expression(depth + 1)
			if err != nil {
				return "", err
			}
			space()
			if cursor >= len(raw) || raw[cursor] != ')' {
				return "", codeError("PERSONAL_PARSE_FAILED")
			}
			cursor++
			return value, nil
		}
		if raw[cursor] != '\'' && raw[cursor] != '"' {
			start := cursor
			for cursor < len(raw) && identifierByte(raw[cursor], cursor > start) {
				cursor++
			}
			if value, ok := bindings[raw[start:cursor]]; ok && cursor > start {
				return value, nil
			}
			return "", codeError("PERSONAL_PARSE_FAILED")
		}
		start, quote := cursor, raw[cursor]
		cursor++
		for cursor < len(raw) && raw[cursor] != quote {
			if raw[cursor] == '\\' {
				cursor++
			}
			cursor++
		}
		if cursor >= len(raw) {
			return "", codeError("PERSONAL_PARSE_FAILED")
		}
		cursor++
		return jsString(raw[start:cursor])
	}
	expression = func(depth int) (string, error) {
		value, err := term(depth)
		if err != nil {
			return "", err
		}
		for {
			space()
			if cursor >= len(raw) || raw[cursor] != '+' {
				return value, nil
			}
			cursor++
			next, err := term(depth)
			if err != nil {
				return "", err
			}
			value += next
		}
	}
	value, err := expression(0)
	space()
	if err != nil || cursor != len(raw) {
		return "", codeError("PERSONAL_PARSE_FAILED")
	}
	return value, nil
}

func identifierByte(ch byte, digit bool) bool {
	return ch == '_' || ch == '$' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || digit && ch >= '0' && ch <= '9'
}

var stringAssignment = regexp.MustCompile(`\b([A-Za-z_$][A-Za-z0-9_$]*)\s*=`)

// Read preceding assignments as bounded string data, never execute script.
// Unsupported reassignment invalidates the binding instead of retaining stale data.
func precedingStringBindings(clean, mask string) map[string]string {
	bindings := map[string]string{}
	for _, loc := range stringAssignment.FindAllStringSubmatchIndex(mask, -1) {
		start := loc[0]
		previous := strings.TrimSpace(mask[:start])
		if strings.HasSuffix(previous, ".") || strings.HasSuffix(previous, "=") || len(bindings) > 1024 {
			continue
		}
		name, begin := mask[loc[2]:loc[3]], loc[1]
		if begin < len(mask) && (mask[begin] == '=' || mask[begin] == '>') {
			continue
		}
		end, depth := begin, 0
		for end < len(mask) {
			ch := mask[end]
			if depth == 0 && (ch == ';' || ch == ',' || ch == '\n' || ch == '\r' || ch == '}') {
				break
			}
			if ch == '(' {
				depth++
			}
			if ch == ')' {
				depth--
				if depth < 0 {
					break
				}
			}
			end++
		}
		value, err := boundStringExpression(clean[begin:end], bindings)
		if err != nil {
			delete(bindings, name)
		} else {
			bindings[name] = value
		}
	}
	return bindings
}

func personalNameFailure(raw string, size, items int) error {
	// Expose grammar only: all literals, identifiers and numbers are replaced.
	// This distinguishes concatenation/grouping/calls without retaining names.
	var shape strings.Builder
	for i := 0; i < len(raw) && shape.Len() < 160; {
		ch := raw[i]
		if ch == '\'' || ch == '"' {
			quote := ch
			i++
			for i < len(raw) && raw[i] != quote {
				if raw[i] == '\\' {
					i++
				}
				i++
			}
			if i < len(raw) {
				i++
			}
			shape.WriteString("STRING")
			continue
		}
		if ch >= '0' && ch <= '9' {
			for i < len(raw) && raw[i] >= '0' && raw[i] <= '9' {
				i++
			}
			shape.WriteString("NUMBER")
			continue
		}
		if ch == '_' || ch == '$' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
			for i < len(raw) && (raw[i] == '_' || raw[i] == '$' || raw[i] >= 'a' && raw[i] <= 'z' || raw[i] >= 'A' && raw[i] <= 'Z' || raw[i] >= '0' && raw[i] <= '9') {
				i++
			}
			shape.WriteString("IDENT")
			continue
		}
		if strings.ContainsRune("+-*/?:(),[].", rune(ch)) {
			shape.WriteByte(ch)
		}
		i++
	}
	slog.Warn("academic timetable expression rejected", "reason", "activity-name-expression", "expression_shape", shape.String(), "response_bytes", size, "structure_items", items)
	return codeError("PERSONAL_PARSE_FAILED")
}
