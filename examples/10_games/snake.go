package main

import (
	"fmt"
	"math/rand"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

type GridPoint struct {
	X int
	Y int
}

type SnakeGame struct {
	Width       int // 22
	Height      int // 15
	Snake       []GridPoint
	Dir         Direction
	NextDir     Direction
	Food        GridPoint
	GoldenFood  GridPoint
	HasGolden   bool
	Score       int
	HighScore   int
	IsGameOver  bool
	IsPaused    bool
	HasStarted  bool
	SpeedPreset string // "Relaxed", "Normal", "Turbo"
}

func NewSnakeGame() *SnakeGame {
	g := &SnakeGame{
		Width:       22,
		Height:      15,
		SpeedPreset: "Normal",
	}
	g.Reset()
	return g
}

func (g *SnakeGame) Reset() {
	g.Snake = []GridPoint{
		{X: 8, Y: 7},
		{X: 7, Y: 7},
		{X: 6, Y: 7},
	}
	g.Dir = DirRight
	g.NextDir = DirRight
	g.Score = 0
	g.IsGameOver = false
	g.IsPaused = false
	g.HasStarted = true
	g.HasGolden = false
	g.spawnFood()
}

func (g *SnakeGame) spawnFood() {
	for {
		fx := rand.Intn(g.Width)
		fy := rand.Intn(g.Height)
		collides := false
		for _, p := range g.Snake {
			if p.X == fx && p.Y == fy {
				collides = true
				break
			}
		}
		if !collides {
			g.Food = GridPoint{X: fx, Y: fy}
			break
		}
	}

	// 25% chance of golden apple
	if rand.Float64() < 0.25 {
		g.HasGolden = true
		g.GoldenFood = GridPoint{X: rand.Intn(g.Width), Y: rand.Intn(g.Height)}
	} else {
		g.HasGolden = false
	}
}

func (g *SnakeGame) ChangeDir(newDir Direction) {
	if !g.HasStarted {
		g.HasStarted = true
		g.IsPaused = false
	}
	// Prevent 180-degree immediate reversal
	if (g.Dir == DirUp && newDir == DirDown) ||
		(g.Dir == DirDown && newDir == DirUp) ||
		(g.Dir == DirLeft && newDir == DirRight) ||
		(g.Dir == DirRight && newDir == DirLeft) {
		return
	}
	g.NextDir = newDir
}

func (g *SnakeGame) Step() bool {
	if g.IsGameOver || g.IsPaused || !g.HasStarted {
		return false
	}

	g.Dir = g.NextDir
	head := g.Snake[0]
	var nextHead GridPoint

	switch g.Dir {
	case DirUp:
		nextHead = GridPoint{X: head.X, Y: head.Y - 1}
	case DirDown:
		nextHead = GridPoint{X: head.X, Y: head.Y + 1}
	case DirLeft:
		nextHead = GridPoint{X: head.X - 1, Y: head.Y}
	case DirRight:
		nextHead = GridPoint{X: head.X + 1, Y: head.Y}
	}

	// Wall Collision
	if nextHead.X < 0 || nextHead.X >= g.Width || nextHead.Y < 0 || nextHead.Y >= g.Height {
		g.IsGameOver = true
		return true
	}

	// Self Collision
	for _, p := range g.Snake {
		if p.X == nextHead.X && p.Y == nextHead.Y {
			g.IsGameOver = true
			return true
		}
	}

	// Move snake
	g.Snake = append([]GridPoint{nextHead}, g.Snake...)

	// Food check
	if nextHead.X == g.Food.X && nextHead.Y == g.Food.Y {
		g.Score += 10
		if g.Score > g.HighScore {
			g.HighScore = g.Score
		}
		g.spawnFood()
	} else if g.HasGolden && nextHead.X == g.GoldenFood.X && nextHead.Y == g.GoldenFood.Y {
		g.Score += 30
		g.HasGolden = false
		if g.Score > g.HighScore {
			g.HighScore = g.Score
		}
	} else {
		// Remove tail
		g.Snake = g.Snake[:len(g.Snake)-1]
	}

	return true
}

func (g *SnakeGame) RenderCanvas() ui.Component {
	cellSize := 24.0
	boardW := float64(g.Width) * cellSize
	boardH := float64(g.Height) * cellSize

	return widgets.Canvas(boardW, boardH, func(canvas *render.Canvas, bounds geom.Rect) {
		canvas.PushClip(bounds)

		// Board background
		canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#0F172A"))
		canvas.StrokeRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#1E293B"), 1.5)

		// Grid lines (subtle)
		for x := 0; x <= g.Width; x++ {
			px := float64(x) * cellSize
			canvas.DrawLine(geom.Pt(px, 0), geom.Pt(px, boardH), color.Hex("#1E293B").WithAlpha(0.3), 1.0)
		}
		for y := 0; y <= g.Height; y++ {
			py := float64(y) * cellSize
			canvas.DrawLine(geom.Pt(0, py), geom.Pt(boardW, py), color.Hex("#1E293B").WithAlpha(0.3), 1.0)
		}

		// Food (Red Apple)
		fx := float64(g.Food.X)*cellSize + 3
		fy := float64(g.Food.Y)*cellSize + 3
		canvas.FillCircle(geom.Pt(fx+cellSize/2-3, fy+cellSize/2-3), cellSize/2-4, color.Hex("#EF4444"))
		// Apple stem
		canvas.FillRoundedRect(geom.NewRect(fx+cellSize/2-4, fy+1, 2, 4), geom.RadiusUniform(1), color.Hex("#10B981"))

		// Golden Food
		if g.HasGolden {
			gx := float64(g.GoldenFood.X)*cellSize + 3
			gy := float64(g.GoldenFood.Y)*cellSize + 3
			canvas.FillCircle(geom.Pt(gx+cellSize/2-3, gy+cellSize/2-3), cellSize/2-3, color.Hex("#FACC15"))
		}

		// Snake Body
		for i, p := range g.Snake {
			sx := float64(p.X)*cellSize + 2
			sy := float64(p.Y)*cellSize + 2
			bodyCol := color.Hex("#10B981") // Green
			if i == 0 {
				bodyCol = color.Hex("#34D399") // Bright Head
			}
			canvas.FillRoundedRect(geom.NewRect(sx, sy, cellSize-4, cellSize-4), geom.RadiusUniform(4), bodyCol)

			// Head Eyes
			if i == 0 {
				eyeCol := color.Hex("#000000")
				switch g.Dir {
				case DirRight:
					canvas.FillCircle(geom.Pt(sx+cellSize-9, sy+5), 2, eyeCol)
					canvas.FillCircle(geom.Pt(sx+cellSize-9, sy+cellSize-9), 2, eyeCol)
				case DirLeft:
					canvas.FillCircle(geom.Pt(sx+5, sy+5), 2, eyeCol)
					canvas.FillCircle(geom.Pt(sx+5, sy+cellSize-9), 2, eyeCol)
				case DirUp:
					canvas.FillCircle(geom.Pt(sx+5, sy+5), 2, eyeCol)
					canvas.FillCircle(geom.Pt(sx+cellSize-9, sy+5), 2, eyeCol)
				case DirDown:
					canvas.FillCircle(geom.Pt(sx+5, sy+cellSize-9), 2, eyeCol)
					canvas.FillCircle(geom.Pt(sx+cellSize-9, sy+cellSize-9), 2, eyeCol)
				}
			}
		}

		// Game Over Canvas Overlay
		if g.IsGameOver {
			canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#000000").WithAlpha(0.65))
		}

		canvas.PopClip()
	})
}

func (g *SnakeGame) Render(onDir func(Direction), onReset func(), onPause func()) ui.Component {
	statusBadge := widgets.Badge(fmt.Sprintf("Score: %d", g.Score)).Success()
	if g.IsGameOver {
		statusBadge = widgets.Badge("GAME OVER! Press [R] to Restart").Error()
	} else if g.IsPaused {
		statusBadge = widgets.Badge("PAUSED (Space to Resume)").Warning()
	}

	dirLabel := "▶ RIGHT"
	switch g.Dir {
	case DirUp:
		dirLabel = "▲ UP"
	case DirDown:
		dirLabel = "▼ DOWN"
	case DirLeft:
		dirLabel = "◀ LEFT"
	}

	dPad := ui.Column(
		ui.Center(widgets.Button("[▲ Up (W)]").Secondary().OnClick(func() { onDir(DirUp) })),
		ui.Row(
			widgets.Button("[◀ Left (A)]").Secondary().OnClick(func() { onDir(DirLeft) }),
			widgets.Button("[▼ Down (S)]").Secondary().OnClick(func() { onDir(DirDown) }),
			widgets.Button("[▶ Right (D)]").Secondary().OnClick(func() { onDir(DirRight) }),
		).GapSpacing(6),
	).GapSpacing(6)

	pauseLabel := "[Pause / Resume (Space)]"
	if g.IsPaused {
		pauseLabel = "[▶ RESUME (Space)]"
	}

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("Retro Snake 2.0").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Average Tier").Info(),
					statusBadge,
					widgets.Badge(fmt.Sprintf("Length: %d", len(g.Snake))).Info(),
					widgets.Badge("Heading: "+dirLabel).Info(),
					ui.Spacer(),
					widgets.Button(pauseLabel).Secondary().OnClick(onPause),
					widgets.Button("[Restart Game (R)]").Primary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					g.RenderCanvas(),
					ui.Container().
						WithWidth(320).
						WithHeight(360).
						WithChild(
							ui.Column(
								widgets.Card("D-Pad Controls", dPad),
								widgets.Card("Keyboard Controls",
									ui.Text("• Arrow Keys / W, A, S, D: Move\n• Spacebar: Pause / Resume\n• R Key: Instant Restart\n• Red Apple: +10 pts\n• Gold Apple: +30 pts").Size(12).Col(color.Hex("#94A3B8")),
								),
							).GapSpacing(10),
						),
				).GapSpacing(20),
			).GapSpacing(12),
		)
}
