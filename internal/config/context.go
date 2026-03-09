package config

import "context"

const (
	mcpAddress    = "mcpAddress"
	daemonVersion = "daemonVersion"
)

// configCtxKey stores individual values on context
// TODO: consider maybe storing everything in *Config and setting that on context?
type configCtxKey struct {
	name string
}

func StoreMcpAddress(ctx context.Context, addr string) context.Context {
	ctx = context.WithValue(ctx, configCtxKey{name: mcpAddress}, addr)
	return ctx
}

func GetMcpAddress(ctx context.Context) string {
	if v := ctx.Value(configCtxKey{mcpAddress}); v != nil {
		if value, ok := v.(string); ok {
			return value
		}
	}
	return ""
}

func StoreDaemonVersion(ctx context.Context, version string) context.Context {
	ctx = context.WithValue(ctx, configCtxKey{name: daemonVersion}, version)
	return ctx
}

func GetDaemonVersion(ctx context.Context) string {
	if v := ctx.Value(configCtxKey{daemonVersion}); v != nil {
		if value, ok := v.(string); ok {
			return value
		}
	}
	return ""
}
