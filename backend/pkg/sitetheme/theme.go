package sitetheme

import (
	_ "embed"
	"encoding/json"
	"errors"
	"regexp"
)

const Key = "theme.palette"

//go:embed defaults.json
var Defaults string

type Palette map[string]map[string]string

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$`)

func Parse(value string) (Palette, error) {
	var palette, defaults Palette
	if len(value) > 16384 || json.Unmarshal([]byte(value), &palette) != nil {
		return nil, errors.New("invalid theme palette")
	}
	if err := json.Unmarshal([]byte(Defaults), &defaults); err != nil {
		return nil, err
	}
	if len(palette) != len(defaults) {
		return nil, errors.New("invalid theme modes")
	}
	for mode, tokens := range defaults {
		actual := palette[mode]
		if len(actual) != len(tokens) {
			return nil, errors.New("invalid theme tokens")
		}
		for token := range tokens {
			if !hexColor.MatchString(actual[token]) {
				return nil, errors.New("invalid theme color")
			}
		}
	}
	return palette, nil
}
