package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// User represents the data structure for App v1
type User struct {
	FullName  string `json:"full_name"`
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

	// HTTP handler for /user
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

			var err error
			
			// Case: get full_name = "Anastasiia Gubska"
			err = db.QueryRow("SELECT full_name FROM users WHERE full_name = $1", fullName).Scan(&fullName)
			if err != nil {
				http.Error(w, "User not found", 404)
				return
			}
			w.WriteHeader(200)
			fmt.Fprintf(w, "%s\n", fullName)
		
		case http.MethodPost:
			var u User
			json.NewDecoder(r.Body).Decode(&u)

			// Check: v1 need an input full_name
			if u.FullName == "" {
				http.Error(w, "full_name is required", 400)
				return
			}

			// v1 ONLY inserts into the legacy full_name column
			_, err = db.Exec("INSERT INTO users (full_name) VALUES ($1)", u.FullName)
			if err != nil {
				http.Error(w, "Insert failed", 500)
				return
			}
			w.WriteHeader(201)
		}
	})

	log.Println("App v1 (Legacy) starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}