package daemon

import (
	"net"
	"time"

	"github.com/jimschubert/hi/internal/config"
)

func IsRunning(conf config.Config) bool {
	conn, err := net.DialTimeout("unix", conf.SocketPath(), 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
