package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainConfg "github.com/wensboy/lightpc/backend/domain/context"
)

// FileSource 用于处理文件配置搜索和存储, 通常为静态文件配置, 无法变更.
type FileSource struct {
	raw  map[string]any // 原始
	flat map[string]any // 平铺
}

func NewFileSource() *FileSource {
	return &FileSource{
		raw: make(map[string]any),
	}
}

func (fs *FileSource) Parse(fsCtx domainConfg.FsContext) {
	fs.raw = make(map[string]any)
	fs.flat = nil
	for _, configPath := range fsCtx.Paths {
		if filepath.Ext(configPath) != ".json" {
			continue
		}
		var content []byte
		var err error
		if fsCtx.Fs != nil {
			content, err = fsCtx.Fs.ReadFile(strings.TrimPrefix(configPath, "/"))
		} else {
			content, err = os.ReadFile(configPath)
		}
		if err != nil {
			continue
		}
		var value map[string]any
		if err := json.Unmarshal(content, &value); err != nil {
			continue
		}
		name := filepath.Base(configPath)
		fs.raw[strings.TrimSuffix(name, filepath.Ext(name))] = value
	}
}

func (fs *FileSource) Flat() {
	type pending struct {
		key   string
		value any
	}
	flat := make(map[string]any)
	stack := make([]pending, 0, len(fs.raw))
	for key, value := range fs.raw {
		stack = append(stack, pending{key: key, value: value})
	}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		children, isObject := current.value.(map[string]any)
		if isObject {
			for key, value := range children {
				stack = append(stack, pending{key: current.key + "." + key, value: value})
			}
			continue
		}
		flat[current.key] = current.value
	}
	fs.flat = flat
}

func (es *FileSource) Kind() uint8 {
	return domainConfig.SourceEnv
}

func (fs *FileSource) Lookup(key string) (any, bool) {
	if fs.flat == nil {
		fs.Flat()
	}
	if value, ok := fs.flat[key]; ok {
		return value, true
	}
	var current any = fs.raw
	for segment := range strings.SplitSeq(key, ".") {
		object, isObject := current.(map[string]any)
		if !isObject {
			return nil, false
		}
		value, ok := object[segment]
		if !ok {
			return nil, false
		}
		current = value
	}
	return current, true
}
