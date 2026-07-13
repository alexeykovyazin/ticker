package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"ticker/internal/config"
	"ticker/internal/firebird"
)

func Run(ctx context.Context, cfg config.Config) error {
	if cfg.Interval <= 0 {
		return fmt.Errorf("invalid interval: %s", cfg.Interval)
	}

	// Run once immediately, then on each tick.
	next := time.NewTimer(0)
	defer next.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-next.C:
			start := time.Now()
			if err := firebird.InsertCurrentTimestampOnce(ctx, cfg); err != nil {
				log.Printf("tick: error: %v", err)
			}

			// Maintain a stable period from the start of the tick attempt,
			// but never schedule negative/zero waits.
			wait := cfg.Interval - time.Since(start)
			if wait < 1*time.Millisecond {
				wait = 1 * time.Millisecond
			}
			next.Reset(wait)
		}
	}
}

