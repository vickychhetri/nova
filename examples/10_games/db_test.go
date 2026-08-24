package main

import (
	"path/filepath"
	"testing"
)

func TestGameDatabase_Operations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_games.db")

	db, err := OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("OpenDatabase failed: %v", err)
	}
	defer db.Close()

	// 1. Initial seeded scores check
	scores, err := db.GetTopScores("snake", 5)
	if err != nil || len(scores) == 0 {
		t.Fatalf("expected seeded scores for snake, got len=%d, err=%v", len(scores), err)
	}

	// 2. Insert new high score
	id, err := db.RecordScore("snake", "ProPlayer", 9999, "Advance", "Length: 55")
	if err != nil || id <= 0 {
		t.Fatalf("RecordScore failed: %v, id=%d", err, id)
	}

	// 3. Verify top score
	hs := db.GetHighScore("snake")
	if hs != 9999 {
		t.Errorf("expected high score 9999, got %d", hs)
	}

	// 4. Verify leaderboard ranking
	topScores, err := db.GetTopScores("snake", 1)
	if err != nil || len(topScores) == 0 || topScores[0].Score != 9999 {
		t.Errorf("expected top score 9999 in leaderboard query")
	}
}
