package config

import (
	"os"
	"strings"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

// EnvSource 实现了 config.Source 接口, 查找环境变量.
type EnvSource struct {
	Prefix string              // PREFIX_xxx
	KeyFn  func(string) string // 自定义 key 转换

	extra map[string]any // 拓展 env, 通常来自 .env 文件加载.
}

func NewEnvSource() *EnvSource {
	return &EnvSource{
		Prefix: "",
		KeyFn: func(key string) string {
			return strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		},
		extra: make(map[string]any),
	}
}

func (es *EnvSource) WithPrefix(prefix string) *EnvSource {
	es.Prefix = prefix
	return es
}

func (es *EnvSource) WithKeyFn(keyFn func(string) string) *EnvSource {
	es.KeyFn = keyFn
	return es
}

func (es *EnvSource) WithExtra(extra map[string]any) *EnvSource {
	es.extra = extra
	return es
}

func (es *EnvSource) Export(fsCtx domainContext.FsContext) {
	for _, envPath := range fsCtx.Paths {
		var content []byte
		var err error
		if fsCtx.Fs != nil {
			content, err = fsCtx.Fs.ReadFile(strings.TrimPrefix(envPath, "/"))
		} else {
			content, err = os.ReadFile(envPath)
		}
		if err != nil {
			continue
		}
		for line := range strings.SplitSeq(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, found := strings.Cut(line, "=")
			if !found {
				continue
			}
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			es.extra[key] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
}

func (es *EnvSource) Kind() uint8 {
	return domainConfig.SourceEnv
}

func (es *EnvSource) Lookup(key string) (any, bool) {
	if es.KeyFn != nil {
		key = es.KeyFn(key)
	}
	key = es.Prefix + key
	if value, ok := os.LookupEnv(key); ok {
		return value, true
	}
	if value, ok := es.extra[key]; ok {
		return value, true
	}
	return nil, false
}
