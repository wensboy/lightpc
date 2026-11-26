package notice

// Client is the surface a caller depends on.
//
// Publishing is the reason the engine exists: any module can raise a notice without knowing how it reaches the
// interface. The listing operations serve a future notice service, which owns the sheet the person sees.
type Client interface {
	// Notify records a notice, announces it on the bus and returns the stored form.
	Notify(input Input) (Record, error)
	// History returns every live notice, newest first.
	History() []Record
	// Remove drops one notice and reports whether it was there.
	Remove(id uint64) bool
	// Clear drops every notice and reports how many were removed.
	Clear() int
}
