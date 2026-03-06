package main

import (
	"context"
	"fmt"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/ipc"
)

type ShutdownCmd struct{}

func (s ShutdownCmd) Run(conf config.Config) error {
	client, err := ipc.NewClient(conf)
	if err != nil {
		return err
	}

	if _, err := client.Shutdown(context.Background()); err != nil {
		return err
	}

	fmt.Println("hi daemon shutdown request was delivered successfully")
	return nil
}
