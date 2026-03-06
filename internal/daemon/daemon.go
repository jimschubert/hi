package daemon

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type Daemon struct {
	// mcpAddr is the address the MCP server is listening on.
	mcpAddr string

	// cancel stops all goroutines cleanly.
	cancel context.CancelFunc

	logger *slog.Logger

	startedAt time.Time
}

func New(mcpAddr string, opts ...Option) *Daemon {
	options := &daemonOpts{
		logLevel: slog.LevelError,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &Daemon{
		mcpAddr: mcpAddr,
		logger:  slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: options.logLevel})),
	}
}

func (d *Daemon) Start(ctx context.Context) error {
	slog.Info("hi daemon started", slog.String("addr", d.mcpAddr))

	// TODO: we'll need a cancel function
	_, cancel := context.WithCancel(ctx)
	d.cancel = cancel
	d.startedAt = time.Now()

	// TODO: start MCP server, register tools, etc.

	slog.Info("hi daemon stopped", slog.String("uptime", time.Since(d.startedAt).String()))
	return nil
}
