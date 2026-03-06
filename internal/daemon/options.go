package daemon

import "log/slog"

type daemonOpts struct {
	logLevel slog.Level
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
