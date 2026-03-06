package config

import (
	"strings"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL,default=warn"`
	// HiLogging allows for scoped logging, e.g. daemon=warn; scopes will be 1:1 with tool subcommand names
	HiLogging string `env:"HI_LOGGING,default=warn"`
}

func (c Config) DaemonLogLevel() string {
	return c.scopedLevel("daemon")
}

func (c Config) scopedLevel(scope string) string {
	logLevel := c.LogLevel
	if c.HiLogging != "" {
		// split on ',', find daemon and apply that level
		for _, s := range strings.Split(c.HiLogging, ",") {
			if foundScope, level, ok := strings.Cut(s, "="); ok && strings.ToLower(strings.TrimSpace(foundScope)) == scope {
				logLevel = strings.TrimSpace(level)
			}
		}
	}
	return logLevel
}
