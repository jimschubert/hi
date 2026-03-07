package daemon

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
	"github.com/jimschubert/hi/internal/proto/gen/hi/v1/v1connect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (d *Daemon) serveIPC(ctx context.Context) any {
	socketPath := d.config.SocketPath()

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	slog.Info("IPC listening", "socket", socketPath)

	// see https://connectrpc.com/docs/go/getting-started
	mux := http.NewServeMux()
	mux.Handle(v1connect.NewHiServiceHandler(d, connect.WithInterceptors(validate.NewInterceptor())))

	// Mount both versions of the gRPC reflection API so tools like grpcurl work.
	// grpcurl and many other tools still use the v1alpha API.
	reflector := grpcreflect.NewStaticReflector(v1connect.HiServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	srv := &http.Server{Handler: h2c.NewHandler(withCORS(mux), &http2.Server{})}
	srv.BaseContext = func(listener net.Listener) context.Context {
		return ctx
	}

	d.shutdownHooks = append(d.shutdownHooks, func() {
		d.logger.Debug("Shutting down IPC server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	})

	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (d *Daemon) Ping(ctx context.Context, request *v1.PingRequest) (*v1.PingResponse, error) {
	return &v1.PingResponse{Version: daemonVersion}, nil
}

func (d *Daemon) GetStatus(ctx context.Context, request *v1.GetStatusRequest) (*v1.GetStatusResponse, error) {
	return &v1.GetStatusResponse{
		Running:         true,
		PendingRequests: 0,
		McpAddress:      d.mcpAddr,
		UptimeSeconds:   int64(time.Since(d.startedAt).Seconds()),
		Version:         daemonVersion,
	}, nil
}

func (d *Daemon) Shutdown(ctx context.Context, request *v1.ShutdownRequest) (*v1.ShutdownResponse, error) {
	defer d.cancel()
	return &v1.ShutdownResponse{}, nil
}

func (d *Daemon) SubmitRequest(ctx context.Context, request *v1.SubmitRequestRequest) (*v1.SubmitRequestResponse, error) {
	// TODO: Implement values. need to enqueue, respond, then return values here.
	backend := RandomResponseBackend{}

	resp := &v1.SubmitRequestResponse{
		TextValue:    "",
		BoolValue:    false,
		ChoiceValues: []string{},
		Cancelled:    false,
	}

	switch request.Type {
	case v1.RequestType_REQUEST_TYPE_CONFIRM:
		confirm, cancel, err := backend.SubmitConfirm(ctx, request.AgentName, request.Title, request.Prompt)
		if err != nil {
			d.logger.Warn("failed: SubmitConfirm", "err", err)
			return nil, errors.New("submission failed")
		}
		resp.BoolValue = confirm
		resp.Cancelled = cancel
	case v1.RequestType_REQUEST_TYPE_TEXT:
		val, cancel, err := backend.SubmitText(ctx, request.AgentName, request.Title, request.Prompt, request.DefaultVal)
		if err != nil {
			d.logger.Warn("failed: SubmitText", "err", err)
			return nil, errors.New("submission failed")
		}
		resp.Cancelled = cancel
		resp.TextValue = val
	case v1.RequestType_REQUEST_TYPE_MULTILINE:
		val, lines, cancel, err := backend.SubmitMultiline(ctx, request.AgentName, request.Title, request.Prompt, request.DefaultVal)
		if err != nil {
			d.logger.Warn("failed: SubmitMultiline", "err", err)
			return nil, errors.New("submission failed")
		}
		d.logger.Debug("SubmitMultiline result", "lines", lines)
		resp.Cancelled = cancel
		resp.TextValue = val
	case v1.RequestType_REQUEST_TYPE_CHOICE:
		selected, cancel, err := backend.SubmitChoice(ctx, request.AgentName, request.Title, request.Prompt, request.Choices, request.MultiSelect)
		if err != nil {
			d.logger.Warn("failed: SubmitChoice", "err", err)
			return nil, errors.New("submission failed")
		}
		resp.Cancelled = cancel
		resp.ChoiceValues = selected
	case v1.RequestType_REQUEST_TYPE_UNSPECIFIED:
		d.logger.Warn("received an unexpected request type",
			"agent", request.AgentName,
			"prompt", request.Prompt,
		)
	}

	return resp, nil
}

// withCORS adds CORS support to a Connect HTTP handler.
func withCORS(h http.Handler) http.Handler {
	// taken from https://connectrpc.com/docs/go/deployment#h2c
	middleware := cors.New(cors.Options{
		AllowedOrigins: []string{"example.com"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	return middleware.Handler(h)
}
