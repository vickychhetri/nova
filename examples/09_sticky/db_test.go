package main

import (
	"path/filepath"
	"testing"
)

func TestStickyDatabase_Operations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_sticky.db")

	db, err := OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("OpenDatabase failed: %v", err)
	}
	defer db.Close()

	// 1. Initial count from seeds
	items, err := db.GetClips("", "All", false)
	if err != nil || len(items) == 0 {
		t.Fatalf("expected seeded clips, got len=%d, err=%v", len(items), err)
	}

	// 2. Add new Sticky Clip
	id, err := db.AddClip("git commit -m 'feat: Add sticky clipboard manager'", "Code", "Blue", true)
	if err != nil || id <= 0 {
		t.Fatalf("AddClip failed: %v, id=%d", err, id)
	}

	// 3. Verify Pinned query
	pinnedItems, err := db.GetClips("", "All", true)
	if err != nil || len(pinnedItems) == 0 {
		t.Fatalf("GetClips pinned failed: %v", err)
	}

	// 4. Test Deduplication
	dupID, err := db.AddClip("git commit -m 'feat: Add sticky clipboard manager'", "Code", "Blue", true)
	if err != nil || dupID != id {
		t.Errorf("expected duplicate to return existing id %d, got %d, err=%v", id, dupID, err)
	}

	// 5. Test Copy Count
	err = db.RecordCopy(id)
	if err != nil {
		t.Errorf("RecordCopy failed: %v", err)
	}

	// 6. Test Color Update
	err = db.UpdateColor(id, "Pink")
	if err != nil {
		t.Errorf("UpdateColor failed: %v", err)
	}

	// 7. Test Search
	searchResults, err := db.GetClips("commit", "All", false)
	if err != nil || len(searchResults) == 0 {
		t.Errorf("expected search match for 'commit', got %d results", len(searchResults))
	}

	// 8. Test Toggle Pin
	err = db.TogglePin(id)
	if err != nil {
		t.Errorf("TogglePin failed: %v", err)
	}

	// 9. Test Delete
	err = db.DeleteClip(id)
	if err != nil {
		t.Errorf("DeleteClip failed: %v", err)
	}
}
