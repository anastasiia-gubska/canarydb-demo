package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// User represents the data structure for App v1
type User struct {
	FullName  string `json:"full_name"`
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
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			name := r.URL.Query().Get("name")
			if name != "" {
				var email string
				db.QueryRow("SELECT email_addr FROM users WHERE full_name = $1", name).Scan(&email)
				json.NewEncoder(w).Encode(map[string]string{"email": email})
			} else {
				rows, _ := db.Query("SELECT full_name, email_addr FROM users")
				var users []User
				for rows.Next() {
					var u User
					rows.Scan(&u.FullName, &u.EmailAddr)
					users = append(users, u)
				}
				json.NewEncoder(w).Encode(users)
			}
		case http.MethodPost:
			var u User
			json.NewDecoder(r.Body).Decode(&u)
			db.Exec("INSERT INTO users (full_name, email_addr) VALUES ($1, $2)", u.FullName, u.EmailAddr)
			w.WriteHeader(201)
		}
	})

	log.Println("App v1 starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
