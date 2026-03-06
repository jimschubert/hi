package main

import (
	"cmp"
	"context"
	"fmt"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/ipc"
	v1 "github.com/jimschubert/hi/internal/proto/gen/hi/v1"
)

type PingCmd struct {
}

func (s PingCmd) Run(conf config.Config) error {
	var resp *v1.PingResponse
	client, err := ipc.NewClient(conf)
	if client != nil {
		resp, err = client.Ping(context.Background())
	}

	if resp == nil || err != nil {
		msg := "no ping response"
		if err != nil {
			msg = msg + fmt.Sprintf(": %v", err)
		}
		fmt.Println(msg)
		return nil
	}

	fmt.Printf("version: %s", cmp.Or(resp.GetVersion(), "<empty>"))
	return nil
}
