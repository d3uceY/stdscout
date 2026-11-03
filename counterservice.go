package main

import (
	"strconv"
	"sync"
	"time"
)

// CounterState is the full snapshot the frontend renders. Every field is
// exported with a json tag so the binding generator can mirror it in JS.
type CounterState struct {
	Count      int    `json:"count"`
	LastAction string `json:"lastAction"`
	UpdatedAt  string `json:"updatedAt"`
}

// CounterService owns the counter. Elm never touches it directly: Elm emits a
// port message, main.js calls one of these methods, and the returned snapshot
// is pushed back into Elm through another port.
type CounterService struct {
	mu    sync.Mutex
	count int
	last  string
	seen  time.Time
}

func (c *CounterService) snapshot() CounterState {
	return CounterState{
		Count:      c.count,
		LastAction: c.last,
		UpdatedAt:  c.seen.Format("15:04:05"),
	}
}

// State returns the current snapshot without changing anything. The frontend
// calls this once on boot to sync.
func (c *CounterService) State() CounterState {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.last == "" {
		c.last = "init"
		c.seen = time.Now()
	}
	return c.snapshot()
}

// Increment adds delta (which may be negative) and returns the new snapshot.
func (c *CounterService) Increment(delta int) CounterState {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.count += delta
	c.last = "increment(" + strconv.Itoa(delta) + ")"
	c.seen = time.Now()
	return c.snapshot()
}

// Reset puts the counter back to zero.
func (c *CounterService) Reset() CounterState {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.count = 0
	c.last = "reset"
	c.seen = time.Now()
	return c.snapshot()
}
