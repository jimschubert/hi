package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jimschubert/hi/internal/config"
	"github.com/sethvargo/go-envconfig"
)

var (
	programName = "hi"
	version     = "dev"
	commit      = "unknown SHA"
)

var CLI struct {
	Daemon  DaemonCmd        `cmd:"" help:"Run the hi daemon."`
	Status  StatusCmd        `cmd:"" help:"Query status of the hi daemon"`
	Version kong.VersionFlag `short:"v" help:"Print version information."`
	Default DefaultCmd       `hidden:"" cmd:"" default:"withargs" help:"Ensures the daemon is running, starting if it's not (default)."`
}

func main() {
	logger := log.New(os.Stdout, "[hi] ", 0)

	conf := processConfig()
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
		kong.Bind(
			logger,
			conf,
		),
	)

	err := ctx.Run(context.Background())
	ctx.FatalIfErrorf(err)
}

func processConfig() config.Config {
	c := config.Config{}
	err := envconfig.Process(context.Background(), &c)
	if err != nil {
		fmt.Printf("error processing config: %s\n", err)
		os.Exit(1)
	}
	return c
}
