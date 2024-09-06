package main

import (
	"sync"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var sessionStore = make(map[string]string)
var storeLock sync.RWMutex

// GenerateToken generates a new session token
func GenerateToken(nickname string) string {
	token := uuid.New().String()
	storeLock.Lock()
	sessionStore[token] = nickname
	storeLock.Unlock()
	return token
}

// ValidateToken validates the session token
func ValidateToken(token string) (string, bool) {
	username, exists := tokenStore[token]
	return username, exists
}



func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
