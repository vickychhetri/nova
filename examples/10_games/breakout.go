package main

import (
	"fmt"
	"math"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/forms"
)

type Brick struct {
	X       float64
	Y       float64
	W       float64
	H       float64
	Points  int
	Color   color.Color
	IsAlive bool
}

type BreakoutGame struct {
	Width       float64 // 560
	Height      float64 // 360
	PaddleX     float64
	PaddleW     float64
	PaddleH     float64
	PaddleSpeed float64

	BallX  float64
	BallY  float64
	BallVX float64
	BallVY float64
	BallR  float64

	Bricks     []Brick
	Score      int
	Lives      int
	IsGameOver bool
	IsWon      bool
	IsPaused   bool
}

func NewBreakoutGame() *BreakoutGame {
	g := &BreakoutGame{
		Width:       560,
		Height:      360,
		PaddleW:     96,
		PaddleH:     14,
		PaddleSpeed: 28,
		BallR:       6,
	}
	g.Reset()
	return g
}

func (g *BreakoutGame) Reset() {
	g.PaddleX = (g.Width - g.PaddleW) / 2.0
	g.BallX = g.Width / 2.0
	g.BallY = g.Height - 60
	g.BallVX = 4.0
	g.BallVY = -5.0
	g.Score = 0
	g.Lives = 3
	g.IsGameOver = false
	g.IsWon = false
	g.IsPaused = false

	// Build 4 rows of 8 bricks
	rows := 4
	cols := 8
	brickW := (g.Width - 40 - float64(cols-1)*6) / float64(cols)
	brickH := 18.0

	g.Bricks = nil
	rowColors := []struct {
		col color.Color
		pts int
	}{
		{color.Hex("#EF4444"), 30}, // Red
		{color.Hex("#F59E0B"), 20}, // Orange
		{color.Hex("#10B981"), 15}, // Green
		{color.Hex("#38BDF8"), 10}, // Blue
	}

	for r := 0; r < rows; r++ {
		rc := rowColors[r]
		for c := 0; c < cols; c++ {
			bx := 20 + float64(c)*(brickW+6)
			by := 40 + float64(r)*(brickH+6)
			g.Bricks = append(g.Bricks, Brick{
				X:       bx,
				Y:       by,
				W:       brickW,
				H:       brickH,
				Points:  rc.pts,
				Color:   rc.col,
				IsAlive: true,
			})
		}
	}
}

func (g *BreakoutGame) MovePaddle(dx float64) {
	if g.IsGameOver || g.IsPaused {
		return
	}
	g.PaddleX += dx
	if g.PaddleX < 0 {
		g.PaddleX = 0
	}
	if g.PaddleX > g.Width-g.PaddleW {
		g.PaddleX = g.Width - g.PaddleW
	}
}

func (g *BreakoutGame) SetPaddlePosition(x float64) {
	if g.IsGameOver || g.IsPaused {
		return
	}
	g.PaddleX = x - g.PaddleW/2
	if g.PaddleX < 0 {
		g.PaddleX = 0
	}
	if g.PaddleX > g.Width-g.PaddleW {
		g.PaddleX = g.Width - g.PaddleW
	}
}

func (g *BreakoutGame) Step() bool {
	if g.IsGameOver || g.IsPaused || g.IsWon {
		return false
	}

	// Update ball position
	g.BallX += g.BallVX
	g.BallY += g.BallVY

	// Wall collisions (Left & Right)
	if g.BallX-g.BallR <= 0 {
		g.BallX = g.BallR
		g.BallVX = -g.BallVX
	} else if g.BallX+g.BallR >= g.Width {
		g.BallX = g.Width - g.BallR
		g.BallVX = -g.BallVX
	}

	// Top ceiling collision
	if g.BallY-g.BallR <= 0 {
		g.BallY = g.BallR
		g.BallVY = -g.BallVY
	}

	// Paddle Collision
	paddleY := g.Height - 35
	if g.BallVY > 0 && g.BallY+g.BallR >= paddleY && g.BallY-g.BallR <= paddleY+g.PaddleH {
		if g.BallX >= g.PaddleX && g.BallX <= g.PaddleX+g.PaddleW {
			// Deflection angle based on where it hit the paddle
			hitOffset := (g.BallX - (g.PaddleX + g.PaddleW/2)) / (g.PaddleW / 2)
			speed := math.Sqrt(g.BallVX*g.BallVX + g.BallVY*g.BallVY)
			maxAngle := math.Pi / 3.2 // 55 degrees
			angle := hitOffset * maxAngle

			g.BallVX = speed * math.Sin(angle)
			g.BallVY = -math.Abs(speed * math.Cos(angle))
			g.BallY = paddleY - g.BallR
		}
	}

	// Brick Collisions
	bricksRemaining := 0
	for i := range g.Bricks {
		b := &g.Bricks[i]
		if !b.IsAlive {
			continue
		}
		bricksRemaining++

		// AABB vs Circle check
		if g.BallX+g.BallR >= b.X && g.BallX-g.BallR <= b.X+b.W &&
			g.BallY+g.BallR >= b.Y && g.BallY-g.BallR <= b.Y+b.H {
			b.IsAlive = false
			g.Score += b.Points
			g.BallVY = -g.BallVY
			bricksRemaining--
			break
		}
	}

	if bricksRemaining == 0 {
		g.IsWon = true
		return true
	}

	// Bottom loss
	if g.BallY > g.Height+20 {
		g.Lives--
		if g.Lives <= 0 {
			g.IsGameOver = true
		} else {
			// Respawn ball on paddle
			g.BallX = g.PaddleX + g.PaddleW/2
			g.BallY = g.Height - 60
			g.BallVX = 4.0
			g.BallVY = -5.0
		}
	}

	return true
}

func (g *BreakoutGame) RenderCanvas() ui.Component {
	return widgets.Canvas(g.Width, g.Height, func(canvas *render.Canvas, bounds geom.Rect) {
		canvas.PushClip(bounds)

		// Board background
		canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#0F172A"))
		canvas.StrokeRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#1E293B"), 1.5)

		// Draw Bricks
		for _, b := range g.Bricks {
			if b.IsAlive {
				bRect := geom.NewRect(b.X, b.Y, b.W, b.H)
				canvas.FillRoundedRect(bRect, geom.RadiusUniform(3), b.Color)
				canvas.StrokeRoundedRect(bRect, geom.RadiusUniform(3), color.Hex("#000000").WithAlpha(0.3), 1.0)
			}
		}

		// Draw Paddle
		paddleY := g.Height - 35
		pRect := geom.NewRect(g.PaddleX, paddleY, g.PaddleW, g.PaddleH)
		canvas.FillRoundedRect(pRect, geom.RadiusUniform(6), color.Hex("#38BDF8"))
		canvas.StrokeRoundedRect(pRect, geom.RadiusUniform(6), color.Hex("#0284C7"), 1.5)

		// Draw Ball
		canvas.FillCircle(geom.Pt(g.BallX, g.BallY), g.BallR, color.Hex("#F8FAFC"))
		canvas.FillCircle(geom.Pt(g.BallX, g.BallY), g.BallR-2, color.Hex("#FACC15"))

		// Overlay if Game Over or Victory
		if g.IsGameOver || g.IsWon {
			canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#000000").WithAlpha(0.65))
		}

		canvas.PopClip()
	})
}

func (g *BreakoutGame) Render(onLeft func(), onRight func(), onReset func(), onPause func(), onSliderMove func(float64)) ui.Component {
	statusBadge := widgets.Badge(fmt.Sprintf("Score: %d", g.Score)).Success()
	if g.IsGameOver {
		statusBadge = widgets.Badge("BALL LOST - GAME OVER! (Press R)").Error()
	} else if g.IsWon {
		statusBadge = widgets.Badge("ALL BRICKS CLEARED - VICTORY!").Success()
	} else if g.IsPaused {
		statusBadge = widgets.Badge("PAUSED (Space to Resume)").Warning()
	}

	slider := forms.Slider(0, g.Width-g.PaddleW).
		WithWidth(540).
		WithStep(5)
	slider.Value.Set(g.PaddleX)
	slider.OnChanged = func(val float64) {
		onSliderMove(val)
	}

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("Brick Breaker (Breakout)").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Advance Tier").Error(),
					statusBadge,
					widgets.Badge(fmt.Sprintf("Lives: %d", g.Lives)).Info(),
					ui.Spacer(),
					widgets.Button("[Pause / Resume (Space)]").Secondary().OnClick(onPause),
					widgets.Button("[Restart Game (R)]").Primary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					ui.Column(
						g.RenderCanvas(),
						ui.Container().
							Bg(color.Hex("#1E293B")).
							Pad(geom.Insets{Top: 8, Bottom: 8, Left: 10, Right: 10}).
							Rounded(geom.RadiusUniform(8)).
							WithChild(
								ui.Row(
									ui.Text("Paddle Slider:").Size(12).Weight(font.WeightBold).Col(color.Hex("#38BDF8")),
									slider,
								).GapSpacing(10),
							),
					).GapSpacing(8),

					ui.Container().
						WithWidth(320).
						WithChild(
							ui.Column(
								widgets.Card("Paddle Controls",
									ui.Column(
										ui.Row(
											widgets.Button("[◀◀ Fast Left]").Secondary().OnClick(func() { onLeft(); onLeft() }),
											widgets.Button("[Fast Right ▶▶]").Secondary().OnClick(func() { onRight(); onRight() }),
										).GapSpacing(6),
										ui.Row(
											widgets.Button("[◀ Left (A)]").Primary().OnClick(onLeft),
											widgets.Button("[Right (D) ▶]").Primary().OnClick(onRight),
										).GapSpacing(6),
									).GapSpacing(8),
								),
								widgets.Card("Keyboard Controls",
									ui.Text("• Arrow Left / Right or A / D: Move\n• Drag the Paddle Slider below\n• Spacebar: Pause / Resume\n• R Key: Restart Game").Size(12).Col(color.Hex("#94A3B8")),
								),
							).GapSpacing(10),
						),
				).GapSpacing(20),
			).GapSpacing(12),
		)
}
