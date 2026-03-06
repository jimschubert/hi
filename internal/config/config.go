package config

import (
	"strings"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL,default=warn"`
	// HiLogging allows for scoped logging, e.g. daemon=warn; scopes will be 1:1 with tool subcommand names
	HiLogging          map[string]string `env:"HI_LOGGING,separator=="`
	HiSocketPath       string            `env:"HI_SOCKET_PATH,default=/tmp/hi.sock"`
	HiClientTimeoutSec int               `env:"HI_CLIENT_TIMEOUT_SEC,default=5"`
}

func (c Config) ClientTimeout() int {
	return max(5, c.HiClientTimeoutSec)
}

func (c Config) SocketPath() string {
	socketPath := c.HiSocketPath
	if socketPath == "" {
		return "/tmp/hi.sock"
	}
	return socketPath
}

func (c Config) DaemonLogLevel() string {
	return c.scopedLevel("daemon")
}

func (c Config) scopedLevel(scope string) string {
	logLevel := c.LogLevel
	if value, ok := c.HiLogging[scope]; ok {
		logLevel = strings.TrimSpace(value)
	}
	return logLevel
}
