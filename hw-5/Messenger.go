package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	cookieAuth           = "auth"
	userID     cookieVal = "ID"
)

type cookieVal string

type User struct {
	Login string
}

type Messenger struct {
	messagesLock     sync.RWMutex
	messages         []byte
	userMessagesLock sync.RWMutex
	userMessages     map[string][]byte
}

func (m *Messenger) addMessage(message string, user cookieVal) {
	message = formatMessage(message, user)

	m.messagesLock.Lock()
	defer m.messagesLock.Unlock()
	m.messages = append(m.messages, []byte(message)...)
}

func (m *Messenger) addUserMessage(message string, user cookieVal, receiverID string) error {
	message = formatMessage(message, user)

	m.userMessagesLock.Lock()
	defer m.userMessagesLock.Unlock()
	_, ok := m.userMessages[receiverID]
	if !ok {
		return errors.New("NO SUCH USER")
	}
	m.userMessages[receiverID] = append(m.userMessages[receiverID], []byte(message)...)
	return nil
}

func formatMessage(message string, user cookieVal) string {
	return fmt.Sprintf("{%s %s} %s\n", time.Now().Format("2006.01.02 15:04:05"), user, message)
}

func NewMessenger() *Messenger {
	var m Messenger
	m.userMessages = make(map[string][]byte)

	return &m
}

func (m *Messenger) PostMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(cookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.addMessage(chi.URLParam(r, "msg"), id)
}

func (m *Messenger) PostPrivateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(cookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := chi.URLParam(r, "id")
	message := chi.URLParam(r, "msg")
	if err := m.addUserMessage(message, id, user); err != nil {
		_, _ = w.Write([]byte(err.Error()))
	}
}

func (m *Messenger) GetMessages(w http.ResponseWriter, r *http.Request) {
	m.messagesLock.RLock()
	messages := m.messages
	m.messagesLock.RUnlock()

	_, _ = w.Write(messages)
}

func (m *Messenger) GetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(cookieVal)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.userMessagesLock.RLock()
	messages := m.userMessages[string(id)]
	m.userMessagesLock.RUnlock()

	_, _ = w.Write(messages)
}

func (m *Messenger) Login(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var u User
	err = json.Unmarshal(d, &u)
	if err != nil || u.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c := &http.Cookie{
		Name:  cookieAuth,
		Value: u.Login,
		Path:  "/",
	}

	m.userMessagesLock.Lock()
	if _, ok := m.userMessages[u.Login]; !ok {
		m.userMessages[u.Login] = []byte("Welcome to the chat!!! \n")
	}
	m.userMessagesLock.Unlock()

	http.SetCookie(w, c)
}

func Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieAuth)
		switch err {
		case nil:
		case http.ErrNoCookie:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if c.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		idCtx := context.WithValue(r.Context(), userID, cookieVal(c.Value))

		handler.ServeHTTP(w, r.WithContext(idCtx))
	}

	return http.HandlerFunc(fn)
}
