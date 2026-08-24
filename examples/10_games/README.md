# 🎮 Nova Game Arcade — Comprehensive Developer & Architecture Guide

A native, high-performance, multi-tier game suite built with the **Nova GUI Toolkit** for Go. Features **8 complete games** organized across three difficulty tiers (**Basic**, **Average**, and **Advance**), powered by a 60 FPS tick engine and persistent offline SQLite leaderboards.

---

## 🌟 Game Lineup by Tier

```
┌───────────────────────────────────────────────────────────────────────────────────┐
│                              NOVA GAME ARCADE SUITE                               │
├───────────────────────────────────────────────────────────────────────────────────┤
│  ★ BASIC TIER (Casual / Easy / Relaxed)                                           │
│    1. Memory Card Match       - Pair matching, Fisher-Yates shuffle, moves counter │
│    2. Tic-Tac-Toe Neon        - Minimax AI (Unbeatable), 2-Player, Win line math  │
│    3. Reflex & Reaction Speed - Millisecond precision timer, S/A/B/C rank system  │
├───────────────────────────────────────────────────────────────────────────────────┤
│  ⚡ AVERAGE TIER (Intermediate / Puzzle / Speed)                                   │
│    1. Retro Snake 2.0         - 22x15 grid canvas, food spawns, speed progression │
│    2. 2048 Merge Puzzle       - 4x4 matrix sliding/merging, score multipliers     │
│    3. Minesweeper Classic     - Safe first-click, flood-fill reveal, flag toggle  │
├───────────────────────────────────────────────────────────────────────────────────┤
│  🔥 ADVANCE TIER (Strategic / Action / High Score Arcade)                         │
│    1. Brick Breaker (Breakout)- Paddle friction deflection physics, brick matrix   │
│    2. Space Defender          - Starfield background, laser pooling, alien waves  │
├───────────────────────────────────────────────────────────────────────────────────┤
│  🏆 LEADERBOARD SYSTEM        - SQLite WAL mode, local persistent high scores     │
└───────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Technical Architecture

### 1. Unified Game Loop (`main.go`)
Real-time arcade games (`Snake`, `Breakout`, `Space Defender`, `Reaction`) execute on a decoupled background goroutine running at ~30–60 FPS (35ms ticker).
```go
go func() {
    ticker := time.NewTicker(35 * time.Millisecond)
    defer ticker.Stop()
    for range ticker.C {
        // Step active game engine
        if activeGame.Step() {
            rebuildView()
        }
    }
}()
```

### 2. Reactive UI & State Engine
Uses Nova's reactive state signals (`state.String()`) and declarative builder pattern (`ui.Container()`, `ui.Row()`, `ui.Column()`, `widgets.Card()`, `widgets.Canvas()`).

### 3. Persistent SQLite Storage (`db.go`)
- High scores, win streaks, and extra metadata (e.g. `Length: 28`, `14 Moves`, `Avg: 215ms`) are saved automatically to `games.db`.
- Configured with `PRAGMA journal_mode=WAL;` and `PRAGMA synchronous=NORMAL;` for ultra-fast, zero-overhead disk persistence.

---

## 🕹️ Deep-Dive: Game Logic & Mathematical Implementations

### 1. Tic-Tac-Toe: Minimax AI Algorithm (`tictactoe.go`)
- **State Representation**: 9-element string array `[9]string`.
- **Minimax Decision Tree**: Recursively evaluates all future game states:
  $$\text{Score}(O) = 10 - \text{depth}, \quad \text{Score}(X) = \text{depth} - 10, \quad \text{Draw} = 0$$
- **AI Modes**:
  - `Easy`: Random selection of unoccupied indices.
  - `Medium`: 50% minimax optimal move, 50% random move.
  - `Unbeatable`: 100% minimax recursion with alpha-beta depth minimization.

### 2. Memory Card Match: Permutation & Lifecycle (`memory.go`)
- **Deck Setup**: 8 paired symbols (`[Go]`, `[Rust]`, `[Py]`, `[JS]`, `[DB]`, `[API]`, `[Git]`, `[UI]`) duplicated into a 16-card slice.
- **Fisher-Yates Shuffling**: $O(n)$ unbiased random permutation:
  ```go
  rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
  ```
- **Turn State Machine**:
  - First card flip: Sets `FirstFlipped = idx`, cards remains face up.
  - Second card flip: Increments `Moves`. If `Cards[first].Symbol == Cards[second].Symbol`, sets `IsMatched = true`. Otherwise, flips both back.

### 3. Reflex & Reaction Speed: Timing Analysis (`reaction.go`)
- **State Sequence**: `StateIdle` $\rightarrow$ `StateWaiting` $\rightarrow$ `StateClickNow` $\rightarrow$ `StateRoundResult` $\rightarrow$ `StateGameOver`.
- **Random Delay**: Random trigger delay between $1500\text{ms}$ and $4000\text{ms}$ to eliminate anticipation.
- **Rank Evaluation**:
  - **S Rank**: $< 210\text{ms}$ (Godlike Reflexes)
  - **A Rank**: $210\text{ms} - 260\text{ms}$ (Pro Gamer)
  - **B Rank**: $260\text{ms} - 340\text{ms}$ (Good Reaction)
  - **C Rank**: $> 340\text{ms}$ (Casual)

### 4. Retro Snake 2.0: Grid Vector Kinematics (`snake.go`)
- **Grid Space**: $22 \times 15$ discrete coordinate space rendered with `render.Canvas`.
- **Direction Lock**: Prevents instantaneous $180^\circ$ self-reversals (e.g. `Left` while moving `Right`).
- **Body Mutation**:
  - New head added: $\vec{H}_{t+1} = \vec{H}_t + \vec{D}$.
  - Tail trimmed unless eating food: $\text{Snake} = [\vec{H}_{t+1}, \vec{B}_1, \dots, \vec{B}_{n-1}]$.
- **Spawning**: Random non-occupied cell selection with $25\%$ chance of Golden Apple ($+30\text{pts}$).

### 5. 2048 Merge Puzzle: Matrix Compression Algorithm (`game2048.go`)
- **Algorithm Step (`slideAndMerge`)**:
  1. Filter out all zeros from row/column slice.
  2. Iterate through filtered elements: if $arr[i] == arr[i+1]$, merge into $arr[i] \times 2$, add to score, and skip next element.
  3. Pad remaining elements with zeros up to size 4.
- **Tile Color Encoding**: Distinct hexadecimal palettes for values $2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048$.
- **Win & Termination**: Detects $\max(\text{tile}) \ge 2048$ for victory, and absence of adjacent equal tiles when board is full for game over.

### 6. Minesweeper: Recursive Flood-Fill (`minesweeper.go`)
- **First-Click Safety Guarantee**: Mines are placed *after* the player's first click, ensuring $(r_0, c_0)$ is never a mine.
- **Adjacent Mine Calculation**: 8-neighbor bounding check $\forall (dr, dc) \in \{-1, 0, 1\}^2 \setminus \{(0,0)\}$.
- **Zero-Neighbor Flood Fill**:
  ```go
  func (g *MinesweeperGame) reveal(r, c int) {
      if cell.IsRevealed || cell.IsFlagged || cell.IsMine { return }
      cell.IsRevealed = true
      if cell.AdjacentMines == 0 {
          for dr := -1; dr <= 1; dr++ {
              for dc := -1; dc <= 1; dc++ {
                  g.reveal(r+dr, c+dc)
              }
          }
      }
  }
  ```

### 7. Brick Breaker: 2D Deflection Physics (`breakout.go`)
- **Ball Kinematics**: $\vec{P}_{t+1} = \vec{P}_t + \vec{V}$.
- **Paddle Friction Deflection**:
  Calculates angle relative to paddle center offset $u \in [-1, 1]$:
  $$\theta = u \cdot \theta_{\max} \quad (\theta_{\max} \approx 55^\circ)$$
  $$V_x = \|\vec{V}\| \sin(\theta), \quad V_y = -\|\vec{V}\| \cos(\theta)$$
- **AABB vs Circle Collision**: Tests brick rectangle $[X, X+W] \times [Y, Y+H]$ against ball radius $R$.

### 8. Space Defender: Fleet Traversal & Projectiles (`space.go`)
- **Armada Step Formation**: 3 rows of 6 aliens moving horizontally at speed $V_{\text{alien}}$.
- **Edge Reversal & Descent**: When any alien touches screen edge ($X \le 20$ or $X+W \ge \text{Width}-20$), direction reverses ($V_{\text{alien}} \leftarrow -V_{\text{alien}}$) and all aliens step downward by $12\text{px}$.
- **Projectile Pooling**: Real-time laser struct pooling for both player cyan cannons and alien red dropping bombs.

---

## 🚀 How to Run and Test

### 1. Run Nova Arcade:
```bash
cd examples/10_games
go run .
```

### 2. Run Complete Unit Test Suite:
```bash
go test -v ./examples/10_games/...
```

### 3. Run Repository-Wide Test Suite:
```bash
go test -v ./...
```

---

## 🛠️ Developer Integration Guide: Adding a New Game

Adding a new game to Nova Arcade takes just 4 simple steps:

1. **Create the Game Engine (`mygame.go`)**:
   Define your struct with `Reset()`, `Step() bool`, and `Render(...) ui.Component`.
2. **Implement Input & Controls**:
   Wire mouse clicks or keyboard signals (`onMove`, `onClick`, `onReset`).
3. **Register in `main.go`**:
   - Instantiate your engine in `main()`.
   - Add a button in `subBar` under your chosen tier (`Basic`, `Average`, or `Advance`).
   - Add a case in `switch curGame` inside `buildMainView()`.
4. **Hook Score Persistence**:
   Call `db.RecordScore("mygame", playerName, score, difficulty, extraMeta)` on Game Over or Victory!
