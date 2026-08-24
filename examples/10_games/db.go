package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// GameScore represents an entry in the game leaderboard.
type GameScore struct {
	ID         int64
	GameID     string // "tictactoe", "memory", "reaction", "snake", "2048", "minesweeper", "breakout", "space"
	PlayerName string
	Score      int
	Difficulty string // "Basic", "Average", "Advance"
	ExtraMeta  string // e.g. "Moves: 12", "Time: 24s", "Waves: 5"
	CreatedAt  time.Time
}

// GameStats represents aggregated game statistics.
type GameStats struct {
	TotalPlays int
	HighScore  int
	BestTime   int // in seconds or ms
	TotalWins  int
}

// Database handles SQLite persistence for Nova Arcade.
type Database struct {
	mu sync.RWMutex
	db *sql.DB
}

// OpenDatabase initializes the arcade SQLite database.
func OpenDatabase(dbPath string) (*Database, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, pragma := range pragmas {
		_, _ = db.Exec(pragma)
	}

	d := &Database{db: db}
	if err := d.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return d, nil
}

func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *Database) initSchema() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	schema := `
	CREATE TABLE IF NOT EXISTS game_scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id TEXT NOT NULL,
		player_name TEXT DEFAULT 'Player 1',
		score INTEGER NOT NULL,
		difficulty TEXT DEFAULT 'Basic',
		extra_meta TEXT DEFAULT '',
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_game_score ON game_scores(game_id, score DESC);
	`
	_, err := d.db.Exec(schema)
	if err != nil {
		return err
	}

	// Seed starter high scores if empty
	var count int
	_ = d.db.QueryRow("SELECT COUNT(*) FROM game_scores").Scan(&count)
	if count == 0 {
		now := time.Now()
		starterScores := []GameScore{
			{GameID: "snake", PlayerName: "RetroGamer", Score: 480, Difficulty: "Average", ExtraMeta: "Length: 28", CreatedAt: now.Add(-2 * time.Hour)},
			{GameID: "2048", PlayerName: "MathMaster", Score: 2460, Difficulty: "Average", ExtraMeta: "Max Tile: 512", CreatedAt: now.Add(-1 * time.Hour)},
			{GameID: "breakout", PlayerName: "BrickSlayer", Score: 850, Difficulty: "Advance", ExtraMeta: "Bricks: 42", CreatedAt: now.Add(-30 * time.Minute)},
			{GameID: "space", PlayerName: "StarAce", Score: 1200, Difficulty: "Advance", ExtraMeta: "Wave 4", CreatedAt: now.Add(-15 * time.Minute)},
			{GameID: "memory", PlayerName: "MindPro", Score: 100, Difficulty: "Basic", ExtraMeta: "14 Moves", CreatedAt: now.Add(-10 * time.Minute)},
			{GameID: "reaction", PlayerName: "Speedy", Score: 95, Difficulty: "Basic", ExtraMeta: "Avg: 220ms", CreatedAt: now.Add(-5 * time.Minute)},
		}

		for _, s := range starterScores {
			_, _ = d.db.Exec(`INSERT INTO game_scores (game_id, player_name, score, difficulty, extra_meta, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
				s.GameID, s.PlayerName, s.Score, s.Difficulty, s.ExtraMeta, s.CreatedAt)
		}
	}

	return nil
}

// RecordScore inserts a new score entry.
func (d *Database) RecordScore(gameID, playerName string, score int, difficulty, extraMeta string) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if playerName == "" {
		playerName = "Player 1"
	}
	if difficulty == "" {
		difficulty = "Basic"
	}

	res, err := d.db.Exec(`INSERT INTO game_scores (game_id, player_name, score, difficulty, extra_meta, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		gameID, playerName, score, difficulty, extraMeta, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTopScores retrieves top N scores for a game.
func (d *Database) GetTopScores(gameID string, limit int) ([]GameScore, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, game_id, player_name, score, difficulty, extra_meta, created_at FROM game_scores WHERE game_id = ? ORDER BY score DESC, created_at ASC LIMIT ?`
	rows, err := d.db.Query(query, gameID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []GameScore
	for rows.Next() {
		var s GameScore
		if err := rows.Scan(&s.ID, &s.GameID, &s.PlayerName, &s.Score, &s.Difficulty, &s.ExtraMeta, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}

// GetHighScore returns the all-time highest score for a game.
func (d *Database) GetHighScore(gameID string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var hs sql.NullInt64
	_ = d.db.QueryRow("SELECT MAX(score) FROM game_scores WHERE game_id = ?", gameID).Scan(&hs)
	if hs.Valid {
		return int(hs.Int64)
	}
	return 0
}
