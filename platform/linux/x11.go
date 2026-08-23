//go:build linux && !headless

package linux

/*
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

static X11Window* create_x11_window(const char* title, int width, int height) {
    Display* display = XOpenDisplay(NULL);
    if (!display) {
        return NULL;
    }

    int screen = DefaultScreen(display);
    Window root = RootWindow(display, screen);

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

    XStoreName(display, win, title);

    Atom wmDelete = XInternAtom(display, "WM_DELETE_WINDOW", False);
    XSetWMProtocols(display, win, &wmDelete, 1);

    GC gc = XCreateGC(display, win, 0, NULL);

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

static void draw_x11_framebuffer(X11Window* w, const unsigned char* rgbaPix, int wWidth, int wHeight) {
    if (!w || !w->display || !rgbaPix) return;

    Visual* visual = DefaultVisual(w->display, w->screen);
    int depth = DefaultDepth(w->display, w->screen);
    int needed = wWidth * wHeight * 4;

    if (!w->bgraBuffer || w->bgraCap < needed) {
        if (w->bgraBuffer) free(w->bgraBuffer);
        w->bgraBuffer = (char*)malloc(needed);
        w->bgraCap = needed;
    }
    if (!w->bgraBuffer) return;

    // Fast 32-bit pixel word RGBA -> BGRA byte swap
    const uint32_t* src = (const uint32_t*)rgbaPix;
    uint32_t* dst = (uint32_t*)w->bgraBuffer;
    int count = wWidth * wHeight;
    for (int i = 0; i < count; i++) {
        uint32_t p = src[i];
        // Little Endian RGBA in memory is 0xAABBGGRR.
        // Little Endian BGRA in memory is 0xAARRGGBB.
        uint32_t r = p & 0x000000FF;
        uint32_t b = (p & 0x00FF0000) >> 16;
        dst[i] = (p & 0xFF00FF00) | (r << 16) | b;
    }

    XImage* img = XCreateImage(
        w->display, visual, depth, ZPixmap, 0,
        w->bgraBuffer, wWidth, wHeight, 32, 0
    );

    if (img) {
        XPutImage(w->display, w->window, w->gc, img, 0, 0, 0, 0, wWidth, wHeight);
        img->data = NULL; // prevent XDestroyImage from freeing persistent buffer
        XDestroyImage(img);
    }

    XFlush(w->display);
}

static int get_xevent_type(XEvent* ev) {
    return ev->type;
}

static void close_x11_window(X11Window* w) {
    if (!w) return;
    if (w->display) {
        XFreeGC(w->display, w->gc);
        XDestroyWindow(w->display, w->window);
        XCloseDisplay(w->display);
    }
    if (w->bgraBuffer) {
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
type X11Platform struct {
	mu sync.Mutex
}

// X11NativeWindow wraps the C X11 window pointer.
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

// NewX11Window creates and maps an X11 window.
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

// BlitRGBA copies RGBA buffer pixels to the X11 window.
func (w *X11NativeWindow) BlitRGBA(img *image.RGBA) {
	if w.ptr == nil || img == nil {
		return
	}
	C.draw_x11_framebuffer(w.ptr, (*C.uchar)(unsafe.Pointer(&img.Pix[0])), C.int(img.Bounds().Dx()), C.int(img.Bounds().Dy()))
}

// PollEvents processes pending X11 events and returns true if window is still open.
func (w *X11NativeWindow) PollEvents() bool {
	if w.ptr == nil || w.ptr.display == nil {
		return false
	}

	for C.XPending(w.ptr.display) > 0 {
		var xev C.XEvent
		C.XNextEvent(w.ptr.display, &xev)

		switch C.get_xevent_type(&xev) {
		case C.Expose:
			if w.OnExpose != nil {
				w.OnExpose()
			}

		case C.ConfigureNotify:
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
			motion := (*C.XMotionEvent)(unsafe.Pointer(&xev))
			if w.OnPointerMove != nil {
				w.OnPointerMove(geom.Pt(float64(motion.x), float64(motion.y)))
			}

		case C.KeyPress:
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
			switch uint32(keySym) {
			case 0xFF08: // XK_BackSpace
				k = input.KeyBackspace
			case 0xFFFF: // XK_Delete
				k = input.KeyDelete
			case 0xFF0D, 0xFF8D: // XK_Return, XK_KP_Enter
				k = input.KeyEnter
			case 0xFF09: // XK_Tab
				k = input.KeyTab
			case 0xFF1B: // XK_Escape
				k = input.KeyEscape
			case 0xFF51: // XK_Left
				k = input.KeyArrowLeft
			case 0xFF52: // XK_Up
				k = input.KeyArrowUp
			case 0xFF53: // XK_Right
				k = input.KeyArrowRight
			case 0xFF54: // XK_Down
				k = input.KeyArrowDown
			case 0xFF50: // XK_Home
				k = input.KeyHome
			case 0xFF57: // XK_End
				k = input.KeyEnd
			case 0xFF55: // XK_Page_Up
				k = input.KeyPageUp
			case 0xFF56: // XK_Page_Down
				k = input.KeyPageDown
			case 0x0020: // XK_space
				k = input.KeySpace
			default:
				if count > 0 && (buf[0] == 8 || buf[0] == 127) {
					k = input.KeyBackspace
				}
			}
			if w.OnKeyDown != nil {
				w.OnKeyDown(&event.KeyEvent{
					Key:  k,
					Rune: r,
				})
			}

		case C.ClientMessage:
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

// Close destroys the X11 window.
func (w *X11NativeWindow) Close() {
	if w.ptr != nil {
		C.close_x11_window(w.ptr)
		w.ptr = nil
	}
}

// SetCallbacks registers event listener handlers.
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
