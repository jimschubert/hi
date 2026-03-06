package daemon

import (
	"log/slog"

	"github.com/jimschubert/hi/internal/config"
)

type daemonOpts struct {
	logLevel slog.Level
	config   config.Config
}

type Option func(*daemonOpts)

func WithLogLevel(logLevel string) Option {
	return func(opts *daemonOpts) {
		var l slog.Level
		if err := l.UnmarshalText([]byte(logLevel)); err != nil {
			l = slog.LevelError
		}
		opts.logLevel = l
	}
}

func WithConfig(config config.Config) Option {
	return func(opts *daemonOpts) {
		opts.config = config
	}
}
