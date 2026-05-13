package lifecycle_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/JaimeStill/curiosity/engine/lifecycle"
)

func TestCoordinator_PreStart_NotReady(t *testing.T) {
	c := lifecycle.New()
	if c.Ready() {
		t.Error("Coordinator reported ready before Start")
	}
}

func TestCoordinator_StartFlipsReady(t *testing.T) {
	c := lifecycle.New()
	c.Start()
	if !c.Ready() {
		t.Error("Coordinator did not report ready after Start")
	}
}

func TestCoordinator_StartDrainsOnStartHooks(t *testing.T) {
	c := lifecycle.New()
	var ran atomic.Bool
	c.OnStart(func() {
		ran.Store(true)
	})
	c.Start()
	if !ran.Load() {
		t.Error("OnStart hook did not run before Start returned")
	}
}

func TestCoordinator_Shutdown_CancelsContext(t *testing.T) {
	c := lifecycle.New()
	ctx := c.Context()
	if err := c.Shutdown(time.Second); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	select {
	case <-ctx.Done():
	default:
		t.Error("context was not cancelled after Shutdown")
	}
}

func TestCoordinator_Shutdown_DrainsOnShutdownHooks(t *testing.T) {
	c := lifecycle.New()
	var ran atomic.Bool
	c.OnShutdown(func() {
		<-c.Context().Done()
		ran.Store(true)
	})
	if err := c.Shutdown(time.Second); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
	if !ran.Load() {
		t.Error("OnShutdown hook did not run before Shutdown returned")
	}
}

func TestCoordinator_Shutdown_TimeoutReturnsError(t *testing.T) {
	c := lifecycle.New()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	c.OnShutdown(func() {
		<-release
	})
	if err := c.Shutdown(10 * time.Millisecond); err == nil {
		t.Error("Shutdown did not return error when hook outlasted timeout")
	}
}
