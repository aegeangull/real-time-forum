package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var tokenStore = make(map[string]string)

type LoginResponse struct {
    Token  string `json:"token"`
    UserID int    `json:"user_id"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err = addUser(db, user)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("User registered successfully"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Fetch user from the database
	dbUser, err := findUserByIdentifier(db, credentials.Identifier)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Login failed", http.StatusUnauthorized)
		return
	}

	token := uuid.New().String()
	tokenStore[token] = dbUser.Nickname // Store the token associated with the username
	w.Header().Set("Content-Type", "application/json")
	response := LoginResponse{
		Token:  token,
		UserID: dbUser.ID,
	}
	json.NewEncoder(w).Encode(response)
}



func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")

		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Initialize the database
	_, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	// Register WebSocket handler
	http.HandleFunc("/ws", enableCORS(messageHandler))
	http.HandleFunc("/", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "This is API!")
	}))
	http.HandleFunc("/get-users", enableCORS(getUsersHandler))
	http.HandleFunc("/login", enableCORS(http.HandlerFunc(loginHandler)))
	http.HandleFunc("/protected", enableCORS(protectedHandler))
	http.HandleFunc("/register", enableCORS(registerHandler))
	// http.Handle("/getMessages/", enableCORS(http.HandlerFunc(getMessagesHandler))) //messageHandler

	// http.HandleFunc("/getPosts", getPostsHandler)


	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getUserFromDB(nickname string) (*User, error) {
	var user User
	query := "SELECT nickname, password FROM users WHERE nickname = ?"
	err := db.QueryRow(query, nickname).Scan(&user.Nickname, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
    // Check the token in the request header
    token := r.Header.Get("Authorization")
    if token == "" {
        http.Error(w, "No token provided", http.StatusUnauthorized)
        return
    }

    // Validate the token
    _, valid := ValidateToken(token)
    if !valid {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    w.Write([]byte("Protected content"))
}

// Find a user strictly by their ID
func findUserByID(db *sql.DB, userID int) (*User, error) {
	var user User
	row := db.QueryRow(`SELECT * FROM users WHERE id = ?`, userID)
	err := row.Scan(&user.ID, &user.Nickname, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
