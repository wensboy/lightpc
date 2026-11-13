package cast

import "testing"

type sample struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func TestToStructUsesTagName(t *testing.T) {
	var target sample
	ToStruct(map[string]any{"name": "app", "port": 8080}, &target, "json")
	if target.Name != "app" || target.Port != 8080 {
		t.Fatalf("target = %+v", target)
	}
}

func TestToStructIgnoresUnknownKeys(t *testing.T) {
	var target sample
	ToStruct(map[string]any{"name": "app", "unknown": true}, &target, "json")
	if target.Name != "app" || target.Port != 0 {
		t.Fatalf("target = %+v", target)
	}
}

func TestToStructWithDefaultTag(t *testing.T) {
	type defaultTagged struct {
		Name string `mapstructure:"name"`
	}
	var target defaultTagged
	ToStruct(map[string]any{"name": "app"}, &target, "")
	if target.Name != "app" {
		t.Fatalf("target = %+v", target)
	}
}

func TestTo(t *testing.T) {
	if value, ok := To[int](1); !ok || value != 1 {
		t.Fatalf("To[int](1) = %v, %v", value, ok)
	}
	if _, ok := To[int](int64(1)); ok {
		t.Fatal("To[int](int64) must fail")
	}
	if value, ok := To[*sample](&sample{}); !ok || value == nil {
		t.Fatalf("To[*sample] = %v, %v", value, ok)
	}
}

func TestMustTo(t *testing.T) {
	if value := MustTo[string]("text"); value != "text" {
		t.Fatalf("MustTo[string] = %v", value)
	}
}

func TestMustToPanicsOnMismatch(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustTo must panic on a type mismatch")
		}
	}()
	MustTo[int]("text")
}

func TestIs(t *testing.T) {
	if !Is[int](1) {
		t.Fatal("Is[int](1) must be true")
	}
	if Is[int](int64(1)) {
		t.Fatal("Is[int](int64) must be false")
	}
	if !Is[*sample](&sample{}) {
		t.Fatal("Is[*sample] must be true")
	}
}

func TestAsPrefersAssertion(t *testing.T) {
	if value, ok := As[int](1); !ok || value != 1 {
		t.Fatalf("As[int](1) = %v, %v", value, ok)
	}
}

func TestAsConverts(t *testing.T) {
	if value, ok := As[int](float64(1)); !ok || value != 1 {
		t.Fatalf("As[int](float64) = %v, %v", value, ok)
	}
	if value, ok := As[int64](int(2)); !ok || value != 2 {
		t.Fatalf("As[int64](int) = %v, %v", value, ok)
	}
	if value, ok := As[float64](int(3)); !ok || value != 3 {
		t.Fatalf("As[float64](int) = %v, %v", value, ok)
	}
	if value, ok := As[string]([]byte("text")); !ok || value != "text" {
		t.Fatalf("As[string]([]byte) = %v, %v", value, ok)
	}
	if value, ok := As[string](rune('a')); !ok || value != "a" {
		t.Fatalf("As[string](rune) = %v, %v", value, ok)
	}
}

func TestAsRejectsNonConvertible(t *testing.T) {
	if _, ok := As[int]("12"); ok {
		t.Fatal("As[int](string) must fail")
	}
	if _, ok := As[bool]("true"); ok {
		t.Fatal("As[bool](string) must fail")
	}
	if _, ok := As[*sample](sample{}); ok {
		t.Fatal("As[*sample](sample) must fail")
	}
	if _, ok := As[int](nil); ok {
		t.Fatal("As[int](nil) must fail")
	}
}
