package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

const (
	dbCreationQuery   = "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)"
	dbGetAllQuery     = "SELECT * FROM users"
	dbCreateUserQuery = "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	//connect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("error connecting db", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	//create the table if it doesn't exist
	_, err = db.Exec(dbCreationQuery)
	if err != nil {
		log.Fatal("error creating db", err)
	}

	//create router
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers(db)).Methods("GET")
	router.HandleFunc("/users", createUser(db)).Methods("POST")

	//start server
	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router)))
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all users
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(dbGetAllQuery)
		if err != nil {
			log.Fatal(err)
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {

			}
		}(rows)

		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
				log.Fatal(err)
			}
			users = append(users, u)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		encodeErr := json.NewEncoder(w).Encode(users)
		if encodeErr != nil {
			return
		}
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var u User
		err := json.NewDecoder(request.Body).Decode(&u)
		if err != nil {
			log.Fatal("unable to parse body", err)
			return
		}
		insertErr := db.QueryRow(dbCreateUserQuery, u.Name, u.Email).Scan(&u.ID)
		if insertErr != nil {
			log.Fatal(insertErr)
		}

		encodeErr := json.NewEncoder(w).Encode(u)
		if encodeErr != nil {
			return
		}
	}
}
