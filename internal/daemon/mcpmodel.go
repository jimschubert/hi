package daemon

type AskInput struct {
	AgentName    string `json:"agent_name,omitempty" jsonschema:"The name of the calling agent (optional)"`
	Title        string `json:"title" jsonschema:"Short title for the prompt dialog"`
	Prompt       string `json:"prompt" jsonschema:"The question or instruction shown to the user"`
	DefaultValue string `json:"default_value,omitempty" jsonschema:"Optional pre-filled text"`
}

type AskOutput struct {
	Value     string `json:"value"     jsonschema:"The text entered by the user"`
	Cancelled bool   `json:"cancelled" jsonschema:"True if the user dismissed the dialog without responding"`
}

type MultilineInput struct {
	AgentName    string `json:"agent_name,omitempty" jsonschema:"The name of the calling agent (optional)"`
	Title        string `json:"title" jsonschema:"Short title for the prompt dialog"`
	Prompt       string `json:"prompt" jsonschema:"The question or instruction shown to the user"`
	DefaultValue string `json:"default_value,omitempty" jsonschema:"Optional pre-filled text"`
}

type MultilineOutput struct {
	Value     string `json:"value" jsonschema:"The full multi-line text entered by the user"`
	LineCount int    `json:"line_count" jsonschema:"Number of lines in the response"`
	Cancelled bool   `json:"cancelled"  jsonschema:"True if the user dismissed the dialog without responding"`
}

type ChooseInput struct {
	AgentName   string   `json:"agent_name,omitempty" jsonschema:"The name of the calling agent (optional)"`
	Title       string   `json:"title" jsonschema:"Short title for the prompt dialog"`
	Prompt      string   `json:"prompt" jsonschema:"The question or instruction shown to the user"`
	Choices     []string `json:"choices" jsonschema:"List of options to present to the user"`
	MultiSelect bool     `json:"multi_select,omitempty" jsonschema:"Allow selecting more than one option"`
}

type ChooseOutput struct {
	Selected  []string `json:"selected"  jsonschema:"The option or options chosen by the user"`
	Cancelled bool     `json:"cancelled" jsonschema:"True if the user dismissed the dialog without responding"`
}

type ConfirmInput struct {
	AgentName string `json:"agent_name,omitempty" jsonschema:"The name of the calling agent (optional)"`
	Title     string `json:"title" jsonschema:"Short title for the prompt dialog"`
	Message   string `json:"message" jsonschema:"The question requiring a yes/no answer"`
}

type ConfirmOutput struct {
	Confirmed bool `json:"confirmed" jsonschema:"True if the user confirmed"`
	Cancelled bool `json:"cancelled" jsonschema:"True if the user dismissed the dialog without responding"`
}

type NotifyInput struct {
	AgentName string `json:"agent_name,omitempty" jsonschema:"The name of the calling agent (optional)"`
	Title     string `json:"title" jsonschema:"Short title for the prompt dialog"`
	Message   string `json:"message" jsonschema:"The full notification message to show the user"`
}

type NotifyOutput struct {
	Sent bool `json:"sent" jsonschema:"True if the notification was successfully delivered"`
}
