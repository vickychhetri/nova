# 📚 Deep Inside Nova: The Architectural Internals & 2D GUI Engine Manual

Welcome to the definitive internal architectural manual for the **Nova GUI Engine**.

This book is written for software engineers, systems programmers, and Go developers who want to understand exactly how a terminal/backend language like Go can draw high-performance, native 60+ FPS graphical user interfaces directly on your desktop—without Electron, Chromium, Qt, GTK, or heavy external C++ frameworks.

---

## 📑 Complete Book Table of Contents

| Chapter | Title | Topics Covered |
|---|---|---|
| **[01. The Mystery of GUI in Go](01_the_mystery_of_gui_in_go.md)** | From Terminal to Pixels | How Go programs talk to the OS, what a pixel is in RAM, memory layouts, framebuffers, and why Go doesn't need Chromium/Electron. |
| **[02. OS Windowing & Display Servers](02_os_windowing_and_display_servers.md)** | The OS Platform Layer | X11, Wayland, Windows Desktop Window Manager (DWM), macOS Quartz/Cocoa. Cgo FFI, `XCreateWindow`, shared memory, and event polling. |
| **[03. Graphics Pipeline & Rasterization](03_graphics_pipeline_and_rasterization.md)** | 2D Software Rasterizer | Bresenham algorithms, scanline polygon filling, subpixel anti-aliasing, coverage buffers, Little-Endian RGBA/BGRA byte swapping, and alpha blending math. |
| **[04. Typography & Vector Fonts](04_typography_and_vector_fonts.md)** | TrueType Typography Engine | TrueType vector outlines, Quadratic & Cubic Bézier curves, font metrics, text shaping, word wrapping algorithms, and glyph atlas caching. |
| **[05. Declarative UI & The Node Graph](05_declarative_ui_and_component_tree.md)** | Declarative Component Tree | `ui.Component` vs `ui.Node`, mount lifecycle, tree reconciliation, depth-first paint traversal, overlays, and floating portals. |
| **[06. Layout Engine & Box Constraints](06_layout_engine_and_box_constraints.md)** | Constraint Solving & Flexbox | Min/Max width/height constraints, tight vs loose layout, Flexbox Row/Column distribution, Spacer expansion, and CrossAxisAlignment math. |
| **[07. Event Loop, Routing & Hit Testing](07_event_loop_and_hit_testing.md)** | Input Dispatch & Threading | Pointer events, recursive spatial hit testing, focus transitions, hover states, keyboard key mapping, and non-blocking mutex boundaries. |
| **[08. Reactive State & Reactivity System](08_reactive_state_and_reactivity.md)** | Fine-Grained Signals | `state.Value[T]`, reactive dependencies, computed signals, effect batching, and lightweight Canvas invalidation vs full tree mounting. |
| **[09. High-Performance Virtualization](09_virtualization_and_high_performance_data.md)** | 1,000,000 Items at 60 FPS | Viewport slicing math, index offsets, zero-allocation list virtualization, and scrollable tabular data grids. |
| **[10. Animation, Physics & Real-Time Loops](10_animation_and_physics_engines.md)** | Animation & Game Physics | 30–60 FPS ticker loops, lerp interpolation, easing curves, 2D ball kinematics, paddle deflection angles, and Minimax AI. |
| **[11. Enterprise Architectural Patterns](11_architectural_patterns_and_real_world_apps.md)** | Real-World Application Design | SQLite WAL persistence, AES-256-GCM encrypted local storage, file dialogs, background worker pipelines, and form validations. |
| **[12. Engine Reference, Benchmarks & Future](12_complete_engine_reference_and_future.md)** | Reference & Future GPU Roadmap | Complete system architecture diagrams, performance benchmarks, memory overhead comparisons, and Vulkan/Metal hardware acceleration. |

---

## 🎯 Target Audience & Reading Approach

- **If you are curious how GUI fundamentally works**: Read **Chapters 1, 2, and 3** to demystify how RAM bytes become illuminated monitor pixels.
- **If you are building complex UI layouts**: Read **Chapters 5, 6, and 8** to master declarative components, flexbox layout, and reactive state.
- **If you want to build high-performance tools or games**: Read **Chapters 9, 10, and 11** for virtualization, 60 FPS loops, and SQLite persistence.
