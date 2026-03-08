package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// User represents the data structure for App v2
type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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
		// GET - read from new columns (first_name and last_name) 
		case http.MethodGet:
			query := r.URL.Query()
			firstName := query.Get("first_name")
			lastName := query.Get("last_name")
			
			if firstName == "" || lastName == "" {
				http.Error(w, "Missing first_name + last_name", 400)
				return
			}

			var err error
			// v3: Read new columns first_name and last_name
			err = db.QueryRow("SELECT first_name, last_name FROM users WHERE first_name = $1 AND last_name = $2", firstName, lastName).Scan(&firstName, &lastName)
			if err != nil {
				http.Error(w, "User not found", 404)
				return
			}
			w.WriteHeader(200)
			fmt.Fprintf(w, "%s %s\n", firstName, lastName)
		
		// POST - write to new columns
		case http.MethodPost:
			var u User
			json.NewDecoder(r.Body).Decode(&u)

			// Check: v3 need both inputs first_name + last_name
			if u.FirstName == "" || u.LastName == "" {
				http.Error(w, "first_name and last_name are required", 400)
				return
			}
			// v3: Only writes to the new schema
			_, err = db.Exec(
				"INSERT INTO users (first_name, last_name) VALUES ($1, $2)",
				u.FirstName, u.LastName)

			if err != nil {
				log.Println("Insert Error:", err)
				http.Error(w, "Database insertion failed", 500)
				return
			}
			w.WriteHeader(201)
		}
	})


	log.Println("App v3 starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
