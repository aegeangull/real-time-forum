package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)
import "sort"
// var clients = make(map[int]*websocket.Conn)  // map user ID to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var (
	clients = make(map[int]*websocket.Conn)
	mu      sync.Mutex // Protects access to clients.
)

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := getAllUsers()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error during WebSocket handshake:", err)
		return
	}
	defer conn.Close()
	var senderID int // Store senderID to remove it from clients on disconnect.

	// Infinite loop to continuously listen to messages.
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading WebSocket message:", err)
			// Remove the user from clients map upon disconnect.
			mu.Lock()
			delete(clients, senderID)
			mu.Unlock()
			return
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(p, &msg); err != nil {
			fmt.Println("Error unmarshalling WebSocket message:", err)
			continue
		}

		switch msg["type"].(string) {
		case "new_message":
			payload := msg["data"].(map[string]interface{})
			fmt.Println("Received data:", payload)

			senderIDFloat, ok := payload["sender_id"].(float64)
			if !ok {
				fmt.Println("Error: sender_id is not valid or missing")
				continue
			}
			senderID := int(senderIDFloat)

			// Store the sender's WebSocket
			mu.Lock()
			clients[senderID] = conn
			mu.Unlock()

			receiverIDFloat, ok := payload["receiver_id"].(float64)
			if !ok {
				fmt.Println("Error: receiver_id is not valid or missing")
				continue
			}
			receiverID := int(receiverIDFloat)

			content := payload["content"].(string)

			message := Message{
				SenderID:   senderID,
				ReceiverID: receiverID,
				Content:    content,
			}

			if err := addMessage(db, message); err != nil {
				fmt.Println("Error saving message to the database:", err)
			}

			sender, err := findUserByID(db, senderID)
			if err != nil {
				fmt.Println("Error fetching sender:", err)
				continue
			}

			enrichedPayload := map[string]interface{}{
				"sender_nickname": sender.Nickname,
				"content":         content,
				"sent_at":       time.Now().Format("2006-01-02 15:04:05"),
				"sender_id":       senderID,
				"receiver_id":     receiverID,
			}

			fmt.Println("Enriched payload:", enrichedPayload)
			response := map[string]interface{}{
				"type":    "new_message",
				"data": enrichedPayload,
			}

			responseBytes, err := json.Marshal(response)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}

			// Send the message only to the sender and receiver.
			mu.Lock()
			for _, id := range []int{senderID, receiverID} {
				if clientConn, ok := clients[id]; ok {
					if err := clientConn.WriteMessage(messageType, responseBytes); err != nil {
						fmt.Printf("Error sending message to client %d: %v\n", id, err)
					}
				}
			}
			mu.Unlock()

		case "create_post":
			// Extract data payload
			payload, ok := msg["data"].(map[string]interface{})
			if !ok {
				fmt.Println("Error: payload is not valid or missing")
				continue
			}
		
			// Validate and extract post details from payload
			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)
		
			category, ok := payload["category"].(string)
			if !ok {
				fmt.Println("Error: category is not valid or missing")
				continue
			}
		
			content, ok := payload["content"].(string)
			if !ok {
				fmt.Println("Error: content is not valid or missing")
				continue
			}
		
			// Create a Post struct instance
			post := Post{
				UserID:   userID,
				Category: category,
				Content:  content,
			}
		
			postID, err := addPost(db, post)
			if err != nil {
				fmt.Println("Error saving post to the database:", err)
			} else {
				fmt.Println("New post ID:", postID)
			}
		
			// Fetching the user's nickname from the database.
			var nickname string
			_err := db.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&nickname)
			if _err != nil {
				fmt.Println("Error fetching user nickname:", _err)
				// Handle error appropriately.
			}
		
			// Send the post to all connected clients
			enrichedPayload := map[string]interface{}{
				"id":  postID,
				"user_id":   userID,
				"category":  category,
				"content":   content,
				"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
				"nickname":  nickname,  // Add the fetched nickname here
			}
		
			fmt.Println("Enriched payload:", enrichedPayload)
			response := map[string]interface{}{
				"type": "new_post",
				"data": enrichedPayload,
			}
		
			responseBytes, err := json.Marshal(response)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}
		
			// Send the message to all connected clients.
			mu.Lock()
			clients[userID] = conn
			for _, clientConn := range clients {
				if err := clientConn.WriteMessage(messageType, responseBytes); err != nil {
					fmt.Println("Error sending message to client:", err)
				}
			}
			fmt.Printf("Message sent to %d clients.\n", len(clients))
			mu.Unlock()
		case "get_post_list":
			// Extract data payload
			payload, ok := msg["data"].(map[string]interface{})
			if !ok {
				fmt.Println("Error: payload is not valid or missing")
				continue
			}
		
			// Validate and extract post details from payload
			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)
			
			// Retrieve the list of posts from the database
			posts, err := getPostList(db)
			if err != nil {
				fmt.Println("Error retrieving post list:", err)
				continue
			}

			// Prepare the response payload
			responsePayload := map[string]interface{}{
				"type": "post_list",
				"data": posts,
			}

			// Convert the payload to JSON
			responseBytes, err := json.Marshal(responsePayload)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}
			// Send the message to all connected clients.
			mu.Lock()
			clients[userID] = conn
			
			for _, clientConn := range clients {
				if err := clientConn.WriteMessage(messageType, responseBytes); err != nil {
					fmt.Println("Error sending message to client:", err)
				}
			}
			fmt.Printf("Message sent to %d clients.\n", len(clients))
			mu.Unlock()
		case "get_comment_list":
			// Extract data payload
			payload, ok := msg["data"].(map[string]interface{})
			if !ok {
				fmt.Println("Error: payload is not valid or missing")
				continue
			}
			// Validate and extract post details from payload
			postIDFloat, ok := payload["post_id"].(float64)
			if !ok {
				fmt.Println("Error: post_id is not valid or missing")
				continue
			}
			postID := int(postIDFloat)

			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)
			// Retrieve the list of comments for the given post from the database
			comments, err := getCommentList(db, postID)
			if err != nil {
				fmt.Println("Error retrieving comment list:", err)
				continue
			}
			// Prepare the response payload
			responsePayload := map[string]interface{}{
				"type": "comment_list",
				"data": comments,
			}
			// Convert the payload to JSON
			responseBytes, err := json.Marshal(responsePayload)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}
			// Send the message to all connected clients.
			mu.Lock()
			clients[userID] = conn
			for _, clientConn := range clients {
				if err := clientConn.WriteMessage(messageType, responseBytes); err != nil {
					fmt.Println("Error sending message to client:", err)
				}
			}
			fmt.Printf("Message sent to %d clients.\n", len(clients))
			mu.Unlock()
		case "create_comment":
			// Extract data payload
			payload, ok := msg["data"].(map[string]interface{})
			if !ok {
				fmt.Println("Error: payload is not valid or missing")
				continue
			}
			// Validate and extract comment details from payload
			postIDFloat, ok := payload["post_id"].(float64)
			if !ok {
				fmt.Println("Error: post_id is not valid or missing")
				continue
			}
			postID := int(postIDFloat)
			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)
			content, ok := payload["content"].(string)
			if !ok {
				fmt.Println("Error: content is not valid or missing")
				continue
			}
			// Create a Comment struct instance
			comment := Comment{
				PostID:   postID,
				UserID:   userID,
				Content:  content,
			}
			// Save the comment to the database
			if err := addComment(db, comment); err != nil {
				fmt.Println("Error saving comment to the database:", err)
			}
			// Fetching the user's nickname from the database.
			var nickname string
			err := db.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&nickname)
			if err != nil {
				fmt.Println("Error fetching user nickname:", err)
				// Handle error appropriately.
			}
			// Send the comment to all connected clients
			enrichedPayload := map[string]interface{}{
				"post_id":   postID,
				"user_id":   userID,
				"content":   content,
				//"timestamp": comment.Timestamp,
				"nickname":  nickname,  // Add the fetched nickname here
			}
			fmt.Println("Enriched payload:", enrichedPayload)
			response := map[string]interface{}{
				"type": "new_comment",
				"data": enrichedPayload,
			}
			responseBytes, err := json.Marshal(response)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}
			// Send the message to all connected clients.
			mu.Lock()
			clients[userID] = conn
			for _, clientConn := range clients {
				if err := clientConn.WriteMessage(messageType, responseBytes); err != nil {
					fmt.Println("Error sending message to client:", err)
				}
			}
			fmt.Printf("Message sent to %d clients.\n", len(clients))
			mu.Unlock()
		case "get_online_users":
			payload, ok := msg["data"].(map[string]interface{})
			if !ok || payload == nil {
				fmt.Println("Error: data is not valid or missing")
				continue
			}
			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)
			fmt.Println("get_online_users:",userIDFloat,clients)
			// Send online users to the user who requested
			users := make([]User, 0, len(clients))
			mu.Lock()
			for id := range clients {
				if id == userID {
					continue // Skip the current user
				}
				user, err := getUserByID(db, id)
				if err != nil {
					fmt.Println("Error fetching user:", err)
					continue
				}
				users = append(users, user)
			}
			mu.Unlock()
			fmt.Println("online user list:", users)
			response := map[string]interface{}{
				"type": "online_users",
				"data": users,
			}
			responseBytes, err := json.Marshal(response)
			fmt.Println("Sending message to clients:", string(responseBytes))
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}
			err = conn.WriteMessage(messageType, responseBytes)
			if err != nil {
				fmt.Println("Error sending online users:", err)
			}
		
		case "get_online_users_sort":
			payload, ok := msg["data"].(map[string]interface{})
			if !ok || payload == nil {
				fmt.Println("Error: data is not valid or missing")
				continue
			}
			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)
			// Retrieve the last message sent to you by each online user
			lastMessages := make(map[int]Message) // Map to store the last message for each us

			allUsers, err := getAllUsers()
			if err != nil {
				http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
				return
			}
			sortedUsers := make([]User, 0, len(allUsers))
			// Send online users to the user who requested
			mu.Lock()
			for id, userObject := range allUsers {
				fmt.Println("User:", userObject, userObject.ID)
				if userObject.ID == userID {
					continue // Skip the current user
				}
				lastMessage, err := getLastMessage(db, userID, userObject.ID)
				if err != nil {
					fmt.Println("Error fetching last message:", err)
					continue
				}
				lastMessages[id] = lastMessage
				user, err := getUserByID(db, userObject.ID)
				if err != nil {
					fmt.Println("Error fetching user:", err)
					continue
				}
				_, exists := clients[userObject.ID]
				if exists {
					// Client exists, you can perform further actions with the conn variable
					user.Online = true;
				} else {
					// Client does not exist
					user.Online = false;
				}
				user.LastMessageSentAt = lastMessage.SentAt
				sortedUsers = append(sortedUsers, user)
			}
			mu.Unlock()

			// Sort alphabetically if no message data available
			sort.Slice(sortedUsers, func(i, j int) bool {
				// Sort by last message sent
				if !sortedUsers[i].LastMessageSentAt.Equal(sortedUsers[j].LastMessageSentAt) {
					return sortedUsers[i].LastMessageSentAt.After(sortedUsers[j].LastMessageSentAt)
				}
				return false
			})
			
			response := map[string]interface{}{
				"type": "online_users",
				"data": sortedUsers,
			}
			responseBytes, err := json.Marshal(response)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}
			err = conn.WriteMessage(messageType, responseBytes)
			if err != nil {
				fmt.Println("Error sending online users:", err)
			}
		case "get_chat_list":
			// Extract data payload
			payload, ok := msg["data"].(map[string]interface{})
			if !ok {
				fmt.Println("Error: payload is not valid or missing")
				continue
			}
			// Validate and extract chat details from payload
			senderIDFloat, ok := payload["sender_id"].(float64)
			if !ok {
				fmt.Println("Error: sender_id is not valid or missing")
				continue
			}
			senderID := int(senderIDFloat)
			receiverIDFloat, ok := payload["receiver_id"].(float64)
			if !ok {
				fmt.Println("Error: receiver_id is not valid or missing")
				continue
			}
			receiverID := int(receiverIDFloat)

			offsetFloat, ok := payload["offset"].(float64)
			if !ok {
				fmt.Println("Error: receiver_id is not valid or missing")
				continue
			}
			offset := int(offsetFloat)

			// Retrieve the chat history between the sender and receiver from the database
			chatHistory, err := getChatHistory(db, senderID, receiverID, 10, offset)
			if err != nil {
				fmt.Println("Error retrieving chat history:", err)
				continue
			}

			// Prepare the response payload
			responsePayload := map[string]interface{}{
				"type": "chat_list",
				"data": chatHistory,
			}

			// Convert the payload to JSON
			responseBytes, err := json.Marshal(responsePayload)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}

			// Send the message to the client
			err = conn.WriteMessage(messageType, responseBytes)
			if err != nil {
				fmt.Println("Error sending chat history:", err)
			}
		case "get_chat_list_scroll":
			// Extract data payload
			payload, ok := msg["data"].(map[string]interface{})
			if !ok {
				fmt.Println("Error: payload is not valid or missing")
				continue
			}
			// Validate and extract chat details from payload
			senderIDFloat, ok := payload["sender_id"].(float64)
			if !ok {
				fmt.Println("Error: sender_id is not valid or missing")
				continue
			}
			senderID := int(senderIDFloat)
			receiverIDFloat, ok := payload["receiver_id"].(float64)
			if !ok {
				fmt.Println("Error: receiver_id is not valid or missing")
				continue
			}
			receiverID := int(receiverIDFloat)

			offsetFloat, ok := payload["offset"].(float64)
			if !ok {
				fmt.Println("Error: receiver_id is not valid or missing")
				continue
			}
			offset := int(offsetFloat)
			
			// Retrieve the chat history between the sender and receiver from the database
			chatHistory, err := getChatHistory(db, senderID, receiverID, 10, offset)
			if err != nil {
				fmt.Println("Error retrieving chat history:", err)
				continue
			}

			// Prepare the response payload
			responsePayload := map[string]interface{}{
				"type": "more_chat_list",
				"data": chatHistory,
			}

			// Convert the payload to JSON
			responseBytes, err := json.Marshal(responsePayload)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}

			// Send the message to the client
			err = conn.WriteMessage(messageType, responseBytes)
			if err != nil {
				fmt.Println("Error sending chat history:", err)
			}
		case "new_user":
			payload := msg["data"].(map[string]interface{})
			fmt.Println("Received data:", payload)

			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)

			// Store the sender's WebSocket
			mu.Lock()
			clients[userID] = conn
			mu.Unlock()

			// Notify all connected users about the new user
			mu.Lock()
			for id, clientConn := range clients {
				if id != userID {
					response := map[string]interface{}{
						"type": "new_user_notification",
						"data": userID,
					}
					responseBytes, err := json.Marshal(response)
					if err != nil {
						fmt.Println("Error marshalling response:", err)
						continue
					}
					if err := clientConn.WriteMessage(messageType, responseBytes); err != nil {
						fmt.Println("Error sending message to client:", err)
					}
				}
			}
			mu.Unlock()
		case "logout":
			payload := msg["data"].(map[string]interface{})
			fmt.Println("Received data:", payload)

			userIDFloat, ok := payload["user_id"].(float64)
			if !ok {
				fmt.Println("Error: user_id is not valid or missing")
				continue
			}
			userID := int(userIDFloat)

			response := map[string]interface{}{
				"type": "logout",
				"data": userID,
			}
			delete(clients, userID)
			responseBytes, err := json.Marshal(response)
			if err != nil {
				fmt.Println("Error marshalling response:", err)
				continue
			}
		
			// Send the message to all connected clients.
			mu.Lock()
			for _, clientConn := range clients {
				if err := clientConn.WriteMessage(messageType, responseBytes); err != nil {
					fmt.Println("Error sending message to client:", err)
				}
			}
			fmt.Printf("Message sent to %d clients.\n", len(clients))
			mu.Unlock()
		default:
			fmt.Println("Unknown message type:", msg["type"].(string))
		}

	}
}
