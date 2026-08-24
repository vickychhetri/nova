//go:build linux && !headless

package linux

/*
// The C section contains the small Xlib-facing layer used by the Go code below.
// Keeping the X11 structs and calls here avoids spreading cgo details throughout
// the event and rendering code while still allowing Go to own the public API.

// The framebuffer path intentionally uses XImage rather than an X11 drawing
// primitive. Nova renders into an RGBA image, so the native layer only needs to
// adapt that pixel memory and submit it to the window.

#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <X11/Xatom.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

typedef struct {
    Display* display;
    int screen;
    Window window;
    GC gc;
    Atom wmDeleteMessage;
    int width;
    int height;
    char* bgraBuffer;
    int bgraCap;
} X11Window;

// create_x11_window opens the process display connection and creates one
// top-level window with the event categories consumed by PollEvents.
//
// The returned structure owns the Display, Window, graphics context, and any
// later framebuffer allocation. A NULL result means that XOpenDisplay failed,
// usually because DISPLAY is unset or no X server is reachable.
static X11Window* create_x11_window(const char* title, int width, int height) {
    Display* display = XOpenDisplay(NULL);
    if (!display) {
        return NULL;
    }

	// Use the display's default screen so this backend works with the user's
	// normal X11 configuration without requiring a screen number from Go.
	int screen = DefaultScreen(display);
    Window root = RootWindow(display, screen);

	// Select every event class that the Go callback surface understands:
	// painting, keyboard input, pointer input, and window-size changes.
	XSetWindowAttributes attrs;
    attrs.background_pixel = BlackPixel(display, screen);
    attrs.event_mask = ExposureMask | KeyPressMask | KeyReleaseMask |
                       ButtonPressMask | ButtonReleaseMask |
                       PointerMotionMask | StructureNotifyMask;

    Window win = XCreateWindow(
        display, root,
        100, 100, width, height, 0,
        DefaultDepth(display, screen),
        InputOutput,
        DefaultVisual(display, screen),
        CWBackPixel | CWEventMask,
        &attrs
    );

	// Disable X11 automatic background erasing. Nova repaints the framebuffer
	// itself, so Xlib clearing the window first would only add a visible blank
	// frame and could produce blinking or flickering during redraws.
    XSetWindowBackgroundPixmap(display, win, None);

	// WM_DELETE_WINDOW is delivered as a ClientMessage. Registering the atom
	// makes the window manager's close button observable by PollEvents instead
	// of allowing the server to destroy the window without a callback.
    XStoreName(display, win, title);

    Atom wmDelete = XInternAtom(display, "WM_DELETE_WINDOW", False);
    XSetWMProtocols(display, win, &wmDelete, 1);

	// The GC is required by XPutImage and is released in close_x11_window.
	GC gc = XCreateGC(display, win, 0, NULL);

	// Mapping makes the window visible. XFlush sends the pending create/map
	// requests immediately rather than waiting for a later Xlib operation.
    XMapWindow(display, win);
    XFlush(display);

    X11Window* w = (X11Window*)malloc(sizeof(X11Window));
    w->display = display;
    w->screen = screen;
    w->window = win;
    w->gc = gc;
    w->wmDeleteMessage = wmDelete;
    w->width = width;
    w->height = height;
    w->bgraBuffer = NULL;
    w->bgraCap = 0;

    return w;
}

// draw_x11_framebuffer submits one complete RGBA frame to the X11 window.
//
// XImage is created over bgraBuffer, which is owned by X11Window and reused
// across frames. XDestroyImage normally frees its data pointer, so the caller
// clears img->data before destroying the temporary XImage wrapper. This keeps
// the persistent buffer alive until close_x11_window releases it.
static void draw_x11_framebuffer(X11Window* w, const unsigned char* rgbaPix, int wWidth, int wHeight) {
    if (!w || !w->display || !rgbaPix) return;

	// Match the visual and depth selected when the window was created. The
	// server's visual determines the byte layout expected by XPutImage.
	Visual* visual = DefaultVisual(w->display, w->screen);
    int depth = DefaultDepth(w->display, w->screen);
    int needed = wWidth * wHeight * 4;

	// Grow only when necessary. Reusing this allocation avoids a malloc/free
	// pair on every frame, which matters for an animation or game loop.
	if (!w->bgraBuffer || w->bgraCap < needed) {
        if (w->bgraBuffer) free(w->bgraBuffer);
        w->bgraBuffer = (char*)malloc(needed);
        w->bgraCap = needed;
    }
    if (!w->bgraBuffer) return;

	// Convert each packed RGBA pixel to the byte order expected by the X11
	// image. This operates on 32-bit words to avoid four separate byte writes
	// per pixel.
    const uint32_t* src = (const uint32_t*)rgbaPix;
    uint32_t* dst = (uint32_t*)w->bgraBuffer;
    int count = wWidth * wHeight;
    for (int i = 0; i < count; i++) {
        uint32_t p = src[i];
		// On the little-endian Linux targets supported here, RGBA bytes in
		// memory form the word 0xAABBGGRR, while BGRA bytes form 0xAARRGGBB.
		// Preserve alpha and green, then exchange the red and blue bytes.
        uint32_t r = p & 0x000000FF;
        uint32_t b = (p & 0x00FF0000) >> 16;
        dst[i] = (p & 0xFF00FF00) | (r << 16) | b;
    }

    XImage* img = XCreateImage(
        w->display, visual, depth, ZPixmap, 0,
        w->bgraBuffer, wWidth, wHeight, 32, 0
    );

	if (img) {
		// XPutImage copies pixels to the server-side window. The XImage object
		// itself is only a temporary descriptor around our reusable buffer.
        XPutImage(w->display, w->window, w->gc, img, 0, 0, 0, 0, wWidth, wHeight);
		// Prevent XDestroyImage from freeing the persistent framebuffer.
		img->data = NULL;
        XDestroyImage(img);
    }

	// Make the frame visible promptly; Xlib otherwise may batch the request.
    XFlush(w->display);
}

// get_xevent_type is a cgo-friendly accessor for XEvent's type field. The
// complete XEvent union is awkward to inspect directly from Go, while this
// helper lets Go use the normal switch statement over event constants.
static int get_xevent_type(XEvent* ev) {
    return ev->type;
}

// close_x11_window releases all resources allocated by create_x11_window and
// draw_x11_framebuffer. It is deliberately safe for a NULL pointer so cleanup
// can remain simple at the Go boundary.
static void close_x11_window(X11Window* w) {
    if (!w) return;
	if (w->display) {
		// Release X11 resources while the display connection is still valid.
        XFreeGC(w->display, w->gc);
        XDestroyWindow(w->display, w->window);
        XCloseDisplay(w->display);
    }
	if (w->bgraBuffer) {
		// The framebuffer is process memory, not an X11 server resource.
        free(w->bgraBuffer);
        w->bgraBuffer = NULL;
    }
    free(w);
}
*/
import "C"
import (
	"image"
	"sync"
	"unsafe"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/input"
)

// X11Platform manages native Linux X11 windows.
//
// The platform value currently reserves a mutex for coordination between
// platform operations. Window-specific state lives in X11NativeWindow, which
// keeps this type suitable for the higher-level platform abstraction.
type X11Platform struct {
	mu sync.Mutex
}

// X11NativeWindow wraps the C X11 window pointer and exposes callbacks in
// Nova's platform-neutral event types.
//
// ptr is the ownership handle for the native resources. Width and Height are
// kept in Go so resize callbacks and callers can use the latest dimensions
// without repeatedly reading the X11 structure. Callback fields are optional;
// an absent callback simply means that event is ignored.
type X11NativeWindow struct {
	ptr           *C.X11Window
	Width         int
	Height        int
	OnExpose      func()
	OnResize      func(w, h int)
	OnPointerDown func(p geom.Point, btn input.MouseButton)
	OnPointerUp   func(p geom.Point, btn input.MouseButton)
	OnPointerMove func(p geom.Point)
	OnKeyDown     func(e *event.KeyEvent)
	OnClose       func()
}

// NewX11Window creates, maps, and returns an X11 window.
//
// The title is copied into temporary C memory for the duration of the Xlib
// call. Once the call returns, X11 owns the stored window title and the C
// string can be released. A nil window and nil error are returned when no X11
// display can be opened; this preserves the existing backend contract for an
// unavailable display.
func NewX11Window(title string, width, height int) (*X11NativeWindow, error) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	wPtr := C.create_x11_window(cTitle, C.int(width), C.int(height))
	if wPtr == nil {
		return nil, nil
	}

	return &X11NativeWindow{
		ptr:    wPtr,
		Width:  width,
		Height: height,
	}, nil
}

// BlitRGBA copies an RGBA image to the X11 window.
//
// The C bridge converts the image into its reusable native-format buffer before
// submitting it with XPutImage. A nil native handle or image is ignored so a
// caller can safely stop rendering while a window is being closed.
func (w *X11NativeWindow) BlitRGBA(img *image.RGBA) {
	if w.ptr == nil || img == nil {
		return
	}
	C.draw_x11_framebuffer(w.ptr, (*C.uchar)(unsafe.Pointer(&img.Pix[0])), C.int(img.Bounds().Dx()), C.int(img.Bounds().Dy()))
}

// PollEvents drains all currently pending X11 events and returns whether the
// window is still open.
//
// This method is deliberately non-blocking: XPending reports only events that
// are already queued, allowing the application's main loop to continue
// rendering when there is no input. XNextEvent removes one event at a time;
// the type-specific Xlib union is then viewed through the matching C event
// structure before the corresponding Nova callback is invoked.
func (w *X11NativeWindow) PollEvents() bool {
	if w.ptr == nil || w.ptr.display == nil {
		return false
	}

	for C.XPending(w.ptr.display) > 0 {
		var xev C.XEvent
		C.XNextEvent(w.ptr.display, &xev)

		switch C.get_xevent_type(&xev) {
		case C.Expose:
			// X11 asks the application to repaint an exposed or uncovered area.
			if w.OnExpose != nil {
				w.OnExpose()
			}

		case C.ConfigureNotify:
			// ConfigureNotify carries the server-confirmed size after a window
			// move or resize. Notify clients only when dimensions actually change.
			cfg := (*C.XConfigureEvent)(unsafe.Pointer(&xev))
			newW := int(cfg.width)
			newH := int(cfg.height)
			if newW != w.Width || newH != w.Height {
				w.Width = newW
				w.Height = newH
				if w.OnResize != nil {
					w.OnResize(newW, newH)
				}
			}

		case C.ButtonPress:
			// X11 button numbers are 1-based: left is 1, middle is 2, and
			// right is 3. Nova's enum uses named values instead.
			btn := (*C.XButtonEvent)(unsafe.Pointer(&xev))
			var b input.MouseButton = input.ButtonLeft
			if btn.button == 3 {
				b = input.ButtonRight
			} else if btn.button == 2 {
				b = input.ButtonMiddle
			}
			if w.OnPointerDown != nil {
				w.OnPointerDown(geom.Pt(float64(btn.x), float64(btn.y)), b)
			}

		case C.ButtonRelease:
			// Use the same button translation as ButtonPress so press/release
			// pairs produce matching Nova input values.
			btn := (*C.XButtonEvent)(unsafe.Pointer(&xev))
			var b input.MouseButton = input.ButtonLeft
			if btn.button == 3 {
				b = input.ButtonRight
			} else if btn.button == 2 {
				b = input.ButtonMiddle
			}
			if w.OnPointerUp != nil {
				w.OnPointerUp(geom.Pt(float64(btn.x), float64(btn.y)), b)
			}

		case C.MotionNotify:
			// Motion coordinates are window-local pixels, matching the coordinate
			// space used by Nova's pointer callbacks.
			motion := (*C.XMotionEvent)(unsafe.Pointer(&xev))
			if w.OnPointerMove != nil {
				w.OnPointerMove(geom.Pt(float64(motion.x), float64(motion.y)))
			}

		case C.KeyPress:
			// XLookupString resolves the X11 key event into both a KeySym for
			// non-printing keys and a short character buffer for text-like keys.
			// The mapping below prefers well-known special keys, then falls back
			// to Nova's ASCII letter, digit, and control-key values.
			keyEv := (*C.XKeyEvent)(unsafe.Pointer(&xev))
			var keySym C.KeySym
			var buf [32]C.char
			var status C.XComposeStatus
			count := C.XLookupString(keyEv, &buf[0], 32, &keySym, &status)
			var r rune
			if count > 0 {
				r = rune(buf[0])
			}
			k := input.KeyUnknown
			sym := uint32(keySym)
			switch {
			case sym == 0xFF08: // XK_BackSpace
				k = input.KeyBackspace
			case sym == 0xFFFF: // XK_Delete
				k = input.KeyDelete
			case sym == 0xFF0D || sym == 0xFF8D: // XK_Return, XK_KP_Enter
				k = input.KeyEnter
			case sym == 0xFF09: // XK_Tab
				k = input.KeyTab
			case sym == 0xFF1B: // XK_Escape
				k = input.KeyEscape
			case sym == 0xFF51: // XK_Left
				k = input.KeyArrowLeft
			case sym == 0xFF52: // XK_Up
				k = input.KeyArrowUp
			case sym == 0xFF53: // XK_Right
				k = input.KeyArrowRight
			case sym == 0xFF54: // XK_Down
				k = input.KeyArrowDown
			case sym == 0xFF50: // XK_Home
				k = input.KeyHome
			case sym == 0xFF57: // XK_End
				k = input.KeyEnd
			case sym == 0xFF55: // XK_Page_Up
				k = input.KeyPageUp
			case sym == 0xFF56: // XK_Page_Down
				k = input.KeyPageDown
			case sym == 0x0020 || sym == 0xFF80: // XK_space, XK_KP_Space
				k = input.KeySpace
			default:
				if count > 0 {
					switch {
					case r >= 'a' && r <= 'z':
						k = input.Key(int(input.KeyA) + int(r-'a'))
					case r >= 'A' && r <= 'Z':
						k = input.Key(int(input.KeyA) + int(r-'A'))
					case r >= '0' && r <= '9':
						k = input.Key(int(input.Key0) + int(r-'0'))
					case r == ' ':
						k = input.KeySpace
					case r == 8 || r == 127:
						k = input.KeyBackspace
					case r == 13 || r == 10:
						k = input.KeyEnter
					}
				}
			}
			if w.OnKeyDown != nil {
				w.OnKeyDown(&event.KeyEvent{
					Key:  k,
					Rune: r,
				})
			}

		case C.ClientMessage:
			// Window managers request a close through WM_DELETE_WINDOW rather
			// than a normal input event. Run the callback before destroying the
			// native resources, then report false to stop the caller's loop.
			cm := (*C.XClientMessageEvent)(unsafe.Pointer(&xev))
			dataPtr := (*[5]C.long)(unsafe.Pointer(&cm.data[0]))
			if C.Atom(dataPtr[0]) == w.ptr.wmDeleteMessage {
				if w.OnClose != nil {
					w.OnClose()
				}
				w.Close()
				return false
			}
		}
	}

	return true
}

// Close destroys the X11 window and releases its display-side and heap-owned
// resources. It is idempotent: after the first call ptr is nil and later calls
// do nothing.
func (w *X11NativeWindow) Close() {
	if w.ptr != nil {
		C.close_x11_window(w.ptr)
		w.ptr = nil
	}
}

// SetCallbacks registers the event handlers used by PollEvents.
//
// Passing nil for any handler disables delivery for that event category. The
// callbacks execute synchronously on the goroutine calling PollEvents, so a
// callback should avoid blocking if the application relies on a responsive
// render loop.
func (w *X11NativeWindow) SetCallbacks(
	onExpose func(),
	onResize func(w, h int),
	onPointerDown func(p geom.Point, btn input.MouseButton),
	onPointerUp func(p geom.Point, btn input.MouseButton),
	onPointerMove func(p geom.Point),
	onKeyDown func(e *event.KeyEvent),
	onClose func(),
) {
	w.OnExpose = onExpose
	w.OnResize = onResize
	w.OnPointerDown = onPointerDown
	w.OnPointerUp = onPointerUp
	w.OnPointerMove = onPointerMove
	w.OnKeyDown = onKeyDown
	w.OnClose = onClose
}
