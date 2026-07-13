//go:build windows

package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"ticker/internal/config"
	"ticker/internal/runner"
)

func IsWindowsService() (bool, error) {
	return svc.IsWindowsService()
}

type tickerService struct {
	cfg config.Config
}

func (m tickerService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := runner.Run(ctx, m.cfg); err != nil && ctx.Err() == nil {
			log.Printf("service runner error: %v", err)
		}
	}()

	s <- svc.Status{State: svc.Running, Accepts: accepted}

	for {
		select {
		case <-done:
			s <- svc.Status{State: svc.StopPending}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending}
				cancel()
				<-done
				return false, 0
			default:
			}
		}
	}
}

func RunAsService(cfg config.Config) error {
	return svc.Run(cfg.Service.Name, tickerService{cfg: cfg})
}

func exePath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(p)
}

func RunServiceCommand(cmd string, cfg config.Config) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to service manager: %w", err)
	}
	defer m.Disconnect()

	name := cfg.Service.Name

	switch cmd {
	case "install":
		exe, err := exePath()
		if err != nil {
			return fmt.Errorf("resolve exe path: %w", err)
		}

		// If it already exists, return a helpful error.
		if s, err := m.OpenService(name); err == nil {
			s.Close()
			return fmt.Errorf("service %q already exists", name)
		}

		s, err := m.CreateService(name, exe, mgr.Config{
			DisplayName: cfg.Service.DisplayName,
			Description: cfg.Service.Description,
			StartType:   mgr.StartAutomatic,
		}, "service")
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}
		defer s.Close()
		return nil

	case "uninstall":
		s, err := m.OpenService(name)
		if err != nil {
			return fmt.Errorf("open service: %w", err)
		}
		defer s.Close()
		if err := s.Delete(); err != nil {
			return fmt.Errorf("delete service: %w", err)
		}
		return nil

	case "start":
		s, err := m.OpenService(name)
		if err != nil {
			return fmt.Errorf("open service: %w", err)
		}
		defer s.Close()
		if err := s.Start(); err != nil {
			return fmt.Errorf("start service: %w", err)
		}
		return nil

	case "stop":
		s, err := m.OpenService(name)
		if err != nil {
			return fmt.Errorf("open service: %w", err)
		}
		defer s.Close()

		st, err := s.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("send stop: %w", err)
		}

		timeout := time.Now().Add(20 * time.Second)
		for st.State != svc.Stopped {
			if time.Now().After(timeout) {
				return fmt.Errorf("timeout waiting for service to stop")
			}
			time.Sleep(300 * time.Millisecond)
			st, err = s.Query()
			if err != nil {
				return fmt.Errorf("query service: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown service command: %s", cmd)
	}
}

