package serviceSetting

// SettingPayloadDto is the settings document handed to the front end: the schema that describes every field
// together with the values used to render it.
type SettingPayloadDto struct {
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
	Themes []SettingThemeDto `json:"themes"`
}

// SettingThemeDto is one palette with both colour modes resolved. Keys are CSS custom properties such as
// "--card-foreground" rather than the document's camel case form.
type SettingThemeDto struct {
	Name  string            `json:"name"`
	Light map[string]string `json:"light"`
	Dark  map[string]string `json:"dark"`
}

// Kind routes a change onto one of the two buses.
//
// It is the service's vocabulary rather than the engine's: the engine stores layers and knows nothing about who listens.
// What a change is *for* - an engine that has to re-read a parameter, or a service and the interface that only need the
// new value - is a business distinction, and it is made once, here.
type Kind string

const (
	// KindService is an ordinary setting: frequent, mostly cosmetic, and it only reaches services and the interface.
	KindService Kind = "service"
	// KindEngine is an advanced setting: rare, and it drives a parameter another engine reads while running.
	KindEngine Kind = "engine"
)

// Known reports whether the kind is one the service can route.
func (k Kind) Known() bool {
	return k == KindService || k == KindEngine
}

// SettingChangedPayload is the payload of EventSettingChanged on the service bus: one key that now resolves to one
// value.
//
// The shape is deliberately a copy of the engine's payload rather than an alias of it. The two buses are isolated, and
// an alias would couple them: whoever edited the engine's struct to serve a new listener would silently change what
// services receive. A copy means the two can evolve apart, and a listener casts to the type of the bus it subscribed to.
type SettingChangedPayload struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}
