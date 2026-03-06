package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/sethvargo/go-envconfig"
)

func processWithMap(t *testing.T, env map[string]string) Config {
	t.Helper()
	var c Config
	if err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target:   &c,
		Lookuper: envconfig.MapLookuper(env),
	}); err != nil {
		t.Fatalf("envconfig.ProcessWith: %v", err)
	}
	return c
}

func TestConfig_Defaults(t *testing.T) {
	t.Parallel()

	c := processWithMap(t, map[string]string{})

	if c.LogLevel != "warn" {
		t.Errorf("LogLevel: got %q, want %q", c.LogLevel, "warn")
	}
	if c.HiSocketPath != "/tmp/hi.sock" {
		t.Errorf("HiSocketPath: got %q, want %q", c.HiSocketPath, "/tmp/hi.sock")
	}
	if len(c.HiLogging) != 0 {
		t.Errorf("HiLogging: got %v, want empty map", c.HiLogging)
	}
}

func TestConfig_SocketPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name:     "default when unset",
			env:      map[string]string{},
			expected: "/tmp/hi.sock",
		},
		{
			name:     "custom path from env",
			env:      map[string]string{"HI_SOCKET_PATH": "/var/run/hi.sock"},
			expected: "/var/run/hi.sock",
		},
		{
			name: "fallback to default when empty string",
			// go-envconfig will use the default= tag value, so an explicit empty
			// string isn't reachable via env, but test the method guard directly.
			env:      map[string]string{"HI_SOCKET_PATH": ""},
			expected: "/tmp/hi.sock",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := processWithMap(t, tc.env)
			if got := c.SocketPath(); got != tc.expected {
				t.Errorf("SocketPath(): got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestConfig_DaemonLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name:     "falls back to global log level",
			env:      map[string]string{"LOG_LEVEL": "info"},
			expected: "info",
		},
		{
			name:     "falls back to default log level when unset",
			env:      map[string]string{},
			expected: "warn",
		},
		{
			name:     "daemon scope overrides global",
			env:      map[string]string{"LOG_LEVEL": "info", "HI_LOGGING": "daemon=debug"},
			expected: "debug",
		},
		{
			name:     "daemon scope value is trimmed",
			env:      map[string]string{"HI_LOGGING": "daemon= error "},
			expected: "error",
		},
		{
			name:     "unrelated scope does not affect daemon",
			env:      map[string]string{"LOG_LEVEL": "warn", "HI_LOGGING": "server=debug"},
			expected: "warn",
		},
		{
			name:     "multiple scopes, daemon wins",
			env:      map[string]string{"LOG_LEVEL": "warn", "HI_LOGGING": "server=debug,daemon=trace"},
			expected: "trace",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := processWithMap(t, tc.env)
			if got := c.DaemonLogLevel(); got != tc.expected {
				t.Errorf("DaemonLogLevel(): got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestConfig_ScopedLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		scope    string
		expected string
	}{
		{
			name:     "unknown scope returns global level",
			env:      map[string]string{"LOG_LEVEL": "error"},
			scope:    "unknown",
			expected: "error",
		},
		{
			name:     "known scope overrides global",
			env:      map[string]string{"LOG_LEVEL": "error", "HI_LOGGING": "myservice=debug"},
			scope:    "myservice",
			expected: "debug",
		},
		{
			name:     "scoped value is trimmed",
			env:      map[string]string{"HI_LOGGING": "myservice=  info  "},
			scope:    "myservice",
			expected: "info",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := processWithMap(t, tc.env)
			if got := c.scopedLevel(tc.scope); got != tc.expected {
				t.Errorf("scopedLevel(%q): got %q, want %q", tc.scope, got, tc.expected)
			}
		})
	}
}

func TestConfig_ClientTimeout(t *testing.T) {
	t.Parallel()

	defaultValue := 5
	defaultStr := fmt.Sprintf("%d", defaultValue)

	tests := []struct {
		name               string
		hiClientTimeoutSec int
		want               int
	}{
		{
			name:               "negative value returns min of " + defaultStr,
			hiClientTimeoutSec: -1,
			want:               defaultValue,
		},
		{
			name:               "zero value returns min of " + defaultStr,
			hiClientTimeoutSec: 0,
			want:               defaultValue,
		},
		{
			name:               defaultStr + " returns " + defaultStr,
			hiClientTimeoutSec: defaultValue,
			want:               defaultValue,
		},

		{
			name:               "value > default returns that value",
			hiClientTimeoutSec: defaultValue + 1,
			want:               defaultValue + 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				HiClientTimeoutSec: tt.hiClientTimeoutSec,
			}
			if got := c.ClientTimeout(); got != tt.want {
				t.Errorf("ClientTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
