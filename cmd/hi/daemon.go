package main

import (
	"context"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/daemon"
)

type DaemonCmd struct {
	Addr string `default:"" help:"TCP address for the MCP HTTP server (example: localhost:45678)."`
	IPC  bool   `default:"true" negatable:"" help:"Whether to start up the IPC server on a Unix socket (default: true)."`
}

func (c *DaemonCmd) Run(conf config.Config) error {
	ctx := context.Background()
	d := daemon.New(c.Addr,
		daemon.WithLogLevel(conf.DaemonLogLevel()),
		daemon.WithConfig(conf),
		daemon.WithIPCEnabled(c.IPC),
	)
	return d.Start(ctx)
}
