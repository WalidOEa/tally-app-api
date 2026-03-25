package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	count int
	mu    sync.Mutex
	port  = 5050
	db    *sql.DB
)

func tallyHandle(w http.ResponseWriter, r *http.Request) {
	var err error

	mu.Lock()
	count++
	curr := count
	slog.Info("Tally incremented",
		"curr_count", curr,
		"remote_addr", r.RemoteAddr,
	)

	_, err = db.Exec("INSERT INTO history (total, created_at, ip_address) VALUES (?, ?, ?)",
		curr, time.Now(), r.RemoteAddr)
	if err != nil {
		slog.Error("Database insert failed", "error", err)
	} else {
		slog.Info("Database sync complete", "saved_total", curr)
	}
	mu.Unlock()

	fmt.Fprintf(w, "%d", curr)
}

func currHandle(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	curr := count
	slog.Info("Fetched current count",
		"curr_count", curr,
		"remote_addr", r.RemoteAddr,
	)
	mu.Unlock()

	fmt.Fprintf(w, "%d", curr)
}

func main() {
	var err error
	db, err = sql.Open("sqlite", "tally.db")
	if err != nil {
		slog.Error("Cannot open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	total INTEGER,
	created_at DATETIME,
	ip_address TEXT
	)`)
	if err != nil {
		slog.Error("Cannot create database", "error", err)
		os.Exit(1)
	} else {
		slog.Info("Created database (if not exists)")
	}

	err = db.QueryRow("SELECT total FROM history ORDER BY id DESC LIMIT 1").Scan(&count)
	if err != nil {
		slog.Info("No history found, starting anew")
	} else {
		slog.Info("Found history")
	}

	http.HandleFunc("/tally", tallyHandle)
	http.HandleFunc("/curr", currHandle)
	slog.Info("Server-",
		"port", port,
		"status", "ready",
	)

	err = (http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	if err != nil {
		slog.Error("Fatal", "error", err)
		os.Exit(1)
	}
}
