package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type User struct {
	FullName  string `json:"full_name"`
	EmailAddr string `json:"email_addr"`
}

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer db.Close()

	// Initial setup: ensure v1 table exists
	//db.Exec("CREATE TABLE IF NOT EXISTS users (full_name TEXT, email_addr TEXT)")

	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// v1 ONLY supports searching by full_name
			fullName := r.URL.Query().Get("full_name")
			if fullName == "" {
				http.Error(w, "v1 requires full_name parameter", 400)
				return
			}

			var email string
			err := db.QueryRow("SELECT email_addr FROM users WHERE full_name = $1", fullName).Scan(&email)
			if err != nil {
				http.Error(w, "User not found", 404)
				return
			}
			fmt.Fprintf(w, "%s\n", email)

		case http.MethodPost:
			var u User
			json.NewDecoder(r.Body).Decode(&u)

			// v1 ONLY inserts into the legacy full_name column
			_, err = db.Exec("INSERT INTO users (full_name, email_addr) VALUES ($1, $2)", u.FullName, u.EmailAddr)
			if err != nil {
				http.Error(w, "Insert failed", 500)
				return
			}
			w.WriteHeader(201)
		}
	})

	http.HandleFunc("/clean", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST to clean", 405)
			return
		}
		db.Exec("DROP TABLE IF EXISTS users")
		db.Exec("CREATE TABLE users (full_name TEXT, email_addr TEXT)")
		fmt.Fprintln(w, "v1 Database Reset: All v2 columns and data removed.")
	})

	log.Println("App v1 (Legacy) starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}