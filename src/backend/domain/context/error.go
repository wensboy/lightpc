package context

// ErrorContext 映射 Code 到 Kind. 非并发安全, 逻辑上确保多协程不会出现并发读写问题.
type ErrorContext map[string]string
