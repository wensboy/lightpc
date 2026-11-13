package cast

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// To asserts the given value to T.
func To[T any](v any) (T, bool) {
	value, ok := v.(T)
	return value, ok
}

// MustTo asserts the given value to T and panics when the type does not match.
func MustTo[T any](v any) T {
	value, ok := v.(T)
	if !ok {
		panic(fmt.Sprintf("cast: cannot convert %T to %T", v, *new(T)))
	}
	return value
}

// Is reports whether the given value is of type T.
func Is[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

// As asserts the given value to T first, and falls back to a direct type conversion when the value permits it.
func As[T any](v any) (T, bool) {
	if value, ok := v.(T); ok {
		return value, true
	}
	var target T
	source := reflect.ValueOf(v)
	destination := reflect.ValueOf(&target).Elem()
	if !source.IsValid() || !source.Type().ConvertibleTo(destination.Type()) {
		return target, false
	}
	destination.Set(source.Convert(destination.Type()))
	return target, true
}

// ToStruct maps a map[string]any into the structure that target points to, resolving fields by the given tag name.
func ToStruct[T any](source map[string]any, target T, tagname string) {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  target,
		TagName: tagname,
	})
	if err != nil {
		return
	}
	_ = decoder.Decode(source)
}
