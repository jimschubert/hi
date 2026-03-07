package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/daemon"
	"github.com/jimschubert/hi/internal/ipc"
	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
)

type TestCmd struct {
	Type string `arg:"" optional:"" enum:"text,multiline,choice,confirm,notify,all" default:"all" help:"Type of test request to send (text, multiline, choice, confirm, notify, or all)."`
}

func (c *TestCmd) Run(conf config.Config, logger *log.Logger) error {
	if !daemon.IsRunning(conf) {
		logger.Println("daemon is not running — start it with: hi daemon")
		return fmt.Errorf("daemon not running")
	}

	client, err := ipc.NewClient(conf)
	if err != nil {
		return fmt.Errorf("connect to daemon: %w", err)
	}

	tests := []struct {
		name string
		req  *v1.SubmitRequestRequest
	}{
		{
			name: "text",
			req: &v1.SubmitRequestRequest{
				AgentName:  "TestAgent",
				Type:       v1.RequestType_REQUEST_TYPE_TEXT,
				Title:      "Text Input Test",
				Prompt:     "Please enter your name:",
				DefaultVal: "John Doe",
			},
		},
		{
			name: "multiline",
			req: &v1.SubmitRequestRequest{
				AgentName:  "TestAgent",
				Type:       v1.RequestType_REQUEST_TYPE_MULTILINE,
				Title:      "Multiline Input Test",
				Prompt:     "Please provide a detailed description:",
				DefaultVal: "This is a\nmultiline\ndefault value.",
			},
		},
		{
			name: "choice",
			req: &v1.SubmitRequestRequest{
				AgentName: "TestAgent",
				Type:      v1.RequestType_REQUEST_TYPE_CHOICE,
				Title:     "Single Choice Test",
				Prompt:    "Select your favorite color:",
				Choices:   []string{"Red", "Green", "Blue", "Yellow"},
			},
		},
		{
			name: "multi-choice",
			req: &v1.SubmitRequestRequest{
				AgentName:   "TestAgent",
				Type:        v1.RequestType_REQUEST_TYPE_CHOICE,
				Title:       "Multiple Choice Test",
				Prompt:      "Select programming languages you know:",
				Choices:     []string{"Go", "Python", "JavaScript", "Rust", "TypeScript"},
				MultiSelect: true,
			},
		},
		{
			name: "confirm",
			req: &v1.SubmitRequestRequest{
				AgentName: "TestAgent",
				Type:      v1.RequestType_REQUEST_TYPE_CONFIRM,
				Title:     "Confirmation Test",
				Prompt:    "Do you want to proceed with this action?",
			},
		},
		{
			name: "notify",
			req: &v1.SubmitRequestRequest{
				AgentName: "TestAgent",
				Type:      v1.RequestType_REQUEST_TYPE_NOTIFY,
				Title:     "Notification Test",
				Prompt:    "This is a test notification message!",
			},
		},
	}

	var testsToRun []struct {
		name string
		req  *v1.SubmitRequestRequest
	}

	if c.Type == "all" {
		testsToRun = tests
	} else {
		for _, t := range tests {
			if t.name == c.Type {
				testsToRun = append(testsToRun, t)
				break
			}
		}
	}

	ctx := context.Background()
	for _, test := range testsToRun {
		logger.Printf("Sending %s test request...\n", test.name)
		resp, err := client.SubmitRequest(ctx, test.req)
		if err != nil {
			logger.Printf("  Error: %v\n", err)
			continue
		}

		if resp.Cancelled {
			logger.Println("  Result: Cancelled")
		} else {
			switch test.req.Type {
			case v1.RequestType_REQUEST_TYPE_TEXT, v1.RequestType_REQUEST_TYPE_MULTILINE:
				logger.Printf("  Result: %q\n", resp.TextValue)
			case v1.RequestType_REQUEST_TYPE_CHOICE:
				logger.Printf("  Result: %v\n", resp.ChoiceValues)
			case v1.RequestType_REQUEST_TYPE_CONFIRM:
				logger.Printf("  Result: %v\n", resp.BoolValue)
			case v1.RequestType_REQUEST_TYPE_NOTIFY:
				logger.Println("  Result: Notification sent")
			}
		}
	}

	return nil
}
