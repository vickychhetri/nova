# Chapter 10: Animation, Physics & Real-Time Loops — Kinetics & AI

> *"How does Nova power real-time physics simulations, 60 FPS animation controllers, and game loops like Brick Breaker, Retro Snake, and Minimax AI in pure Go?"*

---

## 10.1 The Real-Time Tick Loop in Go

To drive real-time animation, Nova decouples simulation state from the OS event pump using a dedicated background ticker goroutine:

```go
go func() {
    ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
    defer ticker.Stop()

    for range ticker.C {
        if activeGame.Step() {
            win.Invalidate() // Fast repaint without recreating UI components!
        }
    }
}()
```

---

## 10.2 Linear & Eased Interpolation (Lerp)

When animating values (e.g. fading an opacity, expanding a sidebar, or moving a card), Nova uses **Linear Interpolation (`Lerp`)** (`animation/animation.go`):

$$\text{Lerp}(a, b, t) = a + (b - a) \times t, \quad 0 \le t \le 1$$

```go
func Lerp(a, b, t float64) float64 {
    return a + (b-a)*math.Max(0, math.Min(1, t))
}
```

### Easing Functions:
To make animations feel organic instead of robotic:

- **Ease-In (Cubic)**: $f(t) = t^3$ (Slow start, accelerates).
- **Ease-Out (Cubic)**: $f(t) = 1 - (1 - t)^3$ (Fast start, smoothly decelerates into place).
- **Ease-InOut**: Combines smooth acceleration and smooth braking.

---

## 10.3 2D Ball Kinematics & Paddle Deflection Physics

In **Brick Breaker (Breakout)** (`examples/10_games/breakout.go`), the ball moves using Euler numerical integration:

$$X_{t+1} = X_t + V_x, \quad Y_{t+1} = Y_t + V_y$$

```
                         Ceiling Reflection: Vy = -Vy
                      +-------------------------------+
                      |   [Brick]   [Brick]   [Brick] |
                      |                               |
                      |          ● (Vx, Vy)           |
                      |         /                     |
Wall Reflection:      |        /                      | Wall Reflection:
Vx = -Vx              |       /                       | Vx = -Vx
                      |      /                        |
                      |     ▼                         |
                      |   [==== Paddle ====]          |
                      +-------------------------------+
```

### Dynamic Paddle Deflection Angle Math:
When the ball hits the paddle, its bounce angle depends on where it struck the paddle relative to the center:

$$\text{Hit Offset} = \frac{X_{\text{ball}} - (X_{\text{paddle}} + \frac{W_{\text{paddle}}}{2})}{\frac{W_{\text{paddle}}}{2}}, \quad -1.0 \le \text{Offset} \le +1.0$$

$$\text{Deflection Angle } \theta = \text{Hit Offset} \times \theta_{\max} \quad (\theta_{\max} = 55^\circ)$$

$$\text{Speed } \|V\| = \sqrt{V_x^2 + V_y^2}$$

$$V_x = \|V\| \cdot \sin(\theta), \quad V_y = -|\|V\| \cdot \cos(\theta)|$$

This gives players surgical control: striking with the edge sends the ball speeding sharply sideways, while hitting dead-center bounces it straight up!

---

## 10.4 Unbeatable Minimax Game AI

In **Tic-Tac-Toe Neon** (`examples/10_games/tictactoe.go`), the AI uses the **Minimax Tree Search Algorithm**:

```
                         Current Board State (AI Turn - Maximizing)
                                     │
                 ┌───────────────────┼───────────────────┐
                 ▼                   ▼                   ▼
            Move at (0,0)       Move at (0,1)       Move at (1,1)
                 │                   │                   │
           (Opponent Turn)     (Opponent Turn)     (Opponent Turn)
                 ▼                   ▼                   ▼
            Min Score: -10      Min Score: 0        Min Score: +10 (BEST)
```

$$\text{Minimax}(B, d, \text{isMax}) = \begin{cases}
+10 - d & \text{if AI wins} \\
-10 + d & \text{if Player wins} \\
0 & \text{if Draw} \\
\max_{m} \text{Minimax}(B \cup \{m\}, d+1, \text{false}) & \text{if isMax} \\
\min_{m} \text{Minimax}(B \cup \{m\}, d+1, \text{true}) & \text{if isMin}
\end{cases}$$

Because the depth $d$ is factored in, the AI selects the fastest winning line while actively blocking player traps.

In Chapter 11, we explore Enterprise Architecture, SQLite WAL storage, and AES-256 encrypted vaults!
