package context

import "embed"

// EmbedContext 持有 embed.FS 数据, 用于透明传递所有 embed.FS 结构并提供持久调用.
// key 按照规范总是以相对 src 的目录路径, 对于单结构多嵌入, key 定义为语义功能.
type EmbedContext map[string]*embed.FS

// FsContext 对 fs 路径处理统一传递, 能够同时兼容 local fs 和 embed fs 的处理函数.
type FsContext struct {
	Paths []string
	Fs    *embed.FS
}
