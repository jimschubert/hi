package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
)

var (
	programName = "hi"
	version     = "dev"
	commit      = "unknown SHA"
)

var CLI struct {
	Version kong.VersionFlag `short:"v" help:"Print version information."`
}

func main() {
	logger := log.New(os.Stdout, "[hi] ", 0)
	ctx := kong.Parse(&CLI,
		kong.Name(programName),
		kong.Description("Human Intelligence — MCP server for human-in-the-loop agent interactions"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": fmt.Sprintf("%s %s (commit: %s)", programName, version, commit),
		},
		kong.Bind(logger),
	)

	if ctx.Command() == "" {
		logger.Println("No command provided. Use --help for usage information.")
		os.Exit(1)
	}

	err := ctx.Run(context.Background())
	ctx.FatalIfErrorf(err)
}
