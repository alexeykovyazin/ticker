package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ticker/internal/config"
	"ticker/internal/runner"
	"ticker/internal/service"
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  ticker.exe run")
	fmt.Fprintln(os.Stderr, "  ticker.exe install")
	fmt.Fprintln(os.Stderr, "  ticker.exe start")
	fmt.Fprintln(os.Stderr, "  ticker.exe stop")
	fmt.Fprintln(os.Stderr, "  ticker.exe uninstall")
}

func runForeground(cfg config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("ticker: starting (interval=%s)", cfg.Interval)
	err := runner.Run(ctx, cfg)
	if err != nil && ctx.Err() == nil {
		return err
	}
	log.Printf("ticker: stopped")
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cmd := "run"
	if len(os.Args) >= 2 {
		cmd = os.Args[1]
	}

	cfg, err := config.LoadNearExecutable("config.json")
	if err != nil {
		log.Printf("config: %v", err)
	}

	switch cmd {
	case "run":
		if err != nil {
			log.Fatal(err)
		}
		if err := runForeground(cfg); err != nil {
			log.Fatal(err)
		}
	case "install", "uninstall", "start", "stop":
		if err != nil {
			log.Fatal(err)
		}
		if err := service.RunServiceCommand(cmd, cfg); err != nil {
			log.Fatal(err)
		}
	case "service":
		// Intended to be used by Windows SCM; runs until stopped.
		if err != nil {
			log.Fatal(err)
		}
		// If launched interactively, keep behavior consistent with "run".
		isSvc, _ := service.IsWindowsService()
		if !isSvc {
			if err := runForeground(cfg); err != nil {
				log.Fatal(err)
			}
			return
		}
		if err := service.RunAsService(cfg); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

