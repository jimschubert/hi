package daemon

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/validate"
	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
	"github.com/jimschubert/hi/internal/proto/gen/hi/v1/v1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func (d *Daemon) serveIPC(ctx context.Context) interface{} {
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

	srv := &http.Server{Handler: h2c.NewHandler(mux, &http2.Server{})}

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
	// TODO implement me
	panic("implement me")
}

func (d *Daemon) Shutdown(ctx context.Context, request *v1.ShutdownRequest) (*v1.ShutdownResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (d *Daemon) SubmitRequest(ctx context.Context, request *v1.SubmitRequestRequest) (*v1.SubmitRequestResponse, error) {
	// TODO implement me
	panic("implement me")
}
