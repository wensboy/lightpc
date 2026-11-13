package config

import (
	"encoding/json"
	"os"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

type FlagSource struct {
	cmd *cli.Command // 完整 command 结构, 从结构中获取 flag
}

func NewFlagSource() *FlagSource {
	return &FlagSource{}
}

func (fs *FlagSource) WithCmd(cmd *cli.Command) *FlagSource {
	fs.cmd = cmd
	return fs
}

func (es *FlagSource) Kind() uint8 {
	return domainConfig.SourceFlag
}

func (fs *FlagSource) Lookup(key string) (any, bool) {
	if fs.cmd == nil {
		return nil, false
	}
	segments := strings.Split(key, ".")
	if segments[0] == fs.cmd.Name {
		segments = segments[1:]
	}
	if len(segments) == 0 {
		return fs.cmd, true
	}
	current := fs.cmd
	for _, segment := range segments[:len(segments)-1] {
		current = current.Command(segment)
		if current == nil {
			return nil, false
		}
	}
	leaf := segments[len(segments)-1]
	for _, flag := range current.Flags {
		if slices.Contains(flag.Names(), leaf) {
			return flag.Get(), true
		}
	}
	if command := current.Command(leaf); command != nil {
		return command, true
	}
	return nil, false
}

func (fs *FlagSource) ParseCommand(fsCtx domainContext.FsContext) error {
	for _, commandPath := range fsCtx.Paths {
		var content []byte
		var err error
		if fsCtx.Fs != nil {
			content, err = fsCtx.Fs.ReadFile(strings.TrimPrefix(commandPath, "/"))
		} else {
			content, err = os.ReadFile(commandPath)
		}
		if err != nil {
			return err
		}
		var raw domainConfig.Cli
		if err := json.Unmarshal(content, &raw); err != nil {
			return err
		}
		root := &cli.Command{
			Name:        raw.Command.Name,
			Usage:       raw.Command.Usage,
			UsageText:   raw.Command.UsageText,
			Description: raw.Command.Description,
			Version:     raw.Version,
			Flags:       fs.parseFlag(raw.Command.Flags),
		}
		type pending struct {
			raw  *domainConfig.Command
			node *cli.Command
		}
		stack := []pending{{raw: &raw.Command, node: root}}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, subCommand := range current.raw.SubCommands {
				if subCommand == nil {
					continue
				}
				child := &cli.Command{
					Name:        subCommand.Name,
					Usage:       subCommand.Usage,
					UsageText:   subCommand.UsageText,
					Description: subCommand.Description,
					Flags:       fs.parseFlag(subCommand.Flags),
				}
				current.node.Commands = append(current.node.Commands, child)
				stack = append(stack, pending{raw: subCommand, node: child})
			}
		}
		fs.cmd = root
	}
	return nil
}

func (fs *FlagSource) parseFlag(rawFlags []domainConfig.Flag) []cli.Flag {
	flags := make([]cli.Flag, 0, len(rawFlags))
	for _, rawFlag := range rawFlags {
		deprecated := ""
		if rawFlag.Deprecated {
			deprecated = "deprecated: " + rawFlag.Name
		}
		switch rawFlag.Value.Type {
		case "bool":
			value, _ := cast.As[bool](rawFlag.Value.Default)
			flags = append(flags, &cli.BoolFlag{
				Name:       rawFlag.Name,
				Usage:      rawFlag.Usage,
				Required:   rawFlag.Required,
				Hidden:     rawFlag.Hidden,
				Deprecated: deprecated,
				Local:      rawFlag.Local,
				OnlyOnce:   rawFlag.OnlyOnce,
				Aliases:    rawFlag.Aliases,
				Value:      value,
			})
		case "int":
			value, _ := cast.As[int](rawFlag.Value.Default)
			flags = append(flags, &cli.IntFlag{
				Name:       rawFlag.Name,
				Usage:      rawFlag.Usage,
				Required:   rawFlag.Required,
				Hidden:     rawFlag.Hidden,
				Deprecated: deprecated,
				Local:      rawFlag.Local,
				OnlyOnce:   rawFlag.OnlyOnce,
				Aliases:    rawFlag.Aliases,
				Value:      value,
			})
		case "int64":
			value, _ := cast.As[int64](rawFlag.Value.Default)
			flags = append(flags, &cli.Int64Flag{
				Name:       rawFlag.Name,
				Usage:      rawFlag.Usage,
				Required:   rawFlag.Required,
				Hidden:     rawFlag.Hidden,
				Deprecated: deprecated,
				Local:      rawFlag.Local,
				OnlyOnce:   rawFlag.OnlyOnce,
				Aliases:    rawFlag.Aliases,
				Value:      value,
			})
		case "uint":
			value, _ := cast.As[uint](rawFlag.Value.Default)
			flags = append(flags, &cli.UintFlag{
				Name:       rawFlag.Name,
				Usage:      rawFlag.Usage,
				Required:   rawFlag.Required,
				Hidden:     rawFlag.Hidden,
				Deprecated: deprecated,
				Local:      rawFlag.Local,
				OnlyOnce:   rawFlag.OnlyOnce,
				Aliases:    rawFlag.Aliases,
				Value:      value,
			})
		case "float64":
			value, _ := cast.As[float64](rawFlag.Value.Default)
			flags = append(flags, &cli.Float64Flag{
				Name:       rawFlag.Name,
				Usage:      rawFlag.Usage,
				Required:   rawFlag.Required,
				Hidden:     rawFlag.Hidden,
				Deprecated: deprecated,
				Local:      rawFlag.Local,
				OnlyOnce:   rawFlag.OnlyOnce,
				Aliases:    rawFlag.Aliases,
				Value:      value,
			})
		default:
			value, _ := cast.As[string](rawFlag.Value.Default)
			flags = append(flags, &cli.StringFlag{
				Name:       rawFlag.Name,
				Usage:      rawFlag.Usage,
				Required:   rawFlag.Required,
				Hidden:     rawFlag.Hidden,
				Deprecated: deprecated,
				Local:      rawFlag.Local,
				OnlyOnce:   rawFlag.OnlyOnce,
				Aliases:    rawFlag.Aliases,
				Value:      value,
			})
		}
	}
	return flags
}
