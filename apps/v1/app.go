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

// User represents the data structure for App v1
type User struct {
	FullName  string `json:"full_name"`
	EmailAddr string `json:"email_addr"`
}

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
	}

	// Open connection to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Could not connect to DB:", err)
	}
	defer db.Close()

	// To prevent the panic. Check if the DB is reachable.
	if err = db.Ping(); err != nil {
		log.Fatal("Cannot reach database. Check if Postgres is running:", err)
	}

	// HTTP handler for /users
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// GET - return full_name and email;
		case http.MethodGet:

			fullName := r.URL.Query().Get("full_name")

			// v1 ONLY supports searching by full_name
			if fullName == "" {
				http.Error(w, "v1 requires full_name parameter", 400)
				return
			}

			var email string
			var err error
			
			// Case: get full_name = "Anastasiia Gubska"
			err = db.QueryRow("SELECT email_addr FROM users WHERE full_name = $1", fullName).Scan(&email)
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

	// HTTP handler for /clean (Reset the DB for the demo)
	http.HandleFunc("/clean", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST to clean", 405)
			return
		}
		// Drop the table and recreate it in the original v1 state
		_, err := db.Exec("DROP TABLE IF EXISTS users")
		if err != nil {
			http.Error(w, "Failed to drop table", 500)
			return
		}

		_, err = db.Exec("CREATE TABLE users (full_name TEXT, email_addr TEXT)")
		if err != nil {
			http.Error(w, "Failed to recreate table", 500)
			return
		}

		fmt.Fprintln(w, "Database cleaned. Back to v1 state and empty (only full_name and email_addr).")
	})

	log.Println("App v1 (Legacy) starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}