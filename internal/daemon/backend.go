package daemon

import (
	"context"
	"strings"

	"github.com/jimschubert/hi/internal/daemon/store"
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

type QueueBackend struct {
	queue    *store.Queue
	notifyFn func(title, body string)
}

func NewQueueBackend(q *store.Queue, notifyFn func(title, body string)) *QueueBackend {
	return &QueueBackend{queue: q, notifyFn: notifyFn}
}

func (b *QueueBackend) SubmitText(ctx context.Context, agentName, title, prompt, defaultVal string) (string, bool, error) {
	b.notifyFn(agentName+" is waiting", "Respond in the dialog that just appeared.")
	resp, err := b.queue.Enqueue(ctx, &store.PendingRequest{
		Type:       store.RequestTypeText,
		AgentName:  agentName,
		Title:      title,
		Prompt:     prompt,
		DefaultVal: defaultVal,
	})
	if err != nil {
		return "", true, err
	}
	return resp.TextValue, resp.Cancelled, nil
}

func (b *QueueBackend) SubmitMultiline(ctx context.Context, agentName, title, prompt, defaultVal string) (string, int, bool, error) {
	b.notifyFn(agentName+" is waiting", "Respond in the dialog that just appeared.")
	resp, err := b.queue.Enqueue(ctx, &store.PendingRequest{
		Type:       store.RequestTypeMultiline,
		AgentName:  agentName,
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

func (b *QueueBackend) SubmitChoice(ctx context.Context, agentName, title, prompt string, choices []string, multiSelect bool) ([]string, bool, error) {
	b.notifyFn(agentName+" is waiting", "Respond in the dialog that just appeared.")
	resp, err := b.queue.Enqueue(ctx, &store.PendingRequest{
		Type:        store.RequestTypeChoice,
		AgentName:   agentName,
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

func (b *QueueBackend) SubmitConfirm(ctx context.Context, agentName, title, message string) (bool, bool, error) {
	b.notifyFn(agentName+" needs confirmation", "Respond in the dialog that just appeared.")
	resp, err := b.queue.Enqueue(ctx, &store.PendingRequest{
		Type:      store.RequestTypeConfirm,
		AgentName: agentName,
		Title:     title,
		Prompt:    message,
	})
	if err != nil {
		return false, true, err
	}
	return resp.BoolValue, resp.Cancelled, nil
}

func (b *QueueBackend) SubmitNotify(ctx context.Context, agentName, title, message string) error {
	if title == "" {
		title = agentName
	}
	b.notifyFn(title, message)
	return nil
}
