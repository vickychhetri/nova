# Chapter 2: OS Windowing & Display Servers — The Platform Abstraction Layer

> *"How does a Go program ask the Linux kernel, Windows, or macOS to create a native window frame with minimize/maximize/close buttons, and how does it deliver raw pixels to the OS?"*

---

## 2.1 The Role of Display Servers

An operating system kernel (such as Linux) does not directly provide buttons or windows. Instead, GUI environments rely on a **Display Server** or **Windowing Compositor**:

```
+-------------------------------------------------------------------------+
|                           User Application (Nova)                       |
+-------------------------------------------------------------------------+
                                     │
                    IPC Protocol / FFI Function Calls
                                     ▼
+-------------------------------------------------------------------------+
|                         OS Display Server                             |
|   • Linux: X11 Server (Xorg) or Wayland Compositor (Mutter, Sway)       |
|   • Windows: Desktop Window Manager (DWM / Win32 GDI / DirectX)         |
|   • macOS: Quartz Compositor (WindowServer / Cocoa NSWindow)            |
+-------------------------------------------------------------------------+
                                     │
                             GPU Driver (DRM / KMS)
                                     ▼
+-------------------------------------------------------------------------+
|                        Physical Monitor Display                         |
+-------------------------------------------------------------------------+
```

Nova defines a clean, platform-agnostic interface in Go (`platform/platform.go`):

```go
type NativeWindow interface {
    BlitRGBA(img *image.RGBA)
    PollEvents() bool
    Close()
    SetCallbacks(
        onExpose func(),
        onResize func(w, h int),
        onPointerDown func(p geom.Point, btn input.MouseButton),
        onPointerUp func(p geom.Point, btn input.MouseButton),
        onPointerMove func(p geom.Point),
        onKeyDown func(e *event.KeyEvent),
        onClose func(),
    )
}
```

---

## 2.2 Linux X11 Deep Dive: Cgo & The Xlib Protocol

On Linux, Nova communicates directly with the X11 server using a lightweight Cgo wrapper over `libX11` (`platform/linux/x11.go`).

### Step 1: Connecting to the X Display
When Nova launches, it connects to the active X11 Unix domain socket (`/tmp/.X11-unix/X0`):

```c
Display* display = XOpenDisplay(NULL);
if (!display) {
    // Handle headless or Wayland fallback
}
int screen = DefaultScreen(display);
Window root = RootWindow(display, screen);
```

### Step 2: Creating the Window
To create a window with full control over input events and background rendering:

```c
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

// CRITICAL FLICKER FIX: Tell X server NEVER to auto-erase background before blitting!
XSetWindowBackgroundPixmap(display, win, None);

// Intercept Window Close Button (X on titlebar)
Atom wmDelete = XInternAtom(display, "WM_DELETE_WINDOW", False);
XSetWMProtocols(display, win, &wmDelete, 1);

XMapWindow(display, win);
XFlush(display);
```

---

## 2.3 The Pixel Delivery Mechanism: `BlitRGBA` & Little-Endian Byte Swapping

Once Nova's software rasterizer finishes rendering a frame in Go into an `*image.RGBA` buffer, it calls:

```go
w.nativeWin.BlitRGBA(w.rasterizer.Buffer())
```

### The Endianness Problem: RGBA vs BGRA

In memory, Go's `image.RGBA` stores bytes in linear order:

$$\text{Go Memory: } [\text{Byte 0: R}, \text{Byte 1: G}, \text{Byte 2: B}, \text{Byte 3: A}]$$

However, on x86-64 Linux systems, X11 32-bit `ZPixmap` visuals interpret 32-bit integers in **Little-Endian format**, where the least significant byte is stored first:

$$\text{X11 Integer in Little Endian: } 0\text{xAARRGGBB} \implies [\text{Byte 0: B}, \text{Byte 1: G}, \text{Byte 2: R}, \text{Byte 3: A}]$$

If you copy Go's RGBA bytes directly into X11 without conversion, all red and blue channels are inverted (red looks blue, and blue looks red!).

Nova solves this with an ultra-fast 32-bit word manipulation in C:

```c
static void draw_x11_framebuffer(X11Window* w, const unsigned char* rgbaPix, int wWidth, int wHeight) {
    const uint32_t* src = (const uint32_t*)rgbaPix;
    uint32_t* dst = (uint32_t*)w->bgraBuffer;
    int count = wWidth * wHeight;

    // Fast 32-bit pixel word bitwise swap
    for (int i = 0; i < count; i++) {
        uint32_t p = src[i];
        uint32_t r = p & 0x000000FF;
        uint32_t b = (p & 0x00FF0000) >> 16;
        dst[i] = (p & 0xFF00FF00) | (r << 16) | b;
    }

    XImage* img = XCreateImage(
        w->display, visual, depth, ZPixmap, 0,
        w->bgraBuffer, wWidth, wHeight, 32, 0
    );

    if (img) {
        // Copy pixel buffer directly into the native X11 window
        XPutImage(w->display, w->window, w->gc, img, 0, 0, 0, 0, wWidth, wHeight);
        img->data = NULL; // prevent XDestroyImage from freeing persistent buffer
        XDestroyImage(img);
    }
    XFlush(w->display);
}
```

---

## 2.4 The Event Polling Loop: Driving the Application

The Nova application loop (`app/app.go`) pumps events from the OS message queue:

```go
func (a *App) Run() error {
    // ...
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
        time.Sleep(8 * time.Millisecond) // ~120 Hz tick rate
    }
    return nil
}
```

In `PollEvents()`, Nova processes X11 events non-blockingly:
- `Expose`: Triggered when the window is uncovered $\implies$ sets `needsRedraw = true`.
- `ConfigureNotify`: Triggered when the user resizes the window $\implies$ updates `w.size` and resizes rasterizer buffer.
- `ButtonPress` / `ButtonRelease`: Mouse clicks $\implies$ routes coordinates $(X, Y)$ to `DispatchPointerDown` / `DispatchPointerUp`.
- `MotionNotify`: Mouse movement $\implies$ routes coordinates $(X, Y)$ to `DispatchPointerMove` for hover states.
- `KeyPress`: Keyboard input $\implies$ translates XKeySym into `input.Key` and passes to `DispatchKeyDown`.
- `ClientMessage` (`WM_DELETE_WINDOW`): Close button clicked $\implies$ invokes `OnClose` and cleans up display resources.

Now that we understand how the OS window is created and updated, let's explore Nova's internal 2D software rasterization pipeline in Chapter 3!
