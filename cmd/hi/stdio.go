package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/daemon"
)

type StdioCmd struct{}

func (c *StdioCmd) Run(conf config.Config, logger *log.Logger) error {
	running := daemon.IsRunning(conf)
	logger.Printf("hi stdio starting (daemon running: %v)", running)

	if !running {
		if err := startDaemonBackground(logger); err != nil {
			return fmt.Errorf("could not start daemon: %w", err)
		}

		logger.Println("daemon process started, waiting a bit for socket availability…")

		for range 20 {
			time.Sleep(250 * time.Millisecond)
			if daemon.IsRunning(conf) {
				break
			}
		}

		if !daemon.IsRunning(conf) {
			return fmt.Errorf("daemon failed to start (socket: %s)", conf.SocketPath())
		}

		logger.Println("daemon is ready")
	}

	logger.Println("starting stdio bridge")
	if err := daemon.RunStdioServer(context.Background(), conf); err != nil {
		logger.Printf("stdio bridge exited with error: %v", err)
		return err
	}

	logger.Println("stdio bridge exited cleanly")
	return nil
}
