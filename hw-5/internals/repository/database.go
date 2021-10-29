package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"tfs-go-hw/hw-5/internals/domain"
)

type Database struct {
	messagesLock     sync.RWMutex
	messages         []byte
	userMessagesLock sync.RWMutex
	userMessages     map[string][]byte
}

func NewDatabase() *Database {
	var db Database
	db.userMessages = make(map[string][]byte)

	return &db
}

func (db *Database) AddMessage(message string, user domain.CookieVal) {
	message = formatMessage(message, user)

	db.messagesLock.Lock()
	defer db.messagesLock.Unlock()
	db.messages = append(db.messages, []byte(message)...)
}

func (db *Database) AddUserMessage(message string, user domain.CookieVal, receiverID string) error {
	message = formatMessage(message, user)

	db.userMessagesLock.Lock()
	defer db.userMessagesLock.Unlock()
	_, ok := db.userMessages[receiverID]
	if !ok {
		return errors.New("NO SUCH USER")
	}
	db.userMessages[receiverID] = append(db.userMessages[receiverID], []byte(message)...)
	return nil
}

func (db *Database) GetMessages() []byte {
	db.messagesLock.RLock()
	defer db.messagesLock.RUnlock()
	messages := append(make([]byte, 0, len(db.messages)), db.messages...)

	return messages
}

func (db *Database) GetPrivateMessages(id domain.CookieVal) ([]byte, error) {
	db.userMessagesLock.RLock()
	defer db.userMessagesLock.RUnlock()
	ids := string(id)
	messages, ok := db.userMessages[ids]
	if !ok {
		return []byte{}, errors.New("NO SUCH USER")
	}
	safeMessages := append(make([]byte, 0, len(messages)), messages...)

	return safeMessages, nil
}

func (db *Database) CreateNewUser(id domain.CookieVal) {
	db.userMessagesLock.Lock()
	defer db.userMessagesLock.Unlock()
	db.userMessages[string(id)] = []byte("Welcome to the chat!!! \n")
}

func formatMessage(message string, user domain.CookieVal) string {
	return fmt.Sprintf("{%s %s} %s\n", time.Now().Format("2006.01.02 15:04:05"), user, message)
}
