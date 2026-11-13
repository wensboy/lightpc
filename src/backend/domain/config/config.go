package config

const (
	SourceFile uint8 = iota + 1
	SourceEnv
	SourceFlag
	SourceConfig
)

type Source interface {
	Kind() uint8
	Lookup(string) (any, bool)
}
