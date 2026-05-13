package engine

import (
	"fmt"
	"time"

	"github.com/JaimeStill/curiosity/engine/lifecycle"
)

// Config carries engine-wide configuration. Currently empty; fields
// are added per subsystem as they land.
type Config struct{}

// Engine is the runtime root. It owns the lifecycle Coordinator and
// exposes the two-phase initialization surface (New for cold start,
// Start for hot start) plus a bounded Shutdown.
type Engine struct {
	coordinator *lifecycle.Coordinator
}

var _ lifecycle.ReadinessChecker = (*Engine)(nil)

// New performs cold start: allocates the lifecycle Coordinator and any
// future engine-owned subsystems. The returned Engine is wired but not
// yet running; call Start to bring it online.
func New(cfg *Config) *Engine {
	return &Engine{
		coordinator: lifecycle.New(),
	}
}

// Start performs hot start: drives the Coordinator through registered
// startup hooks and flips the readiness flag.
func (e *Engine) Start() {
	e.coordinator.Start()
}

// Shutdown cancels the lifecycle context and waits up to timeout for
// registered shutdown hooks to drain. Returns nil on clean shutdown or
// an error wrapping the Coordinator's timeout.
func (e *Engine) Shutdown(timeout time.Duration) error {
	if err := e.coordinator.Shutdown(timeout); err != nil {
		return fmt.Errorf("engine shutdown: %w", err)
	}
	return nil
}

// Ready returns true after Start has completed all registered startup
// hooks. Satisfies lifecycle.ReadinessChecker.
func (e *Engine) Ready() bool {
	return e.coordinator.Ready()
}
