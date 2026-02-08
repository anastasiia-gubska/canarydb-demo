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

// User represents the data structure for App v1
type User struct {
	FullName  string `json:"full_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	EmailAddr string `json:"email_addr"`
}

func main() {
	// 1. Get database connection string from Environment Variables
	// If running locally, you'll use: postgres://postgres:postgres@localhost:5433/testdb?sslmode=disable
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
	}

	// 2. Open connection to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Could not connect to DB:", err)
	}
	defer db.Close()

	// This prevents the panic. It checks if the DB is actually reachable.
	if err = db.Ping(); err != nil {
		log.Fatal("Cannot reach database. Check if Postgres is running:", err)
	}

	// 3. Define the HTTP handler for /users
http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// GET - return full_name and email;
		case http.MethodGet:
			query := r.URL.Query()
			fullName := query.Get("full_name")
			firstName := query.Get("first_name")
			lastName := query.Get("last_name")

			var email string
			var err error

			if fullName != "" {
				// Case: get full_name = "Anastasiia Gubska"
        		err = db.QueryRow("SELECT email_addr FROM users WHERE full_name = $1", fullName).Scan(&email)
			}else if firstName != "" && lastName != ""{
				// Case: get first_name = "Anastasiia" AND last_name = "Gubska"
				err = db.QueryRow("SELECT email_addr FROM users WHERE first_name = $1 AND last_name = $2", firstName, lastName).Scan(&email)
			}else {
        		http.Error(w, "Missing name parameters", 400)
        		return
    		}
			if err != nil {
        		http.Error(w, "User not found", 404)
        		return
    		}
			fmt.Fprintf(w, "%s\n", email)
		case http.MethodPost:
			var u User
			json.NewDecoder(r.Body).Decode(&u)
			// Check if we have EVERYTHING (v1 + v2 data)
			if u.FullName != "" && u.FirstName != "" && u.LastName != "" {
				// Upgrade schema if needed
				db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name TEXT, ADD COLUMN IF NOT EXISTS last_name TEXT")
				
				// Insert into all columns
				_, err = db.Exec("INSERT INTO users (full_name, first_name, last_name, email_addr) VALUES ($1, $2, $3, $4)", 
					u.FullName, u.FirstName, u.LastName, u.EmailAddr)

			// Check if we have only v2 data
			} else if u.FirstName != "" && u.LastName != "" {
				db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name TEXT, ADD COLUMN IF NOT EXISTS last_name TEXT")
				_, err = db.Exec("INSERT INTO users (first_name, last_name, email_addr) VALUES ($1, $2, $3)", 
					u.FirstName, u.LastName, u.EmailAddr)

			// Fallback to v1 data
			} else {
				_, err = db.Exec("INSERT INTO users (full_name, email_addr) VALUES ($1, $2)", 
					u.FullName, u.EmailAddr)
			}

			if err != nil {
				log.Println("Insert Error:", err)
				http.Error(w, "Database insertion failed", 500)
				return
			}
			w.WriteHeader(201)
		}
	})

	// 3. Define the HTTP handler for /clean (Reset the DB for the demo)
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

	log.Println("App v1 starting on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
