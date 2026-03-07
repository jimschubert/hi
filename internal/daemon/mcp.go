package daemon

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (d *Daemon) serveMCP(_ context.Context) error {
	if d.mcpAddr != "" {
		mcpServer := mcp.NewServer(&mcp.Implementation{
			Name:    "hi",
			Version: daemonVersion,
		}, &mcp.ServerOptions{Logger: d.logger})

		registerTools(mcpServer, RandomResponseBackend{})

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
	return nil
}

func registerTools(server *mcp.Server, backend RequestBackend) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name: "hi_ask",
			Description: "Ask the human developer for a single-line text response. " +
				"Use this instead of stopping the session to ask a clarifying question.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, in AskInput) (*mcp.CallToolResult, AskOutput, error) {
			agentName := cmp.Or(in.AgentName, "Agent")
			value, cancelled, err := backend.SubmitText(ctx, agentName, in.Title, in.Prompt, in.DefaultValue)
			if err != nil {
				return nil, AskOutput{Cancelled: true}, nil
			}
			return nil, AskOutput{Value: value, Cancelled: cancelled}, nil
		},
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "hi_multiline",
			Description: "Ask the human developer for a multi-line text response, " +
				"such as a description, code snippet, or detailed feedback.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, in MultilineInput) (*mcp.CallToolResult, MultilineOutput, error) {
			agentName := cmp.Or(in.AgentName, "Agent")
			value, lines, cancelled, err := backend.SubmitMultiline(ctx, agentName, in.Title, in.Prompt, in.DefaultValue)
			if err != nil {
				return nil, MultilineOutput{Cancelled: true}, nil
			}
			return nil, MultilineOutput{
				Value:     value,
				LineCount: lines,
				Cancelled: cancelled,
			}, nil
		},
	)

	mcp.AddTool(server,
		&mcp.Tool{
			Name: "hi_choose",
			Description: "Present a list of choices to the human developer and return " +
				"their selection(s). Use for branching decisions.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, in ChooseInput) (*mcp.CallToolResult, ChooseOutput, error) {
			agentName := cmp.Or(in.AgentName, "Agent")
			selected, cancelled, err := backend.SubmitChoice(ctx, agentName, in.Title, in.Prompt, in.Choices, in.MultiSelect)
			if err != nil {
				return nil, ChooseOutput{Cancelled: true}, nil
			}
			return nil, ChooseOutput{
				Selected:  selected,
				Cancelled: cancelled,
			}, nil
		},
	)
}
