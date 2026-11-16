package config

// Value 映射 command:flag:value 结构用于 value 处理.
type Value struct {
	Type    string `json:"type"`
	Default any    `json:"default"`
}

// Flag 映射 command:flag 结构用于解析处理 flag.
type Flag struct {
	Name       string   `json:"name"`
	Key        string   `json:"key"`
	Usage      string   `json:"usage"`
	Required   bool     `json:"required"`
	Hidden     bool     `json:"hidden"`
	Deprecated bool     `json:"deprecated"`
	Local      bool     `json:"local"`
	OnlyOnce   bool     `json:"onlyOnce"`
	Aliases    []string `json:"aliases"`
	Value      Value    `json:"value"`
}

// Command 映射 command:Command 结构用于解析处理 command.
type Command struct {
	Name        string     `json:"name"`
	Key         string     `json:"key"`
	Usage       string     `json:"usage"`
	UsageText   string     `json:"usageText"`
	Description string     `json:"description"`
	Flags       []Flag     `json:"flags"`
	SubCommands []*Command `json:"subCommands"`
}

// Cli 映射整个 command file 结构.
type Cli struct {
	Version     string  `json:"version"`
	Description string  `json:"description"`
	Command     Command `json:"command"`
}

type LogSerializer struct {
	Type         string `json:"type"`
	EnableSource bool   `json:"enableSource"`
}

type LogOutput struct {
	Serializer LogSerializer `json:"serializer"`
	Out        string        `json:"out"`
	Path       string        `json:"path"`
	Mode       string        `json:"mode"`
}

// LogItem is one entry of runtime.log.<group>. Output is an object for the direct group and an array
// for the fanout and pattern groups, so it stays untyped until the log engine narrows it.
type LogItem struct {
	Name    string `json:"name"`
	Level   string `json:"level"`
	Handler string `json:"handler"`
	Output  any    `json:"output"`
}
