package app

import (
	"sync"

	"github.com/vickychhetri/nova/window"
)

// App manages global application state, windows, and lifecycle loop.
type App struct {
	mu      sync.Mutex
	windows []*window.Window
	running bool
}

// New creates a new Nova App instance.
func New() *App {
	return &App{
		windows: make([]*window.Window, 0, 4),
	}
}

// Window creates and registers a new window in the application.
func (a *App) Window(options ...window.WindowOption) *window.Window {
	win := window.NewWindow(options...)
	a.mu.Lock()
	a.windows = append(a.windows, win)
	a.mu.Unlock()
	return win
}

// Windows returns all active windows.
func (a *App) Windows() []*window.Window {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.windows
}

// Run starts the application loop and renders active windows.
func (a *App) Run() error {
	a.mu.Lock()
	a.running = true
	windows := append([]*window.Window{}, a.windows...)
	a.mu.Unlock()

	for _, win := range windows {
		win.RenderFrame()
	}

	return nil
}

// Quit terminates the application.
func (a *App) Quit() {
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()
}
