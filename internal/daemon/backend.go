package daemon

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
)

type RequestBackend interface {
	// SubmitText will submit details for hi_ask tooling
	SubmitText(ctx context.Context, agentName, title, prompt, defaultVal string) (value string, cancelled bool, err error)

	// SubmitMultiline will submit details for hi_multiline tooling
	SubmitMultiline(ctx context.Context, agentName, title, prompt, defaultVal string) (value string, lineCount int, cancelled bool, err error)

	// SubmitChoice will submit details for hi_choice tooling
	SubmitChoice(ctx context.Context, agentName, title, prompt string, choices []string, multiSelect bool) (selected []string, cancelled bool, err error)

	// SubmitConfirm will submit details for hi_confirm tooling
	SubmitConfirm(ctx context.Context, agentName, title, message string) (confirmed, cancelled bool, err error)

	// SubmitNotify will submit details for hi_notify tooling
	SubmitNotify(ctx context.Context, agentName, title, message string) error
}

// RandomResponseBackend allows for fake data to work out API interactions before human interactions
type RandomResponseBackend struct {
}

func (r RandomResponseBackend) randomlyShould() bool {
	return rand.IntN(10) == 0 || rand.IntN(20) >= 19
}

func (r RandomResponseBackend) SubmitText(ctx context.Context, agentName, title, prompt, defaultVal string) (value string, cancelled bool, err error) {
	if r.randomlyShould() {
		return "", true, nil
	}
	return randomSentence(rand.IntN(8) + 1), false, nil
}

func (r RandomResponseBackend) SubmitMultiline(ctx context.Context, agentName, title, prompt, defaultVal string) (value string, lineCount int, cancelled bool, err error) {
	if r.randomlyShould() {
		return "", 0, true, nil
	}

	n := rand.IntN(4) + 1
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("%d: %s", i+1, randomSentence(rand.IntN(5)+2))
	}
	out := strings.Join(lines, "\n")
	return out, n, false, nil
}

func (r RandomResponseBackend) SubmitChoice(ctx context.Context, agentName, title, prompt string, choices []string, multiSelect bool) (selected []string, cancelled bool, err error) {
	if len(choices) == 0 || r.randomlyShould() {
		return nil, true, nil
	}

	if !multiSelect {
		// pick one at random
		return []string{choices[rand.IntN(len(choices))]}, false, nil
	}

	// pick some at random
	shuffled := make([]string, len(choices))
	copy(shuffled, choices)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	n := rand.IntN(len(shuffled)) + 1
	return shuffled[:n], false, nil
}

func (r RandomResponseBackend) SubmitConfirm(ctx context.Context, agentName, title, message string) (confirmed, cancelled bool, err error) {
	if r.randomlyShould() {
		return false, true, nil
	}
	return rand.IntN(2) == 1, false, nil
}

func (r RandomResponseBackend) SubmitNotify(ctx context.Context, agentName, title, message string) error {
	// notifications are fire-and-forget; nothing to randomise
	return nil
}

type IPCClient interface {
	SubmitRequest(ctx context.Context, req *v1.SubmitRequestRequest) (*v1.SubmitRequestResponse, error)
}

type IPCBackend struct {
	client IPCClient
}

func NewIPCBackend(client IPCClient) RequestBackend {
	return &IPCBackend{client: client}
}

func (b *IPCBackend) SubmitText(ctx context.Context, agentName, title, prompt, defaultVal string) (string, bool, error) {
	resp, err := b.client.SubmitRequest(ctx, &v1.SubmitRequestRequest{
		AgentName:  agentName,
		Type:       v1.RequestType_REQUEST_TYPE_TEXT,
		Title:      title,
		Prompt:     prompt,
		DefaultVal: defaultVal,
	})
	if err != nil {
		return "", true, err
	}
	return resp.TextValue, resp.Cancelled, nil
}

func (b *IPCBackend) SubmitMultiline(ctx context.Context, agentName, title, prompt, defaultVal string) (string, int, bool, error) {
	resp, err := b.client.SubmitRequest(ctx, &v1.SubmitRequestRequest{
		AgentName:  agentName,
		Type:       v1.RequestType_REQUEST_TYPE_MULTILINE,
		Title:      title,
		Prompt:     prompt,
		DefaultVal: defaultVal,
	})
	if err != nil {
		return "", 0, true, err
	}
	lines := strings.Count(resp.TextValue, "\n") + 1
	return resp.TextValue, lines, resp.Cancelled, nil
}

func (b *IPCBackend) SubmitChoice(ctx context.Context, agentName, title, prompt string, choices []string, multiSelect bool) ([]string, bool, error) {
	resp, err := b.client.SubmitRequest(ctx, &v1.SubmitRequestRequest{
		AgentName:   agentName,
		Type:        v1.RequestType_REQUEST_TYPE_CHOICE,
		Title:       title,
		Prompt:      prompt,
		Choices:     choices,
		MultiSelect: multiSelect,
	})
	if err != nil {
		return nil, true, err
	}
	return resp.ChoiceValues, resp.Cancelled, nil
}

func (b *IPCBackend) SubmitConfirm(ctx context.Context, agentName, title, message string) (bool, bool, error) {
	resp, err := b.client.SubmitRequest(ctx, &v1.SubmitRequestRequest{
		AgentName: agentName,
		Type:      v1.RequestType_REQUEST_TYPE_CONFIRM,
		Title:     title,
		Prompt:    message,
	})
	if err != nil {
		return false, true, err
	}
	return resp.BoolValue, resp.Cancelled, nil
}

func (b *IPCBackend) SubmitNotify(ctx context.Context, agentName, title, message string) error {
	_, err := b.client.SubmitRequest(ctx, &v1.SubmitRequestRequest{
		AgentName: agentName,
		Type:      v1.RequestType_REQUEST_TYPE_NOTIFY,
		Title:     title,
		Prompt:    message,
	})
	return err
}
