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
