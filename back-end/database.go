package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int    `json:"id"`
	Nickname  string `json:"nickname"`
	Age       int    `json:"age,string"`
	Gender    string `json:"gender"`
	FirstName string `json:"first-name"`
	LastName  string `json:"last-name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	LastMessage   string `json:"last-message"`
	LastMessageSentAt time.Time `json:"last-message-sent-at"`
	Online    bool `json:"online"`
}

type Credentials struct {
	Identifier string `json:"login-identifier"`
	Password   string `json:"login-password"`
}

type Post struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Category string `json:"category"`
	Content  string `json:"content"`
	Nickname string `json:"nickname"`
}

type Comment struct {
	ID      int	  	`json:"id"`
	UserID   int    `json:"user_id"`
	PostID  int		`json:"post_id"`
	Content string	`json:"content"`
	Nickname string `json:"nickname"`
}

type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sent_at"` 
	Nickname   string `json:"nickname"`
}

var db *sql.DB
var err error

func initializeDatabase() (*sql.DB, error) {
	db, err = sql.Open("sqlite3", "forum.db")
	if err != nil {
		return nil, err
	}

	// Create Users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nickname TEXT UNIQUE,
		age INTEGER,
		gender TEXT,
		first_name TEXT,
		last_name TEXT,
		email TEXT UNIQUE,
		password TEXT
	);
	`
	_, err = db.Exec(createUsersTable)
	if err != nil {
		return nil, fmt.Errorf("error creating users table: %v", err)
	}

	// Create Posts table
	createPostsTable := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		category TEXT,
		content TEXT,
		FOREIGN KEY (user_id) REFERENCES users (id)
	);
	`
	_, err = db.Exec(createPostsTable)
	if err != nil {
		return nil, fmt.Errorf("error creating posts table: %v", err)
	}

	// Create Posts table
	createCommentsTable := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		post_id INTEGER,
		content TEXT,
		FOREIGN KEY (user_id) REFERENCES users (id)
		FOREIGN KEY (post_id) REFERENCES users (id)
	);
	`
	_, err = db.Exec(createCommentsTable)
	if err != nil {
		return nil, fmt.Errorf("error creating comments table: %v", err)
	}

	// Create Messages table
	createMessagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER,
		receiver_id INTEGER,
		content TEXT,
		sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (sender_id) REFERENCES users (id),
		FOREIGN KEY (receiver_id) REFERENCES users (id)
	);
	`
	_, err = db.Exec(createMessagesTable)
	if err != nil {
		return nil, fmt.Errorf("error creating messages table: %v", err)
	}

	return db, nil
}
func sendPostToClient(postData map[string]interface{}, userID int, db *sql.DB) {
    // Fetching the user's nickname from the database.
    var nickname string
    err := db.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&nickname)
    if err != nil {
        log.Println("Error fetching user nickname:", err)
        
    }

    postData["sender_nickname"] = nickname

}

// Add a new user to the database
func addUser(db *sql.DB, user User) error {
	log.Printf("Received addUser data: %+v", user)
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		return err
	}

	// Insert into the database
	log.Printf("Inserting data into DB: %+v", user)

	_, err = db.Exec(`INSERT INTO users (nickname, age, gender, first_name, last_name, email, password) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.Nickname, user.Age, user.Gender, user.FirstName, user.LastName, user.Email, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err)
	}

	return nil
}

// Find a user by identifier (either nickname or email)
func findUserByIdentifier(db *sql.DB, identifier string) (*User, error) {
	var user User
	row := db.QueryRow(`SELECT * FROM users WHERE nickname = ? OR email = ?`, identifier, identifier)
	err := row.Scan(&user.ID, &user.Nickname, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Add a new post to the database
func addPost(db *sql.DB, post Post) (int, error) {
	result, err := db.Exec(`INSERT INTO posts (user_id, category, content) VALUES (?, ?, ?)`,
		post.UserID, post.Category, post.Content)
	if err != nil {
		return 0, err
	}
	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(postID), nil
}

func addComment(db *sql.DB, comment Comment) error {
	_, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)",
		comment.PostID, comment.UserID, comment.Content)
	if err != nil {
		return err
	}
	return nil
}

// Fetch all posts from the database
func getPostList(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`SELECT posts.id, posts.user_id, posts.category, posts.content, u.nickname FROM posts LEFT JOIN users u ON posts.user_id = u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Category, &post.Content, &post.Nickname); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// Add a new message to the database
func addMessage(db *sql.DB, message Message) error {
	_, err := db.Exec(`INSERT INTO messages (sender_id, receiver_id, content, sent_at) VALUES (?, ?, ?, ?)`,
		message.SenderID, message.ReceiverID, message.Content, time.Now()) 
	if err != nil {
		return err
	}
	return nil
}

// Fetch all registered users from the database
func getAllUsers() ([]User, error) {
    // Initialize an empty slice to store users
    var users []User

    // Query the database
    rows, err := db.Query("SELECT * FROM users ORDER BY nickname COLLATE NOCASE")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Iterate through the result rows and append to the users slice
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Nickname, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.LastName, &user.Password); err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    // Check for errors from iterating over rows
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}

// Fetch all comments by post_id from the database
func getCommentList(db *sql.DB, postID int) ([]Comment, error) {
	rows, err := db.Query("SELECT comments.id, post_id, user_id, content, u.nickname FROM comments LEFT JOIN users u ON comments.user_id = u.id WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Nickname); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}


func getUserByID(db *sql.DB, userID int) (User, error) {
	fmt.Println("user id :", userID)
	var user User
	row := db.QueryRow(`SELECT * FROM users WHERE id = ?`, userID)
	err := row.Scan(&user.ID, &user.Nickname, &user.Age, &user.Gender, &user.FirstName, &user.LastName, &user.Email, &user.Password)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func getChatHistory(db *sql.DB, senderID, receiverID int, limit, offset int) ([]Message, error) {
	rows, err := db.Query("SELECT sent_at, sender_id, receiver_id, content, users.nickname FROM messages LEFT JOIN users ON messages.sender_id = users.id  WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) ORDER BY sent_at DESC LIMIT ? OFFSET ?", senderID, receiverID, receiverID, senderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chatHistory []Message
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.SentAt, &message.SenderID, &message.ReceiverID, &message.Content, &message.Nickname); err != nil {
			return nil, err
		}
		//chatHistory = append(chatHistory, message)
		chatHistory = append([]Message{message}, chatHistory...)
	}

	return chatHistory, nil
}

func getLastMessage(db *sql.DB, senderID, receiverID int) (Message, error) {
	var message Message
	err := db.QueryRow("SELECT sent_at FROM messages WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) ORDER BY sent_at DESC LIMIT 1",
		senderID, receiverID, receiverID, senderID).Scan(&message.SentAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle case when no message is found
			return message, nil
		}
		return message, err
	}
	return message, nil
}