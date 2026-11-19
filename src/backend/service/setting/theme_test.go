package setting

import "testing"

func palettes(list []any) map[string]any {
	return map[string]any{themePalettesKey: list}
}

func themeByName(t *testing.T, merged map[string]any, name string) map[string]any {
	t.Helper()
	list, ok := merged[themePalettesKey].([]any)
	if !ok {
		t.Fatalf("merged = %T", merged[themePalettesKey])
	}
	for _, entry := range list {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if item["name"] == name {
			return item
		}
	}
	t.Fatalf("theme %q missing", name)
	return nil
}

func TestMergeAppendsUnknownTheme(t *testing.T) {
	defaults := palettes([]any{
		map[string]any{"name": "Default", "light": map[string]any{"background": "white"}},
	})
	values := palettes([]any{
		map[string]any{"name": "Custom", "light": map[string]any{"background": "black"}},
	})

	mergeThemePalettes(values, defaults)

	if got := themeByName(t, values, "Custom")["light"].(map[string]any)["background"]; got != "black" {
		t.Fatalf("custom background = %v", got)
	}
	if got := themeByName(t, values, "Default")["light"].(map[string]any)["background"]; got != "white" {
		t.Fatalf("default background = %v", got)
	}
}

func TestMergeOverridesNamedThemePerKey(t *testing.T) {
	defaults := palettes([]any{
		map[string]any{
			"name":  "Default",
			"light": map[string]any{"background": "white", "foreground": "black"},
			"dark":  map[string]any{"background": "black"},
		},
	})
	values := palettes([]any{
		map[string]any{
			"name":  "Default",
			"light": map[string]any{"background": "cream"},
			"dark":  map[string]any{"foreground": "white"},
		},
	})

	mergeThemePalettes(values, defaults)

	merged := themeByName(t, values, "Default")
	light := merged["light"].(map[string]any)
	if light["background"] != "cream" {
		t.Fatalf("background must be overridden, got %v", light["background"])
	}
	if light["foreground"] != "black" {
		t.Fatalf("untouched key must keep its shipped value, got %v", light["foreground"])
	}
	dark := merged["dark"].(map[string]any)
	if dark["foreground"] != "white" || dark["background"] != "black" {
		t.Fatalf("dark = %v", dark)
	}
}
