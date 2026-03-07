package daemon

import (
	"context"
	"fmt"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/ipc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RunStdioServer(ctx context.Context, conf config.Config) error {
	if !IsRunning(conf) {
		return fmt.Errorf("hi daemon is not running")
	}

	client, err := ipc.NewClient(conf)
	if err != nil {
		return fmt.Errorf("failed to create IPC client: %w", err)
	}

	// communicate with daemon over IPC
	backend := NewIPCBackend(client)

	// create and register stdio mcp server. this allows for communication over command line
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "hi",
		Version: daemonVersion,
	}, nil)

	registerTools(server, backend)

	return server.Run(ctx, &mcp.StdioTransport{})
}
