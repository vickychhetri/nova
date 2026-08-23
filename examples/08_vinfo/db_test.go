package main

import (
	"path/filepath"
	"testing"
)

func TestDatabase_VaultFull(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_vault.db")

	db, err := OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("OpenDatabase failed: %v", err)
	}
	defer db.Close()

	// 1. Verify default login code is 2212
	code, err := db.GetLoginCode()
	if err != nil || code != "2212" {
		t.Errorf("expected default passcode 2212, got %q, err=%v", code, err)
	}

	// 2. Add Task, Secret, Contact, Evidence
	tid, err := db.AddTask(TaskItem{
		CaseID:      "PRJ-999",
		Title:       "Deploy production cloud instance",
		Category:    "Work",
		Notes:       "High security environment",
		Priority:    "High",
		IsCompleted: false,
	})
	if err != nil || tid <= 0 {
		t.Fatalf("AddTask failed: %v", tid)
	}

	sid, err := db.AddSecret(VaultSecret{
		Title:       "Database Master Credentials",
		SecretType:  "Credential",
		Username:    "admin",
		SecretValue: "Omega#9921!",
		TargetURI:   "db.internal:5432",
		Notes:       "Confidential",
	})
	if err != nil || sid <= 0 {
		t.Fatalf("AddSecret failed: %v", sid)
	}

	cid, err := db.AddContact(PersonContact{
		FullName:          "Victor Vance",
		Alias:             "'Vic'",
		Role:              "Key Contact",
		Phone:             "+1 (555) 999-1122",
		Email:             "vance@enterprise.org",
		Address:           "742 Innovation Way",
		Organization:      "Vance Enterprises",
		ConfidentialIntel: "Direct project partner. Primary technical lead.",
	})
	if err != nil || cid <= 0 {
		t.Fatalf("AddContact failed: %v", cid)
	}

	eid, err := db.AddEvidence(EvidenceFile{
		CaseNumber:  "DOC-999",
		FileName:    "report.pdf",
		FileType:    "PDF/Report",
		FileSize:    2048,
		FileHash:    "a1b2c3d4e5",
		FileContent: "CONFIDENTIAL FINANCIAL AUDIT DATA",
		FileDetails: "Audit documentation",
	})
	if err != nil || eid <= 0 {
		t.Fatalf("AddEvidence failed: %v", eid)
	}

	// 3. Verify SQLite stored values are encrypted with 'enc:'
	var rawSecretVal string
	_ = db.db.QueryRow("SELECT secret_value FROM vault_secrets WHERE id = ?", sid).Scan(&rawSecretVal)
	if len(rawSecretVal) < 4 || rawSecretVal[:4] != "enc:" {
		t.Errorf("expected secret_value to be encrypted starting with 'enc:', got %q", rawSecretVal)
	}

	var rawIntel string
	_ = db.db.QueryRow("SELECT confidential_intel FROM contacts WHERE id = ?", cid).Scan(&rawIntel)
	if len(rawIntel) < 4 || rawIntel[:4] != "enc:" {
		t.Errorf("expected confidential_intel to be encrypted starting with 'enc:', got %q", rawIntel)
	}

	// 4. Test Decrypted Queries
	secrets, err := db.GetSecrets("Credential", "")
	if err != nil || len(secrets) == 0 {
		t.Fatalf("GetSecrets failed: %v", err)
	}
	if secrets[0].SecretValue != "Omega#9921!" {
		t.Errorf("expected decrypted secret 'Omega#9921!', got %q", secrets[0].SecretValue)
	}

	contacts, err := db.GetContacts("Key Contact", "")
	if err != nil || len(contacts) == 0 {
		t.Fatalf("GetContacts failed: %v", err)
	}
	if contacts[0].FullName != "Victor Vance" {
		t.Errorf("expected decrypted contact 'Victor Vance', got %q", contacts[0].FullName)
	}

	// 5. Test Passcode Update & Re-encryption
	err = db.UpdateLoginCode("2212", "7788")
	if err != nil {
		t.Fatalf("UpdateLoginCode failed: %v", err)
	}

	ok, _, err := db.VerifyPIN("7788")
	if !ok || err != nil {
		t.Errorf("expected new PIN 7788 to verify, got ok=%v, err=%v", ok, err)
	}

	// Verify decrypted items after passcode change
	secretsAfter, _ := db.GetSecrets("", "Omega")
	if len(secretsAfter) == 0 || secretsAfter[0].SecretValue != "Omega#9921!" {
		t.Errorf("expected secret to decrypt cleanly with new key after passcode change")
	}

	// 6. Test 3-Strike Auto-Wipe
	_, _, _ = db.VerifyPIN("bad1")
	_, _, _ = db.VerifyPIN("bad2")
	ok, rem, err := db.VerifyPIN("bad3")
	if ok || rem != 0 || err != ErrVaultWiped {
		t.Errorf("expected 3rd bad attempt to wipe database, got ok=%v, rem=%d, err=%v", ok, rem, err)
	}

	// Verify all tables are wiped
	st, _ := db.GetStats(dbPath)
	if st.TotalTasks != 0 || st.TotalSecrets != 0 || st.TotalContacts != 0 || st.TotalEvidence != 0 {
		t.Errorf("expected all tables to have 0 records after wipe, got %+v", st)
	}
}
