package context

import "context"

// AppContext 代表应用上下文, 所有应用需要的上下文集合, 被运行时持有, 用于统一管理应用的上下文持有和传递.
type AppContext struct {
	context.Context
	EmbedContext
	ConfigContext
	ErrorContext
}
