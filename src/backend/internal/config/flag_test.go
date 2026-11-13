package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

func TestParseFlagByType(t *testing.T) {
	source := NewFlagSource()
	flags := source.parseFlag([]domainConfig.Flag{
		{Name: "config", Usage: "config path", Required: true, Aliases: []string{"c"}, Value: domainConfig.Value{Type: "string", Default: "app.json"}},
		{Name: "verbose", Deprecated: true, Value: domainConfig.Value{Type: "bool", Default: true}},
		{Name: "port", Value: domainConfig.Value{Type: "int", Default: float64(8080)}},
		{Name: "limit", Value: domainConfig.Value{Type: "int64", Default: float64(10)}},
		{Name: "count", Value: domainConfig.Value{Type: "uint", Default: float64(3)}},
		{Name: "ratio", Value: domainConfig.Value{Type: "float64", Default: 0.5}},
		{Name: "plain", Value: domainConfig.Value{}},
	})

	if len(flags) != 7 {
		t.Fatalf("len(flags) = %d, want 7", len(flags))
	}
	if _, ok := flags[0].(*cli.StringFlag); !ok {
		t.Fatalf("flags[0] = %T, want *cli.StringFlag", flags[0])
	}
	if _, ok := flags[1].(*cli.BoolFlag); !ok {
		t.Fatalf("flags[1] = %T, want *cli.BoolFlag", flags[1])
	}
	if _, ok := flags[2].(*cli.IntFlag); !ok {
		t.Fatalf("flags[2] = %T, want *cli.IntFlag", flags[2])
	}
	if _, ok := flags[3].(*cli.Int64Flag); !ok {
		t.Fatalf("flags[3] = %T, want *cli.Int64Flag", flags[3])
	}
	if _, ok := flags[4].(*cli.UintFlag); !ok {
		t.Fatalf("flags[4] = %T, want *cli.UintFlag", flags[4])
	}
	if _, ok := flags[5].(*cli.Float64Flag); !ok {
		t.Fatalf("flags[5] = %T, want *cli.Float64Flag", flags[5])
	}
	if _, ok := flags[6].(*cli.StringFlag); !ok {
		t.Fatalf("flags[6] = %T, want *cli.StringFlag", flags[6])
	}
	if value := flags[0].Get(); value != "app.json" {
		t.Fatalf("flags[0].Get() = %v", value)
	}
	if value := flags[1].Get(); value != true {
		t.Fatalf("flags[1].Get() = %v", value)
	}
	if value := flags[2].Get(); value != 8080 {
		t.Fatalf("flags[2].Get() = %v", value)
	}
	if value := flags[3].Get(); value != int64(10) {
		t.Fatalf("flags[3].Get() = %v", value)
	}
	if value := flags[4].Get(); value != uint(3) {
		t.Fatalf("flags[4].Get() = %v", value)
	}
	if value := flags[5].Get(); value != 0.5 {
		t.Fatalf("flags[5].Get() = %v", value)
	}
}

func TestParseFlagDistinguishesNumberTypes(t *testing.T) {
	source := NewFlagSource()
	flags := source.parseFlag([]domainConfig.Flag{
		{Name: "count", Value: domainConfig.Value{Type: "int", Default: float64(1)}},
		{Name: "ratio", Value: domainConfig.Value{Type: "float64", Default: float64(1)}},
	})

	if _, ok := flags[0].(*cli.IntFlag); !ok {
		t.Fatalf("flags[0] = %T, want *cli.IntFlag", flags[0])
	}
	if _, ok := flags[1].(*cli.Float64Flag); !ok {
		t.Fatalf("flags[1] = %T, want *cli.Float64Flag", flags[1])
	}
	if value := flags[0].Get(); value != 1 {
		t.Fatalf("flags[0].Get() = %v", value)
	}
	if value := flags[1].Get(); value != 1.0 {
		t.Fatalf("flags[1].Get() = %v", value)
	}
}

func TestParseCommandAndLookup(t *testing.T) {
	commandPath := filepath.Join(t.TempDir(), "command.json")
	content := `{
	"version": "1.0.0",
	"command": {
		"name": "app",
		"usage": "run the app",
		"flags": [{"name": "config", "value": {"type": "string", "default": "app.json"}}],
		"subCommands": [
			{"name": "serve", "flags": [{"name": "port", "value": {"type": "int", "default": 8080}}]}
		]
	}
}`
	if err := os.WriteFile(commandPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	source := NewFlagSource()
	if err := source.ParseCommand(domainContext.FsContext{Paths: []string{commandPath}}); err != nil {
		t.Fatal(err)
	}

	root, ok := source.Lookup("app")
	rootCommand, isCommand := root.(*cli.Command)
	if !ok || !isCommand || rootCommand.Name != "app" {
		t.Fatalf("app = %v, %v", root, ok)
	}
	if value, ok := source.Lookup("app.config"); !ok || value != "app.json" {
		t.Fatalf("app.config = %v, %v", value, ok)
	}
	if value, ok := source.Lookup("app.serve.port"); !ok || value != 8080 {
		t.Fatalf("app.serve.port = %v, %v", value, ok)
	}
	command, ok := source.Lookup("app.serve")
	subCommand, isCommand := command.(*cli.Command)
	if !ok || !isCommand || subCommand.Name != "serve" {
		t.Fatalf("app.serve = %v, %v", command, ok)
	}
	if _, ok := source.Lookup("app.missing"); ok {
		t.Fatal("a missing key must not be found")
	}
	if _, ok := source.Lookup("nope.config"); ok {
		t.Fatal("a missing command must not be found")
	}
}

func TestFlagSourceWithoutCommand(t *testing.T) {
	if _, ok := NewFlagSource().Lookup("app.config"); ok {
		t.Fatal("an empty source must not find anything")
	}
}

func TestParseCommandMissingFile(t *testing.T) {
	source := NewFlagSource()
	if err := source.ParseCommand(domainContext.FsContext{Paths: []string{filepath.Join(t.TempDir(), "absent.json")}}); err == nil {
		t.Fatal("a missing file must return an error")
	}
}
