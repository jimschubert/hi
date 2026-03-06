package main

import (
	"cmp"
	"context"
	"fmt"
	"time"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/ipc"
	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
	"github.com/lithammer/dedent"
)

type StatusCmd struct {
}

func (s StatusCmd) Run(conf config.Config) error {
	var status *v1.GetStatusResponse
	client, err := ipc.NewClient(conf)
	if client != nil {
		status, err = client.Status(context.Background())
	}

	if status == nil || err != nil {
		msg := fmt.Sprintf(`
			hi daemon: not running
				socket: %s
			`, conf.SocketPath())

		fmt.Print(dedent.Dedent(msg))
		return nil
	}

	msg := fmt.Sprintf(`
		hi daemon: running
			version : %s	
			uptime  : %s
			socket  : %s
			pending : %d
			address : %s
		`,
		status.GetVersion(),
		time.Duration(status.UptimeSeconds)*time.Second,
		conf.SocketPath(),
		status.GetPendingRequests(),
		cmp.Or(status.GetMcpAddress(), "<undefined>"),
	)

	fmt.Print(dedent.Dedent(msg))
	return nil
}
