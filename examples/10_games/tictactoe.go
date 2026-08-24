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

type TicTacToeGame struct {
	Board      [9]string // "", "X", "O"
	CurrentTurn string    // "X" or "O"
	Winner     string    // "", "X", "O", "Draw"
	WinningLine []int     // e.g. [0, 1, 2]
	VsAI       bool
	AIDifficulty string // "Easy", "Medium", "Unbeatable"
	XWins      int
	OWins      int
	Draws      int
}

func NewTicTacToeGame() *TicTacToeGame {
	g := &TicTacToeGame{
		CurrentTurn:  "X",
		VsAI:         true,
		AIDifficulty: "Unbeatable",
	}
	g.Reset()
	return g
}

func (g *TicTacToeGame) Reset() {
	g.Board = [9]string{}
	g.CurrentTurn = "X"
	g.Winner = ""
	g.WinningLine = nil
}

func (g *TicTacToeGame) PlayMove(idx int) bool {
	if idx < 0 || idx >= 9 || g.Board[idx] != "" || g.Winner != "" {
		return false
	}

	g.Board[idx] = g.CurrentTurn
	g.checkGameState()

	if g.Winner == "" {
		if g.CurrentTurn == "X" {
			g.CurrentTurn = "O"
			if g.VsAI {
				g.playAIMove()
			}
		} else {
			g.CurrentTurn = "X"
		}
	}
	return true
}

func (g *TicTacToeGame) playAIMove() {
	if g.Winner != "" {
		return
	}

	var bestMove int
	switch g.AIDifficulty {
	case "Easy":
		// Random available move
		var avail []int
		for i, v := range g.Board {
			if v == "" {
				avail = append(avail, i)
			}
		}
		if len(avail) > 0 {
			bestMove = avail[rand.Intn(len(avail))]
		}
	case "Medium":
		// 50% minimax, 50% random
		if rand.Float64() < 0.5 {
			bestMove = g.getBestMinimaxMove("O")
		} else {
			var avail []int
			for i, v := range g.Board {
				if v == "" {
					avail = append(avail, i)
				}
			}
			if len(avail) > 0 {
				bestMove = avail[rand.Intn(len(avail))]
			}
		}
	default: // "Unbeatable" Minimax
		bestMove = g.getBestMinimaxMove("O")
	}

	g.Board[bestMove] = "O"
	g.checkGameState()
	if g.Winner == "" {
		g.CurrentTurn = "X"
	}
}

func (g *TicTacToeGame) getBestMinimaxMove(player string) int {
	bestScore := -1000
	bestMove := -1

	for i := 0; i < 9; i++ {
		if g.Board[i] == "" {
			g.Board[i] = player
			score := g.minimax(0, false)
			g.Board[i] = ""
			if score > bestScore {
				bestScore = score
				bestMove = i
			}
		}
	}
	if bestMove == -1 {
		for i := 0; i < 9; i++ {
			if g.Board[i] == "" {
				return i
			}
		}
	}
	return bestMove
}

func (g *TicTacToeGame) minimax(depth int, isMaximizing bool) int {
	winner, _ := g.evaluateWinner()
	if winner == "O" {
		return 10 - depth
	}
	if winner == "X" {
		return depth - 10
	}
	if g.isBoardFull() {
		return 0
	}

	if isMaximizing {
		bestScore := -1000
		for i := 0; i < 9; i++ {
			if g.Board[i] == "" {
				g.Board[i] = "O"
				score := g.minimax(depth+1, false)
				g.Board[i] = ""
				if score > bestScore {
					bestScore = score
				}
			}
		}
		return bestScore
	} else {
		bestScore := 1000
		for i := 0; i < 9; i++ {
			if g.Board[i] == "" {
				g.Board[i] = "X"
				score := g.minimax(depth+1, true)
				g.Board[i] = ""
				if score < bestScore {
					bestScore = score
				}
			}
		}
		return bestScore
	}
}

func (g *TicTacToeGame) isBoardFull() bool {
	for _, v := range g.Board {
		if v == "" {
			return false
		}
	}
	return true
}

func (g *TicTacToeGame) evaluateWinner() (string, []int) {
	lines := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // Rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // Cols
		{0, 4, 8}, {2, 4, 6},             // Diagonals
	}

	for _, l := range lines {
		if g.Board[l[0]] != "" && g.Board[l[0]] == g.Board[l[1]] && g.Board[l[1]] == g.Board[l[2]] {
			return g.Board[l[0]], l
		}
	}
	return "", nil
}

func (g *TicTacToeGame) checkGameState() {
	winner, line := g.evaluateWinner()
	if winner != "" {
		g.Winner = winner
		g.WinningLine = line
		if winner == "X" {
			g.XWins++
		} else if winner == "O" {
			g.OWins++
		}
		return
	}

	if g.isBoardFull() {
		g.Winner = "Draw"
		g.Draws++
	}
}

// RenderComponent builds the visual Tic-Tac-Toe UI
func (g *TicTacToeGame) Render(onMove func(int), onReset func()) ui.Component {
	var rows []ui.Component

	for r := 0; r < 3; r++ {
		var cellCols []ui.Component
		for c := 0; c < 3; c++ {
			idx := r*3 + c
			val := g.Board[idx]

			// Highlight winning line cells
			isWinCell := false
			if g.WinningLine != nil {
				for _, wIdx := range g.WinningLine {
					if wIdx == idx {
						isWinCell = true
						break
					}
				}
			}

			cellBg := color.Hex("#1E293B")
			cellBorder := color.Hex("#334155")
			if isWinCell {
				cellBg = color.Hex("#065F46")
				cellBorder = color.Hex("#10B981")
			}

			btnLabel := "[ . ]"
			if val == "X" {
				btnLabel = "[ X ]"
			} else if val == "O" {
				btnLabel = "[ O ]"
			}

			cellIdx := idx
			btn := widgets.Button(btnLabel).OnClick(func() {
				onMove(cellIdx)
			})

			cellComp := ui.Container().
				Size(120, 100).
				Bg(cellBg).
				Border(cellBorder, 2.0).
				Rounded(geom.RadiusUniform(8)).
				WithChild(
					ui.Center(btn),
				)

			cellCols = append(cellCols, cellComp)
		}
		rows = append(rows, ui.Row(cellCols...).GapSpacing(8))
	}

	statusText := "Current Turn: " + g.CurrentTurn
	statusBadge := widgets.Badge(statusText).Info()
	if g.Winner == "X" {
		statusBadge = widgets.Badge("Winner: Player X!").Success()
	} else if g.Winner == "O" {
		statusBadge = widgets.Badge("Winner: Player O / AI!").Error()
	} else if g.Winner == "Draw" {
		statusBadge = widgets.Badge("Game Ended in a Draw!").Warning()
	}

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("Tic-Tac-Toe Neon").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Basic Tier").Success(),
					statusBadge,
					ui.Spacer(),
					widgets.Button("[Restart Game]").Primary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					widgets.Badge(fmt.Sprintf("X Wins: %d", g.XWins)).Info(),
					widgets.Badge(fmt.Sprintf("O Wins: %d", g.OWins)).Error(),
					widgets.Badge(fmt.Sprintf("Draws: %d", g.Draws)).Warning(),
					ui.Spacer(),
					ui.Text("AI Mode: "+g.AIDifficulty).Size(12).Col(color.Hex("#94A3B8")),
				).GapSpacing(8),

				ui.Center(
					ui.Column(rows...).GapSpacing(8),
				),
			).GapSpacing(12),
		)
}
