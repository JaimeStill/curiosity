package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReadinessChecker reports whether a subsystem is ready to serve work.
type ReadinessChecker interface {
	Ready() bool
}

// Coordinator manages startup and shutdown hooks for the application
// lifecycle. It exposes a cancellation context for hooks to observe, a
// readiness flag that flips once all startup hooks complete, and a
// bounded Shutdown that waits for shutdown hooks to drain.
type Coordinator struct {
	ctx        context.Context
	cancel     context.CancelFunc
	startupWg  sync.WaitGroup
	shutdownWg sync.WaitGroup
	ready      bool
	readyMu    sync.RWMutex
}

// New creates a Coordinator with a cancellable context. The context is
// cancelled by Shutdown; registered shutdown hooks observe the
// cancellation via Context before running cleanup.
func New() *Coordinator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Coordinator{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context returns the coordinator's context, cancelled on Shutdown.
func (c *Coordinator) Context() context.Context {
	return c.ctx
}

// OnStart registers fn to run concurrently when Start is called. It
// contributes to the startup WaitGroup that Start drains before
// flipping the ready flag.
func (c *Coordinator) OnStart(fn func()) {
	c.startupWg.Go(fn)
}

// OnShutdown registers fn to run concurrently when Shutdown is called.
// Hooks should block on <-c.Context().Done() before executing cleanup
// so they observe the shutdown signal before tearing down.
func (c *Coordinator) OnShutdown(fn func()) {
	c.shutdownWg.Go(fn)
}

// Ready returns true after Start has completed all registered startup
// hooks.
func (c *Coordinator) Ready() bool {
	c.readyMu.RLock()
	defer c.readyMu.RUnlock()
	return c.ready
}

// Start blocks until all registered OnStart hooks complete, then marks
// the coordinator as ready. Call after all OnStart registrations.
func (c *Coordinator) Start() {
	c.startupWg.Wait()
	c.readyMu.Lock()
	c.ready = true
	c.readyMu.Unlock()
}

// Shutdown cancels the coordinator's context and waits up to timeout
// for shutdown hooks to drain. Returns nil on clean shutdown, or an
// error wrapping the timeout if hooks fail to complete in time.
func (c *Coordinator) Shutdown(timeout time.Duration) error {
	c.cancel()

	done := make(chan struct{})
	go func() {
		c.shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout after %v", timeout)
	}
}
