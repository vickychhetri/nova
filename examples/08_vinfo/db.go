package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrVaultWiped  = errors.New("maximum failed attempts exceeded: all vault data has been permanently wiped")
	ErrInvalidPIN  = errors.New("incorrect passcode")
	ErrShortPIN    = errors.New("passcode must be at least 4 characters long")
	ErrWrongOldPIN = errors.New("current passcode is incorrect")
)

// TaskItem represents an actionable task, project milestone, or reminder.
type TaskItem struct {
	ID          int64
	CaseID      string // Project or Reference Code (e.g. "PRJ-2026-089")
	Title       string
	Category    string // "Work", "Urgent", "Personal", "Legal", "Research", "Vault"
	Notes       string
	Priority    string // "High", "Medium", "Low"
	IsCompleted bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// VaultSecret represents a confidential password, credential, access key, or secure note.
type VaultSecret struct {
	ID          int64
	Title       string
	SecretType  string // "Credential", "Access Key", "Secure Note", "Recovery Key", "API Token"
	Username    string
	SecretValue string // Passwords, PINs, classified strings
	TargetURI   string // System, Service, Portal, Server
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PersonContact represents a contact, associate, client, partner, or executive.
type PersonContact struct {
	ID                int64
	FullName          string
	Alias             string // Preferred name or alias
	Role              string // "Key Contact", "Associate", "Client", "Partner", "Executive", "Emergency Contact"
	Phone             string
	Email             string
	Address           string
	Organization      string
	ConfidentialIntel string // Sensitive notes (hidden behind passcode)
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// EvidenceFile represents an encrypted digital file, image, document, or attachment.
type EvidenceFile struct {
	ID          int64
	CaseNumber  string // Tag / Reference Code
	FileName    string
	FileType    string // "Image/Photo", "PDF/Report", "Audio/Recording", "Document", "Data/Archive"
	FileSize    int64
	FileHash    string // SHA-256 checksum
	FileContent string // Encrypted file payload / data
	FileDetails string // Description & confidential notes
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// VaultStats contains overview counts for the secure vault hub.
type VaultStats struct {
	TotalTasks        int
	PendingTasks      int
	CompletedTasks    int
	TotalSecrets      int
	TotalContacts     int
	TotalEvidence     int
	HighPriorityTasks int
	DatabasePath      string
	DatabaseBytes     int64
}

// Database encapsulates SQLite operations with AES-256-GCM encryption and 3-attempt auto-wipe.
type Database struct {
	mu sync.RWMutex
	db *sql.DB
}

// OpenDatabase initializes the SQLite database at the specified path and creates schema.
func OpenDatabase(dbPath string) (*Database, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
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
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return d, nil
}

// Close closes the database connection.
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Derive AES-256 key from PIN and persistent salt
func deriveKey(pin string, salt string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(pin))
	hasher.Write([]byte(salt))
	hasher.Write([]byte("vinfo-secure-vault-intel-2026"))
	return hasher.Sum(nil)
}

// EncryptText encrypts plaintext using AES-256-GCM and returns "enc:<base64>"
func encryptText(plainText string, key []byte) (string, error) {
	if plainText == "" {
		return "", nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return "enc:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptText decrypts ciphertext prefixed with "enc:<base64>" using AES-256-GCM
func decryptText(cipherText string, key []byte) (string, error) {
	if cipherText == "" {
		return "", nil
	}
	if !strings.HasPrefix(cipherText, "enc:") {
		return cipherText, nil
	}
	rawB64 := strings.TrimPrefix(cipherText, "enc:")
	data, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		return cipherText, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return cipherText, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return cipherText, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return cipherText, errors.New("malformed ciphertext")
	}
	nonce, actualCipher := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, actualCipher, nil)
	if err != nil {
		return cipherText, err
	}
	return string(plain), nil
}

func (d *Database) getSalt() (string, error) {
	var salt string
	err := d.db.QueryRow("SELECT value FROM auth_config WHERE key = 'vault_salt'").Scan(&salt)
	if err == sql.ErrNoRows {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		salt = base64.StdEncoding.EncodeToString(b)
		_, err = d.db.Exec("INSERT INTO auth_config (key, value, updated_at) VALUES ('vault_salt', ?, ?)", salt, time.Now())
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return salt, nil
}

func (d *Database) getActiveKey() ([]byte, error) {
	pin, err := d.GetLoginCode()
	if err != nil {
		return nil, err
	}
	salt, err := d.getSalt()
	if err != nil {
		return nil, err
	}
	return deriveKey(pin, salt), nil
}

func (d *Database) initSchema() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tables := []string{
		`CREATE TABLE IF NOT EXISTS auth_config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			case_id TEXT NOT NULL,
			title TEXT NOT NULL,
			category TEXT NOT NULL,
			notes TEXT DEFAULT '',
			priority TEXT DEFAULT 'Medium',
			is_completed INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS vault_secrets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			secret_type TEXT NOT NULL,
			username TEXT DEFAULT '',
			secret_value TEXT NOT NULL,
			target_uri TEXT DEFAULT '',
			notes TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			full_name TEXT NOT NULL,
			alias TEXT DEFAULT '',
			role TEXT NOT NULL,
			phone TEXT DEFAULT '',
			email TEXT DEFAULT '',
			address TEXT DEFAULT '',
			organization TEXT DEFAULT '',
			confidential_intel TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS evidence_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			case_number TEXT NOT NULL,
			file_name TEXT NOT NULL,
			file_type TEXT NOT NULL,
			file_size INTEGER DEFAULT 0,
			file_hash TEXT NOT NULL,
			file_content TEXT DEFAULT '',
			file_details TEXT DEFAULT '',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);`,
	}

	for _, stmt := range tables {
		if _, err := d.db.Exec(stmt); err != nil {
			return err
		}
	}

	// Default PIN: 2212
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM auth_config WHERE key = 'login_code'").Scan(&count)
	if err != nil {
		return err
	}
	now := time.Now()
	if count == 0 {
		_, err = d.db.Exec("INSERT INTO auth_config (key, value, updated_at) VALUES ('login_code', '2212', ?)", now)
		if err != nil {
			return err
		}
	}

	// Initialize failed attempts & salt
	var failedCount int
	_ = d.db.QueryRow("SELECT COUNT(*) FROM auth_config WHERE key = 'failed_attempts'").Scan(&failedCount)
	if failedCount == 0 {
		_, _ = d.db.Exec("INSERT INTO auth_config (key, value, updated_at) VALUES ('failed_attempts', '0', ?)", now)
	}

	var saltCount int
	_ = d.db.QueryRow("SELECT COUNT(*) FROM auth_config WHERE key = 'vault_salt'").Scan(&saltCount)
	if saltCount == 0 {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		salt := base64.StdEncoding.EncodeToString(b)
		_, _ = d.db.Exec("INSERT INTO auth_config (key, value, updated_at) VALUES ('vault_salt', ?, ?)", salt, now)
	}

	// Seed Sample Vault Data if empty
	var taskCount int
	_ = d.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if taskCount == 0 {
		salt, _ := d.getSalt()
		key := deriveKey("2212", salt)

		// Seed Tasks
		sampleTasks := []TaskItem{
			{
				CaseID:      "PRJ-2026-089",
				Title:       "Finalize quarterly financial audit and tax filings",
				Category:    "Work",
				Notes:       "Coordinate with accounting team. Review expense reports.",
				Priority:    "High",
				IsCompleted: false,
				CreatedAt:   now.Add(-2 * time.Hour),
				UpdatedAt:   now.Add(-2 * time.Hour),
			},
			{
				CaseID:      "PRJ-2026-104",
				Title:       "Conduct confidential security architecture review",
				Category:    "Research",
				Notes:       "Review end-to-end encryption keys, rotation policy, and access control.",
				Priority:    "Medium",
				IsCompleted: false,
				CreatedAt:   now.Add(-1 * time.Hour),
				UpdatedAt:   now.Add(-1 * time.Hour),
			},
			{
				CaseID:      "PRJ-2026-077",
				Title:       "Backup personal documents and private credentials",
				Category:    "Personal",
				Notes:       "Encrypted backup generated and verified.",
				Priority:    "High",
				IsCompleted: true,
				CreatedAt:   now.Add(-4 * time.Hour),
				UpdatedAt:   now.Add(-4 * time.Hour),
			},
		}
		for _, t := range sampleTasks {
			encCase, _ := encryptText(t.CaseID, key)
			encTitle, _ := encryptText(t.Title, key)
			encCat, _ := encryptText(t.Category, key)
			encNotes, _ := encryptText(t.Notes, key)
			encPri, _ := encryptText(t.Priority, key)
			comp := 0
			if t.IsCompleted {
				comp = 1
			}
			_, _ = d.db.Exec(`INSERT INTO tasks (case_id, title, category, notes, priority, is_completed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				encCase, encTitle, encCat, encNotes, encPri, comp, t.CreatedAt, t.UpdatedAt)
		}

		// Seed Secrets & Credentials
		sampleSecrets := []VaultSecret{
			{
				Title:       "Production Cloud Infrastructure Master Key",
				SecretType:  "Credential",
				Username:    "admin.lead",
				SecretValue: "K9$Secure#Pass8812!",
				TargetURI:   "https://cloud-console.internal:8443",
				Notes:       "Authorized for system administrators. Rotated quarterly.",
				CreatedAt:   now.Add(-3 * time.Hour),
				UpdatedAt:   now.Add(-3 * time.Hour),
			},
			{
				Title:       "Encrypted Recovery Seed Phrase",
				SecretType:  "Recovery Key",
				Username:    "MasterWallet",
				SecretValue: "echo velvet quantum horizon matrix beacon solar pulse galaxy orbit",
				TargetURI:   "Cold Storage Hardware Wallet",
				Notes:       "24-word root key. Keep strictly confidential.",
				CreatedAt:   now.Add(-2 * time.Hour),
				UpdatedAt:   now.Add(-2 * time.Hour),
			},
			{
				Title:       "Office Secure Entry Pad Code",
				SecretType:  "Access Key",
				Username:    "security_door",
				SecretValue: "8899-7711-X4",
				TargetURI:   "Building 4 Level 3",
				Notes:       "Keypad access for private lab.",
				CreatedAt:   now.Add(-1 * time.Hour),
				UpdatedAt:   now.Add(-1 * time.Hour),
			},
		}
		for _, s := range sampleSecrets {
			encTitle, _ := encryptText(s.Title, key)
			encType, _ := encryptText(s.SecretType, key)
			encUser, _ := encryptText(s.Username, key)
			encVal, _ := encryptText(s.SecretValue, key)
			encURI, _ := encryptText(s.TargetURI, key)
			encNotes, _ := encryptText(s.Notes, key)
			_, _ = d.db.Exec(`INSERT INTO vault_secrets (title, secret_type, username, secret_value, target_uri, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				encTitle, encType, encUser, encVal, encURI, encNotes, s.CreatedAt, s.UpdatedAt)
		}

		// Seed Contacts & Associates
		sampleContacts := []PersonContact{
			{
				FullName:          "Alexander Vance",
				Alias:             "'Alex'",
				Role:              "Key Contact",
				Phone:             "+1 (555) 019-2834",
				Email:             "a.vance@encrypted-mail.org",
				Address:           "742 Innovation Way, Tech Park",
				Organization:      "Vance Global Logistics",
				ConfidentialIntel: "Primary technical contact for enterprise logistics. Available Monday through Thursday.",
				CreatedAt:         now.Add(-5 * time.Hour),
				UpdatedAt:         now.Add(-5 * time.Hour),
			},
			{
				FullName:          "Elena Rostova",
				Alias:             "'Legal Counsel'",
				Role:              "Partner",
				Phone:             "+1 (555) 018-9921",
				Email:             "elena.legal@proton.me",
				Address:           "District 3 Corporate Plaza",
				Organization:      "Rostova & Associates Legal",
				ConfidentialIntel: "Direct legal advisor for international contracts and regulatory compliance.",
				CreatedAt:         now.Add(-3 * time.Hour),
				UpdatedAt:         now.Add(-3 * time.Hour),
			},
			{
				FullName:          "Sarah Jenkins",
				Alias:             "Lead Architect",
				Role:              "Executive",
				Phone:             "+1 (555) 010-4400",
				Email:             "sjenkins@enterprise.com",
				Address:           "Innovation Center, Suite 302",
				Organization:      "Executive Technology Board",
				ConfidentialIntel: "Oversees research and development projects.",
				CreatedAt:         now.Add(-6 * time.Hour),
				UpdatedAt:         now.Add(-6 * time.Hour),
			},
		}
		for _, c := range sampleContacts {
			encName, _ := encryptText(c.FullName, key)
			encAlias, _ := encryptText(c.Alias, key)
			encRole, _ := encryptText(c.Role, key)
			encPhone, _ := encryptText(c.Phone, key)
			encEmail, _ := encryptText(c.Email, key)
			encAddr, _ := encryptText(c.Address, key)
			encOrg, _ := encryptText(c.Organization, key)
			encIntel, _ := encryptText(c.ConfidentialIntel, key)
			_, _ = d.db.Exec(`INSERT INTO contacts (full_name, alias, role, phone, email, address, organization, confidential_intel, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				encName, encAlias, encRole, encPhone, encEmail, encAddr, encOrg, encIntel, c.CreatedAt, c.UpdatedAt)
		}

		// Seed Encrypted Files
		sampleFiles := []EvidenceFile{
			{
				CaseNumber:  "DOC-2026-104",
				FileName:    "Quarterly_Financial_Report.pdf",
				FileType:    "PDF/Report",
				FileSize:    48500,
				FileHash:    "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4",
				FileContent: "EXECUTIVE FINANCIAL SUMMARY:\nTotal Assets: $4,250,000\nOperational Expenses: $890,000\nNet Growth: +18.4%\nStatus: Verified and Sealed.",
				FileDetails: "Confidential audit statement verified by CPA.",
				CreatedAt:   now.Add(-1 * time.Hour),
				UpdatedAt:   now.Add(-1 * time.Hour),
			},
			{
				CaseNumber:  "DOC-2026-112",
				FileName:    "MFD_Postman_Collection.json",
				FileType:    "Document",
				FileSize:    26010,
				FileHash:    "39dd5124c7699116eab839b2275f4dea2761aeeb4921d18952813721e9047e599e",
				FileContent: `{"info":{"_postman_id":"034fbbfd-181f-4a19-a043-cf5e4b0e91a5","name":"MFD_API","schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},"item":[{"name":"Get Vault Stats","request":{"method":"GET","header":[],"url":{"raw":"https://api.vault.internal/v1/stats"}}},{"name":"Authenticate Session","request":{"method":"POST","header":[{"key":"Content-Type","value":"application/json"}],"body":{"mode":"raw","raw":"{\"pin\":\"2212\"}"},"url":{"raw":"https://api.vault.internal/v1/auth"}}}]}`,
				FileDetails: "Encrypted API configuration & test collection.",
				CreatedAt:   now.Add(-2 * time.Hour),
				UpdatedAt:   now.Add(-2 * time.Hour),
			},
			{
				CaseNumber:  "DOC-2026-095",
				FileName:    "Executive_Briefing_Audio.mp3",
				FileType:    "Audio/Recording",
				FileSize:    345000,
				FileHash:    "c7a1098e9124fbcf89324089aeeb8912784b1239012489012384912093849012",
				FileContent: "EXECUTIVE AUDIO LOG: Recording verified. All security perimeter nodes active. Encryption rotation confirmed.",
				FileDetails: "Confidential strategy meeting voice notes.",
				CreatedAt:   now.Add(-3 * time.Hour),
				UpdatedAt:   now.Add(-3 * time.Hour),
			},
		}
		for _, f := range sampleFiles {
			encCase, _ := encryptText(f.CaseNumber, key)
			encName, _ := encryptText(f.FileName, key)
			encType, _ := encryptText(f.FileType, key)
			encHash, _ := encryptText(f.FileHash, key)
			encCnt, _ := encryptText(f.FileContent, key)
			encDetails, _ := encryptText(f.FileDetails, key)
			_, _ = d.db.Exec(`INSERT INTO evidence_files (case_number, file_name, file_type, file_size, file_hash, file_content, file_details, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				encCase, encName, encType, f.FileSize, encHash, encCnt, encDetails, f.CreatedAt, f.UpdatedAt)
		}
	}

	return nil
}

// GetLoginCode returns the stored master passcode.
func (d *Database) GetLoginCode() (string, error) {
	var code string
	err := d.db.QueryRow("SELECT value FROM auth_config WHERE key = 'login_code'").Scan(&code)
	if err != nil {
		return "", err
	}
	return code, nil
}

// VerifyPIN verifies the passcode. If wrong 3 times, triggers WipeAllData and returns ErrVaultWiped.
func (d *Database) VerifyPIN(pin string) (bool, int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var actual string
	err := d.db.QueryRow("SELECT value FROM auth_config WHERE key = 'login_code'").Scan(&actual)
	if err != nil {
		return false, 0, err
	}

	var failedStr string
	err = d.db.QueryRow("SELECT value FROM auth_config WHERE key = 'failed_attempts'").Scan(&failedStr)
	failedAttempts := 0
	if err == nil {
		failedAttempts, _ = strconv.Atoi(failedStr)
	}

	if pin == actual {
		_, _ = d.db.Exec("UPDATE auth_config SET value = '0', updated_at = ? WHERE key = 'failed_attempts'", time.Now())
		return true, 3, nil
	}

	failedAttempts++
	now := time.Now()
	_, _ = d.db.Exec("UPDATE auth_config SET value = ?, updated_at = ? WHERE key = 'failed_attempts'", strconv.Itoa(failedAttempts), now)

	if failedAttempts >= 3 {
		// 3-Strike Wipe
		_, _ = d.db.Exec("DELETE FROM tasks")
		_, _ = d.db.Exec("DELETE FROM vault_secrets")
		_, _ = d.db.Exec("DELETE FROM contacts")
		_, _ = d.db.Exec("DELETE FROM evidence_files")
		_, _ = d.db.Exec("VACUUM")
		_, _ = d.db.Exec("UPDATE auth_config SET value = '2212', updated_at = ? WHERE key = 'login_code'", now)
		_, _ = d.db.Exec("UPDATE auth_config SET value = '0', updated_at = ? WHERE key = 'failed_attempts'", now)

		b := make([]byte, 16)
		_, _ = rand.Read(b)
		salt := base64.StdEncoding.EncodeToString(b)
		_, _ = d.db.Exec("UPDATE auth_config SET value = ?, updated_at = ? WHERE key = 'vault_salt'", salt, now)

		return false, 0, ErrVaultWiped
	}

	remaining := 3 - failedAttempts
	return false, remaining, ErrInvalidPIN
}

// WipeAllData manually wipes all vault intelligence tables and resets to default 2212.
func (d *Database) WipeAllData() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, _ = d.db.Exec("DELETE FROM tasks")
	_, _ = d.db.Exec("DELETE FROM vault_secrets")
	_, _ = d.db.Exec("DELETE FROM contacts")
	_, _ = d.db.Exec("DELETE FROM evidence_files")
	_, _ = d.db.Exec("VACUUM")

	now := time.Now()
	_, _ = d.db.Exec("UPDATE auth_config SET value = '2212', updated_at = ? WHERE key = 'login_code'", now)
	_, _ = d.db.Exec("UPDATE auth_config SET value = '0', updated_at = ? WHERE key = 'failed_attempts'", now)

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	salt := base64.StdEncoding.EncodeToString(b)
	_, _ = d.db.Exec("UPDATE auth_config SET value = ?, updated_at = ? WHERE key = 'vault_salt'", salt, now)

	return nil
}

// UpdateLoginCode updates passcode and re-encrypts all database records with new key.
func (d *Database) UpdateLoginCode(oldPIN, newPIN string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var actual string
	err := d.db.QueryRow("SELECT value FROM auth_config WHERE key = 'login_code'").Scan(&actual)
	if err != nil {
		return err
	}

	if oldPIN != actual {
		return ErrWrongOldPIN
	}

	if len(newPIN) < 4 {
		return ErrShortPIN
	}

	// 1. Old and New Keys
	var salt string
	_ = d.db.QueryRow("SELECT value FROM auth_config WHERE key = 'vault_salt'").Scan(&salt)
	oldKey := deriveKey(oldPIN, salt)

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	newSalt := base64.StdEncoding.EncodeToString(b)
	newKey := deriveKey(newPIN, newSalt)

	// 2. Re-encrypt Tasks
	taskRows, err := d.db.Query("SELECT id, case_id, title, category, notes, priority FROM tasks")
	if err == nil {
		type reTask struct {
			id, caseID, title, cat, notes, pri string
		}
		var tasks []reTask
		for taskRows.Next() {
			var id int64
			var c, t, cat, n, p string
			if err := taskRows.Scan(&id, &c, &t, &cat, &n, &p); err == nil {
				decC, _ := decryptText(c, oldKey)
				decT, _ := decryptText(t, oldKey)
				decCat, _ := decryptText(cat, oldKey)
				decN, _ := decryptText(n, oldKey)
				decP, _ := decryptText(p, oldKey)
				tasks = append(tasks, reTask{strconv.FormatInt(id, 10), decC, decT, decCat, decN, decP})
			}
		}
		taskRows.Close()
		for _, it := range tasks {
			encC, _ := encryptText(it.caseID, newKey)
			encT, _ := encryptText(it.title, newKey)
			encCat, _ := encryptText(it.cat, newKey)
			encN, _ := encryptText(it.notes, newKey)
			encP, _ := encryptText(it.pri, newKey)
			_, _ = d.db.Exec("UPDATE tasks SET case_id = ?, title = ?, category = ?, notes = ?, priority = ? WHERE id = ?", encC, encT, encCat, encN, encP, it.id)
		}
	}

	// 3. Re-encrypt Secrets
	secRows, err := d.db.Query("SELECT id, title, secret_type, username, secret_value, target_uri, notes FROM vault_secrets")
	if err == nil {
		type reSec struct {
			id, title, sType, user, val, uri, notes string
		}
		var secrets []reSec
		for secRows.Next() {
			var id int64
			var t, st, u, v, uri, n string
			if err := secRows.Scan(&id, &t, &st, &u, &v, &uri, &n); err == nil {
				decT, _ := decryptText(t, oldKey)
				decST, _ := decryptText(st, oldKey)
				decU, _ := decryptText(u, oldKey)
				decV, _ := decryptText(v, oldKey)
				decURI, _ := decryptText(uri, oldKey)
				decN, _ := decryptText(n, oldKey)
				secrets = append(secrets, reSec{strconv.FormatInt(id, 10), decT, decST, decU, decV, decURI, decN})
			}
		}
		secRows.Close()
		for _, s := range secrets {
			encT, _ := encryptText(s.title, newKey)
			encST, _ := encryptText(s.sType, newKey)
			encU, _ := encryptText(s.user, newKey)
			encV, _ := encryptText(s.val, newKey)
			encURI, _ := encryptText(s.uri, newKey)
			encN, _ := encryptText(s.notes, newKey)
			_, _ = d.db.Exec("UPDATE vault_secrets SET title = ?, secret_type = ?, username = ?, secret_value = ?, target_uri = ?, notes = ? WHERE id = ?", encT, encST, encU, encV, encURI, encN, s.id)
		}
	}

	// 4. Re-encrypt Contacts
	cntRows, err := d.db.Query("SELECT id, full_name, alias, role, phone, email, address, organization, confidential_intel FROM contacts")
	if err == nil {
		type reCnt struct {
			id, name, alias, role, phone, email, addr, org, intel string
		}
		var contacts []reCnt
		for cntRows.Next() {
			var id int64
			var name, alias, role, ph, em, addr, org, intel string
			if err := cntRows.Scan(&id, &name, &alias, &role, &ph, &em, &addr, &org, &intel); err == nil {
				contacts = append(contacts, reCnt{
					strconv.FormatInt(id, 10),
					mustDec(name, oldKey),
					mustDec(alias, oldKey),
					mustDec(role, oldKey),
					mustDec(ph, oldKey),
					mustDec(em, oldKey),
					mustDec(addr, oldKey),
					mustDec(org, oldKey),
					mustDec(intel, oldKey),
				})
			}
		}
		cntRows.Close()
		for _, c := range contacts {
			_, _ = d.db.Exec(`UPDATE contacts SET full_name = ?, alias = ?, role = ?, phone = ?, email = ?, address = ?, organization = ?, confidential_intel = ? WHERE id = ?`,
				mustEnc(c.name, newKey), mustEnc(c.alias, newKey), mustEnc(c.role, newKey), mustEnc(c.phone, newKey), mustEnc(c.email, newKey),
				mustEnc(c.addr, newKey), mustEnc(c.org, newKey), mustEnc(c.intel, newKey), c.id)
		}
	}

	// 5. Re-encrypt Evidence Files
	evRows, err := d.db.Query("SELECT id, case_number, file_name, file_type, file_hash, file_content, file_details FROM evidence_files")
	if err == nil {
		type reEv struct {
			id, cNum, name, fType, hash, content, details string
		}
		var files []reEv
		for evRows.Next() {
			var id int64
			var cNum, name, fType, hash, content, details string
			if err := evRows.Scan(&id, &cNum, &name, &fType, &hash, &content, &details); err == nil {
				files = append(files, reEv{
					strconv.FormatInt(id, 10),
					mustDec(cNum, oldKey),
					mustDec(name, oldKey),
					mustDec(fType, oldKey),
					mustDec(hash, oldKey),
					mustDec(content, oldKey),
					mustDec(details, oldKey),
				})
			}
		}
		evRows.Close()
		for _, f := range files {
			_, _ = d.db.Exec(`UPDATE evidence_files SET case_number = ?, file_name = ?, file_type = ?, file_hash = ?, file_content = ?, file_details = ? WHERE id = ?`,
				mustEnc(f.cNum, newKey), mustEnc(f.name, newKey), mustEnc(f.fType, newKey), mustEnc(f.hash, newKey), mustEnc(f.content, newKey), mustEnc(f.details, newKey), f.id)
		}
	}

	// 6. Update auth_config
	now := time.Now()
	_, _ = d.db.Exec("UPDATE auth_config SET value = ?, updated_at = ? WHERE key = 'login_code'", newPIN, now)
	_, _ = d.db.Exec("UPDATE auth_config SET value = ?, updated_at = ? WHERE key = 'vault_salt'", newSalt, now)
	_, _ = d.db.Exec("UPDATE auth_config SET value = '0', updated_at = ? WHERE key = 'failed_attempts'", now)

	return nil
}

func mustEnc(val string, key []byte) string {
	s, _ := encryptText(val, key)
	return s
}

func mustDec(val string, key []byte) string {
	s, _ := decryptText(val, key)
	return s
}

// ------------------------------------------------------------
// Task CRUD
// ------------------------------------------------------------

func (d *Database) AddTask(t TaskItem) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key, err := d.getActiveKey()
	if err != nil {
		return 0, err
	}

	encCase, _ := encryptText(t.CaseID, key)
	encTitle, _ := encryptText(t.Title, key)
	encCat, _ := encryptText(t.Category, key)
	encNotes, _ := encryptText(t.Notes, key)
	encPri, _ := encryptText(t.Priority, key)

	comp := 0
	if t.IsCompleted {
		comp = 1
	}
	now := time.Now()

	res, err := d.db.Exec(`INSERT INTO tasks (case_id, title, category, notes, priority, is_completed, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		encCase, encTitle, encCat, encNotes, encPri, comp, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Database) GetTasks(categoryFilter, searchFilter string, onlyPending bool) ([]TaskItem, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key, err := d.getActiveKey()
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(`SELECT id, case_id, title, category, notes, priority, is_completed, created_at, updated_at FROM tasks ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TaskItem
	searchLower := strings.ToLower(strings.TrimSpace(searchFilter))

	for rows.Next() {
		var t TaskItem
		var comp int
		var encCase, encTitle, encCat, encNotes, encPri string

		if err := rows.Scan(&t.ID, &encCase, &encTitle, &encCat, &encNotes, &encPri, &comp, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}

		t.CaseID, _ = decryptText(encCase, key)
		t.Title, _ = decryptText(encTitle, key)
		t.Category, _ = decryptText(encCat, key)
		t.Notes, _ = decryptText(encNotes, key)
		t.Priority, _ = decryptText(encPri, key)
		t.IsCompleted = (comp == 1)

		if categoryFilter != "" && categoryFilter != "All" {
			if !strings.EqualFold(t.Category, categoryFilter) {
				continue
			}
		}

		if onlyPending && t.IsCompleted {
			continue
		}

		if searchLower != "" {
			cMatch := strings.Contains(strings.ToLower(t.CaseID), searchLower)
			tMatch := strings.Contains(strings.ToLower(t.Title), searchLower)
			nMatch := strings.Contains(strings.ToLower(t.Notes), searchLower)
			if !cMatch && !tMatch && !nMatch {
				continue
			}
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (d *Database) ToggleTaskCompletion(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec(`UPDATE tasks SET is_completed = CASE WHEN is_completed = 1 THEN 0 ELSE 1 END, updated_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

func (d *Database) DeleteTask(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// ------------------------------------------------------------
// Vault Secrets CRUD (Passwords, Credentials, Keys)
// ------------------------------------------------------------

func (d *Database) AddSecret(s VaultSecret) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key, err := d.getActiveKey()
	if err != nil {
		return 0, err
	}

	encTitle, _ := encryptText(s.Title, key)
	encType, _ := encryptText(s.SecretType, key)
	encUser, _ := encryptText(s.Username, key)
	encVal, _ := encryptText(s.SecretValue, key)
	encURI, _ := encryptText(s.TargetURI, key)
	encNotes, _ := encryptText(s.Notes, key)

	now := time.Now()
	res, err := d.db.Exec(`INSERT INTO vault_secrets (title, secret_type, username, secret_value, target_uri, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		encTitle, encType, encUser, encVal, encURI, encNotes, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Database) GetSecrets(typeFilter, searchFilter string) ([]VaultSecret, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key, err := d.getActiveKey()
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(`SELECT id, title, secret_type, username, secret_value, target_uri, notes, created_at, updated_at FROM vault_secrets ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []VaultSecret
	searchLower := strings.ToLower(strings.TrimSpace(searchFilter))

	for rows.Next() {
		var s VaultSecret
		var encTitle, encType, encUser, encVal, encURI, encNotes string

		if err := rows.Scan(&s.ID, &encTitle, &encType, &encUser, &encVal, &encURI, &encNotes, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}

		s.Title, _ = decryptText(encTitle, key)
		s.SecretType, _ = decryptText(encType, key)
		s.Username, _ = decryptText(encUser, key)
		s.SecretValue, _ = decryptText(encVal, key)
		s.TargetURI, _ = decryptText(encURI, key)
		s.Notes, _ = decryptText(encNotes, key)

		if typeFilter != "" && typeFilter != "All" {
			if !strings.EqualFold(s.SecretType, typeFilter) {
				continue
			}
		}

		if searchLower != "" {
			tMatch := strings.Contains(strings.ToLower(s.Title), searchLower)
			uMatch := strings.Contains(strings.ToLower(s.Username), searchLower)
			vMatch := strings.Contains(strings.ToLower(s.SecretValue), searchLower)
			nMatch := strings.Contains(strings.ToLower(s.Notes), searchLower)
			if !tMatch && !uMatch && !vMatch && !nMatch {
				continue
			}
		}

		secrets = append(secrets, s)
	}

	return secrets, nil
}

func (d *Database) DeleteSecret(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("DELETE FROM vault_secrets WHERE id = ?", id)
	return err
}

// ------------------------------------------------------------
// Contacts & Associates CRUD
// ------------------------------------------------------------

func (d *Database) AddContact(c PersonContact) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key, err := d.getActiveKey()
	if err != nil {
		return 0, err
	}

	encName, _ := encryptText(c.FullName, key)
	encAlias, _ := encryptText(c.Alias, key)
	encRole, _ := encryptText(c.Role, key)
	encPhone, _ := encryptText(c.Phone, key)
	encEmail, _ := encryptText(c.Email, key)
	encAddr, _ := encryptText(c.Address, key)
	encOrg, _ := encryptText(c.Organization, key)
	encIntel, _ := encryptText(c.ConfidentialIntel, key)

	now := time.Now()
	res, err := d.db.Exec(`INSERT INTO contacts (full_name, alias, role, phone, email, address, organization, confidential_intel, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		encName, encAlias, encRole, encPhone, encEmail, encAddr, encOrg, encIntel, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Database) GetContacts(roleFilter, searchFilter string) ([]PersonContact, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key, err := d.getActiveKey()
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(`SELECT id, full_name, alias, role, phone, email, address, organization, confidential_intel, created_at, updated_at FROM contacts ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []PersonContact
	searchLower := strings.ToLower(strings.TrimSpace(searchFilter))

	for rows.Next() {
		var c PersonContact
		var encName, encAlias, encRole, encPhone, encEmail, encAddr, encOrg, encIntel string

		if err := rows.Scan(&c.ID, &encName, &encAlias, &encRole, &encPhone, &encEmail, &encAddr, &encOrg, &encIntel, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}

		c.FullName, _ = decryptText(encName, key)
		c.Alias, _ = decryptText(encAlias, key)
		c.Role, _ = decryptText(encRole, key)
		c.Phone, _ = decryptText(encPhone, key)
		c.Email, _ = decryptText(encEmail, key)
		c.Address, _ = decryptText(encAddr, key)
		c.Organization, _ = decryptText(encOrg, key)
		c.ConfidentialIntel, _ = decryptText(encIntel, key)

		if roleFilter != "" && roleFilter != "All" {
			if !strings.EqualFold(c.Role, roleFilter) {
				continue
			}
		}

		if searchLower != "" {
			nMatch := strings.Contains(strings.ToLower(c.FullName), searchLower)
			aMatch := strings.Contains(strings.ToLower(c.Alias), searchLower)
			oMatch := strings.Contains(strings.ToLower(c.Organization), searchLower)
			eMatch := strings.Contains(strings.ToLower(c.Email), searchLower)
			if !nMatch && !aMatch && !oMatch && !eMatch {
				continue
			}
		}

		contacts = append(contacts, c)
	}

	return contacts, nil
}

func (d *Database) DeleteContact(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("DELETE FROM contacts WHERE id = ?", id)
	return err
}

// ------------------------------------------------------------
// Evidence & Files CRUD (with encrypted file content & upload)
// ------------------------------------------------------------

func (d *Database) AddEvidence(f EvidenceFile) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key, err := d.getActiveKey()
	if err != nil {
		return 0, err
	}

	encCase, _ := encryptText(f.CaseNumber, key)
	encName, _ := encryptText(f.FileName, key)
	encType, _ := encryptText(f.FileType, key)
	encHash, _ := encryptText(f.FileHash, key)
	encCnt, _ := encryptText(f.FileContent, key)
	encDetails, _ := encryptText(f.FileDetails, key)

	now := time.Now()
	res, err := d.db.Exec(`INSERT INTO evidence_files (case_number, file_name, file_type, file_size, file_hash, file_content, file_details, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		encCase, encName, encType, f.FileSize, encHash, encCnt, encDetails, now, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UploadAndEncryptFile imports a local file from disk, hashes it with SHA-256, and encrypts it into the vault.
func (d *Database) UploadAndEncryptFile(caseNumber, filePath, details string) (*EvidenceFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	baseName := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))
	fType := "Document"
	switch ext {
	case ".png", ".jpg", ".jpeg", ".bmp", ".gif", ".webp":
		fType = "Image/Photo"
	case ".pdf":
		fType = "PDF/Report"
	case ".mp3", ".wav", ".m4a", ".ogg":
		fType = "Audio/Recording"
	case ".raw", ".bin", ".dat", ".pcap", ".zip", ".tar.gz":
		fType = "Data/Archive"
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	filePayload := string(data)
	if fType == "Image/Photo" || fType == "Audio/Recording" || fType == "Data/Archive" {
		filePayload = base64.StdEncoding.EncodeToString(data)
	}

	ev := EvidenceFile{
		CaseNumber:  caseNumber,
		FileName:    baseName,
		FileType:    fType,
		FileSize:    int64(len(data)),
		FileHash:    hash,
		FileContent: filePayload,
		FileDetails: details,
	}

	id, err := d.AddEvidence(ev)
	if err != nil {
		return nil, err
	}
	ev.ID = id
	return &ev, nil
}

func (d *Database) GetEvidence(typeFilter, searchFilter string) ([]EvidenceFile, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key, err := d.getActiveKey()
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(`SELECT id, case_number, file_name, file_type, file_size, file_hash, file_content, file_details, created_at, updated_at FROM evidence_files ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []EvidenceFile
	searchLower := strings.ToLower(strings.TrimSpace(searchFilter))

	for rows.Next() {
		var f EvidenceFile
		var encCase, encName, encType, encHash, encCnt, encDetails string

		if err := rows.Scan(&f.ID, &encCase, &encName, &encType, &f.FileSize, &encHash, &encCnt, &encDetails, &f.CreatedAt, &f.UpdatedAt); err != nil {
			continue
		}

		f.CaseNumber, _ = decryptText(encCase, key)
		f.FileName, _ = decryptText(encName, key)
		f.FileType, _ = decryptText(encType, key)
		f.FileHash, _ = decryptText(encHash, key)
		f.FileContent, _ = decryptText(encCnt, key)
		f.FileDetails, _ = decryptText(encDetails, key)

		if typeFilter != "" && typeFilter != "All" {
			if !strings.EqualFold(f.FileType, typeFilter) {
				continue
			}
		}

		if searchLower != "" {
			cMatch := strings.Contains(strings.ToLower(f.CaseNumber), searchLower)
			nMatch := strings.Contains(strings.ToLower(f.FileName), searchLower)
			dMatch := strings.Contains(strings.ToLower(f.FileDetails), searchLower)
			if !cMatch && !nMatch && !dMatch {
				continue
			}
		}

		files = append(files, f)
	}

	return files, nil
}

func (d *Database) DeleteEvidence(id int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, err := d.db.Exec("DELETE FROM evidence_files WHERE id = ?", id)
	return err
}

// ------------------------------------------------------------
// Summary Statistics
// ------------------------------------------------------------

func (d *Database) GetStats(dbPath string) (VaultStats, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var st VaultStats
	st.DatabasePath = dbPath

	if fi, err := os.Stat(dbPath); err == nil {
		st.DatabaseBytes = fi.Size()
	}

	_ = d.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&st.TotalTasks)
	_ = d.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE is_completed = 0").Scan(&st.PendingTasks)
	_ = d.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE is_completed = 1").Scan(&st.CompletedTasks)
	_ = d.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE priority = 'High' AND is_completed = 0").Scan(&st.HighPriorityTasks)

	_ = d.db.QueryRow("SELECT COUNT(*) FROM vault_secrets").Scan(&st.TotalSecrets)
	_ = d.db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&st.TotalContacts)
	_ = d.db.QueryRow("SELECT COUNT(*) FROM evidence_files").Scan(&st.TotalEvidence)

	return st, nil
}
