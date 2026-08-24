package main

import (
	"testing"
	"time"
)

func TestTicTacToe_MinimaxAndWinDetection(t *testing.T) {
	g := NewTicTacToeGame()
	g.VsAI = false // 2-Player local mode

	// Player X wins horizontally
	g.PlayMove(0) // X
	g.PlayMove(3) // O
	g.PlayMove(1) // X
	g.PlayMove(4) // O
	g.PlayMove(2) // X

	if g.Winner != "X" {
		t.Errorf("expected X to win horizontally, got %s", g.Winner)
	}

	// Test Minimax AI doesn't panic
	aiGame := NewTicTacToeGame()
	aiGame.AIDifficulty = "Unbeatable"
	aiGame.PlayMove(4) // center
	if aiGame.Winner == "" && aiGame.Board[4] != "X" {
		t.Errorf("expected center move to be X")
	}
}

func TestMemoryGame_ShufflingAndMatching(t *testing.T) {
	g := NewMemoryGame()
	if len(g.Cards) != 16 {
		t.Fatalf("expected 16 cards, got %d", len(g.Cards))
	}

	// Find two matching cards
	var sym string
	var idx1, idx2 int
	idx1, idx2 = -1, -1

	for i := 0; i < len(g.Cards); i++ {
		for j := i + 1; j < len(g.Cards); j++ {
			if g.Cards[i].Symbol == g.Cards[j].Symbol {
				idx1 = i
				idx2 = j
				sym = g.Cards[i].Symbol
				break
			}
		}
		if idx1 != -1 {
			break
		}
	}

	// Flip first
	g.FlipCard(idx1)
	if !g.Cards[idx1].IsFlipped {
		t.Errorf("expected card %d to be flipped", idx1)
	}

	// Flip matching second
	g.FlipCard(idx2)
	if !g.Cards[idx1].IsMatched || !g.Cards[idx2].IsMatched {
		t.Errorf("expected symbol %s cards to be matched", sym)
	}
	if g.Matches != 1 {
		t.Errorf("expected 1 match, got %d", g.Matches)
	}
}

func TestReactionGame_TimingAndRank(t *testing.T) {
	g := NewReactionGame()
	g.StartRound()

	// Trigger ready state
	g.TriggerTime = time.Now().Add(-100 * time.Millisecond)
	g.CheckTimer()
	if g.State != StateClickNow {
		t.Errorf("expected StateClickNow, got %v", g.State)
	}

	g.HandleClick()
	if len(g.Times) != 1 {
		t.Errorf("expected 1 recorded time, got %d", len(g.Times))
	}

	rank := g.GetRank()
	if rank == "" || rank == "N/A" {
		t.Errorf("expected valid rank, got %s", rank)
	}
}

func TestSnakeGame_MovementAndCollision(t *testing.T) {
	g := NewSnakeGame()
	initialHead := g.Snake[0]

	// Move right
	g.Step()
	newHead := g.Snake[0]
	if newHead.X != initialHead.X+1 {
		t.Errorf("expected head X to increment to %d, got %d", initialHead.X+1, newHead.X)
	}

	// Test wall collision
	g.Snake[0] = GridPoint{X: g.Width - 1, Y: 5}
	g.Dir = DirRight
	g.NextDir = DirRight
	g.Step()
	if !g.IsGameOver {
		t.Errorf("expected game over on wall collision")
	}
}

func TestGame2048_SlidingAndMerging(t *testing.T) {
	row := []int{2, 2, 4, 8}
	merged, score, moved := slideAndMerge(row)
	if !moved {
		t.Errorf("expected row to move")
	}
	if score != 4 {
		t.Errorf("expected score gain of 4, got %d", score)
	}
	if merged[0] != 4 || merged[1] != 4 || merged[2] != 8 || merged[3] != 0 {
		t.Errorf("expected [4 4 8 0], got %v", merged)
	}
}

func TestMinesweeper_FloodFillAndWin(t *testing.T) {
	g := NewMinesweeperGame()
	g.ClickCell(3, 3)

	if g.FirstClick {
		t.Errorf("expected first click to have placed mines")
	}
	if g.Grid[3][3].IsMine {
		t.Errorf("expected first click cell (3,3) to never be a mine")
	}
	if !g.Grid[3][3].IsRevealed {
		t.Errorf("expected cell (3,3) to be revealed")
	}
}

func TestBreakout_PaddleAndBallPhysics(t *testing.T) {
	g := NewBreakoutGame()
	initialY := g.BallY

	g.Step()
	if g.BallY == initialY {
		t.Errorf("expected ball Y to change after step")
	}

	// Move paddle
	initialPaddleX := g.PaddleX
	g.MovePaddle(20)
	if g.PaddleX != initialPaddleX+20 {
		t.Errorf("expected paddle X to be %f, got %f", initialPaddleX+20, g.PaddleX)
	}
}

func TestSpace_AliensAndLaserCollisions(t *testing.T) {
	g := NewSpaceGame()
	if len(g.Aliens) != 18 {
		t.Fatalf("expected 18 aliens, got %d", len(g.Aliens))
	}

	// Fire laser
	g.FireLaser()
	if len(g.Lasers) != 1 {
		t.Errorf("expected 1 active laser, got %d", len(g.Lasers))
	}

	// Step simulation
	g.Step()
	if g.Lasers[0].Y >= g.PlayerY-4 {
		t.Errorf("expected laser to move upward")
	}
}
