package main

import (
	"context"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/daemon"
)

type DaemonCmd struct {
	Addr string `default:"" help:"TCP address for the MCP HTTP server (example: localhost:45678)."`
}

func (c *DaemonCmd) Run(conf config.Config) error {
	ctx := context.Background()
	d := daemon.New(c.Addr, daemon.WithLogLevel(conf.DaemonLogLevel()))
	return d.Start(ctx)
}
