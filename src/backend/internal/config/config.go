package config

import (
	"sync"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
)

// Config 对外提供所有配置的管理和查找, 遵循统一格式.
type Config struct {
	mu sync.RWMutex

	sources  []domainConfig.Source // 静态配置, 包含: 配置文件, 环境变量, flag 参数.
	registry map[string]any        // 动态配置, 包含: 手动注册, 动态上下文注册.
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithSources(ss ...domainConfig.Source) *Config {
	c.sources = ss
	return c
}

func (c *Config) WithRegistry(registry map[string]any) *Config {
	c.registry = registry
	return c
}

func (c *Config) Kind() uint8 {
	return domainConfig.SourceConfig
}

func (c *Config) Lookup(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, source := range c.sources {
		if value, ok := source.Lookup(key); ok {
			return value, true
		}
	}
	if value, ok := c.registry[key]; ok {
		return value, true
	}
	return nil, false
}

// Sources 返回与定义顺序一致的所有同 kind 的 Source 实例.
func (c *Config) Sources(kind uint8) []domainConfig.Source {
	ss := make([]domainConfig.Source, 0, len(c.sources))
	for _, s := range c.sources {
		if s.Kind() == kind {
			ss = append(ss, s)
		}
	}
	return ss
}
