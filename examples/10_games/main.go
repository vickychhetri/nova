package main

import (
	"fmt"
	"time"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/input"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

func main() {
	// 1. Initialize SQLite Database
	dbPath := "games.db"
	db, err := OpenDatabase(dbPath)
	if err != nil {
		fmt.Printf("Error opening games database: %v\n", err)
		return
	}
	defer db.Close()

	// 2. Initialize Nova Application
	app := nova.New()
	win := app.Window(
		nova.Title("Nova Game Arcade - Keyboard and Mouse Multi-Tier Games"),
		nova.Size(1120, 840),
		nova.Theme(theme.Dark()),
	)

	// 3. Application State Signals
	activeTier := state.String("Basic")  // "Basic", "Average", "Advance", "Leaderboard", "Changelog"
	activeGame := state.String("snake") // default active game
	difficultyMod := state.String("Average") // "Basic", "Average", "Advance"
	toastMsg := state.String("Keyboard Controls Active! Arrow keys / WASD / Space / R work in all games.")

	// Game Engine Instances
	tttGame := NewTicTacToeGame()
	memGame := NewMemoryGame()
	reactGame := NewReactionGame()
	snakeGame := NewSnakeGame()
	game2048 := NewGame2048()
	mineGame := NewMinesweeperGame()
	breakoutGame := NewBreakoutGame()
	spaceGame := NewSpaceGame()

	var rebuildView func()

	// 4. Logo Component Helper (Arcade Gamepad Emblem)
	renderArcadeLogo := func(size float64) ui.Component {
		return widgets.Canvas(size, size, func(canvas *render.Canvas, bounds geom.Rect) {
			w := bounds.Width
			h := bounds.Height
			radius := geom.RadiusUniform(w * 0.25)

			// Controller Body
			rect := geom.NewRect(0, h*0.15, w, h*0.7)
			canvas.FillRoundedRect(rect, radius, color.Hex("#6366F1"))
			canvas.StrokeRoundedRect(rect, radius, color.Hex("#818CF8"), 1.5)

			// D-Pad Cross (Left)
			canvas.FillRoundedRect(geom.NewRect(w*0.18, h*0.35, w*0.08, h*0.3), geom.RadiusUniform(2), color.Hex("#1E1B4B"))
			canvas.FillRoundedRect(geom.NewRect(w*0.12, h*0.44, w*0.2, h*0.12), geom.RadiusUniform(2), color.Hex("#1E1B4B"))

			// Action Buttons (Right)
			canvas.FillCircle(geom.Pt(w*0.75, h*0.4), w*0.06, color.Hex("#EF4444")) // Red
			canvas.FillCircle(geom.Pt(w*0.85, h*0.5), w*0.06, color.Hex("#FACC15")) // Yellow
			canvas.FillCircle(geom.Pt(w*0.65, h*0.5), w*0.06, color.Hex("#38BDF8")) // Blue
			canvas.FillCircle(geom.Pt(w*0.75, h*0.6), w*0.06, color.Hex("#10B981")) // Green
		})
	}

	// 5. Global Keyboard Input Dispatcher
	win.OnKeyDown(func(e *event.KeyEvent) {
		curGame := activeGame.Get()
		k := e.Key
		r := e.Rune

		switch curGame {
		case "snake":
			if k == input.KeyArrowUp || r == 'w' || r == 'W' {
				snakeGame.ChangeDir(DirUp)
			} else if k == input.KeyArrowDown || r == 's' || r == 'S' {
				snakeGame.ChangeDir(DirDown)
			} else if k == input.KeyArrowLeft || r == 'a' || r == 'A' {
				snakeGame.ChangeDir(DirLeft)
			} else if k == input.KeyArrowRight || r == 'd' || r == 'D' {
				snakeGame.ChangeDir(DirRight)
			} else if k == input.KeySpace || r == 'p' || r == 'P' {
				snakeGame.IsPaused = !snakeGame.IsPaused
			} else if r == 'r' || r == 'R' {
				snakeGame.Reset()
			}
			if rebuildView != nil {
				rebuildView()
			}

		case "2048":
			if k == input.KeyArrowLeft || r == 'a' || r == 'A' {
				game2048.MoveLeft()
			} else if k == input.KeyArrowRight || r == 'd' || r == 'D' {
				game2048.MoveRight()
			} else if k == input.KeyArrowUp || r == 'w' || r == 'W' {
				game2048.MoveUp()
			} else if k == input.KeyArrowDown || r == 's' || r == 'S' {
				game2048.MoveDown()
			} else if r == 'r' || r == 'R' {
				game2048.Reset()
			}
			if rebuildView != nil {
				rebuildView()
			}

		case "breakout":
			if k == input.KeyArrowLeft || r == 'a' || r == 'A' {
				breakoutGame.MovePaddle(-breakoutGame.PaddleSpeed)
			} else if k == input.KeyArrowRight || r == 'd' || r == 'D' {
				breakoutGame.MovePaddle(breakoutGame.PaddleSpeed)
			} else if k == input.KeySpace || r == 'p' || r == 'P' {
				breakoutGame.IsPaused = !breakoutGame.IsPaused
			} else if r == 'r' || r == 'R' {
				breakoutGame.Reset()
			}
			if rebuildView != nil {
				rebuildView()
			}

		case "space":
			if k == input.KeyArrowLeft || r == 'a' || r == 'A' {
				spaceGame.MovePlayer(-25)
			} else if k == input.KeyArrowRight || r == 'd' || r == 'D' {
				spaceGame.MovePlayer(25)
			} else if k == input.KeySpace {
				spaceGame.FireLaser()
			} else if r == 'p' || r == 'P' {
				spaceGame.IsPaused = !spaceGame.IsPaused
			} else if r == 'r' || r == 'R' {
				spaceGame.Reset()
			}
			if rebuildView != nil {
				rebuildView()
			}

		case "reaction":
			if k == input.KeySpace || k == input.KeyEnter {
				reactGame.HandleClick()
			} else if r == 'r' || r == 'R' {
				reactGame.Reset()
			}
			if rebuildView != nil {
				rebuildView()
			}

		case "minesweeper":
			if r == 'f' || r == 'F' {
				mineGame.FlagMode = !mineGame.FlagMode
			} else if r == 'r' || r == 'R' {
				mineGame.Reset()
			}
			if rebuildView != nil {
				rebuildView()
			}
		}
	})

	// 6. Game Loop Tick Engine (for real-time games: Snake, Breakout, Space)
	go func() {
		ticker := time.NewTicker(35 * time.Millisecond) // ~30 FPS loop
		defer ticker.Stop()

		snakeTickCounter := 0

		for range ticker.C {
			currentGame := activeGame.Get()
			needsRedraw := false

			switch currentGame {
			case "reaction":
				if reactGame.CheckTimer() {
					needsRedraw = true
				}

			case "snake":
				snakeTickCounter++
				// Speed factor
				snakeSpeedTicks := 3
				if difficultyMod.Get() == "Basic" {
					snakeSpeedTicks = 4
				} else if difficultyMod.Get() == "Advance" {
					snakeSpeedTicks = 2
				}

				if snakeTickCounter >= snakeSpeedTicks {
					snakeTickCounter = 0
					if snakeGame.Step() {
						needsRedraw = true
						if snakeGame.IsGameOver && snakeGame.Score > 0 {
							_, _ = db.RecordScore("snake", "Player 1", snakeGame.Score, difficultyMod.Get(), fmt.Sprintf("Length: %d", len(snakeGame.Snake)))
						}
					}
				}

			case "breakout":
				if breakoutGame.Step() {
					needsRedraw = true
					if (breakoutGame.IsGameOver || breakoutGame.IsWon) && breakoutGame.Score > 0 {
						_, _ = db.RecordScore("breakout", "Player 1", breakoutGame.Score, difficultyMod.Get(), fmt.Sprintf("Lives: %d", breakoutGame.Lives))
						if rebuildView != nil {
							rebuildView()
						}
					}
				}

			case "space":
				if spaceGame.Step() {
					needsRedraw = true
					if spaceGame.IsGameOver && spaceGame.Score > 0 {
						_, _ = db.RecordScore("space", "Player 1", spaceGame.Score, difficultyMod.Get(), fmt.Sprintf("Wave: %d", spaceGame.Wave))
						if rebuildView != nil {
							rebuildView()
						}
					}
				}
			}

			if needsRedraw {
				win.Invalidate()
			}
		}
	}()

	// 7. Build Main Application UI
	buildMainView := func() ui.Component {
		curTier := activeTier.Get()
		curGame := activeGame.Get()

		makeTierBtn := func(label, val string) *ui.ButtonComponent {
			if curTier == val {
				return widgets.Button(label).Primary()
			}
			btn := widgets.Button(label).Secondary()
			btn.OnClick(func() {
				activeTier.Set(val)
				if val == "Basic" {
					activeGame.Set("memory")
				} else if val == "Average" {
					activeGame.Set("snake")
				} else if val == "Advance" {
					activeGame.Set("breakout")
				}
				rebuildView()
			})
			return btn
		}

		makeGameBtn := func(label, val string) *ui.ButtonComponent {
			if curGame == val {
				return widgets.Button(label).Primary()
			}
			btn := widgets.Button(label).Ghost()
			btn.OnClick(func() {
				activeGame.Set(val)
				rebuildView()
			})
			return btn
		}

		// Top Header Bar
		headerBar := widgets.Card("",
			ui.Row(
				renderArcadeLogo(44),
				ui.Column(
					ui.Row(
						ui.Text("Nova Game Arcade").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
						widgets.Badge("Native Go GUI Engine").Info(),
						widgets.Badge("Full Keyboard & Mouse").Success(),
					).GapSpacing(8),
					ui.Text("Enjoy Casual, Puzzle, and High-Score Arcade Games with Offline Leaderboards.").Size(12).Col(color.Hex("#94A3B8")),
				).GapSpacing(2),

				ui.Spacer(),

				ui.Row(
					makeTierBtn("Basic Games", "Basic"),
					makeTierBtn("Average Games", "Average"),
					makeTierBtn("Advance Games", "Advance"),
					makeTierBtn("Leaderboard", "Leaderboard"),
					makeTierBtn("Changelog", "Changelog"),
				).GapSpacing(8),
			).GapSpacing(12),
		)

		// Sub-Header Game Switcher Row
		var subBar ui.Component
		switch curTier {
		case "Basic":
			subBar = ui.Row(
				ui.Text("Select Basic Game:").Weight(font.WeightBold).Col(color.Hex("#E2E8F0")),
				makeGameBtn("Memory Card Match", "memory"),
				makeGameBtn("Tic-Tac-Toe Neon", "tictactoe"),
				makeGameBtn("Reaction Speed Test", "reaction"),
				ui.Spacer(),
				widgets.Badge("Difficulty: Casual").Success(),
			).GapSpacing(8)

		case "Average":
			subBar = ui.Row(
				ui.Text("Select Average Game:").Weight(font.WeightBold).Col(color.Hex("#E2E8F0")),
				makeGameBtn("Retro Snake 2.0", "snake"),
				makeGameBtn("2048 Merge Puzzle", "2048"),
				makeGameBtn("Minesweeper Classic", "minesweeper"),
				ui.Spacer(),
				widgets.Badge("Difficulty: Intermediate").Info(),
			).GapSpacing(8)

		case "Advance":
			subBar = ui.Row(
				ui.Text("Select Advance Game:").Weight(font.WeightBold).Col(color.Hex("#E2E8F0")),
				makeGameBtn("Brick Breaker (Breakout)", "breakout"),
				makeGameBtn("Space Defender (Invaders)", "space"),
				ui.Spacer(),
				widgets.Badge("Difficulty: Hardcore Arcade").Error(),
			).GapSpacing(8)
		}

		// Main Content Area (Active Game, Leaderboard, or Changelog)
		var mainContent ui.Component

		if curTier == "Changelog" {
			// In-App Changelog View
			mainContent = ui.Container().
				Bg(color.Hex("#0F172A")).
				Border(color.Hex("#1E293B"), 1.5).
				Pad(geom.All(20)).
				Rounded(geom.RadiusUniform(12)).
				WithChild(
					ui.Column(
						ui.Row(
							ui.Text("Nova Game Arcade — Changelog & Updates").Size(20).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
							widgets.Badge("v1.1.0 Latest").Success(),
						).GapSpacing(10),

						widgets.Card("Version 1.1.0 — Full Keyboard & Gameplay Overhaul (Current)",
							ui.Column(
								ui.Text("• Full Keyboard Controls: Instant non-blocking keyboard input with Arrow Keys, WASD, Spacebar, and R across all games.").Size(13).Col(color.Hex("#E2E8F0")),
								ui.Text("• Brick Breaker Paddle Slider: Added interactive continuous slider directly under canvas to drag paddle to any position.").Size(13).Col(color.Hex("#E2E8F0")),
								ui.Text("• Retro Snake 2.0 Redesign: Dynamic snake head with directional eyes, apple stems, golden food, and live heading badge.").Size(13).Col(color.Hex("#E2E8F0")),
								ui.Text("• Responsive Layout: Optimized sidebar widths and grid dimensions to prevent canvas clipping on all display sizes.").Size(13).Col(color.Hex("#E2E8F0")),
								ui.Text("• In-App Changelog: New release history and feature logs embedded directly in the application.").Size(13).Col(color.Hex("#E2E8F0")),
							).GapSpacing(6),
						),

						widgets.Card("Version 1.0.0 — Initial Multi-Tier Release",
							ui.Column(
								ui.Text("• 8 Native Go Games: Memory Match, Tic-Tac-Toe Neon, Reaction Speed, Snake 2.0, 2048 Merge, Minesweeper, Breakout, Space Defender.").Size(13).Col(color.Hex("#94A3B8")),
								ui.Text("• Offline SQLite Leaderboard: Persistent high score tracking with WAL mode for all games.").Size(13).Col(color.Hex("#94A3B8")),
								ui.Text("• Multi-Tier Architecture: Organized by Basic, Average, and Advance tiers with instant tab switching.").Size(13).Col(color.Hex("#94A3B8")),
							).GapSpacing(6),
						),
					).GapSpacing(14),
				)

		} else if curTier == "Leaderboard" {
			// Leaderboard View
			makeLeaderboardCard := func(gameTitle, gID string) ui.Component {
				scores, _ := db.GetTopScores(gID, 5)
				var rows []ui.Component
				for i, s := range scores {
					rank := fmt.Sprintf("#%d", i+1)
					rows = append(rows, ui.Row(
						widgets.Badge(rank).Info(),
						ui.Text(s.PlayerName).Size(12).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
						ui.Spacer(),
						widgets.Badge(fmt.Sprintf("%d Pts", s.Score)).Success(),
						ui.Text(s.ExtraMeta).Size(11).Col(color.Hex("#94A3B8")),
					).GapSpacing(6))
				}
				if len(rows) == 0 {
					rows = append(rows, ui.Text("No scores logged yet.").Size(11).Col(color.Hex("#64748B")))
				}
				return ui.Container().
					WithWidth(520).
					WithChild(widgets.Card(gameTitle, ui.Column(rows...).GapSpacing(6)))
			}

			mainContent = ui.Column(
				ui.Row(
					makeLeaderboardCard("Retro Snake 2.0", "snake"),
					makeLeaderboardCard("2048 Merge Puzzle", "2048"),
				).GapSpacing(16),
				ui.Row(
					makeLeaderboardCard("Brick Breaker (Breakout)", "breakout"),
					makeLeaderboardCard("Space Defender (Invaders)", "space"),
				).GapSpacing(16),
			).GapSpacing(16)

		} else {
			// Game View
			switch curGame {
			case "tictactoe":
				mainContent = tttGame.Render(func(idx int) {
					tttGame.PlayMove(idx)
					rebuildView()
				}, func() {
					tttGame.Reset()
					rebuildView()
				})

			case "memory":
				mainContent = memGame.Render(func(idx int) {
					memGame.FlipCard(idx)
					if memGame.IsWon {
						_, _ = db.RecordScore("memory", "Player 1", 100, difficultyMod.Get(), fmt.Sprintf("%d Moves", memGame.Moves))
					}
					rebuildView()
				}, func() {
					memGame.Reset()
					rebuildView()
				})

			case "reaction":
				mainContent = reactGame.Render(func() {
					reactGame.HandleClick()
					if reactGame.State == StateGameOver {
						_, _ = db.RecordScore("reaction", "Player 1", reactGame.BestTime, difficultyMod.Get(), fmt.Sprintf("Avg: %dms", reactGame.AvgTime))
					}
					rebuildView()
				}, func() {
					reactGame.Reset()
					rebuildView()
				})

			case "snake":
				mainContent = snakeGame.Render(func(d Direction) {
					snakeGame.ChangeDir(d)
					rebuildView()
				}, func() {
					snakeGame.Reset()
					rebuildView()
				}, func() {
					snakeGame.IsPaused = !snakeGame.IsPaused
					rebuildView()
				})

			case "2048":
				mainContent = game2048.Render(func(dir string) {
					switch dir {
					case "left":
						game2048.MoveLeft()
					case "right":
						game2048.MoveRight()
					case "up":
						game2048.MoveUp()
					case "down":
						game2048.MoveDown()
					}
					if game2048.IsWon || game2048.IsGameOver {
						_, _ = db.RecordScore("2048", "Player 1", game2048.Score, difficultyMod.Get(), fmt.Sprintf("Best: %d", game2048.BestScore))
					}
					rebuildView()
				}, func() {
					game2048.Reset()
					rebuildView()
				})

			case "minesweeper":
				mainContent = mineGame.Render(func(r, c int) {
					mineGame.ClickCell(r, c)
					if mineGame.IsWon {
						_, _ = db.RecordScore("minesweeper", "Player 1", 100, difficultyMod.Get(), "Cleared")
					}
					rebuildView()
				}, func() {
					mineGame.FlagMode = !mineGame.FlagMode
					rebuildView()
				}, func() {
					mineGame.Reset()
					rebuildView()
				})

			case "breakout":
				mainContent = breakoutGame.Render(func() {
					breakoutGame.MovePaddle(-breakoutGame.PaddleSpeed)
					rebuildView()
				}, func() {
					breakoutGame.MovePaddle(breakoutGame.PaddleSpeed)
					rebuildView()
				}, func() {
					breakoutGame.Reset()
					rebuildView()
				}, func() {
					breakoutGame.IsPaused = !breakoutGame.IsPaused
					rebuildView()
				}, func(val float64) {
					breakoutGame.PaddleX = val
					rebuildView()
				})

			case "space":
				mainContent = spaceGame.Render(func() {
					spaceGame.MovePlayer(-20)
					rebuildView()
				}, func() {
					spaceGame.MovePlayer(20)
					rebuildView()
				}, func() {
					spaceGame.FireLaser()
					rebuildView()
				}, func() {
					spaceGame.Reset()
					rebuildView()
				}, func() {
					spaceGame.IsPaused = !spaceGame.IsPaused
					rebuildView()
				})
			}
		}

		// Toast / Status Banner
		toastBanner := ui.Container().
			Bg(color.Hex("#0B0F19")).
			Border(color.Hex("#1E293B"), 1.0).
			Pad(geom.Insets{Top: 6, Bottom: 6, Left: 14, Right: 14}).
			Rounded(geom.RadiusUniform(6)).
			WithChild(
				ui.Row(
					ui.Text(toastMsg.Get()).Size(12).Weight(font.WeightMedium).Col(color.Hex("#38BDF8")),
					ui.Spacer(),
					ui.Text("Nova Game Arcade v1.1.0 | Offline SQLite Persistent Scores").Size(11).Col(color.Hex("#64748B")),
				),
			)

		return ui.Padding(geom.Insets{Top: 10, Bottom: 10, Left: 16, Right: 16},
			ui.Column(
				headerBar,
				ifLen(subBar != nil, subBar, ui.Spacer()),
				mainContent,
				ui.Spacer(),
				toastBanner,
			).GapSpacing(10),
		)
	}

	// 8. Dynamic Rebuild Function
	rebuildView = func() {
		win.Content(buildMainView())
	}

	rebuildView()

	// 9. Generate Preview Screenshots for all Games & Views
	activeTier.Set("Average")
	activeGame.Set("snake")
	rebuildView()
	_ = win.SaveScreenshot("arcade_snake_preview.png")
	_ = win.SaveScreenshot("arcade_average_preview.png")

	activeTier.Set("Advance")
	activeGame.Set("breakout")
	rebuildView()
	_ = win.SaveScreenshot("arcade_breakout_preview.png")

	activeTier.Set("Changelog")
	rebuildView()
	_ = win.SaveScreenshot("arcade_changelog_preview.png")

	// Reset to Snake
	activeTier.Set("Average")
	activeGame.Set("snake")
	rebuildView()

	fmt.Println("🚀 Running Nova Game Arcade...")
	fmt.Println("🎮 Full Keyboard Controls & Brick Breaker Slider Active!")

	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func ifLen(cond bool, a, b ui.Component) ui.Component {
	if cond {
		return a
	}
	return b
}
