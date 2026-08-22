package state

import (
	"sync"
	"sync/atomic"
)

// Computed creates a memoized reactive computation that re-evaluates only when tracked dependencies update.
type Computed[T any] struct {
	mu           sync.RWMutex
	computeFn    func() T
	cachedValue  T
	isDirty      bool
	listeners    map[uint64]func(T)
	unsubscribes []func()
}

// Compute creates a new Computed reactive signal.
func Compute[T any](fn func() T) *Computed[T] {
	c := &Computed[T]{
		computeFn: fn,
		isDirty:   true,
		listeners: make(map[uint64]func(T)),
	}
	c.recompute()
	return c
}

func (c *Computed[T]) addDependency() {
	trackerMutex.Lock()
	if activeTracker != nil {
		activeTracker.dependencies = append(activeTracker.dependencies, c)
	}
	trackerMutex.Unlock()
}

func (c *Computed[T]) recompute() {
	for _, unsub := range c.unsubscribes {
		unsub()
	}
	c.unsubscribes = nil

	ctx := &trackerContext{}
	trackerMutex.Lock()
	ctx.parent = activeTracker
	activeTracker = ctx
	trackerMutex.Unlock()

	newVal := c.computeFn()

	trackerMutex.Lock()
	activeTracker = ctx.parent
	deps := ctx.dependencies
	trackerMutex.Unlock()

	c.mu.Lock()
	c.cachedValue = newVal
	c.isDirty = false
	c.mu.Unlock()

	for _, dep := range deps {
		switch d := dep.(type) {
		case *Value[int]:
			unsub := d.Subscribe(func(int) { c.invalidate() })
			c.unsubscribes = append(c.unsubscribes, unsub)
		case *Value[string]:
			unsub := d.Subscribe(func(string) { c.invalidate() })
			c.unsubscribes = append(c.unsubscribes, unsub)
		case *Value[bool]:
			unsub := d.Subscribe(func(bool) { c.invalidate() })
			c.unsubscribes = append(c.unsubscribes, unsub)
		case *Value[float64]:
			unsub := d.Subscribe(func(float64) { c.invalidate() })
			c.unsubscribes = append(c.unsubscribes, unsub)
		case subscribableWithCallback:
			unsub := d.subscribeGeneric(func() { c.invalidate() })
			c.unsubscribes = append(c.unsubscribes, unsub)
		}
	}
}

type subscribableWithCallback interface {
	subscribable
	subscribeGeneric(fn func()) func()
}

func (c *Computed[T]) subscribeGeneric(fn func()) func() {
	return c.Subscribe(func(T) { fn() })
}

func (v *Value[T]) subscribeGeneric(fn func()) func() {
	return v.Subscribe(func(T) { fn() })
}

func (c *Computed[T]) invalidate() {
	c.mu.Lock()
	c.isDirty = true
	c.mu.Unlock()

	c.recompute()

	c.mu.RLock()
	val := c.cachedValue
	callbacks := make([]func(T), 0, len(c.listeners))
	for _, fn := range c.listeners {
		callbacks = append(callbacks, fn)
	}
	c.mu.RUnlock()

	for _, fn := range callbacks {
		fn(val)
	}
}

// Get returns the computed value.
func (c *Computed[T]) Get() T {
	c.addDependency()
	c.mu.RLock()
	if !c.isDirty {
		val := c.cachedValue
		c.mu.RUnlock()
		return val
	}
	c.mu.RUnlock()

	c.recompute()
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cachedValue
}

// Peek returns the computed value without dependency tracking.
func (c *Computed[T]) Peek() T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cachedValue
}

// Subscribe registers a callback whenever the computed value updates.
func (c *Computed[T]) Subscribe(fn func(T)) func() {
	id := atomic.AddUint64(&listenerIDGen, 1)
	c.mu.Lock()
	c.listeners[id] = fn
	c.mu.Unlock()

	return func() {
		c.mu.Lock()
		delete(c.listeners, id)
		c.mu.Unlock()
	}
}
