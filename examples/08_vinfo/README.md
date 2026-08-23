# V-Info — Personal Information Vault & Task Hub

[![Nova Framework](https://img.shields.io/badge/Built%20With-Nova%20Go%20UI-00ADD8.svg)](https://github.com/vickychhetri/nova)
[![Database](https://img.shields.io/badge/Database-SQLite%20(Pure%20Go)-blue.svg)](#)
[![Security](https://img.shields.io/badge/Security-PIN%20Protected%20(2212)-success.svg)](#)

> **V-Info** is a native Go desktop application for organizing to-dos, tracking tasks, and securely storing personal information, notes, credentials, and snippets for later — backed by a local SQLite database (`vinfo.db`).

---

## Key Features

1. 🔐 **Passcode Authentication & Lock Screen**:
   - Master PIN authentication on startup.
   - **Default Login Code**: `2212`.
   - Dedicated key pad & password input with instant feedback.
   - One-click `🔒 Lock Vault` button to securely lock your workspace at any time.

2. ⚙️ **In-App Passcode Management**:
   - Dedicated **"Security & Change PIN"** tab inside the application.
   - Validate current PIN, set a new 4+ character passcode, and update directly in the SQLite `auth_config` table.

3. 📋 **To-Do & Task Management**:
   - Add new tasks with Title, Category (`Todo`, `Work`, `Personal`, `Ideas`), Priority (`High`, `Medium`, `Low`), and detailed notes.
   - Filter by category and instant search bar.
   - Toggle completion status (`Mark Done` / `Completed`) and remove tasks with `🗑 Delete`.

4. 🗄️ **Information Vault (Save for Later)**:
   - Securely save important information, server keys, recovery codes, snippets, and meeting notes.
   - Secret masking mode (`🔒 [Protected Secret] ••••••••••••••••`).

5. 📊 **Dashboard & SQLite Telemetry**:
   - Real-time stat cards (Total records, Active tasks, Finished tasks, Protected secrets).
   - Live SQLite database file size and WAL engine status.

6. 🎨 **Custom V-Info Branding**:
   - High-performance vector canvas shield logo with cyan/violet branding.

---

## Running the Application

To run the V-Info application:

```bash
go run ./examples/08_vinfo
```

### Run Unit Tests:

```bash
go test -v ./examples/08_vinfo/...
```

---

## Screenshots

### Lock Screen
![V-Info Lock Screen](vinfo_lock_preview.png)

### Tasks & To-Dos
![V-Info Tasks](vinfo_preview.png)

### Information Vault
![V-Info Vault](vinfo_vault_preview.png)

### Dashboard & Analytics
![V-Info Dashboard](vinfo_dashboard_preview.png)

### Security & Passcode Settings
![V-Info Security](vinfo_security_preview.png)
