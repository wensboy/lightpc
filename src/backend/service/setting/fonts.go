package setting

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	gotextfont "github.com/go-text/typesetting/font"
	ot "github.com/go-text/typesetting/font/opentype"
)

// fontExtensions lists the file extensions a usable font can carry. Collections are included because the
// CJK families ship as a single .ttc holding several faces.
var fontExtensions = []string{".ttf", ".otf", ".ttc", ".otc"}

// styleWords are the tokens a file name appends to the family. They are matched against the tail so a name
// without the usual dash separator still yields the family instead of the family plus its style.
var styleWords = []string{
	"bolditalic", "boldoblique", "semibold", "extrabold", "ultrabold",
	"extralight", "ultralight", "regular", "italic", "oblique", "medium",
	"light", "black", "thin", "solid", "bold", "book", "roman",
}

// systemFontDirs lists the directories a font family is installed in on Linux. A path starting with "~"
// is expanded against the user home directory.
var systemFontDirs = []string{
	"/usr/share/fonts",
	"/usr/local/share/fonts",
	"~/.local/share/fonts",
	"~/.fonts",
}

// SystemFontFamilies lists the font families installed on Linux, sorted and deduplicated.
//
// The family comes from the font's own name table, so the entries are the ones a browser can resolve. A
// file that cannot be parsed falls back to its name, which keeps an unusual install visible instead of
// dropping it.
func SystemFontFamilies() []string {
	seen := map[string]struct{}{}
	families := make([]string, 0, 256)

	for _, dir := range systemFontDirs {
		root := dir
		if strings.HasPrefix(root, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			root = filepath.Join(home, strings.TrimPrefix(root, "~"))
		}
		_ = filepath.WalkDir(root, func(entryPath string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			if !slices.Contains(fontExtensions, strings.ToLower(filepath.Ext(entry.Name()))) {
				return nil
			}
			for _, family := range familiesFromPath(entryPath, entry.Name()) {
				if _, found := seen[family]; found {
					continue
				}
				seen[family] = struct{}{}
				families = append(families, family)
			}
			return nil
		})
	}

	slices.Sort(families)
	return families
}

// familiesFromPath reads the families of one font file, falling back to the file name when the tables
// cannot be read. A collection yields one entry per face.
func familiesFromPath(path, name string) []string {
	data, err := os.ReadFile(path)
	if err == nil {
		if families := describeFamilies(data); len(families) > 0 {
			return families
		}
	}
	if fallback := familyFromName(name); fallback != "" {
		return []string{fallback}
	}
	return nil
}

// describeFamilies returns every family declared by the file. A collection carries several faces, a plain
// font carries one, so loaders are always read as a list.
func describeFamilies(data []byte) []string {
	loaders, err := ot.NewLoaders(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	families := make([]string, 0, len(loaders))
	for _, loader := range loaders {
		description, _ := gotextfont.Describe(loader, nil)
		family := strings.Join(strings.Fields(description.Family), " ")
		if family != "" {
			families = append(families, family)
		}
	}
	return families
}

// familyFromName derives the family from the file name for files without a readable name table. A font
// file is normally "<Family>-<Style>.<ext>", so the style suffix is dropped and camel case is split back
// into words.
func familyFromName(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if index := strings.LastIndex(base, "-"); index > 0 {
		base = base[:index]
	}
	// A dash promises to separate the style, but some families still carry it in the name part, so the
	// trailing style words are removed either way.
	base = trimStyleSuffix(base)
	return splitCamelCase(strings.ReplaceAll(base, "-", ""))
}

// trimStyleSuffix removes a trailing style word that a dash would otherwise have separated.
func trimStyleSuffix(base string) string {
	trimmed := base
	for {
		next := trimmed
		lower := strings.ToLower(next)
		for _, style := range styleWords {
			if strings.HasSuffix(lower, style) && len(next) > len(style) {
				next = next[:len(next)-len(style)]
				break
			}
		}
		if next == trimmed {
			return trimmed
		}
		trimmed = next
	}
}

// splitCamelCase separates words at a lower to upper transition and around a digit group, so
// "AdwaitaMono" becomes "Adwaita Mono" while "D050000L" keeps its shape.
func splitCamelCase(text string) string {
	var builder strings.Builder
	runes := []rune(text)
	for i, r := range runes {
		if i > 0 {
			previous := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			// A word can end in a run of capitals: "LGLDZNerdFont" splits before "Nerd".
			afterAcronym := unicode.IsUpper(r) && unicode.IsUpper(previous) && nextLower
			// A digit starts a word when letters follow it, so "7Brands" splits but the trailing initial
			// of "D050000L" stays attached.
			afterDigit := unicode.IsUpper(r) && unicode.IsDigit(previous) && i+1 < len(runes)
			beforeDigit := unicode.IsDigit(r) && unicode.IsLower(previous)
			if unicode.IsLower(previous) && unicode.IsUpper(r) || afterAcronym || afterDigit || beforeDigit {
				builder.WriteRune(' ')
			}
		}
		builder.WriteRune(r)
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

// bundledFamilies lists the families that actually ship with the frontend bundle.
func (s *SettingService) bundledFamilies() map[string]struct{} {
	families := map[string]struct{}{}
	if s.appCtx == nil {
		return families
	}
	dist, found := s.appCtx.EmbedContext["/frontend/dist"]
	if !found || dist == nil {
		return families
	}
	entries, err := fs.ReadDir(dist, "frontend/dist/fonts")
	if err != nil {
		return families
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !slices.Contains(fontExtensions, strings.ToLower(filepath.Ext(entry.Name()))) {
			continue
		}
		names := []string{familyFromName(entry.Name())}
		if data, err := fs.ReadFile(dist, "frontend/dist/fonts/"+entry.Name()); err == nil {
			if described := describeFamilies(data); len(described) > 0 {
				names = described
			}
		}
		for _, name := range names {
			if name != "" {
				families[normaliseFamily(name)] = struct{}{}
			}
		}
	}
	return families
}

// normaliseFamily folds a family name for comparison.
func normaliseFamily(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), ""))
}
