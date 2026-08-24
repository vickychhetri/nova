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

type MineCell struct {
	IsMine        bool
	AdjacentMines int
	IsRevealed    bool
	IsFlagged     bool
}

type MinesweeperGame struct {
	Rows       int // 8
	Cols       int // 8
	TotalMines int // 8
	Grid       [][]MineCell
	IsGameOver bool
	IsWon      bool
	FlagMode   bool
	FirstClick bool
	FlagsPlaced int
}

func NewMinesweeperGame() *MinesweeperGame {
	g := &MinesweeperGame{
		Rows:       8,
		Cols:       8,
		TotalMines: 8,
	}
	g.Reset()
	return g
}

func (g *MinesweeperGame) Reset() {
	g.Grid = make([][]MineCell, g.Rows)
	for r := 0; r < g.Rows; r++ {
		g.Grid[r] = make([]MineCell, g.Cols)
	}
	g.IsGameOver = false
	g.IsWon = false
	g.FirstClick = true
	g.FlagsPlaced = 0
	g.FlagMode = false
}

func (g *MinesweeperGame) placeMines(safeR, safeC int) {
	placed := 0
	for placed < g.TotalMines {
		r := rand.Intn(g.Rows)
		c := rand.Intn(g.Cols)
		if (r == safeR && c == safeC) || g.Grid[r][c].IsMine {
			continue
		}
		g.Grid[r][c].IsMine = true
		placed++
	}

	// Compute adjacent mines
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if g.Grid[r][c].IsMine {
				continue
			}
			count := 0
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					nr, nc := r+dr, c+dc
					if nr >= 0 && nr < g.Rows && nc >= 0 && nc < g.Cols && g.Grid[nr][nc].IsMine {
						count++
					}
				}
			}
			g.Grid[r][c].AdjacentMines = count
		}
	}
}

func (g *MinesweeperGame) ClickCell(r, c int) {
	if g.IsGameOver || g.IsWon || r < 0 || r >= g.Rows || c < 0 || c >= g.Cols {
		return
	}

	cell := &g.Grid[r][c]

	if g.FlagMode {
		if !cell.IsRevealed {
			cell.IsFlagged = !cell.IsFlagged
			if cell.IsFlagged {
				g.FlagsPlaced++
			} else {
				g.FlagsPlaced--
			}
		}
		return
	}

	if cell.IsFlagged || cell.IsRevealed {
		return
	}

	if g.FirstClick {
		g.FirstClick = false
		g.placeMines(r, c)
	}

	if cell.IsMine {
		// Boom! Game Over
		g.IsGameOver = true
		// Reveal all mines
		for row := 0; row < g.Rows; row++ {
			for col := 0; col < g.Cols; col++ {
				if g.Grid[row][col].IsMine {
					g.Grid[row][col].IsRevealed = true
				}
			}
		}
		return
	}

	g.reveal(r, c)
	g.checkWin()
}

func (g *MinesweeperGame) reveal(r, c int) {
	if r < 0 || r >= g.Rows || c < 0 || c >= g.Cols {
		return
	}
	cell := &g.Grid[r][c]
	if cell.IsRevealed || cell.IsFlagged || cell.IsMine {
		return
	}

	cell.IsRevealed = true

	if cell.AdjacentMines == 0 {
		// Flood fill neighbors
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				if dr != 0 || dc != 0 {
					g.reveal(r+dr, c+dc)
				}
			}
		}
	}
}

func (g *MinesweeperGame) checkWin() {
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if !g.Grid[r][c].IsMine && !g.Grid[r][c].IsRevealed {
				return
			}
		}
	}
	g.IsWon = true
}

func (g *MinesweeperGame) Render(onClick func(int, int), onToggleFlag func(), onReset func()) ui.Component {
	var rows []ui.Component

	for r := 0; r < g.Rows; r++ {
		var cols []ui.Component
		for c := 0; c < g.Cols; c++ {
			cell := g.Grid[r][c]
			txt := "[ . ]"
			cellBg := color.Hex("#334155")
			cellBorder := color.Hex("#475569")

			if cell.IsRevealed {
				if cell.IsMine {
					txt = "[*]"
					cellBg = color.Hex("#7F1D1D")
					cellBorder = color.Hex("#DC2626")
				} else {
					cellBg = color.Hex("#1E293B")
					cellBorder = color.Hex("#0F172A")
					if cell.AdjacentMines > 0 {
						txt = fmt.Sprintf(" %d ", cell.AdjacentMines)
					} else {
						txt = " - "
					}
				}
			} else if cell.IsFlagged {
				txt = "[P]"
				cellBg = color.Hex("#451A03")
				cellBorder = color.Hex("#D97706")
			}

			rowIdx, colIdx := r, c
			btn := widgets.Button(txt).OnClick(func() {
				onClick(rowIdx, colIdx)
			})

			cellBox := ui.Container().
				Size(52, 44).
				Bg(cellBg).
				Border(cellBorder, 1.0).
				Rounded(geom.RadiusUniform(4)).
				WithChild(
					ui.Center(btn),
				)

			cols = append(cols, cellBox)
		}
		rows = append(rows, ui.Row(cols...).GapSpacing(4))
	}

	statusBadge := widgets.Badge(fmt.Sprintf("Mines Left: %d", g.TotalMines-g.FlagsPlaced)).Info()
	if g.IsGameOver {
		statusBadge = widgets.Badge("MINE EXPLODED - GAME OVER!").Error()
	} else if g.IsWon {
		statusBadge = widgets.Badge("ALL MINES CLEARED - VICTORY!").Success()
	}

	flagBtn := widgets.Button("[Mode: Reveal (Click)]").Secondary()
	if g.FlagMode {
		flagBtn = widgets.Button("[Mode: Flag Mines [P]]").Primary()
	}
	flagBtn.OnClick(onToggleFlag)

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("Minesweeper Classic").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Average Tier").Info(),
					statusBadge,
					ui.Spacer(),
					flagBtn,
					widgets.Button("[Restart Game]").Primary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					ui.Container().
						Bg(color.Hex("#0F172A")).
						Pad(geom.All(8)).
						Rounded(geom.RadiusUniform(6)).
						WithChild(
							ui.Column(rows...).GapSpacing(4),
						),
					ui.Container().
						WithWidth(320).
						WithChild(
							ui.Column(
								widgets.Card("Game Rules",
									ui.Text("Instructions:\n• Click cells to reveal\n• Toggle Flag Mode to mark mines\n• Number indicates adjacent mines\n• Clear all 56 safe cells to win!").Size(12).Col(color.Hex("#94A3B8")),
								),
							),
						),
				).GapSpacing(20),
			).GapSpacing(12),
		)
}
