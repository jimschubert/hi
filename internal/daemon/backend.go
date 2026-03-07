package daemon

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
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
