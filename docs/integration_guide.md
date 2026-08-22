# Nova Integration Guide

This guide explains how to integrate Nova with backend services, databases, external Go packages, background goroutines, and operating system integrations.

---

## Table of Contents

1. [Integrating with Go Backends & Microservices](#integrating-with-go-backends--microservices)
2. [Database Integration (SQL, PostgreSQL, SQLite, GORM)](#database-integration-sql-postgresql-sqlite-gorm)
3. [Background Workers & Thread-Safe UI Updates](#background-workers--thread-safe-ui-updates)
4. [Forms & Data Validation Integration](#forms--data-validation-integration)
5. [Integrating REST & gRPC Clients](#integrating-rest--grpc-clients)
6. [Asset & Font Bundling (`embed.FS`)](#asset--font-bundling-embedfs)
7. [Packaging & Distribution](#packaging--distribution)

---

## Integrating with Go Backends & Microservices

Because Nova is **100% native Go**, there are no IPC serialization bridges (like Tauri/Electron IPC). Your UI code shares direct memory pointers with your backend business logic, database connectors, and network clients.

---

## Database Integration (SQL, PostgreSQL, SQLite, GORM)

### Connecting Nova to Database Queries

```go
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/widgets"
)

type User struct {
	ID    int
	Email string
	Role  string
}

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/mydb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	users := state.New([]User{})
	isLoading := state.Bool(false)

	loadUsers := func() {
		isLoading.Set(true)
		go func() {
			rows, _ := db.Query("SELECT id, email, role FROM users LIMIT 1000")
			defer rows.Close()

			var list []User
			for rows.Next() {
				var u User
				rows.Scan(&u.ID, &u.Email, &u.Role)
				list = append(list, u)
			}
			users.Set(list)
			isLoading.Set(false)
		}()
	}

	app := nova.New()
	win := app.Window(nova.Title("Database Viewer"), nova.Size(800, 600))

	// Bind virtualized table to loaded database records
	win.Content(
		widgets.Table([]widgets.TableColumn{
			{Title: "ID", Width: 60},
			{Title: "Email", Width: 200},
			{Title: "Role", Width: 120},
		}, len(users.Get()), func(row, col int) string {
			u := users.Get()[row]
			switch col {
			case 0:
				return fmt.Sprintf("#%d", u.ID)
			case 1:
				return u.Email
			case 2:
				return u.Role
			default:
				return ""
			}
		}),
	)

	app.Run()
}
```

---

## Background Workers & Thread-Safe UI Updates

Nova's reactive signals (`state.Value[T]`) are **fully thread-safe**. You can mutate state from any background goroutine, and registered UI listeners are notified seamlessly.

```go
progress := state.Float(0.0)
statusText := state.String("Idle")

go func() {
    statusText.Set("Processing large dataset...")
    for i := 1; i <= 100; i++ {
        time.Sleep(20 * time.Millisecond)
        progress.Set(float64(i) / 100.0)
    }
    statusText.Set("Complete!")
}()
```

---

## Forms & Data Validation Integration

Use `widgets/forms` with custom schema validators and model binding:

```go
formState := forms.NewFormState()
formState.RegisterField("email", forms.Required(), forms.Email())
formState.RegisterField("password", forms.Required(), forms.MinLength(8))

formState.OnSubmit = func(data map[string]any) {
    // Send to backend service or auth provider
    authService.Login(data["email"].(string), data["password"].(string))
}
```

---

## Integrating REST & gRPC Clients

```go
func fetchMetrics(metricsSignal *state.Value[Metrics]) {
    go func() {
        resp, err := http.Get("https://api.internal/metrics")
        if err == nil {
            defer resp.Body.Close()
            var m Metrics
            json.NewDecoder(resp.Body).Decode(&m)
            metricsSignal.Set(m)
        }
    }()
}
```

---

## Asset & Font Bundling (`embed.FS`)

Leverage standard Go `embed` to bundle application assets directly into the standalone binary:

```go
//go:embed assets/*
var assetsFS embed.FS
```

---

## Packaging & Distribution

```bash
# Build standalone binary with optimizations
go build -ldflags="-s -w" -o myapp .

# Using Nova CLI
nova build
```
