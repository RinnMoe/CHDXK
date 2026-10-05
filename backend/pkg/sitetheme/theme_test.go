package sitetheme

import (
	"encoding/json"
	"testing"
)

func TestParse(t *testing.T) {
	if _, err := Parse(Defaults); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(Palette){
		func(p Palette) { delete(p, "dark") },
		func(p Palette) { delete(p["light"], "primary") },
		func(p Palette) { p["light"]["extra"] = "#ffffff" },
		func(p Palette) { p["dark"]["primary"] = "red; } body {display:none}" },
	} {
		p, _ := Parse(Defaults)
		change(p)
		value, _ := json.Marshal(p)
		if _, err := Parse(string(value)); err == nil {
			t.Fatal("accepted invalid palette")
		}
	}
}
