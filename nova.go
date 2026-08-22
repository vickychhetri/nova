package nova

import (
	"github.com/vickychhetri/nova/app"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/window"
)

// New initializes and creates a new Nova application instance.
func New() *app.App {
	return app.New()
}

// Title configures the window title.
func Title(title string) window.WindowOption {
	return window.WithTitle(title)
}

// Size configures the initial window width and height.
func Size(width, height float64) window.WindowOption {
	return window.WithSize(width, height)
}

// Theme sets the active theme for the window.
func Theme(t *theme.Theme) window.WindowOption {
	return window.WithTheme(t)
}
