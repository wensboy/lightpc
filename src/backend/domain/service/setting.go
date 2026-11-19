package service

// SettingPayload is the settings document handed to the front end: the schema that describes every field
// together with the values used to render it.
type SettingPayload struct {
	Schema map[string]any `json:"schema"`
	Values map[string]any `json:"values"`
	// Defaults is the default layer snapshot. The renderer binds it when a user value violates the schema,
	// because a zero value may itself be invalid (an empty string is not a valid size).
	Defaults map[string]any `json:"defaults"`
}

// SettingClient is the surface the front end consumes to render and edit the settings page.
type SettingClient interface {
	// Payload returns the settings schema and the aggregated values.
	Payload() (SettingPayload, error)
	// Save writes one value into the runtime layer and persists it without blocking the caller.
	Save(key string, value any) error
}
