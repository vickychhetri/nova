# Nova Sample Desktop Application

This is a complete, standalone, production-ready desktop application built with the **Nova Go Native Desktop Framework** (`github.com/vickychhetri/nova`).

---

## What This Sample Application Demonstrates

1. **Window Management & Lifecycle**:
   - Window Title, 1200x820 dimensions, automatic centering, Dark Theme design tokens.
2. **Dashboard & Analytics Tab**:
   - Stat metric cards, active user counters, GPU render pipeline status alerts, progress bars, badges, and avatars.
3. **Complete Form Controls & Validation Tab**:
   - `TextField`, `PasswordField` (with eye toggle), `TextArea`, `NumberInput` (stepper).
   - `Checkbox`, `Switch` / `Toggle`.
   - `Slider`, `Select`/`Dropdown` with options.
   - `DatePicker`, `ColorPicker` (HEX/RGB swatch), `FilePicker` (upload dropzone).
   - Automatic `forms.NewFormState()` validation engine with real-time schema error reporting.
4. **100,000 Rows Virtualized Table Explorer Tab**:
   - High-throughput virtualization engine rendering 100k records at sub-millisecond speeds.
5. **SQL Studio & Code Editor Tab**:
   - Syntax-highlighted code editor with keyword tokens, line numbers gutter, and run query actions.
6. **Navigation & Overlays**:
   - Collapsible left navigation sidebar, tabbed views, resizable split-pane layout, and confirmation dialog modal.

---

## How to Run

### Development Mode

```bash
cd sample_app
go run main.go
```
or:
```bash
./run.sh
```

---

## How to Build Standalone Desktop Binary

```bash
cd sample_app
go build -ldflags="-s -w" -o bin/desktop_app .
```
or:
```bash
./build.sh
```

Execute the compiled desktop application binary:
```bash
./bin/desktop_app
```

---

## Project Structure

```
sample_app/
├── go.mod           # Module definition with local Nova replace directive
├── main.go          # Full desktop application source code
├── build.sh         # Production binary build script
├── run.sh           # Development run script
├── app_preview.png  # Rendered UI window snapshot
└── README.md        # Documentation
```
