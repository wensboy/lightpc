package service

// SettingPayload is the settings document handed to the front end: the schema that describes every field
// together with the values used to render it.
type SettingPayload struct {
	Schema map[string]any `json:"schema"`
	Values map[string]any `json:"values"`
	// Defaults is the default layer snapshot. The renderer binds it when a user value violates the schema,
	// because a zero value may itself be invalid (an empty string is not a valid size).
	Defaults map[string]any `json:"defaults"`
	// SystemFonts lists the families installed on the machine, so the picker can offer them while
	// useSystem is on without the renderer knowing how to enumerate fonts.
	SystemFonts []string `json:"systemFonts"`
	// Themes holds every selectable palette, merged and keyed as CSS custom properties, so the renderer
	// only has to write them onto the root element.
	Themes []SettingTheme `json:"themes"`
}

// SettingTheme is one palette with both colour modes resolved. Keys are CSS custom properties such as
// "--card-foreground" rather than the document's camel case form.
type SettingTheme struct {
	Name  string            `json:"name"`
	Light map[string]string `json:"light"`
	Dark  map[string]string `json:"dark"`
}

// SettingClient is the surface the front end consumes to render and edit the settings page.
type SettingClient interface {
	// Payload returns the settings schema and the aggregated values.
	Payload() (SettingPayload, error)
	// Save writes one value into the runtime layer and persists it without blocking the caller.
	Save(key string, value any) error
}
