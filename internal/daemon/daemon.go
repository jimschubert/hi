package daemon

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/daemon/store"
	"github.com/jimschubert/hi/internal/daemon/ui"
)

var (
	daemonVersion = "dev"
)

type Daemon struct {
	// mcpAddr is the address the MCP server is listening on.
	mcpAddr string

	// enableIPC determines whether to enable IPC over unix socket
	enableIPC bool

	// cancel stops all goroutines cleanly.
	cancel context.CancelFunc

	config config.Config
	logger *slog.Logger

	app *ui.Hi

	queue    *store.Queue
	notifyFn func(title, body string)

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
		mcpAddr:   mcpAddr,
		config:    options.config,
		enableIPC: options.enableIPC,
		logger: slog.New(slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: options.logLevel,
			},
		)),
		shutdownHooks: make([]func(), 0),
		queue:         store.NewQueue(),
		notifyFn:      func(_, _ string) {},
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
	if err := d.validate(); err != nil {
		return err
	}

	d.logger.Info("hi daemon started", "addr", cmp.Or(d.mcpAddr, "<empty>"))
	ctx = config.StoreMcpAddress(ctx, d.mcpAddr)
	ctx = config.StoreDaemonVersion(ctx, daemonVersion)

	ctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel
	d.startedAt = time.Now()

	// can only run once instance of daemon
	socketPath := d.config.SocketPath()
	_ = os.Remove(socketPath)

	if err := d.serveMCP(ctx); err != nil {
		return err
	}

	if err := d.serveIPC(ctx); err != nil {
		return err
	}

	go func() {
		err := d.handleSignals(ctx)
		if err != nil {
			d.logger.Error("handle signals returned error", "error", err)
		}
	}()

	d.app = ui.NewHi(ctx, d.queue)

	d.notifyFn = d.app.Notify

	// Fyne UI (blocks main thread until quit)
	d.app.Run()

	d.logger.Info("hi daemon stopped", "uptime", time.Since(d.startedAt).String())
	return nil
}

func (d *Daemon) validate() error {
	if d.mcpAddr == "" && !d.enableIPC {
		return fmt.Errorf("at least one of MCP or IPC must be enabled")
	}
	return nil
}
