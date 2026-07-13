//go:build !windows

package service

import (
	"fmt"

	"ticker/internal/config"
)

func IsWindowsService() (bool, error) {
	return false, nil
}

func RunAsService(cfg config.Config) error {
	return fmt.Errorf("service mode is only supported on windows")
}

func RunServiceCommand(cmd string, cfg config.Config) error {
	return fmt.Errorf("%s is only supported on windows", cmd)
}

