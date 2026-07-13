package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ServiceConfig struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type Config struct {
	Host       string        `json:"host"`
	Port       int           `json:"port"`
	Database   string        `json:"database"`
	User       string        `json:"user"`
	Password   string        `json:"password"`
	Table      string        `json:"table"`
	Column     string        `json:"column"`
	IntervalMS int           `json:"interval_ms"`
	Interval   time.Duration `json:"-"`
	Service    ServiceConfig `json:"service"`
}

func Default() Config {
	return Config{
		Port:       3050,
		Table:      "TICKS",
		Column:     "TICK",
		IntervalMS: 1000,
		Service: ServiceConfig{
			Name:        "Ticker",
			DisplayName: "Ticker",
			Description: "Inserts timestamps into Firebird at a fixed interval",
		},
	}
}

func (c *Config) applyDefaults() {
	d := Default()
	if c.Port == 0 {
		c.Port = d.Port
	}
	if strings.TrimSpace(c.Table) == "" {
		c.Table = d.Table
	}
	if strings.TrimSpace(c.Column) == "" {
		c.Column = d.Column
	}
	if c.IntervalMS == 0 {
		c.IntervalMS = d.IntervalMS
	}
	if strings.TrimSpace(c.Service.Name) == "" {
		c.Service.Name = d.Service.Name
	}
	if strings.TrimSpace(c.Service.DisplayName) == "" {
		c.Service.DisplayName = d.Service.DisplayName
	}
	if strings.TrimSpace(c.Service.Description) == "" {
		c.Service.Description = d.Service.Description
	}
	c.Interval = time.Duration(c.IntervalMS) * time.Millisecond
}

func (c Config) validate() error {
	var errs []string
	if strings.TrimSpace(c.Host) == "" {
		errs = append(errs, "host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		errs = append(errs, "port must be 1..65535")
	}
	if strings.TrimSpace(c.Database) == "" {
		errs = append(errs, "database is required")
	}
	if strings.TrimSpace(c.User) == "" {
		errs = append(errs, "user is required")
	}
	if strings.TrimSpace(c.Password) == "" {
		errs = append(errs, "password is required")
	}
	if c.Interval <= 0 {
		errs = append(errs, "interval_ms must be > 0")
	}
	if strings.TrimSpace(c.Service.Name) == "" {
		errs = append(errs, "service.name is required")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func executableDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return "", fmt.Errorf("abs executable path: %w", err)
	}
	return filepath.Dir(exe), nil
}

func LoadNearExecutable(filename string) (Config, error) {
	dir, err := executableDir()
	if err != nil {
		return Config{}, err
	}
	path := filepath.Join(dir, filename)

	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}

	cfg := Default()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("invalid %s: %w", path, err)
	}
	return cfg, nil
}

