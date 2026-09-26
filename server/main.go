// scorecard-server is the hosted backend for the scorecard app: it serves
// the static UI and persists the same JSON document the github.io version
// keeps in localStorage, but in SQLite, so history survives across devices.
//
// There's one row: a single shared scoreboard, matching what the static
// version already gives one device. No auth — this sits behind
// scores.app.hyde.ca, which resolves only on the tailnet.
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS store (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	data TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`

func main() {
	dbPath := env("DB_PATH", "/data/scores.db")
	webDir := env("WEB_DIR", "./web")
	addr := ":" + env("PORT", "8080")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	// SQLite has one writer at a time; capping the pool avoids concurrent
	// writers tripping "database is locked" under the driver instead of
	// queuing cleanly.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "db unreachable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/db", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(db, w, r)
		case http.MethodPut:
			handlePut(db, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	log.Printf("scorecard-server listening on %s (db=%s web=%s)", addr, dbPath, webDir)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func handleGet(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var data, updatedAt string
	err := db.QueryRow(`SELECT data, updated_at FROM store WHERE id = 1`).Scan(&data, &updatedAt)
	w.Header().Set("Content-Type", "application/json")
	switch {
	case errors.Is(err, sql.ErrNoRows):
		w.Write([]byte(`{"data":null,"updatedAt":null}`))
	case err != nil:
		http.Error(w, "read failed", http.StatusInternalServerError)
	default:
		// data is already-valid JSON text from a prior handlePut, so it's
		// safe to splice directly rather than round-tripping through Go
		// structs.
		w.Write([]byte(`{"data":` + data + `,"updatedAt":` + jsonString(updatedAt) + `}`))
	}
}

func handlePut(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 2<<20)).Decode(&body); err != nil || len(body.Data) == 0 {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if !json.Valid(body.Data) {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`
		INSERT INTO store (id, data, updated_at) VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET data = excluded.data, updated_at = excluded.updated_at`,
		string(body.Data), now)
	if err != nil {
		http.Error(w, "write failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"updatedAt":` + jsonString(now) + `}`))
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
