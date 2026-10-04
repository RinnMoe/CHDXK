package sitecopy

import (
	"context"
	_ "embed"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

//go:embed catalog.json
var catalogJSON []byte

type Entry struct {
	Text     string `json:"text"`
	Audience string `json:"audience"`
	Location string `json:"location"`
	Optional bool   `json:"optional,omitempty"`
}

var entries = func() map[string]Entry {
	var result map[string]Entry
	if err := json.Unmarshal(catalogJSON, &result); err != nil {
		panic(err)
	}
	return result
}()
var placeholders = regexp.MustCompile(`\{v\d+\}`)

type contextKey struct{}
type valuesKey struct{}

func Entries() map[string]Entry {
	result := make(map[string]Entry, len(entries))
	for key, item := range entries {
		result[key] = item
	}
	return result
}
func Parameters(text string) []string {
	unique := map[string]bool{}
	for _, item := range placeholders.FindAllString(text, -1) {
		unique[item] = true
	}
	result := make([]string, 0, len(unique))
	for item := range unique {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}
func ValidValue(key, value string) bool {
	entry, ok := entries[key]
	return ok && (entry.Optional || strings.TrimSpace(value) != "") && utf8.RuneCountInString(value) <= 8000 && strings.Join(Parameters(value), ",") == strings.Join(Parameters(entry.Text), ",")
}
func WithValues(ctx context.Context, values map[string]string) context.Context {
	translated := map[string]string{}
	for key, item := range entries {
		if strings.HasPrefix(key, "copy.server.") {
			if value, ok := values[key]; ok {
				translated[item.Text] = value
			}
		}
	}
	ctx = context.WithValue(ctx, valuesKey{}, values)
	return context.WithValue(ctx, contextKey{}, translated)
}

func Value(ctx context.Context, key string) string {
	if values, ok := ctx.Value(valuesKey{}).(map[string]string); ok {
		if value, exists := values[key]; exists {
			return value
		}
	}
	if entry, ok := entries[key]; ok {
		return entry.Text
	}
	return key
}
func Text(ctx context.Context, text string) string {
	if values, ok := ctx.Value(contextKey{}).(map[string]string); ok {
		if value, exists := values[text]; exists {
			return value
		}
	}
	return text
}
