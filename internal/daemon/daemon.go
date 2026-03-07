package daemon

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jimschubert/hi/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	// TODO: Move to another function?
	if d.mcpAddr != "" {
		mcpServer := mcp.NewServer(&mcp.Implementation{
			Name:    "hi",
			Version: daemonVersion,
		}, &mcp.ServerOptions{Logger: d.logger})

		RegisterTools(mcpServer, RandomResponseBackend{})

		handler := mcp.NewStreamableHTTPHandler(
			func(_ *http.Request) *mcp.Server { return mcpServer },
			&mcp.StreamableHTTPOptions{Logger: d.logger},
		)

		mcpHttpServer := &http.Server{
			Addr: d.mcpAddr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/mcp", "/mcp/":
					handler.ServeHTTP(w, r)
				case "/health":
					w.WriteHeader(http.StatusOK)
					fmt.Println(w, `{"ok":true}`)
				default:
					http.NotFound(w, r)
				}
			}),
		}

		// listen _first_ so we don't fail in goroutine
		ln, err := net.Listen("tcp", d.mcpAddr)
		if err != nil {
			d.logger.Warn("hi: couldn't MCP address, not starting remaining services.", "addr", d.mcpAddr, "error", err)
			return fmt.Errorf("cannot bind MCP address: %s %w", d.mcpAddr, err)
		}

		go func() {
			d.logger.Info("MCP server listening", "addr", d.mcpAddr, "path", "/mcp")
			if err := mcpHttpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				// NOTE: if *this* fails, we want to keep going to serve IPC
				d.logger.Error("MCP HTTP server error", "err", err)
			}
		}()

		d.shutdownHooks = append(d.shutdownHooks, func() {
			shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutCancel()
			_ = mcpHttpServer.Shutdown(shutCtx)
		})
	}

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
