package engine_test

import (
	"testing"
	"time"

	"github.com/JaimeStill/curiosity/engine"
)

func TestLifecycle_PreStart(t *testing.T) {
	e := engine.New(&engine.Config{})

	if e.Ready() {
		t.Fatal("engine reported ready before Start")
	}

	if err := e.Shutdown(time.Second); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}
}
