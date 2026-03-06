package daemon

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jimschubert/hi/internal/config"
)

var (
	daemonVersion = "dev"
)

type Daemon struct {
	// mcpAddr is the address the MCP server is listening on.
	mcpAddr string

	// cancel stops all goroutines cleanly.
	cancel context.CancelFunc

	config config.Config
	logger *slog.Logger

	startedAt time.Time

	shutdownHooks []func()
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
		config:  options.config,
		logger: slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: options.logLevel,
			},
		)),
		shutdownHooks: make([]func(), 0),
	}
}

func (d *Daemon) handleSignals(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	for {
		select {
		case sig := <-sigChan:
			d.logger.Warn("Signal received", "signal", sig)
			if d.cancel != nil {
				d.cancel()
				continue
			}

		case <-ctx.Done():
			d.logger.Debug("Shutting down...")
			for _, hook := range d.shutdownHooks {
				hook()
			}

			return nil
		}
	}
}

func (d *Daemon) Start(ctx context.Context) error {
	slog.Info("hi daemon started", "addr", cmp.Or(d.mcpAddr, "<empty>"))

	ctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel
	d.startedAt = time.Now()

	// can only run once instance of daemon
	socketPath := d.config.SocketPath()
	_ = os.Remove(socketPath)

	// TODO: start MCP server, register tools, etc.

	go func() {
		if err := d.serveIPC(ctx); err != nil {
			slog.Error("IPC server error", "err", err)
		}
	}()

	// block until interrupt or ctx.Done(); calls shutdown hooks
	_ = d.handleSignals(ctx)

	slog.Info("hi daemon stopped", "uptime", time.Since(d.startedAt).String())
	return nil
}
