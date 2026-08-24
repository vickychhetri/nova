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

type Alien struct {
	X       float64
	Y       float64
	W       float64
	H       float64
	Points  int
	Color   color.Color
	IsAlive bool
}

type Laser struct {
	X        float64
	Y        float64
	VY       float64
	IsAlien  bool
	IsActive bool
}

type SpaceGame struct {
	Width      float64 // 600
	Height     float64 // 400
	PlayerX    float64
	PlayerY    float64
	PlayerW    float64
	PlayerH    float64
	Aliens     []Alien
	AlienDir   float64 // 1.0 (right) or -1.0 (left)
	AlienSpeed float64
	Lasers     []Laser
	Score      int
	Lives      int
	Wave       int
	IsGameOver bool
	IsWon      bool
	IsPaused   bool
	Stars      []geom.Point
}

func NewSpaceGame() *SpaceGame {
	g := &SpaceGame{
		Width:      560,
		Height:     360,
		PlayerW:    34,
		PlayerH:    20,
		AlienDir:   1.0,
		AlienSpeed: 1.5,
	}

	// Generate static starfield
	for i := 0; i < 45; i++ {
		g.Stars = append(g.Stars, geom.Pt(rand.Float64()*560, rand.Float64()*360))
	}

	g.Reset()
	return g
}

func (g *SpaceGame) Reset() {
	g.PlayerX = (g.Width - g.PlayerW) / 2.0
	g.PlayerY = g.Height - 35
	g.Score = 0
	g.Lives = 3
	g.Wave = 1
	g.IsGameOver = false
	g.IsWon = false
	g.IsPaused = false
	g.Lasers = nil
	g.spawnAlienWave()
}

func (g *SpaceGame) spawnAlienWave() {
	g.Aliens = nil
	rows := 3
	cols := 6
	alienW := 30.0
	alienH := 18.0

	rowColors := []struct {
		col color.Color
		pts int
	}{
		{color.Hex("#F43F5E"), 40}, // Red Leader
		{color.Hex("#A855F7"), 20}, // Purple Cruiser
		{color.Hex("#38BDF8"), 10}, // Blue Drone
	}

	for r := 0; r < rows; r++ {
		rc := rowColors[r]
		for c := 0; c < cols; c++ {
			ax := 40 + float64(c)*60
			ay := 30 + float64(r)*32
			g.Aliens = append(g.Aliens, Alien{
				X:       ax,
				Y:       ay,
				W:       alienW,
				H:       alienH,
				Points:  rc.pts,
				Color:   rc.col,
				IsAlive: true,
			})
		}
	}
	g.AlienDir = 1.0
}

func (g *SpaceGame) MovePlayer(dx float64) {
	if g.IsGameOver || g.IsPaused {
		return
	}
	g.PlayerX += dx
	if g.PlayerX < 10 {
		g.PlayerX = 10
	}
	if g.PlayerX > g.Width-g.PlayerW-10 {
		g.PlayerX = g.Width - g.PlayerW - 10
	}
}

func (g *SpaceGame) FireLaser() {
	if g.IsGameOver || g.IsPaused {
		return
	}
	// Limit player laser spam to max 3 on screen
	playerLasers := 0
	for _, l := range g.Lasers {
		if l.IsActive && !l.IsAlien {
			playerLasers++
		}
	}
	if playerLasers < 3 {
		g.Lasers = append(g.Lasers, Laser{
			X:        g.PlayerX + g.PlayerW/2,
			Y:        g.PlayerY - 4,
			VY:       -8.0,
			IsAlien:  false,
			IsActive: true,
		})
	}
}

func (g *SpaceGame) Step() bool {
	if g.IsGameOver || g.IsPaused || g.IsWon {
		return false
	}

	// 1. Move Aliens
	edgeReached := false
	for _, a := range g.Aliens {
		if a.IsAlive {
			if (g.AlienDir > 0 && a.X+a.W >= g.Width-20) || (g.AlienDir < 0 && a.X <= 20) {
				edgeReached = true
				break
			}
		}
	}

	if edgeReached {
		g.AlienDir = -g.AlienDir
		for i := range g.Aliens {
			if g.Aliens[i].IsAlive {
				g.Aliens[i].Y += 12
				// Invasion breach
				if g.Aliens[i].Y+g.Aliens[i].H >= g.PlayerY {
					g.IsGameOver = true
					return true
				}
			}
		}
	} else {
		for i := range g.Aliens {
			if g.Aliens[i].IsAlive {
				g.Aliens[i].X += g.AlienDir * g.AlienSpeed
			}
		}
	}

	// Random Alien firing
	if rand.Float64() < 0.04 {
		var livingAliens []int
		for i, a := range g.Aliens {
			if a.IsAlive {
				livingAliens = append(livingAliens, i)
			}
		}
		if len(livingAliens) > 0 {
			shooter := livingAliens[rand.Intn(len(livingAliens))]
			g.Lasers = append(g.Lasers, Laser{
				X:        g.Aliens[shooter].X + g.Aliens[shooter].W/2,
				Y:        g.Aliens[shooter].Y + g.Aliens[shooter].H,
				VY:       4.5,
				IsAlien:  true,
				IsActive: true,
			})
		}
	}

	// 2. Update Lasers & Collisions
	livingAliensCount := 0
	for i := range g.Aliens {
		if g.Aliens[i].IsAlive {
			livingAliensCount++
		}
	}

	for i := range g.Lasers {
		l := &g.Lasers[i]
		if !l.IsActive {
			continue
		}
		l.Y += l.VY

		// Screen exit
		if l.Y < 0 || l.Y > g.Height {
			l.IsActive = false
			continue
		}

		if !l.IsAlien {
			// Player Laser hitting Alien
			for j := range g.Aliens {
				a := &g.Aliens[j]
				if a.IsAlive && l.X >= a.X && l.X <= a.X+a.W && l.Y >= a.Y && l.Y <= a.Y+a.H {
					a.IsAlive = false
					l.IsActive = false
					g.Score += a.Points
					livingAliensCount--
					break
				}
			}
		} else {
			// Alien Laser hitting Player
			if l.X >= g.PlayerX && l.X <= g.PlayerX+g.PlayerW && l.Y >= g.PlayerY && l.Y <= g.PlayerY+g.PlayerH {
				l.IsActive = false
				g.Lives--
				if g.Lives <= 0 {
					g.IsGameOver = true
				}
			}
		}
	}

	// Next Wave check
	if livingAliensCount == 0 {
		g.Wave++
		g.Score += 200
		g.AlienSpeed += 0.5
		g.spawnAlienWave()
	}

	return true
}

func (g *SpaceGame) RenderCanvas() ui.Component {
	return widgets.Canvas(g.Width, g.Height, func(canvas *render.Canvas, bounds geom.Rect) {
		canvas.PushClip(bounds)

		// Deep Space Background
		canvas.FillRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#050814"))
		canvas.StrokeRoundedRect(bounds, geom.RadiusUniform(8), color.Hex("#1E293B"), 1.5)

		// Stars
		for _, st := range g.Stars {
			canvas.FillCircle(st, 1.2, color.Hex("#FFFFFF").WithAlpha(0.6))
		}

		// Draw Aliens
		for _, a := range g.Aliens {
			if a.IsAlive {
				aRect := geom.NewRect(a.X, a.Y, a.W, a.H)
				canvas.FillRoundedRect(aRect, geom.RadiusUniform(4), a.Color)
				// Alien eye details
				canvas.FillCircle(geom.Pt(a.X+8, a.Y+8), 2.5, color.Hex("#FFFFFF"))
				canvas.FillCircle(geom.Pt(a.X+a.W-8, a.Y+8), 2.5, color.Hex("#FFFFFF"))
			}
		}

		// Draw Player Starship
		pRect := geom.NewRect(g.PlayerX, g.PlayerY, g.PlayerW, g.PlayerH)
		canvas.FillRoundedRect(pRect, geom.RadiusUniform(4), color.Hex("#10B981"))
		// Cannon nose
		canvas.FillRoundedRect(geom.NewRect(g.PlayerX+g.PlayerW/2-3, g.PlayerY-6, 6, 8), geom.RadiusUniform(2), color.Hex("#34D399"))

		// Draw Lasers
		for _, l := range g.Lasers {
			if l.IsActive {
				laserCol := color.Hex("#38BDF8") // Player Cyan Laser
				if l.IsAlien {
					laserCol = color.Hex("#EF4444") // Alien Red Laser
				}
				canvas.FillRoundedRect(geom.NewRect(l.X-1.5, l.Y-4, 3, 10), geom.RadiusUniform(1.5), laserCol)
			}
		}

		canvas.PopClip()
	})
}

func (g *SpaceGame) Render(onLeft func(), onRight func(), onFire func(), onReset func(), onPause func()) ui.Component {
	statusBadge := widgets.Badge(fmt.Sprintf("Score: %d", g.Score)).Success()
	if g.IsGameOver {
		statusBadge = widgets.Badge("FLEET DESTROYED - GAME OVER!").Error()
	}

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("Space Defender").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Advance Tier").Error(),
					statusBadge,
					widgets.Badge(fmt.Sprintf("Wave %d", g.Wave)).Info(),
					widgets.Badge(fmt.Sprintf("Lives: %d", g.Lives)).Info(),
					ui.Spacer(),
					widgets.Button("[Pause / Resume]").Secondary().OnClick(onPause),
					widgets.Button("[Restart Game]").Primary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					g.RenderCanvas(),
					ui.Container().
						WithWidth(320).
						WithChild(
							ui.Column(
								widgets.Card("Ship Controls",
									ui.Column(
										widgets.Button("[FIRE CANNON (Space)]").Primary().OnClick(onFire),
										ui.Row(
											widgets.Button("[◀ Left]").Secondary().OnClick(onLeft),
											widgets.Button("[Right ▶]").Secondary().OnClick(onRight),
										).GapSpacing(6),
									).GapSpacing(8),
								),
								ui.Text("Controls:\n• Left / Right Arrow keys\n• Spacebar to Fire\n• Destroy the alien armada!").Size(11).Col(color.Hex("#94A3B8")),
							).GapSpacing(10),
						),
				).GapSpacing(20),
			).GapSpacing(12),
		)
}
