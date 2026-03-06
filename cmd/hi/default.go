package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/jimschubert/hi/internal/config"
	"github.com/jimschubert/hi/internal/daemon"
)

// DefaultCmd runs when the user doesn't specify a command; verifies daemon is running and starts if it isn't.
type DefaultCmd struct{}

func (c *DefaultCmd) Run(logger *log.Logger, conf config.Config) error {
	if daemon.IsRunning(conf) {
		fmt.Println("Daemon is running.")
		return nil
	}

	logger.Println("daemon is not running — starting it now…")
	if err := startDaemonBackground(logger); err != nil {
		fmt.Fprintf(os.Stderr, "could not start daemon: %v\n", err)
		os.Exit(1)
	}

	for range 10 {
		time.Sleep(250 * time.Millisecond)
		if daemon.IsRunning(conf) {
			logger.Println("daemon started.")
			return nil
		}
	}

	fmt.Fprintln(os.Stderr, "daemon did not start in time")
	os.Exit(1)
	return nil
}

func startDaemonBackground(logger *log.Logger) error {
	program, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(program, "daemon")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	logger.Printf("spawning daemon: %s daemon", program)
	return cmd.Start()
}
