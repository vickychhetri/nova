# 📜 Nova Game Arcade — Changelog & Version History

All notable updates and improvements to Nova Game Arcade are documented in this file.

---

## [v1.1.2] - 2026-08-24

### 🔒 Mutex Re-entrancy & Deadlock Elimination
- **Non-blocking Event Handlers**: Fixed a critical mutex re-entrancy deadlock in `window/window.go` where `DispatchPointerDown`, `DispatchPointerUp`, `DispatchPointerMove`, and `DispatchKeyDown` held `w.mu.Lock()` while invoking user click and key callbacks (which subsequently called `win.Content()` / `rebuildView()`).
- **Clean Lock Boundary**: UI event callbacks (`OnClick`, `OnPointerDown`, `OnKeyDown`, `OnPointerMove`) now execute completely outside mutex locks, ensuring immediate, smooth responsiveness without window freezing or unresponsiveness.
- **Window Close & X11 WM Protocol Fix**: Verified X11 `WM_DELETE_WINDOW` lifecycle so windows close cleanly and terminate the event loop without hangs.
- **Window Title Encoding**: Cleaned up title string formatting for full compatibility across Linux desktop environments.

---

## [v1.1.1] - 2026-08-24

### ⚡ Screen Flickering & Blinking Resolution
- **X11 Background Erase Suppression**: Enabled `XSetWindowBackgroundPixmap(display, win, None)` in `platform/linux/x11.go` to prevent the X server from wiping the background between software frame blits.
- **Thread-Safe Render Synchronization**: Added `sync.RWMutex` to `window.Window` protecting concurrent UI tree mounting and frame rendering.
- **High-Performance Canvas Invalidation**: Decoupled the 35ms game tick engine from full UI tree reconstruction—games now trigger lightweight `win.Invalidate()` canvas repaints instead of remounting all widgets every frame, eliminating frame tearing and blinking.

---

## [v1.1.0] - 2026-08-23

### ⌨️ Full Keyboard Controls & Engine Upgrades
- **Global Window Keyboard Dispatcher**: Added `win.OnKeyDown()` allowing instant, non-blocking keyboard capture across all games.
- **X11 Low-Level Key Mapper**: Updated `platform/linux/x11.go` to map full ASCII character sets (`A-Z`, `a-z`, `0-9`), arrows (`Up`, `Down`, `Left`, `Right`), Spacebar, Enter, Escape, and Backspace.
- **Universal Game Keys**:
  - **Retro Snake 2.0**: `Arrow Keys` or `W / A / S / D` to steer, `Spacebar` for Pause/Resume, `R` for instant restart.
  - **2048 Merge Puzzle**: `Arrow Keys` or `W / A / S / D` for 4-way tile sliding, `R` to restart.
  - **Brick Breaker (Breakout)**: `Left / Right Arrows` or `A / D` for paddle movement, `Spacebar` for Pause/Resume, `R` to restart.
  - **Space Defender (Invaders)**: `Left / Right Arrows` or `A / D` to steer ship, `Spacebar` to fire laser cannons, `R` to restart.
  - **Reflex / Reaction Speed**: `Spacebar` or `Enter` for instant millisecond trigger click.
  - **Minesweeper Classic**: `F` key to toggle flag mode, `R` to restart.

### 🕹️ Brick Breaker (Breakout) Enhancements
- **Interactive Paddle Slider**: Added a dedicated `forms.Slider` directly underneath the game canvas to allow dragging the paddle to any position.
- **Fast Step Controls**: Added `[◀◀ Fast Left]` and `[Fast Right ▶▶]` buttons alongside standard step buttons.
- **Enhanced Physics & Deflection**: Improved paddle boundary clamping and angle calculations.

### 🐍 Retro Snake 2.0 UI & Gameplay Redesign
- **Visual Head & Directional Eyes**: The snake head now renders dynamic eyes pointing in the current direction of movement (`Up`, `Down`, `Left`, `Right`).
- **Apple Stems & Golden Fruits**: Redesigned apple sprites with green stems and vibrant golden apples ($+30\text{pts}$).
- **Status & Heading Badges**: Added live direction badge (`Heading: ▶ RIGHT`) and game state indicators.
- **Responsive Layout**: Re-engineered board to $22\times 15$ ($528\times 360\text{px}$) with dedicated control sidebars.

### 📋 In-App Changelog System
- Added a new **`[Changelog]`** tab in the main arcade header to view release notes and updates directly inside the application.

---

## [v1.0.0] - 2026-08-23

### 🌟 Initial Multi-Tier Release
- **8 Complete Native Go Games**:
  - **Basic Tier**: Memory Card Match, Tic-Tac-Toe Neon (with Minimax AI), Reflex & Reaction Speed.
  - **Average Tier**: Retro Snake 2.0, 2048 Merge Puzzle, Minesweeper Classic.
  - **Advance Tier**: Brick Breaker (Breakout), Space Defender (Invaders).
- **Offline SQLite Leaderboard System**: Persistent high score tracking with WAL mode for all games.
- **Declarative Reactive UI**: Built entirely with Nova GUI components, animations, and immediate software canvas rasterization.
