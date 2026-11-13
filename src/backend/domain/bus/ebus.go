package bus

// EventHandler 事件处理函数.
type EventHandler func(any)

type Bus interface {
	On(string, EventHandler)
	Emit(string, any)
}
