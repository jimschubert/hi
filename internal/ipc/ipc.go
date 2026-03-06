package ipc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/jimschubert/hi/internal/config"
	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
	"github.com/jimschubert/hi/internal/proto/gen/hi/v1/v1connect"
)

type Client struct {
	rpc     v1connect.HiServiceClient
	timeout time.Duration
}

func NewClient(config config.Config) (*Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", config.SocketPath())
			},
		},
	}

	rpc := v1connect.NewHiServiceClient(httpClient,
		"http://localhost",
		connect.WithGRPC(),
	)

	return &Client{
		rpc:     rpc,
		timeout: time.Duration(config.ClientTimeout()) * time.Second,
	}, nil
}

func (c *Client) Ping(ctx context.Context) (*v1.PingResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.rpc.Ping(ctx, &v1.PingRequest{})
	if err != nil {
		return nil, fmt.Errorf("error calling ping: %w", err)
	}
	return resp, nil
}

func (c *Client) Status(ctx context.Context) (*v1.GetStatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.rpc.GetStatus(ctx, &v1.GetStatusRequest{})
	if err != nil {
		return nil, fmt.Errorf("error calling status: %w", err)
	}
	return resp, nil
}
