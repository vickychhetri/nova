package main

import (
	"fmt"
	"math/rand"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

type Game2048 struct {
	Board      [4][4]int
	Score      int
	BestScore  int
	IsWon      bool
	IsGameOver bool
}

func NewGame2048() *Game2048 {
	g := &Game2048{}
	g.Reset()
	return g
}

func (g *Game2048) Reset() {
	g.Board = [4][4]int{}
	g.Score = 0
	g.IsWon = false
	g.IsGameOver = false
	g.spawnTile()
	g.spawnTile()
}

func (g *Game2048) spawnTile() {
	type pos struct{ r, c int }
	var empty []pos

	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if g.Board[r][c] == 0 {
				empty = append(empty, pos{r, c})
			}
		}
	}

	if len(empty) == 0 {
		return
	}

	chosen := empty[rand.Intn(len(empty))]
	val := 2
	if rand.Float64() < 0.1 {
		val = 4
	}
	g.Board[chosen.r][chosen.c] = val
}

func (g *Game2048) MoveLeft() bool {
	changed := false
	for r := 0; r < 4; r++ {
		row := g.Board[r]
		newRow, scoreGain, moved := slideAndMerge(row[:])
		if moved {
			changed = true
			copy(g.Board[r][:], newRow)
			g.Score += scoreGain
		}
	}
	if changed {
		g.afterMove()
	}
	return changed
}

func (g *Game2048) MoveRight() bool {
	changed := false
	for r := 0; r < 4; r++ {
		row := g.Board[r]
		// Reverse
		rev := []int{row[3], row[2], row[1], row[0]}
		newRow, scoreGain, moved := slideAndMerge(rev)
		if moved {
			changed = true
			g.Board[r] = [4]int{newRow[3], newRow[2], newRow[1], newRow[0]}
			g.Score += scoreGain
		}
	}
	if changed {
		g.afterMove()
	}
	return changed
}

func (g *Game2048) MoveUp() bool {
	changed := false
	for c := 0; c < 4; c++ {
		col := []int{g.Board[0][c], g.Board[1][c], g.Board[2][c], g.Board[3][c]}
		newCol, scoreGain, moved := slideAndMerge(col)
		if moved {
			changed = true
			for r := 0; r < 4; r++ {
				g.Board[r][c] = newCol[r]
			}
			g.Score += scoreGain
		}
	}
	if changed {
		g.afterMove()
	}
	return changed
}

func (g *Game2048) MoveDown() bool {
	changed := false
	for c := 0; c < 4; c++ {
		col := []int{g.Board[3][c], g.Board[2][c], g.Board[1][c], g.Board[0][c]}
		newCol, scoreGain, moved := slideAndMerge(col)
		if moved {
			changed = true
			for r := 0; r < 4; r++ {
				g.Board[3-r][c] = newCol[r]
			}
			g.Score += scoreGain
		}
	}
	if changed {
		g.afterMove()
	}
	return changed
}

func slideAndMerge(line []int) ([]int, int, bool) {
	// Filter non-zeros
	var filtered []int
	for _, v := range line {
		if v != 0 {
			filtered = append(filtered, v)
		}
	}

	scoreGain := 0
	var merged []int
	skip := false

	for i := 0; i < len(filtered); i++ {
		if skip {
			skip = false
			continue
		}
		if i+1 < len(filtered) && filtered[i] == filtered[i+1] {
			val := filtered[i] * 2
			merged = append(merged, val)
			scoreGain += val
			skip = true
		} else {
			merged = append(merged, filtered[i])
		}
	}

	// Pad with zeros to 4
	for len(merged) < 4 {
		merged = append(merged, 0)
	}

	moved := false
	for i := 0; i < 4; i++ {
		if line[i] != merged[i] {
			moved = true
			break
		}
	}

	return merged, scoreGain, moved
}

func (g *Game2048) afterMove() {
	g.spawnTile()
	if g.Score > g.BestScore {
		g.BestScore = g.Score
	}

	// Check win (2048)
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if g.Board[r][c] >= 2048 {
				g.IsWon = true
			}
		}
	}

	// Check game over
	g.IsGameOver = g.checkGameOver()
}

func (g *Game2048) checkGameOver() bool {
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if g.Board[r][c] == 0 {
				return false
			}
			if c+1 < 4 && g.Board[r][c] == g.Board[r][c+1] {
				return false
			}
			if r+1 < 4 && g.Board[r][c] == g.Board[r+1][c] {
				return false
			}
		}
	}
	return true
}

func getTileColor(val int) (color.Color, color.Color) {
	switch val {
	case 2:
		return color.Hex("#EEE4DA"), color.Hex("#776E65")
	case 4:
		return color.Hex("#EDE0C8"), color.Hex("#776E65")
	case 8:
		return color.Hex("#F2B179"), color.Hex("#F9F6F2")
	case 16:
		return color.Hex("#F59563"), color.Hex("#F9F6F2")
	case 32:
		return color.Hex("#F67C5F"), color.Hex("#F9F6F2")
	case 64:
		return color.Hex("#F65E3B"), color.Hex("#F9F6F2")
	case 128:
		return color.Hex("#EDCF72"), color.Hex("#F9F6F2")
	case 256:
		return color.Hex("#EDCC61"), color.Hex("#F9F6F2")
	case 512:
		return color.Hex("#EDC850"), color.Hex("#F9F6F2")
	case 1024:
		return color.Hex("#EDC53F"), color.Hex("#F9F6F2")
	case 2048:
		return color.Hex("#ECC400"), color.Hex("#FFFFFF")
	default:
		if val > 2048 {
			return color.Hex("#3C3A32"), color.Hex("#FFFFFF")
		}
		return color.Hex("#CDC1B4"), color.Hex("#776E65")
	}
}

func (g *Game2048) Render(onMove func(string), onReset func()) ui.Component {
	var rows []ui.Component

	for r := 0; r < 4; r++ {
		var cols []ui.Component
		for c := 0; c < 4; c++ {
			val := g.Board[r][c]
			bgCol, txtCol := getTileColor(val)

			txtStr := ""
			if val > 0 {
				txtStr = fmt.Sprintf("%d", val)
			}

			tileBox := ui.Container().
				Size(80, 80).
				Bg(bgCol).
				Border(color.Hex("#BBADA0"), 2.0).
				Rounded(geom.RadiusUniform(6)).
				WithChild(
					ui.Center(
						ui.Text(txtStr).Size(22).Weight(font.WeightBold).Col(txtCol),
					),
				)

			cols = append(cols, tileBox)
		}
		rows = append(rows, ui.Row(cols...).GapSpacing(8))
	}

	statusBadge := widgets.Badge(fmt.Sprintf("Score: %d", g.Score)).Success()
	if g.IsGameOver {
		statusBadge = widgets.Badge("NO MOVES LEFT - GAME OVER!").Error()
	} else if g.IsWon {
		statusBadge = widgets.Badge("2048 ACHIEVED - VICTORY!").Success()
	}

	dPad := ui.Column(
		ui.Center(widgets.Button("[▲ Up]").Secondary().OnClick(func() { onMove("up") })),
		ui.Row(
			widgets.Button("[◀ Left]").Secondary().OnClick(func() { onMove("left") }),
			widgets.Button("[▼ Down]").Secondary().OnClick(func() { onMove("down") }),
			widgets.Button("[▶ Right]").Secondary().OnClick(func() { onMove("right") }),
		).GapSpacing(6),
	).GapSpacing(6)

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("2048 Merge Puzzle").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Average Tier").Info(),
					statusBadge,
					widgets.Badge(fmt.Sprintf("Best: %d", g.BestScore)).Info(),
					ui.Spacer(),
					widgets.Button("[Restart Game]").Primary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					ui.Container().
						Bg(color.Hex("#BBADA0")).
						Pad(geom.All(12)).
						Rounded(geom.RadiusUniform(8)).
						WithChild(
							ui.Column(rows...).GapSpacing(8),
						),
					ui.Container().
						WithWidth(320).
						WithChild(
							ui.Column(
								widgets.Card("Arrow Controls", dPad),
								ui.Text("Instructions:\n• Arrow Keys / WASD\n• Merge matching tiles\n• Reach 2048 to win!").Size(11).Col(color.Hex("#94A3B8")),
							).GapSpacing(10),
						),
				).GapSpacing(20),
			).GapSpacing(12),
		)
}
