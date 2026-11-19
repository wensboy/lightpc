package setting

import "testing"

func TestFamilyFromFile(t *testing.T) {
	cases := map[string]string{
		// Dash separated style: everything after the last dash is dropped.
		"Inter-Medium.ttf":            "Inter",
		"Cantarell-Regular.otf":       "Cantarell",
		"C059-BdIta.otf":              "C059",
		"BitstreamVeraSans-Roman.ttf": "Bitstream Vera Sans",
		// Camel case splits at a lower to upper transition.
		"AdwaitaMono-Bold.ttf": "Adwaita Mono",
		// A run of capitals ends a word, so the split lands before the next capitalised word.
		"MesloLGLDZNerdFont-Regular.ttf": "Meslo LGLDZ Nerd Font",
		// A digit starts a word when letters follow it.
		"FontAwesome7Brands-Regular.otf": "Font Awesome 7 Brands",
		// A trailing initial stays attached to the digits before it.
		"D050000L.otf": "D050000L",
		// Without a dash the trailing style is still removed.
		"FontAwesome7FreeSolid.otf": "Font Awesome 7 Free",
	}

	for name, want := range cases {
		if got := familyFromName(name); got != want {
			t.Errorf("familyFromName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestSplitCamelCase(t *testing.T) {
	cases := map[string]string{
		"Cantarell":   "Cantarell",
		"D050000L":    "D050000L",
		"C059":        "C059",
		"AdwaitaSans": "Adwaita Sans",
	}
	for input, want := range cases {
		if got := splitCamelCase(input); got != want {
			t.Errorf("splitCamelCase(%q) = %q, want %q", input, got, want)
		}
	}
}
