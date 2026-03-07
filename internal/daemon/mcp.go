package daemon

import (
	"cmp"
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AskInput struct {
	AgentName    string `json:"agent_name,omitempty" jsonschema:"The name of the calling agent (optional)"`
	Title        string `json:"title" jsonschema:"Short title for the prompt dialog"`
	Prompt       string `json:"prompt" jsonschema:"The question or instruction shown to the user"`
	DefaultValue string `json:"default_value,omitempty" jsonschema:"Optional pre-filled text"`
}

type AskOutput struct {
	Value     string `json:"value"`
	Cancelled bool   `json:"cancelled"`
}

func RegisterTools(server *mcp.Server, backend RequestBackend) {
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

}
