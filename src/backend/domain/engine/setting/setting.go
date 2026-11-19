package setting

// Layer is one configuration layer. Layers are ordered by priority: a higher index wins and the
// resolved view reads from the last layer backwards.
type Layer struct {
	Name string
	Mode Mode
	Path string // store.fs path; paths carrying embed:// are read only
}

// Entry is one layer contribution to a key. The value type stays generic so a caller can cast the
// untyped contribution back into the shape it stored; the engine itself never inspects Value.
type Entry[T any] struct {
	Layer string
	Mode  Mode
	Value T
}

// Setting is the layered settings centre of the application, comparable to an editor's user
// settings: several layers contribute values for the same key and the highest priority one wins.
//
// Values differ between keys and cannot be typed at compile time, so contributions travel as any.
// Every element returned by Layered or Resolve is an Entry[T]; callers cast it back with the shape
// they wrote. Resolving the payload itself is the caller's job, not the engine's.
type Setting interface {
	// Layers lists the registered layers in ascending priority.
	Layers() []Layer
	// Layered returns every contribution of key in ascending priority.
	Layered(key string) []any
	// Resolve returns the highest priority contribution of key.
	Resolve(key string) (any, bool)
	// Set writes value into the named layer; the layer must accept writes.
	Set(layer, key string, value any) error
	// Load replaces the whole content of the named layer.
	Load(layer string, values map[string]any) error
	// Keys lists every key that has at least one contribution, ordered alphabetically.
	Keys() []string
	// Flush sinks the runtime layer into the layer that owns persistence and empties the runtime layer.
	// It is a no-op when nothing was edited since the last flush.
	Flush() error
}
