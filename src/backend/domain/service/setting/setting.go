package serviceSetting

// SettingClient is the surface the front end consumes to render and edit the settings page.
type SettingClient interface {
	// Payload returns the settings schema and the aggregated values.
	Payload() (SettingPayloadDto, error)
	// Save writes one value into the runtime layer and persists it without blocking the caller.
	Save(key string, value any, kind Kind) error
}
