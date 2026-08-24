# Chapter 11: Enterprise Architectural Patterns — Persistence & Security

> *"How do production-grade Nova desktop applications manage SQLite database transactions, AES-256 encrypted vaults, OS file dialogs, and multi-backend clipboards?"*

---

## 11.1 SQLite in Desktop GUI Applications: WAL Mode

When a GUI application performs database queries on the main UI thread or in background goroutines, standard SQLite locking can cause the window to stutter or freeze.

Nova applications (like **VInfo** `examples/08_vinfo` and **Sticky Notes** `examples/09_sticky`) configure SQLite in **WAL (Write-Ahead Logging)** mode:

```go
func OpenDatabase(dbPath string) (*Database, error) {
    // Enable WAL mode and NORMAL synchronous writes for concurrent high-speed GUI queries:
    dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)", dbPath)
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, err
    }
    // ... run schema migrations ...
}
```

### Why WAL Mode Matters for GUIs:
- **Concurrent Readers & Writers**: Background sync threads can insert hundreds of records while the UI thread queries and renders data tables simultaneously without blocking!
- **Fast Transactions**: Writes append to a sequential `.db-wal` log file rather than rewriting database pages.

---

## 11.2 AES-256-GCM Encrypted Local Storage

In security applications (such as **VInfo Vault** `examples/08_vinfo/db.go`), sensitive files, notes, and records are stored with **AES-256-GCM (Galois/Counter Mode)** authenticated encryption:

```
[User Master Password] + [Random Salt]
                  │
        PBKDF2 (100,000 SHA-256 Iterations)
                  ▼
         [256-bit Secret Key]
                  │
        AES-GCM Authenticated Encryption
                  ▼
  [12-byte Nonce] + [Encrypted Ciphertext] + [16-byte Auth Tag]
```

```go
func EncryptData(plaintext []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    // Seal appends ciphertext + authentication tag to the nonce:
    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
```

If anyone tampers with the SQLite database file outside the application, AES-GCM authentication instantly rejects decryption!

---

## 11.3 Multi-Platform System Clipboard Integration

Copy-pasting across different Linux desktop managers (GNOME on Wayland, KDE on X11, Sway, XFCE) requires detecting the active clipboard protocol:

```go
func ReadClipboard() string {
    // 1. Wayland protocol (wl-paste)
    if out, err := exec.Command("wl-paste", "--no-newline").Output(); err == nil && len(out) > 0 {
        return string(out)
    }
    // 2. X11 clipboard (xclip)
    if out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output(); err == nil && len(out) > 0 {
        return string(out)
    }
    // 3. X11 primary selection (xsel)
    if out, err := exec.Command("xsel", "-b", "-o").Output(); err == nil && len(out) > 0 {
        return string(out)
    }
    // 4. macOS native (pbpaste)
    if out, err := exec.Command("pbpaste").Output(); err == nil && len(out) > 0 {
        return string(out)
    }
    // 5. In-memory internal fallback
    return internalClipboard
}
```

---

## 11.4 Native OS File Open Dialogs

When opening files, Nova leverages lightweight native platform dialog tools (`zenity`, `kdialog`, `AppleScript`):

```go
func OpenNativeFileDialog() (string, error) {
    // Try Zenity on Linux GNOME / GTK:
    cmd := exec.Command("zenity", "--file-selection", "--title=Select Document")
    out, err := cmd.Output()
    if err == nil {
        return strings.TrimSpace(string(out)), nil
    }
    // Try KDialog on Linux KDE:
    cmd2 := exec.Command("kdialog", "--getopenfilename")
    out2, err2 := cmd2.Output()
    if err2 == nil {
        return strings.TrimSpace(string(out2)), nil
    }
    return "", fmt.Errorf("no native dialog provider available")
}
```

In Chapter 12, we conclude with complete system diagrams, performance benchmarks, and future GPU acceleration roadmaps!
