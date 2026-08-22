package state

import (
	"sync"
	"sync/atomic"
)

var (
	listenerIDGen uint64
	activeTracker *trackerContext
	trackerMutex  sync.Mutex
	batchDepth    int
	pendingNotify []func()
)

type trackerContext struct {
	dependencies []subscribable
	parent       *trackerContext
}

type subscribable interface {
	addDependency()
}

// Signal represents a readable reactive value.
type Signal[T any] interface {
	Get() T
	Subscribe(fn func(T)) (unsubscribe func())
}

// Value is a thread-safe reactive state container with dependency tracking.
type Value[T any] struct {
	mu        sync.RWMutex
	val       T
	listeners map[uint64]func(T)
}

// New creates a new reactive Value of type T.
func New[T any](initial T) *Value[T] {
	return &Value[T]{
		val:       initial,
		listeners: make(map[uint64]func(T)),
	}
}

// Int creates a reactive integer.
func Int(initial int) *Value[int] {
	return New(initial)
}

// String creates a reactive string.
func String(initial string) *Value[string] {
	return New(initial)
}

// Bool creates a reactive boolean.
func Bool(initial bool) *Value[bool] {
	return New(initial)
}

// Float creates a reactive float64.
func Float(initial float64) *Value[float64] {
	return New(initial)
}

// Get reads the current value and records dependency if called inside a reactive context.
func (v *Value[T]) Get() T {
	v.addDependency()
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.val
}

// Peek reads value without registering dependency.
func (v *Value[T]) Peek() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.val
}

func (v *Value[T]) addDependency() {
	trackerMutex.Lock()
	if activeTracker != nil {
		activeTracker.dependencies = append(activeTracker.dependencies, v)
	}
	trackerMutex.Unlock()
}

// Set updates the value and triggers listeners if changed.
func (v *Value[T]) Set(newVal T) {
	var callbacks []func(T)
	v.mu.Lock()
	v.val = newVal
	callbacks = make([]func(T), 0, len(v.listeners))
	for _, fn := range v.listeners {
		callbacks = append(callbacks, fn)
	}
	v.mu.Unlock()

	notify := func() {
		for _, fn := range callbacks {
			fn(newVal)
		}
	}

	trackerMutex.Lock()
	if batchDepth > 0 {
		pendingNotify = append(pendingNotify, notify)
		trackerMutex.Unlock()
	} else {
		trackerMutex.Unlock()
		notify()
	}
}

// Update mutates the value using a transformer function.
func (v *Value[T]) Update(fn func(current T) T) {
	v.mu.Lock()
	newVal := fn(v.val)
	v.val = newVal
	callbacks := make([]func(T), 0, len(v.listeners))
	for _, f := range v.listeners {
		callbacks = append(callbacks, f)
	}
	v.mu.Unlock()

	notify := func() {
		for _, f := range callbacks {
			f(newVal)
		}
	}

	trackerMutex.Lock()
	if batchDepth > 0 {
		pendingNotify = append(pendingNotify, notify)
		trackerMutex.Unlock()
	} else {
		trackerMutex.Unlock()
		notify()
	}
}

// Subscribe registers a listener callback invoked whenever the value changes.
// It returns an unsubscribe function.
func (v *Value[T]) Subscribe(fn func(T)) func() {
	id := atomic.AddUint64(&listenerIDGen, 1)
	v.mu.Lock()
	v.listeners[id] = fn
	v.mu.Unlock()

	return func() {
		v.mu.Lock()
		delete(v.listeners, id)
		v.mu.Unlock()
	}
}
