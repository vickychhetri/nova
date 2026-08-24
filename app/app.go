package app

import (
	"sync"
	"time"

	"github.com/vickychhetri/nova/platform"
	"github.com/vickychhetri/nova/window"
)

// App manages application-wide state, registered windows, and the lifecycle
// loop that connects Nova windows to native platform windows.
//
// The application owns the registration list and the running flag. Individual
// window rendering and event behavior remains implemented by window.Window.
type App struct {
	// mu protects windows and running during application-level mutations.
	mu sync.Mutex
	// windows contains windows registered through Window.
	windows []*window.Window
	// running controls the native event-pump loop.
	running bool
}

// New creates an empty Nova application with initial capacity for four windows.
// The capacity is only an allocation hint; the registry grows automatically.
func New() *App {
	return &App{
		windows: make([]*window.Window, 0, 4),
	}
}

// Window creates a window using options, registers it with the application, and
// returns it to the caller.
//
// Registration is protected by the application mutex. The native platform
// window is not created here; native creation is deferred until Run so all
// configured windows can be initialized together.
func (a *App) Window(options ...window.WindowOption) *window.Window {
	win := window.NewWindow(options...)
	a.mu.Lock()
	a.windows = append(a.windows, win)
	a.mu.Unlock()
	return win
}

// Windows returns the currently registered window slice.
//
// The returned slice is the application's internal slice rather than a copy.
// Callers should treat it as read-only and should not retain it across calls to
// Window unless they also coordinate with the application lifecycle.
func (a *App) Windows() []*window.Window {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.windows
}

// Run starts the application lifecycle, creates native windows, renders their
// initial frames, and processes interaction events until the loop stops.
//
// Run snapshots the registered window slice before native creation. This keeps
// one stable set of windows in the current run even if application-level code
// registers another window later. Native creation failures are tolerated by the
// current implementation; the affected window still receives its initial
// RenderFrame call but does not participate in native event polling.
func (a *App) Run() error {
	a.mu.Lock()
	a.running = true
	windows := append([]*window.Window{}, a.windows...)
	a.mu.Unlock()

	hasNative := false

	for _, win := range windows {
		// Native resources are created from the configured title and logical
		// size. The window object remains the owner of the attached native handle.
		nw, err := platform.CreatePlatformWindow(win.Title(), int(win.Size().Width), int(win.Size().Height))
		if err == nil && nw != nil {
			win.AttachNative(nw)
			hasNative = true
		}
		// Render once before entering the event loop so a newly created window
		// has content immediately.
		win.RenderFrame()
	}

	// If at least one native window exists, enter the non-blocking event pump.
	if hasNative {
		for a.running {
			anyOpen := false
			for _, win := range windows {
				if win.NativeWindow() != nil {
					// PollEvents returns false when that native window has been
					// closed. Redraw only when the window reports dirty state.
					if win.NativeWindow().PollEvents() {
						anyOpen = true
					}
					if win.NeedsRedraw() {
						win.RenderFrame()
					}
				}
			}
			// Avoid a tight busy loop while still keeping input and redraw
			// latency low for the native backend.
			if !anyOpen {
				break
			}
			time.Sleep(8 * time.Millisecond)
		}
	}

	return nil
}

// Quit requests termination of the application loop.
//
// It is safe to call from another goroutine. The current Run loop observes the
// flag between event-polling iterations and returns after the current work
// completes.
func (a *App) Quit() {
	a.mu.Lock()
	a.running = false
	a.mu.Unlock()
}
