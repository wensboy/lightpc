package engineSetting

import "testing"

func TestRuntimesMapsTheConfiguredLayers(t *testing.T) {
	config := Config{Layers: []LayerConfig{
		{Layer: "default", Path: "embed://data/conf/setting.default.json", Mode: "ro"},
		{Layer: "user", Path: "local://~/.config/lightpc/setting.json", Mode: "rw"},
		{Layer: "", Path: "local:///tmp/unnamed.json", Mode: "rw"},
	}}

	layers := config.GenLayers()
	// An entry without a name cannot be addressed, so it is dropped rather than producing a layer.
	if len(layers) != 2 {
		t.Fatalf("layers = %d", len(layers))
	}
	if layers[0].Name != "default" || layers[0].Mode != ModeReadOnly {
		t.Fatalf("first layer = %+v", layers[0])
	}
	if layers[1].Name != "user" || layers[1].Mode != ModeReadWrite {
		t.Fatalf("second layer = %+v", layers[1])
	}
}

func TestModeOfFallsBackToReadOnly(t *testing.T) {
	cases := map[string]Mode{
		"rw":    ModeReadWrite,
		"wo":    ModeWriteOnly,
		"ro":    ModeReadOnly,
		"":      ModeReadOnly,
		"write": ModeReadOnly,
	}
	for text, want := range cases {
		if got := ModeOf(text); got != want {
			t.Fatalf("ModeOf(%q) = %v, want %v", text, got, want)
		}
	}
}
