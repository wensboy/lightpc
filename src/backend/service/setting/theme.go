package serviceSettingImpl

import (
	serviceSetting "github.com/wensboy/lightpc/backend/domain/service/setting"
	"strings"
	"unicode"

	"github.com/wensboy/lightpc/backend/internal/cast"
)

// cssProperty turns a storage key into the custom property the stylesheet expects: cardForeground becomes
// --card-foreground.
func cssProperty(key string) string {
	var builder strings.Builder
	builder.WriteString("--")
	for _, letter := range key {
		if unicode.IsUpper(letter) {
			builder.WriteByte('-')
			letter = unicode.ToLower(letter)
		}
		builder.WriteRune(letter)
	}
	return builder.String()
}

// buildThemes resolves the merged palette list into something the renderer can apply directly: every entry
// carries both colour modes and keys already converted to custom properties.
func buildThemes(merged map[string]any) []serviceSetting.SettingTheme {
	// An empty slice rather than nil: the renderer validates an array, and a null would be rejected even
	// when there is simply nothing to list.
	themes := make([]serviceSetting.SettingTheme, 0, 4)
	list, ok := cast.To[[]any](merged[themePalettesKey])
	if !ok {
		return themes
	}
	for _, entry := range list {
		item, ok := cast.To[map[string]any](entry)
		if !ok {
			continue
		}
		name, ok := cast.To[string](item["name"])
		if !ok || name == "" {
			continue
		}
		themes = append(themes, serviceSetting.SettingTheme{
			Name:  name,
			Light: paletteProperties(item["light"]),
			Dark:  paletteProperties(item["dark"]),
		})
	}
	return themes
}

// paletteProperties converts one palette into custom property names, skipping entries that are not text.
func paletteProperties(raw any) map[string]string {
	source, ok := cast.To[map[string]any](raw)
	if !ok {
		return nil
	}
	properties := make(map[string]string, len(source))
	for key, value := range source {
		if text, ok := cast.To[string](value); ok {
			properties[cssProperty(key)] = text
		}
	}
	return properties
}
