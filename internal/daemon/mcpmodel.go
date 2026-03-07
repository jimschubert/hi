package daemon

type AskInput struct {
	AgentName    string `json:"agent_name,omitempty" jsonschema:"description=The name of the calling agent (optional)"`
	Title        string `json:"title" jsonschema:"description=Short title for the prompt dialog"`
	Prompt       string `json:"prompt" jsonschema:"description=The question or instruction shown to the user"`
	DefaultValue string `json:"default_value,omitempty" jsonschema:"description=Optional pre-filled text"`
}

type AskOutput struct {
	Value     string `json:"value"     jsonschema:"description=The text entered by the user"`
	Cancelled bool   `json:"cancelled" jsonschema:"description=True if the user dismissed the dialog without responding"`
}

type MultilineInput struct {
	AgentName    string `json:"agent_name,omitempty" jsonschema:"description=The name of the calling agent (optional)"`
	Title        string `json:"title" jsonschema:"description=Short title for the prompt dialog"`
	Prompt       string `json:"prompt" jsonschema:"description=The question or instruction shown to the user"`
	DefaultValue string `json:"default_value,omitempty" jsonschema:"description=Optional pre-filled text"`
}

type MultilineOutput struct {
	Value     string `json:"value" jsonschema:"description=The full multi-line text entered by the user"`
	LineCount int    `json:"line_count" jsonschema:"description=Number of lines in the response"`
	Cancelled bool   `json:"cancelled"  jsonschema:"description=True if the user dismissed the dialog without responding"`
}

type ChooseInput struct {
	AgentName   string   `json:"agent_name,omitempty" jsonschema:"description=The name of the calling agent (optional)"`
	Title       string   `json:"title" jsonschema:"description=Short title for the prompt dialog"`
	Prompt      string   `json:"prompt" jsonschema:"description=The question or instruction shown to the user"`
	Choices     []string `json:"choices" jsonschema:"description=List of options to present to the user"`
	MultiSelect bool     `json:"multi_select,omitempty" jsonschema:"description=Allow selecting more than one option"`
}

type ChooseOutput struct {
	Selected  []string `json:"selected"  jsonschema:"description=The option or options chosen by the user"`
	Cancelled bool     `json:"cancelled" jsonschema:"description=True if the user dismissed the dialog without responding"`
}

type ConfirmInput struct {
	AgentName string `json:"agent_name,omitempty" jsonschema:"description=The name of the calling agent (optional)"`
	Title     string `json:"title" jsonschema:"description=Short title for the prompt dialog"`
	Message   string `json:"message" jsonschema:"description=The question requiring a yes/no answer"`
}

type ConfirmOutput struct {
	Confirmed bool `json:"confirmed" jsonschema:"description=True if the user confirmed"`
	Cancelled bool `json:"cancelled" jsonschema:"description=True if the user dismissed the dialog without responding"`
}
