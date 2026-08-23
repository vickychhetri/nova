package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// StickyClip represents a sticky note or captured clipboard history entry.
type StickyClip struct {
	ID        int64
	Content   string
	Category  string // "Note", "Code", "Link", "Clipboard", "Draft"
	ColorTag  string // "Yellow", "Blue", "Green", "Pink", "Purple", "Neutral"
	IsPinned  bool
	CopyCount int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Database provides persistent SQLite storage for Sticky Notes & Clipboard History.
type Database struct {
	mu sync.RWMutex
	db *sql.DB
}

// OpenDatabase initializes or opens the SQLite database.
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
	CREATE TABLE IF NOT EXISTS sticky_clips (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		category TEXT DEFAULT 'Clipboard',
		color_tag TEXT DEFAULT 'Yellow',
		is_pinned INTEGER DEFAULT 0,
		copy_count INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_sticky_pinned ON sticky_clips(is_pinned, updated_at DESC);
	`
	_, err := d.db.Exec(schema)
	if err != nil {
		return err
	}

	// Seed starter sample notes if empty
	var count int
	_ = d.db.QueryRow("SELECT COUNT(*) FROM sticky_clips").Scan(&count)
	if count == 0 {
		now := time.Now()
		starterNotes := []StickyClip{
			{
				Content:   "Welcome to Nova Sticky!\nCopy anything in your system (Ctrl+C), and it will instantly show up here.\nClick any sticky note to copy it back to your clipboard ready to paste anywhere!",
				Category:  "Note",
				ColorTag:  "Yellow",
				IsPinned:  true,
				CreatedAt: now.Add(-60 * time.Minute),
				UpdatedAt: now.Add(-60 * time.Minute),
			},
			{
				Content:   "func CalculateMetrics(data []float64) (mean, stdDev float64) {\n    // Core numerical statistics helper\n    return 42.0, 1.414\n}",
				Category:  "Code",
				ColorTag:  "Blue",
				IsPinned:  true,
				CreatedAt: now.Add(-50 * time.Minute),
				UpdatedAt: now.Add(-50 * time.Minute),
			},
			{
				Content:   "https://github.com/vickychhetri/nova\nFast, lightweight, native GUI toolkit for Go.",
				Category:  "Link",
				ColorTag:  "Green",
				IsPinned:  false,
				CreatedAt: now.Add(-40 * time.Minute),
				UpdatedAt: now.Add(-40 * time.Minute),
			},
			{
				Content:   "Project Checklist:\n- [x] Integrate live clipboard history watcher\n- [x] Fast one-click copy & paste\n- [x] Color-coded pastel sticky themes\n- [x] Scrollable multi-page clipboard list\n- [ ] Release v1.0",
				Category:  "Note",
				ColorTag:  "Pink",
				IsPinned:  false,
				CreatedAt: now.Add(-30 * time.Minute),
				UpdatedAt: now.Add(-30 * time.Minute),
			},
			{
				Content:   "SSH Command Quick Access:\nssh -i ~/.ssh/id_ed25519 user@192.168.1.100 -p 2222",
				Category:  "Code",
				ColorTag:  "Purple",
				IsPinned:  false,
				CreatedAt: now.Add(-25 * time.Minute),
				UpdatedAt: now.Add(-25 * time.Minute),
			},
			{
				Content:   "Meeting Notes (Q3 Sync):\n1. Review UI layout and scrollable cards\n2. Optimize clipboard polling interval\n3. Verify zero external dependencies",
				Category:  "Note",
				ColorTag:  "Yellow",
				IsPinned:  false,
				CreatedAt: now.Add(-20 * time.Minute),
				UpdatedAt: now.Add(-20 * time.Minute),
			},
			{
				Content:   "docker run -d --name postgres-db -e POSTGRES_PASSWORD=secret -p 5432:5432 postgres:16-alpine",
				Category:  "Code",
				ColorTag:  "Blue",
				IsPinned:  false,
				CreatedAt: now.Add(-15 * time.Minute),
				UpdatedAt: now.Add(-15 * time.Minute),
			},
			{
				Content:   "https://pkg.go.dev/modernc.org/sqlite\nPure Go SQLite driver without CGO requirements.",
				Category:  "Link",
				ColorTag:  "Green",
				IsPinned:  false,
				CreatedAt: now.Add(-10 * time.Minute),
				UpdatedAt: now.Add(-10 * time.Minute),
			},
			{
				Content:   "Color Palette Hex Codes:\nYellow: #FEFCE8\nBlue: #F0F9FF\nGreen: #F0FDF4\nPink: #FFF1F2\nPurple: #FAF5FF",
				Category:  "Note",
				ColorTag:  "Pink",
				IsPinned:  false,
				CreatedAt: now.Add(-5 * time.Minute),
				UpdatedAt: now.Add(-5 * time.Minute),
			},
		}

		for _, it := range starterNotes {
			pinned := 0
			if it.IsPinned {
				pinned = 1
			}
			_, _ = d.db.Exec(`INSERT INTO sticky_clips (content, category, color_tag, is_pinned, copy_count, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, ?)`,
				it.Content, it.Category, it.ColorTag, pinned, it.CreatedAt, it.UpdatedAt)
		}
	}

	return nil
}

// AddClip inserts a new clip or sticky note. Returns id.
func (d *Database) AddClip(content, category, colorTag string, isPinned bool) (int64, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return 0, fmt.Errorf("content cannot be empty")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if the most recent clip is identical to avoid rapid duplicate spam
	var lastID int64
	var lastContent string
	_ = d.db.QueryRow("SELECT id, content FROM sticky_clips ORDER BY id DESC LIMIT 1").Scan(&lastID, &lastContent)
	if strings.TrimSpace(lastContent) == trimmed {
		// Update timestamp so it moves to top
		_, _ = d.db.Exec("UPDATE sticky_clips SET updated_at = ? WHERE id = ?", time.Now(), lastID)
		return lastID, nil
	}

	if colorTag == "" {
		colorTag = "Yellow"
	}
	if category == "" {
		category = detectCategory(trimmed)
	}

	pinnedInt := 0
	if isPinned {
		pinnedInt = 1
	}

	now := time.Now()
	res, err := d.db.Exec(`INSERT INTO sticky_clips (content, category, color_tag, is_pinned, copy_count, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, ?)`,
		trimmed, category, colorTag, pinnedInt, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// RecordCopy increments copy counter and updates timestamp
func (d *Database) RecordCopy(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("UPDATE sticky_clips SET copy_count = copy_count + 1, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}

// TogglePin toggles pinned status of note
func (d *Database) TogglePin(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("UPDATE sticky_clips SET is_pinned = CASE WHEN is_pinned = 1 THEN 0 ELSE 1 END, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}

// UpdateColor changes color tag
func (d *Database) UpdateColor(id int64, newColor string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("UPDATE sticky_clips SET color_tag = ?, updated_at = ? WHERE id = ?", newColor, time.Now(), id)
	return err
}

// DeleteClip removes a clip by ID
func (d *Database) DeleteClip(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("DELETE FROM sticky_clips WHERE id = ?", id)
	return err
}

// ClearHistory removes non-pinned clips
func (d *Database) ClearHistory(onlyUnpinned bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if onlyUnpinned {
		_, err := d.db.Exec("DELETE FROM sticky_clips WHERE is_pinned = 0")
		return err
	}
	_, err := d.db.Exec("DELETE FROM sticky_clips")
	return err
}

// GetClips fetches clips matching filters
func (d *Database) GetClips(searchQuery, colorFilter string, onlyPinned bool) ([]StickyClip, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	query := `SELECT id, content, category, color_tag, is_pinned, copy_count, created_at, updated_at FROM sticky_clips ORDER BY is_pinned DESC, updated_at DESC`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []StickyClip
	sLower := strings.ToLower(strings.TrimSpace(searchQuery))

	for rows.Next() {
		var it StickyClip
		var pinned int
		if err := rows.Scan(&it.ID, &it.Content, &it.Category, &it.ColorTag, &pinned, &it.CopyCount, &it.CreatedAt, &it.UpdatedAt); err != nil {
			continue
		}
		it.IsPinned = (pinned == 1)

		if onlyPinned && !it.IsPinned {
			continue
		}

		if colorFilter != "" && colorFilter != "All" {
			if !strings.EqualFold(it.ColorTag, colorFilter) {
				continue
			}
		}

		if sLower != "" {
			if !strings.Contains(strings.ToLower(it.Content), sLower) && !strings.Contains(strings.ToLower(it.Category), sLower) {
				continue
			}
		}

		items = append(items, it)
	}

	return items, nil
}

func detectCategory(str string) string {
	s := strings.TrimSpace(str)
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return "Link"
	}
	if strings.Contains(s, "func ") || strings.Contains(s, "package ") || strings.Contains(s, "import ") || strings.Contains(s, "const ") || strings.Contains(s, "{") && strings.Contains(s, "}") {
		return "Code"
	}
	if strings.Contains(s, "\n") {
		return "Note"
	}
	return "Clipboard"
}
