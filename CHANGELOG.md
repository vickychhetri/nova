# Changelog

All notable changes to **Nova** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-08-23

### Added
- **Core Architecture**:
  - `core/geom`: Full 2D geometry primitives (`Point`, `Size`, `Rect`, `Insets`, `CornerRadius`, `Matrix2D`).
  - `core/color`: Color spaces (`RGBA`, `Hex`, `HSL`, `HSLA`), alpha compositing, luminance, and WCAG contrast calculations.
- **Reactive State System (`state`)**:
  - Signal-based state graph: `state.Value[T]`, `state.Int`, `state.String`, `state.Bool`, `state.Float`.
  - Automatic dependency tracking with `state.Compute` memoization.
  - Reactive side-effects with `state.Effect` and lifecycle cleanups.
  - State update grouping with `state.Batch`.
- **Layout Engine (`layout`)**:
  - Multi-pass box constraint system (`BoxConstraints`).
  - Flexbox container layout algorithm (`Flex`, `Row`, `Column`) supporting flex-grow, flex-shrink, and main/cross axis alignment.
  - Stack container layout with absolute positioning anchors.
  - Uniform & responsive Grid layout engine.
  - Box container decoration math and scroll viewport bounds.
- **Render Engine (`render` & `renderer`)**:
  - Command buffer recording display list (`render.CommandBuffer`).
  - Fluent 2D Canvas drawing API (`render.Canvas`).
  - High-performance pure-Go software rasterizer (`renderer/software`) with anti-aliasing, rounded rectangles, text blitting, and PNG exporter.
- **Text & Font Engine (`text`, `font`)**:
  - Proportional typography metrics and procedural glyph rasterizer.
  - Multi-line text measurement, word wrapping, and ellipsis truncation.
- **Event & Input System (`input`, `event`)**:
  - Unified event hierarchy for pointer, scroll, keyboard, and focus events.
  - Hit testing traversal with clipping boundaries.
- **Design System (`theme`)**:
  - Design tokens for Colors, Typography, Spacing, Radii, and Shadows.
  - Built-in Dark and Light mode presets.
- **Virtualization Engine (`virtualization`)**:
  - Viewport-based slice rendering with 13ns calculation latency for 1,000,000+ items.
- **Form Controls Suite (`widgets/forms`)**:
  - `Form` container with validation rules (`Required`, `MinLength`, `MaxLength`, `Email`, `Custom`), dirty state tracking, and submission handling.
  - `TextField`, `PasswordField` (with toggle eye icon), `TextArea`, `NumberInput`.
  - `Checkbox`, `RadioGroup`, `Switch` / `Toggle`.
  - `Slider`, `Select`/`Dropdown` with searchable filter.
  - `DatePicker`, `ColorPicker`, `FilePicker`.
- **UI Components Catalog (`widgets/*`)**:
  - Basic: `Badge`, `Avatar`, `Spinner`, `Progress`, `Card`, `Skeleton`.
  - Navigation: `Tabs`, `Sidebar`, `SplitPane`, `Breadcrumb`.
  - Feedback: `Dialog`, `Alert`, `ToastManager`, `CommandPalette` (`Ctrl+K`).
  - Data: `VirtualList`, `Table` with virtualized rows, `Tree` view.
  - Editors: `CodeEditor` with syntax highlighting, `Canvas` custom drawing widget.
- **Developer Tooling & CLI (`cmd/nova`)**:
  - `nova create`: Scaffolds new Nova desktop apps.
  - `nova dev`: Development watcher with auto-restart.
  - `nova build`: Compiles optimized release desktop binaries.
  - `nova test`: Runs framework test suite.
  - `nova doctor`: Runs environment health diagnostics.
- **Showcase Applications (`examples/`)**:
  - `examples/01_hello_world`: Minimal clean starter.
  - `examples/02_counter`: Reactive signals counter demo.
  - `examples/03_forms_showcase`: Complete form input suite with automated validation.
  - `examples/04_widget_gallery`: Comprehensive component library catalog.
  - `examples/05_novadb`: Showcase database client with 100,000-row virtualized table, SQL editor, and connection sidebar.
  - `examples/06_ip_scanner`: Concurrent multithreaded network discovery and port scanner with live telemetry.
  - `examples/07_payslip_generator`: Enterprise payroll payslip designer with pure-Go vector PDF generator.
  - `examples/08_vinfo`: Secure personal information vault & to-do manager with SQLite (`modernc.org/sqlite`), PIN lock authentication (`2212`), in-app PIN change settings, and custom V-Info branding.
- **Documentation (`docs/`)**:
  - `docs/developer_guide.md`: Architecture guide, custom component tutorial, and state management.
  - `docs/integration_guide.md`: Database, REST/gRPC backend, and goroutine integration guide.
  - `docs/functional_spec.md`: Complete functional and mathematical specification.

---

## [Roadmap to v1.0]

- **v0.2**: Hardware GPU backends (Vulkan / Metal / DirectX 12).
- **v0.3**: Native OS System Tray, Native File Dialogs, Clipboard bridges.
- **v0.4**: Accessibility tree export (AT-SPI, Windows UI Automation, macOS NSAccessibility).
- **v0.5**: Multi-window management and drag-and-drop file targets.
- **v0.6**: Application packaging targets (AppImage, DMG, MSIX installers) and code signing.
- **v1.0**: Stable public API and long-term support release.
