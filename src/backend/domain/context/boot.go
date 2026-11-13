package context

const (
	ModeTui uint8 = iota
	ModeWeb
	ModeGui
)

type BootContext struct {
	Mode uint8
	AppContext
}
