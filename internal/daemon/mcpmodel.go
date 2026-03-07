package daemon

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

type MultilineInput struct {
	AgentName    string `json:"agent_name,omitempty"`
	Title        string `json:"title"`
	Prompt       string `json:"prompt"`
	DefaultValue string `json:"default_value,omitempty"`
}

type MultilineOutput struct {
	Value     string `json:"value"`
	LineCount int    `json:"line_count"`
	Cancelled bool   `json:"cancelled"`
}
