package app

import (
	"sync"
	"time"

	"github.com/vickychhetri/nova/platform"
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

// Run starts the application loop, creates native OS windows, and processes interaction events.
func (a *App) Run() error {
	a.mu.Lock()
	a.running = true
	windows := append([]*window.Window{}, a.windows...)
	a.mu.Unlock()

	hasNative := false

	for _, win := range windows {
		nw, err := platform.CreatePlatformWindow(win.Title(), int(win.Size().Width), int(win.Size().Height))
		if err == nil && nw != nil {
			win.AttachNative(nw)
			hasNative = true
		}
		win.RenderFrame()
	}

	// If running with a native OS windowing display, enter the event pump loop
	if hasNative {
		for a.running {
			anyOpen := false
			for _, win := range windows {
				if win.NativeWindow() != nil {
					if win.NativeWindow().PollEvents() {
						anyOpen = true
					}
					if win.NeedsRedraw() {
						win.RenderFrame()
					}
				}
			}
			if !anyOpen {
				break
			}
			time.Sleep(8 * time.Millisecond)
		}
	}

	return nil
}

// Quit terminates the application.
func (a *App) Quit() {
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()
}
