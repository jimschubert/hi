package daemon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	"github.com/jimschubert/hi/internal/daemon/store"
	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
	"github.com/jimschubert/hi/internal/proto/gen/hi/v1/v1connect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (d *Daemon) serveIPC(ctx context.Context) error {
	if !d.enableIPC {
		return nil
	}

	socketPath := d.config.SocketPath()

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	d.logger.Info("IPC listening", "socket", socketPath)

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

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// NOTE: if *this* fails, we want to keep going to serve MCP
			d.logger.Error("IPC server error", "err", err)
		}
	}()

	d.shutdownHooks = append(d.shutdownHooks, func() {
		d.logger.Debug("Shutting down IPC server")
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	})

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
	d.notifyFn(request.Title, "Agent is awaiting your response.")

	// NOTE: this file is daemon/ipc.go, which is the right place (every time I look at it, I think it's wrong).
	// This is the daemon's IPC handler, so we need to add any incoming request to the queue so the user
	// can provide feedback. Enqueue internally waits for the result, so we can just return it directly here.
	resp, err := d.queue.Enqueue(ctx, &store.PendingRequest{
		Type:        protoToRequestType(request.Type),
		AgentName:   request.AgentName,
		Title:       request.Title,
		Prompt:      request.Prompt,
		DefaultVal:  request.DefaultVal,
		Choices:     request.Choices,
		MultiSelect: request.MultiSelect,
	})
	if err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	return &v1.SubmitRequestResponse{
		TextValue:    resp.TextValue,
		BoolValue:    resp.BoolValue,
		ChoiceValues: resp.ChoiceValues,
		Cancelled:    resp.Cancelled,
	}, nil
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

func protoToRequestType(t v1.RequestType) store.RequestType {
	switch t {
	case v1.RequestType_REQUEST_TYPE_TEXT:
		return store.RequestTypeText
	case v1.RequestType_REQUEST_TYPE_MULTILINE:
		return store.RequestTypeMultiline
	case v1.RequestType_REQUEST_TYPE_CHOICE:
		return store.RequestTypeChoice
	case v1.RequestType_REQUEST_TYPE_CONFIRM:
		return store.RequestTypeConfirm
	case v1.RequestType_REQUEST_TYPE_NOTIFY:
		return store.RequestTypeNotify
	default:
		return store.RequestTypeText
	}
}
