package main

import (
	"embed"
	"os"

	"github.com/wensboy/lightpc/backend/api/cmd"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

//go:embed all:frontend/dist
var dist embed.FS

//go:embed all:data
var data embed.FS

func main() {
	os.Exit(cmd.Execute(domainContext.EmbedContext{
		"/frontend/dist": &dist,
		"/data":          &data,
	}))
}
